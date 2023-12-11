package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reservations-service/data"
	error2 "reservations-service/error"
	"reservations-service/services"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//type KeyProduct struct{}

type AvailabilityHandler struct {
	availabilityService services.AvailabilityService
	DB                  *mongo.Collection
	logger              *log.Logger
}

func NewAvailabilityHandler(availabilityService services.AvailabilityService, db *mongo.Collection, lg *log.Logger) AvailabilityHandler {
	return AvailabilityHandler{availabilityService, db, lg}
}
func (s *AvailabilityHandler) CreateAvailability(rw http.ResponseWriter, h *http.Request) {
	// rw := c.Writer
	// h := c.Request
	vars := mux.Vars(h)
	accIdParam := vars["id"]
	accId, err := primitive.ObjectIDFromHex(accIdParam)
	if err != nil {
		panic(err)
	}

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
			ID       string        `json:"id"`
			UserRole data.UserRole `json:"userRole"`
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

	if userRole != data.Host {
		error2.ReturnJSONError(rw, "Permission denied. Only hosts can create availabilitys.", http.StatusForbidden)
		return
	}

	// availability, exists := c.Get("availability")
	// if !exists {
	// 	error2.ReturnJSONError(rw, "Availability not found in context", http.StatusBadRequest)
	// 	return
	// }
	// avail, ok := availability.(data.Availability)
	// if !ok {
	// 	error2.ReturnJSONError(rw, "Invalid type for Availability", http.StatusBadRequest)
	// 	return
	// }

	avail := h.Context().Value(KeyProduct{}).(*data.Availability)

	avail.AccommodationID = accId
	insertedAvail, err := s.availabilityService.InsertAvailability(avail)
	if err != nil {
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	jsonResponse, err1 := json.Marshal(insertedAvail)
	if err1 != nil {
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}
	rw.Write(jsonResponse)
}

// func (h *AvailabilityHandler) IsAccommodationAvailable(rw http.ResponseWriter, r *http.Request) {
//     accIdParam := mux.Vars(r)["accommodationID"]
// 	accommodationID, err := primitive.ObjectIDFromHex(accIdParam)
// 	if err != nil {
// 		panic(err)
// 	}
//     startDate, _ := time.Parse(time.RFC3339, r.URL.Query().Get("start_date"))
//     endDate, _ := time.Parse(time.RFC3339, r.URL.Query().Get("end_date"))
//     availabilities, err := h.availabilityService.GetAvailableAvailabilitiesBetweenDates(accommodationID, startDate, endDate)
//     if err != nil {
//         http.Error(rw, err.Error(), http.StatusInternalServerError)
//         return
//     }
//     if len(availabilities) == 0 {
//         http.Error(rw, "No availability of type 'Available' found between the provided dates", http.StatusBadRequest)
//         return
//     }
//     // Continue processing...
// }

func (s *AvailabilityHandler) MiddlewareAvailabilityDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		patient := &data.Availability{}
		err := patient.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			s.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, patient)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

// func (s *AvailabilityHandler) GetAllAvailability(c *gin.Context) {
// 	availabilitys, err := s.availabilityService.GetAllAvailability()
// 	if err != nil {
// 		error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	c.JSON(http.StatusOK, availabilitys)
// }

// func (s *AvailabilityHandler) GetAvailabilityByID(c *gin.Context) {
// 	availabilityID := c.Param("id")

// 	availability, err := s.availabilityService.GetAvailabilityByID(availabilityID)
// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			error2.ReturnJSONError(c.Writer, "Availability not found", http.StatusNotFound)
// 		} else {
// 			error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
// 		}
// 		return
// 	}

// 	c.JSON(http.StatusOK, availability)
// }
// func (s *AvailabilityHandler) GetAvailabilityByHostId(c *gin.Context) {
// 	hostID := c.Param("hostId")

// 	accs, err := s.availabilityService.GetAvailabilityByHostId(hostID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get availabilitys"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Availability successfully obtained", "availabilitys": accs})
// }

func (s *AvailabilityHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
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
