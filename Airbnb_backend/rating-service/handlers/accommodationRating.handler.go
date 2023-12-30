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
	"strings"
	"time"
)

type AccommodationRatingHandler struct {
	accommodationRatingService services.AccommodationRatingService
	DB                         *mongo.Collection
	Tracer                     trace.Tracer
}

func NewAccommodationRatingHandler(accommodationRatingService services.AccommodationRatingService, db *mongo.Collection, tr trace.Tracer) AccommodationRatingHandler {
	return AccommodationRatingHandler{accommodationRatingService, db, tr}
}

func (s *AccommodationRatingHandler) RateAccommodation(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.RateAccommodation")
	defer span.End()

	accommodationID := c.Param("accommodationId")

	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain current user information. Try again later")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information. Try again later"})
		return
	}
	println(accommodationID)
	accommodation, err := s.getAccommodationByIDFromAccommodationService(accommodationID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain accommodation information.Try again later.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain accommodation information.Try again later."})
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
		span.SetStatus(codes.Error, "You cannot rate this accommodation. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rate this accommodation. You don't have reservations from him"})
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
		span.SetStatus(codes.Error, "You cannot rate this accommodation. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rate this accommodation. You don't have reservations from him"})
		return
	}

	canRate := false
	for _, reservation := range reservations {
		if reservation.AccommodationId == accommodationID {
			canRate = true
			break
		}
	}

	if !canRate {
		span.SetStatus(codes.Error, "You cannot rate this accommodation. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rate this accommodation. You don't have reservations from him"})
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
	println(accommodation)
	newRateAccommodation := &domain.RateAccommodation{
		ID:            id,
		Accommodation: accommodationID,
		Guest:         currentUser,
		DateAndTime:   currentDateTime,
		Rating:        domain.Rating(requestBody.Rating),
	}

	err, update := s.accommodationRatingService.SaveRating(newRateAccommodation, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to save rating")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to save rating"})
		return
	}
	urlAccommodationCheck := "https://acc-server:8083/api/accommodations/get/" + accommodationID

	timeout = 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(spanCtx, token, urlAccommodationCheck)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Accommodation service is not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Accommodation service not available."})
			return
		}
		span.SetStatus(codes.Error, "Accommodation service is not available.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Accommodation service not available."})
		return
	}
	defer resp.Body.Close()

	statusCodeAccommodation := resp.StatusCode
	fmt.Println(statusCodeAccommodation)
	if statusCodeAccommodation != 200 {
		span.SetStatus(codes.Error, "Accommodation with that id does not exist.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Accommodation with that id does not exist."})
		return
	}

	var responseAccommodation struct {
		AccommodationName      string `json:"accommodation_name"`
		AccommodationLocation  string `json:"accommodation_location"`
		AccommodationHostId    string `json:"host_id"`
		AccommodationMinGuests int    `json:"accommodation_min_guests"`
		AccommodationMaxGuests int    `json:"accommodation_max_guests"`
		AccommodationId        string `json:"_id"`
	}
	decoder = json.NewDecoder(resp.Body)

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseAccommodation); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			span.SetStatus(codes.Error, "Invalid date format.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format."})
			return
		}
		span.SetStatus(codes.Error, "Error decoding JSON response"+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding json response."})
		return
	}

	urlHostCheck := "https://auth-server:8080/api/users/getById/" + responseAccommodation.AccommodationHostId

	resp, err = s.HTTPSPerformAuthorizationRequestWithContext(spanCtx, token, urlHostCheck)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Authorization service is not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization service not available."})
			return
		}
		span.SetStatus(codes.Error, "Authorization service is not available.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization service not available."})
		return
	}
	defer resp.Body.Close()

	statusCodeHostCheck := resp.StatusCode
	if statusCodeHostCheck != 200 {
		span.SetStatus(codes.Error, "Host with that id does not exist.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host with that id does not exist."})
		return
	}

	decoder = json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var responseHost struct {
		Host struct {
			Email    string `json:"email"`
			Username string `json:"username"`
			HostID   string `json:"id"`
		} `json:"user"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseHost); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			span.SetStatus(codes.Error, "Invalid date format.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format."})
			return
		}
		span.SetStatus(codes.Error, "Error decoding JSON response"+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding json response."})
		return
	}

	notificationPayload := map[string]interface{}{
		"host_id":           responseAccommodation.AccommodationHostId,
		"host_email":        responseHost.Host.Email,
		"notification_text": "Dear " + responseHost.Host.Username + ", \n your accommodation " + responseAccommodation.AccommodationName + " has been rated. It got " + strconv.Itoa(requestBody.Rating) + " stars by " + currentUser.Username + "!",
	}

	notificationPayloadJSON, err := json.Marshal(notificationPayload)
	if err != nil {
		span.SetStatus(codes.Error, "Error creating notification payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification payload"})
		return
	}

	notificationURL := "https://notifications-server:8089/api/notifications/create"

	timeout = 2000 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err = s.HTTPSperformAuthorizationRequestWithContextAndBodyAcc(spanCtx, token, notificationURL, "POST", notificationPayloadJSON)
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

	if update {
		eventURL := "https://res-server:8082/api/event/store"

		timeout = 2000 * time.Second
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()

		eventPayload := map[string]interface{}{
			"event":            "Accommodation rating",
			"accommodation_id": responseAccommodation.AccommodationId,
		}

		eventPayloadJSON, err := json.Marshal(eventPayload)
		if err != nil {
			span.SetStatus(codes.Error, "Error creating notification payload")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification payload"})
			return
		}

		resp, err = s.HTTPSperformAuthorizationRequestWithContextAndBodyAcc(spanCtx, token, eventURL, "POST", eventPayloadJSON)
		if err != nil {
			span.SetStatus(codes.Error, "Error creating event request")
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event request"})
				return
			}
			span.SetStatus(codes.Error, "Reservation service not available.")
			fmt.Println(err)
			fmt.Println("here")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation service while event handling not available."})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			span.SetStatus(codes.Error, "Error creating event")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rating successfully saved", "rating": newRateAccommodation})
}

func (s *AccommodationRatingHandler) DeleteRatingAccommodation(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.DeleteRatingAccommodation")
	defer span.End()

	accommodationID := c.Param("accommodationId")
	println(accommodationID)
	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain current user information.  Try again later.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information.  Try again later."})
		return
	}
	guestID := currentUser.ID.Hex()

	err = s.accommodationRatingService.DeleteRating(accommodationID, guestID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Rating successfully deleted")

	eventURL := "https://res-server:8082/api/event/store"

	timeout := 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	eventPayload := map[string]interface{}{
		"event":            "Accommodation rating delete",
		"accommodation_id": accommodationID,
	}

	eventPayloadJSON, err := json.Marshal(eventPayload)
	if err != nil {
		span.SetStatus(codes.Error, "Error creating notification payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification payload"})
		return
	}

	resp, err := s.HTTPSperformAuthorizationRequestWithContextAndBodyAcc(spanCtx, token, eventURL, "POST", eventPayloadJSON)
	if err != nil {
		span.SetStatus(codes.Error, "Error creating event request")
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event request"})
			return
		}
		span.SetStatus(codes.Error, "Reservation service not available.")
		fmt.Println(err)
		fmt.Println("here")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation service while event handling not available."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		span.SetStatus(codes.Error, "Error creating event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating successfully deleted"})
}

func (s *AccommodationRatingHandler) GetByAccommodationAndGuest(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.GetByAccommodationAndGuest")
	defer span.End()
	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(
			codes.Error, "Failed to obtain current user information")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information"})
		return
	}
	guestID := currentUser.ID.Hex()

	accommodationID := c.Param("accommodationId")

	ratings, err := s.accommodationRatingService.GetByAccommodationAndGuest(accommodationID, guestID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Got ratings by accommodation and guest successfully")
	c.JSON(http.StatusOK, gin.H{"ratings": ratings})
}

func (s *AccommodationRatingHandler) getAccommodationByIDFromAccommodationService(accommodationID string, c context.Context) (*domain.Accommodation, error) {
	spanCtx, span := s.Tracer.Start(c, "AccommodationRatingHandler.getAccommodationByIDFromAccommodationService")
	defer span.End()
	url := "https://acc-server:8083/api/accommodations/get/" + accommodationID

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
		span.SetStatus(codes.Error, "Accommodation not found")
		return nil, errors.New("Accommodation not found")
	}

	var accommodationResponse domain.AccommodationResponse
	if err := json.NewDecoder(resp.Body).Decode(&accommodationResponse); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "Got accommodation by id from auth service")
	accommodation := domain.ConvertToDomainAccommodation(accommodationResponse)

	return &accommodation, nil
}

func (s *AccommodationRatingHandler) getCurrentUserFromAuthService(token string, c context.Context) (*domain.User, error) {
	spanCtx, span := s.Tracer.Start(c, "AccommodationRatingHandler.getCurrentUserFromAuthService")
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
	span.SetStatus(codes.Ok, "Got current user from auth service")
	user := domain.ConvertToDomainUser(userResponse)

	return &user, nil
}

func (s *AccommodationRatingHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	_, span := s.Tracer.Start(ctx, "AccommodationRatingHandler.HTTPSPerformAuthorizationRequestWithContext")
	defer span.End()
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *AccommodationRatingHandler) HTTPSperformAuthorizationRequestWithContextAndBodyAcc(
	ctx context.Context, token string, url string, method string, requestBody []byte,
) (*http.Response, error) {
	_, span := s.Tracer.Start(ctx, "AccommodationRatingHandler.HTTPSperformAuthorizationRequestWithContextAndBody")
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

func (s *AccommodationRatingHandler) GetAllRatingsAccommodation(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.GetAllRatingsAccommodation")
	defer span.End()

	ratings, averageRating, err := s.accommodationRatingService.GetAllRatingsAccommodation(spanCtx)
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
