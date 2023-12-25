package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"rating-service/domain"
	"rating-service/services"
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

	err = s.accommodationRatingService.SaveRating(newRateAccommodation, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to save rating")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to save rating"})
		return
	}
	span.SetStatus(codes.Ok, "Rating successfully saved")
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
	c.JSON(http.StatusOK, gin.H{"message": "Rating successfully deleted"})
}

func (s *AccommodationRatingHandler) GetAllRatings(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.GetAllRatings")
	defer span.End()

	ratings, averageRating, err := s.accommodationRatingService.GetAllRatings(spanCtx)
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

//func (s *AccommodationRatingHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
//	tr := http.DefaultTransport.(*http.Transport).Clone()
//	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
//
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return nil, err
//	}
//	req.Header.Set("Authorization", token)
//
//	client := &http.Client{Transport: tr}
//	resp, err := client.Do(req.WithContext(ctx))
//	if err != nil {
//		return nil, err
//	}
//
//	return resp, nil
//}
//
//func (s *AccommodationRatingHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
//	tr := http.DefaultTransport.(*http.Transport).Clone()
//	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
//
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return nil, err
//	}
//	req.Header.Set("Authorization", token)
//
//	client := &http.Client{Transport: tr}
//	resp, err := client.Do(req.WithContext(ctx))
//	if err != nil {
//		return nil, err
//	}
//
//	return resp, nil
//}
