package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"rating-service/domain"
	"rating-service/services"
	"strconv"
	"time"
)

type HostRatingHandler struct {
	hostRatingService services.HostRatingService
	DB                *mongo.Collection
	Tracer            trace.Tracer
}

func NewHostRatingHandler(hostRatingService services.HostRatingService, db *mongo.Collection, tr trace.Tracer) HostRatingHandler {
	return HostRatingHandler{hostRatingService, db, tr}
}

func (s *HostRatingHandler) RateHost(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "HostRatingHandler.RateHost")
	defer span.End()

	hostID := c.Param("hostId")

	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain current user information. Try again later")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information. Try again later"})
		return
	}

	hostUser, err := s.getUserByIDFromAuthService(hostID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain host information.Try again later.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain host information.Try again later."})
		return
	}

	urlCheckReservations := "https://res-server:8082/api/reservations/getAll"

	timeout := 2000 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	respRes, errRes := s.HTTPSPerformAuthorizationRequestWithContext(spanCtx, token, urlCheckReservations)
	if errRes != nil {
		span.SetStatus(codes.Error, "Failed to get reservations. Try again later.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get reservations. Try again later."})
		return
	}

	defer respRes.Body.Close()

	if respRes.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "You cannot rate this host. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rate this host. You don't have reservations from him"})
		return
	}

	decoder := json.NewDecoder(respRes.Body)
	var reservations []domain.ReservationByGuest
	if err := decoder.Decode(&reservations); err != nil {
		fmt.Println(err)
		span.SetStatus(codes.Error, "Failed to decode reservations")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode reservations"})
		return
	}

	if len(reservations) == 0 {
		span.SetStatus(codes.Error, "You cannot rate this host. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rate this host. You don't have reservations from him"})
		return
	}

	canRate := false
	for _, reservation := range reservations {
		if reservation.AccommodationHostId == hostID {
			canRate = true
			break
		}
	}

	if !canRate {
		span.SetStatus(codes.Error, "You cannot rate this host. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rate this host. You don't have reservations from him"})
		return
	}

	var requestBody struct {
		Rating int `json:"rating"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		span.SetStatus(codes.Error, "Invalid JSON request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request"})
		return
	}

	currentDateTime := primitive.NewDateTimeFromTime(time.Now())
	id := primitive.NewObjectID()

	newRateHost := &domain.RateHost{
		ID:          id,
		Host:        hostUser,
		Guest:       currentUser,
		DateAndTime: currentDateTime,
		Rating:      domain.Rating(requestBody.Rating),
	}

	err = s.hostRatingService.SaveRating(newRateHost, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to save rating")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to save rating"})
		return
	}

	hostIDString := hostUser.ID.Hex()

	notificationPayload := map[string]interface{}{
		"host_id":           hostIDString,
		"host_email":        hostUser.Email,
		"notification_text": "Dear " + hostUser.Username + "\n you have been rated. You got " + strconv.Itoa(requestBody.Rating) + " stars",
	}

	notificationPayloadJSON, err := json.Marshal(notificationPayload)
	if err != nil {
		span.SetStatus(codes.Error, "Error creating notification payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification payload"})
		return
	}

	notificationURL := "https://notifications-server:8089/api/notifications/create"

	timeout = 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContextAndBody(spanCtx, token, notificationURL, "POST", notificationPayloadJSON)
	if err != nil {
		span.SetStatus(codes.Error, "Error creating notification request")
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification request"})
			return
		}
		span.SetStatus(codes.Error, "Notification service not available.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification service not available."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		span.SetStatus(codes.Error, "Error creating notification")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification"})
		return
	}
	span.SetStatus(codes.Ok, "Rating successfully saved")
	c.JSON(http.StatusCreated, gin.H{"message": "Rating successfully saved", "rating": newRateHost})
}

func (s *HostRatingHandler) DeleteRating(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "HostRatingHandler.DeleteRating")
	defer span.End()

	hostID := c.Param("hostId")

	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain current user information.  Try again later.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information.  Try again later."})
		return
	}
	guestID := currentUser.ID.Hex()

	err = s.hostRatingService.DeleteRating(hostID, guestID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Rating successfully deleted")
	c.JSON(http.StatusOK, gin.H{"message": "Rating successfully deleted"})
}

func (s *HostRatingHandler) GetAllRatings(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "HostRatingHandler.GetAllRatings")
	defer span.End()

	ratings, averageRating, err := s.hostRatingService.GetAllRatings(spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"ratings":       ratings,
		"averageRating": averageRating,
	}
	span.SetStatus(codes.Ok, "Got all ratings successfully")
	c.JSON(http.StatusOK, response)
}

func (s *HostRatingHandler) GetByHostAndGuest(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "HostRatingHandler.GetByHostAndGuest")
	defer span.End()

	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain current user information")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information"})
		return
	}
	guestID := currentUser.ID.Hex()

	hostID := c.Param("hostId")

	ratings, err := s.hostRatingService.GetByHostAndGuest(hostID, guestID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Got ratings by host and guest successfully")
	c.JSON(http.StatusOK, gin.H{"ratings": ratings})
}

func (s *HostRatingHandler) getUserByIDFromAuthService(userID string, c context.Context) (*domain.User, error) {
	spanCtx, span := s.Tracer.Start(c, "HostRatingHandler.getUserByIDFromAuthService")
	defer span.End()

	url := "https://auth-server:8080/api/users/getById/" + userID

	timeout := 2000 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(spanCtx, "", url)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "User not found")
		return nil, errors.New("User not found")
	}

	var userResponse domain.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	user := domain.ConvertToDomainUser(userResponse)
	return &user, nil
}

func (s *HostRatingHandler) getCurrentUserFromAuthService(token string, c context.Context) (*domain.User, error) {
	spanCtx, span := s.Tracer.Start(c, "HostRatingHandler.getCurrentUserFromAuthService")
	defer span.End()
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(spanCtx, token, url)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "Unauthorized")
		return nil, errors.New("Unauthorized")
	}

	var userResponse domain.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	user := domain.ConvertToDomainUser(userResponse)
	return &user, nil
}

func (s *HostRatingHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	_, span := s.Tracer.Start(ctx, "HostRatingHandler.HTTPSPerformAuthorizationRequestWithContext")
	defer span.End()
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", token)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *HostRatingHandler) HTTPSperformAuthorizationRequestWithContextAndBody(
	ctx context.Context, token string, url string, method string, requestBody []byte,
) (*http.Response, error) {
	_, span := s.Tracer.Start(ctx, "HostRatingHandler.HTTPSperformAuthorizationRequestWithContextAndBody")
	defer span.End()
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", token)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return resp, nil
}

func ExtractTraceInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
