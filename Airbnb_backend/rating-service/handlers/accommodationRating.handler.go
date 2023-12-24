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
	"net/http"
	"rating-service/domain"
	"rating-service/services"
	"time"
)

type AccommodationRatingHandler struct {
	accommodationRatingService services.AccommodationRatingService
	DB                         *mongo.Collection
}

func NewAccommodationRatingHandler(accommodationRatingService services.AccommodationRatingService, db *mongo.Collection) AccommodationRatingHandler {
	return AccommodationRatingHandler{accommodationRatingService, db}
}

func (s *AccommodationRatingHandler) RateAccommodation(c *gin.Context) {
	accommodationID := c.Param("accommodationId")

	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information. Try again later"})
		return
	}
	println(accommodationID)
	accommodation, err := s.getAccommodationByIDFromAccommodationService(accommodationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain accommodation information.Try again later."})
		return
	}

	urlCheckReservations := "https://res-server:8082/api/reservations/getAll"

	timeout := 2000 * time.Second
	ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	respRes, errRes := s.HTTPSPerformAuthorizationRequestWithContext(ctxRest, token, urlCheckReservations)
	if errRes != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get reservations. Try again later."})
		return
	}

	defer respRes.Body.Close()

	if respRes.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rate this accommodation. You don't have reservations from him"})
		return
	}

	decoder := json.NewDecoder(respRes.Body)
	var reservations []domain.ReservationByGuest
	if err := decoder.Decode(&reservations); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode reservations"})
		return
	}

	if len(reservations) == 0 {
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rate this accommodation. You don't have reservations from him"})
		return
	}

	var requestBody struct {
		Rating int `json:"rating"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
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

	err = s.accommodationRatingService.SaveRating(newRateAccommodation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to save rating"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rating successfully saved", "rating": newRateAccommodation})
}

func (s *AccommodationRatingHandler) DeleteRatingAccommodation(c *gin.Context) {
	accommodationID := c.Param("accommodationId")
	println(accommodationID)
	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information.  Try again later."})
		return
	}
	guestID := currentUser.ID.Hex()

	err = s.accommodationRatingService.DeleteRating(accommodationID, guestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating successfully deleted"})
}

func (s *AccommodationRatingHandler) GetAllRatings(c *gin.Context) {
	ratings, averageRating, err := s.accommodationRatingService.GetAllRatings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"ratings":       ratings,
		"averageRating": averageRating,
	}

	c.JSON(http.StatusOK, response)
}

func (s *AccommodationRatingHandler) GetByAccommodationAndGuest(c *gin.Context) {
	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information"})
		return
	}
	guestID := currentUser.ID.Hex()

	accommodationID := c.Param("accommodationId")

	ratings, err := s.accommodationRatingService.GetByAccommodationAndGuest(accommodationID, guestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ratings": ratings})
}

func (s *AccommodationRatingHandler) getAccommodationByIDFromAccommodationService(accommodationID string) (*domain.Accommodation, error) {
	println(accommodationID)
	url := "https://acc-server:8083/api/accommodations/get/" + accommodationID

	timeout := 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, "", url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Accommodation not found")
	}
	println("ovde sam")
	println(resp.Body)
	var accommodationResponse domain.AccommodationResponse
	if err := json.NewDecoder(resp.Body).Decode(&accommodationResponse); err != nil {
		return nil, err
	}
	println("evo ovde")
	accommodation := domain.ConvertToDomainAccommodation(accommodationResponse)

	return &accommodation, nil
}

func (s *AccommodationRatingHandler) getCurrentUserFromAuthService(token string) (*domain.User, error) {
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Unauthorized")
	}

	var userResponse domain.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return nil, err
	}

	user := domain.ConvertToDomainUser(userResponse)

	return &user, nil
}

func (s *AccommodationRatingHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
