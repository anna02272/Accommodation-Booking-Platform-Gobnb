package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"reservations-service/data"
	error2 "reservations-service/error"
	"reservations-service/repository"
	"strings"
	"time"
)

var validateFields = validator.New()

type KeyProduct struct{}

type ReservationsHandler struct {
	logger *log.Logger
	// NoSQL: injecting student repository
	repo *repository.ReservationRepo
}

func NewReservationsHandler(l *log.Logger, r *repository.ReservationRepo) *ReservationsHandler {
	return &ReservationsHandler{l, r}
}

func (s *ReservationsHandler) CreateReservationForGuest(rw http.ResponseWriter, h *http.Request) {
	token := h.Header.Get("Authorization")
	//token = html.EscapeString(token)

	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusInternalServerError)
			return
		}

		error2.ReturnJSONError(rw, "Error performing authorization request", http.StatusInternalServerError)
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
	guestId := response.LoggedInUser.ID
	//guestId = html.EscapeString(guestId)

	userRole := response.LoggedInUser.UserRole

	if userRole != data.Guest {
		error2.ReturnJSONError(rw, "Permission denied. Only guests can create reservations.", http.StatusForbidden)
		return
	}

	guestReservation := h.Context().Value(KeyProduct{}).(*data.ReservationByGuestCreate)

	accId := guestReservation.AccommodationId.String()
	urlAccommodationCheck := "https://acc-server:8083/api/accommodations/get/" + accId

	resp, err = s.HTTPSperformAuthorizationRequestWithContext(ctx, token, urlAccommodationCheck)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Accommodation service is not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusInternalServerError)
			return
		}

		errorMsg := map[string]string{"error": "Error performing accommodation request."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	statusCodeAccommodation := resp.StatusCode
	if statusCodeAccommodation != 200 {
		errorMsg := map[string]string{"error": "Accommodation with that id does not exist."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	// Validate empty fields
	if err := validateFields.Struct(guestReservation); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			validationErrors[field] = fmt.Sprintf("Field %s is required", field)
		}
		error2.ReturnJSONError(rw, validationErrors, http.StatusBadRequest)
		return
	}

	if guestReservation.CheckInDate.IsZero() || guestReservation.CheckOutDate.IsZero() {
		errorMsg := map[string]string{"error": "Check-in and check-out dates are required."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	//if !utils.IsValidDateFormat(guestReservation.CheckInDate.String()) ||
	//	!utils.IsValidDateFormat(guestReservation.CheckOutDate.String()) {
	//	error2.ReturnJSONError(rw, "Invalid date format. Use '2006-01-02T15:04:05Z'", http.StatusBadRequest)
	//	return
	//}

	if guestReservation.CheckInDate.Before(time.Now()) {
		errorMsg := map[string]string{"error": "Check-in date must be in the future."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	if guestReservation.CheckInDate.After(guestReservation.CheckOutDate) {
		errorMsg := map[string]string{"error": "Check-in date must be before check out date."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	decoder = json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var responseAccommodation struct {
		AccommodationName     string `json:"accommodation_name"`
		AccommodationLocation string `json:"accommodation_location"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseAccommodation); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			errorMsg := map[string]string{"error": "Invalid date format."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}
	//
	//responseAccommodation.AccommodationName = html.EscapeString(responseAccommodation.AccommodationName)
	//responseAccommodation.AccommodationLocation = html.EscapeString(responseAccommodation.AccommodationLocation)

	errReservation := s.repo.InsertReservationByGuest(guestReservation, guestId,
		responseAccommodation.AccommodationName, responseAccommodation.AccommodationLocation)
	if errReservation != nil {
		s.logger.Print("Database exception: ", errReservation)
		errorMsg := map[string]string{"error": "Cannot reserve. Please double check if you already reserved exactly the accommodation and check in date"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	responseJSON, err := json.Marshal(guestReservation)
	if err != nil {
		error2.ReturnJSONError(rw, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	rw.Write(responseJSON)
}

//func (s *ReservationsHandler) GetReservationsByGuest(rw http.ResponseWriter, h *http.Request) {
//	vars := mux.Vars(h)
//	guestId := vars["id"]
//
//	reservationsByGuest, err := s.repo.GetReservationsByGuest(guestId)
//	if err != nil {
//		s.logger.Print("Database exception: ", err)
//	}
//
//	if reservationsByGuest == nil {
//		return
//	}
//
//	err = reservationsByGuest.ToJSON(rw)
//	if err != nil {
//		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
//		s.logger.Fatal("Unable to convert to json :", err)
//		return
//	}
//}

func (s *ReservationsHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		s.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func (s *ReservationsHandler) MiddlewareReservationForGuestDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		patient := &data.ReservationByGuestCreate{}
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

func (s *ReservationsHandler) HTTPSperformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
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

func (s *ReservationsHandler) performAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
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
