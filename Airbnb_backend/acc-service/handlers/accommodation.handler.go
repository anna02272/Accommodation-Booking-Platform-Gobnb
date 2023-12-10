package handlers

import (
	"acc-service/domain"
	error2 "acc-service/error"
	"acc-service/services"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strings"
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
func (s *AccommodationHandler) CreateAccommodations(c *gin.Context) {
	rw := c.Writer
	h := c.Request

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, "Error performing authorization request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		errorMsg := map[string]string{"error": "Unauthorized."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	// Read the response body
	// Create a JSON decoder for the response body
	decoder := json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var response struct {
		LoggedInUser struct {
			ID       string          `json:"id"`
			UserRole domain.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole

	if userRole != domain.Host {
		error2.ReturnJSONError(rw, "Permission denied. Only hosts can create accommodations.", http.StatusForbidden)
		return
	}

	id := primitive.NewObjectID()

	accommodation, exists := c.Get("accommodation")
	if !exists {
		error2.ReturnJSONError(rw, "Accommodation not found in context", http.StatusBadRequest)
		return
	}
	acc, ok := accommodation.(domain.Accommodation)
	if !ok {
		error2.ReturnJSONError(rw, "Invalid type for Accommodation", http.StatusBadRequest)
		return
	}
	acc.ID = id

	insertedAcc, _, err := s.accommodationService.InsertAccommodation(&acc, response.LoggedInUser.ID)
	if err != nil {
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	jsonResponse, err1 := json.Marshal(insertedAcc)
	if err1 != nil {
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}
	rw.Write(jsonResponse)
}

func (s *AccommodationHandler) GetAllAccommodations(c *gin.Context) {
	accommodations, err := s.accommodationService.GetAllAccommodations()
	if err != nil {
		error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, accommodations)
}

func (s *AccommodationHandler) GetAccommodationByID(c *gin.Context) {
	accommodationID := c.Param("id")

	accommodation, err := s.accommodationService.GetAccommodationByID(accommodationID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			error2.ReturnJSONError(c.Writer, "Accommodation not found", http.StatusNotFound)
		} else {
			error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, accommodation)
}
func (s *AccommodationHandler) GetAccommodationsByHostId(c *gin.Context) {
	hostID := c.Param("hostId")

	accs, err := s.accommodationService.GetAccommodationsByHostId(hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get accommodations"})
		return
	}

	if len(accs) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No accommodations found for this host", "accommodations": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Accommodations successfully obtained", "accommodations": accs})
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
