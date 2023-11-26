package handlers

import (
	"accomodations-service/domain"
	error2 "accomodations-service/error"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	//"reservations-service/data"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type AccommodationsHandler struct {
	logger *log.Logger
	// NoSQL: injecting student repository
	repo *domain.AccommodationRepo
}

func NewAccommodationsHandler(l *log.Logger, r *domain.AccommodationRepo) *AccommodationsHandler {
	return &AccommodationsHandler{l, r}
}

func (s *AccommodationsHandler) CreateAccommodations(rw http.ResponseWriter, h *http.Request) {

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 5 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
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

	accommodation := h.Context().Value(KeyProduct{}).(*domain.Accommodation)
	acc, err := s.repo.InsertAccommodation(accommodation)
	if err != nil {
		s.logger.Print("Database exception: ", err)
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	jsonResponse, err1 := json.Marshal(acc)
	if err1 == nil {
		rw.Write(jsonResponse)
	}
}

func (s *AccommodationsHandler) GetAccommodationById(rw http.ResponseWriter, h *http.Request) {

	vars := mux.Vars(h)
	accommodationID := vars["id"]

	//idRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{4}-[8|9|aA|bB][a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	//if !idRegex.MatchString(accommodationID) {
	//	errJson.ReturnJSONError(rw, "Invalid UUID format.", http.StatusBadRequest)
	//	return
	//}

	accommodations, err := s.repo.GetAccommodations(accommodationID)
	if err != nil {
		s.logger.Print("Exception: ", err)
		error2.ReturnJSONError(rw, err, http.StatusBadRequest)
		return
	}

	if len(accommodations) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Assuming you want to return the first accommodation found
	accommodation := accommodations[0]

	// Send the JSON response
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := accommodation.ToJSON(rw); err != nil {
		s.logger.Println("Error encoding JSON:", err)
	}
}

func (s *AccommodationsHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		s.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func (s *AccommodationsHandler) MiddlewareAccommodationDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		patient := &domain.Accommodation{}
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

/*
func (s *AccommodationsHandler) SetAvailability(rw http.ResponseWriter, r *http.Request) {
	var availabilityRequest AvailabilityRequest

	err := json.NewDecoder(r.Body).Decode(&availabilityRequest)
	if err != nil {
		s.logger.Println("Error decoding JSON:", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	accommodationID := availabilityRequest.AccommodationID
	if accommodationID == "" {
		s.logger.Println("Accommodation ID is required.")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	accommodation, err := s.repo.GetAccommodations(accommodationID)
	if accommodation == nil {
		s.logger.Printf("Accommodation with ID %s not found.", accommodationID)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	for date, available := range availabilityRequest.Dates {
		accommodation[0].Availability[date] = available
	}

	// Čuvanje promena u bazi podataka (prilagoditi prema vašoj implementaciji)
	// s.repo.UpdateAccommodation(accommodation)

	// Slanje uspešnog odgovora
	rw.WriteHeader(http.StatusOK)
}

type AvailabilityRequest struct {
	AccommodationID string             `json:"accommodation_id"`
	Dates           map[time.Time]bool `json:"dates"`
}
*/

func (s *AccommodationsHandler) SetAccommodationAvailability(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	accommodationID := vars["id"]
	var availability map[time.Time]bool

	err := json.NewDecoder(h.Body).Decode(&availability)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	nil := s.repo.UpdateAccommodationAvailability(accommodationID, availability)
	if nil != nil {
		s.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (s *AccommodationsHandler) SetAccommodationPrice(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	accommodationID := vars["id"]
	var price map[time.Time]string

	err := json.NewDecoder(h.Body).Decode(&price)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	nil := s.repo.UpdateAccommodationPrice(accommodationID, price)
	if nil != nil {
		s.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (s *AccommodationsHandler) HTTPSperformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	// Perform the request with the provided context
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *AccommodationsHandler) performAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	// Perform the request with the provided context
	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
