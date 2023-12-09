package handlers

import (
	"acc-service/domain"
	"acc-service/services"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type AccommodationHandler struct {
	accommodationService services.AccommodationService
	DB                   *mongo.Collection
}

func NewAccommodationHandler(accommodationService services.AccommodationService, db *mongo.Collection) AccommodationHandler {
	return AccommodationHandler{accommodationService, db}
}

func (s *AccommodationHandler) AddAccommodation(c *gin.Context) {
	var acc *domain.Accommodation
	//hostID := c.Param("hostId")

	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to obtain current user information"})
		return
	}

	hostID := currentUser.ID

	// hostUser, err := s.getUserByIDFromAuthService(hostID)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to obtain host user information"})
	// 	return
	// }

	//currentDateTime := primitive.NewDateTimeFromTime(time.Now())

	// rating := domain.Rating("5")
	// newRateHost := &domain.RateHost{
	// 	Host:        hostUser,
	// 	Guest:       currentUser,
	// 	DateAndTime: currentDateTime,
	// 	Rating:      rating,
	// }

	if err := c.ShouldBindJSON(&acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	acc.HostId = hostID

	err = s.accommodationService.SaveAccommodation(acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save accommodation"})
		return
	}

	// err = s.accommodationService.SaveRating(newRateHost)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rating"})
	// 	return
	// }

	c.JSON(http.StatusCreated, gin.H{"message": "Rating successfully saved", "rating": acc})
}

// func (s *AccommodationHandler) getUserByIDFromAuthService(userID string) (*domain.User, error) {
// 	url := "https://auth-server:8080/api/users/getById/" + userID

// 	timeout := 2000 * time.Second
// 	ctx, cancel := context.WithTimeout(context.Background(), timeout)
// 	defer cancel()

// 	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, "", url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, errors.New("User not found")
// 	}

// 	var user domain.User
// 	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
// 		return nil, err
// 	}

// 	return &user, nil
// }

func (s *AccommodationHandler) getCurrentUserFromAuthService(token string) (*domain.User, error) {
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
	var user domain.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AccommodationHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
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
