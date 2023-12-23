package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reservations-service/data"
	error2 "reservations-service/error"
	"reservations-service/repository"
	"reservations-service/services"
	"reservations-service/utils"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var validateFields = validator.New()

type KeyProduct struct{}

type ReservationsHandler struct {
	logger    *log.Logger
	repo      *repository.ReservationRepo
	serviceAv services.AvailabilityService
	DB        *mongo.Collection
}

func NewReservationsHandler(l *log.Logger, srv services.AvailabilityService,
	r *repository.ReservationRepo, db *mongo.Collection) *ReservationsHandler {
	return &ReservationsHandler{l, r, srv, db}
}

func (s *ReservationsHandler) CreateReservationForGuest(rw http.ResponseWriter, h *http.Request) {
	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Authorization service not available.."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusInternalServerError)
			return
		}

		errorMsg := map[string]string{"error": "Authorization service not available.."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusInternalServerError)
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
		errorMsg := map[string]string{"error": "Permission denied. Only guests can create reservations"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusForbidden)
		return
	}

	guestReservation := h.Context().Value(KeyProduct{}).(*data.ReservationByGuestCreate)

	accId := guestReservation.AccommodationId
	urlAccommodationCheck := "https://acc-server:8083/api/accommodations/get/" + accId

	resp, err = s.HTTPSperformAuthorizationRequestWithContext(ctx, token, urlAccommodationCheck)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Accommodation service is not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		errorMsg := map[string]string{"error": "Accommodation service is not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCodeAccommodation := resp.StatusCode
	fmt.Println(statusCodeAccommodation)
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

	if !utils.IsValidInteger(guestReservation.NumberOfGuests) {
		errorMsg := map[string]string{"error": "Invalid field number_of_guests. It's a whole number."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	// Define a struct to represent the JSON structure of user
	var responseAccommodation struct {
		AccommodationName      string `json:"accommodation_name"`
		AccommodationLocation  string `json:"accommodation_location"`
		AccommodationHostId    string `json:"host_id"`
		AccommodationMinGuests int    `json:"accommodation_min_guests"`
		AccommodationMaxGuests int    `json:"accommodation_max_guests"`
	}
	decoder = json.NewDecoder(resp.Body)

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseAccommodation); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			errorMsg := map[string]string{"error": "Invalid date format."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		fmt.Println("Acommodaiton errored")
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	if responseAccommodation.AccommodationMaxGuests < guestReservation.NumberOfGuests {
		errorMsg := map[string]string{"error": "Too much guests.Double check the capacity of accommodation."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	guestRsvPrimitive, err := primitive.ObjectIDFromHex(guestReservation.AccommodationId)
	isAvailable, err := s.serviceAv.IsAvailable(guestRsvPrimitive, guestReservation.CheckInDate, guestReservation.CheckOutDate)

	if !isAvailable {
		errorMsg := map[string]string{"error": "Accommodation is not available for the specified dates"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	errReservation := s.repo.InsertReservationByGuest(guestReservation, guestId,
		responseAccommodation.AccommodationName, responseAccommodation.AccommodationLocation, responseAccommodation.AccommodationHostId)
	if errReservation != nil {
		s.logger.Print("Database exception: ", errReservation)
		errorMsg := map[string]string{"error": "Cannot reserve. Please double check if you already reserved exactly the accommodation and check in date"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	errBookAccommodation := s.serviceAv.BookAccommodation(guestRsvPrimitive, guestReservation.CheckInDate, guestReservation.CheckOutDate)
	if errBookAccommodation != nil {
		errorMsg := map[string]string{"error": "Error booking accommodation."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	fmt.Println(responseAccommodation.AccommodationHostId)
	urlHostCheck := "https://auth-server:8080/api/users/getById/" + responseAccommodation.AccommodationHostId

	resp, err = s.HTTPSperformAuthorizationRequestWithContext(ctx, token, urlHostCheck)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Authorization service is not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		errorMsg := map[string]string{"error": "Authorization service is not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCodeHostCheck := resp.StatusCode
	fmt.Println(statusCodeHostCheck)
	if statusCodeHostCheck != 200 {
		errorMsg := map[string]string{"error": "Host with that id does not exist."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	decoder = json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var responseHost struct {
		Host struct {
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"user"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseHost); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			errorMsg := map[string]string{"error": "Invalid date format."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		fmt.Println("User has errored")
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	notificationPayload := map[string]interface{}{
		"host_id":    responseAccommodation.AccommodationHostId,
		"host_email": responseHost.Host.Email,
		"notification_text": "Dear " + responseHost.Host.Username + "\n you have a new reservation! Your " +
			responseAccommodation.AccommodationName + " has been reserved.",
	}

	notificationPayloadJSON, err := json.Marshal(notificationPayload)
	if err != nil {
		errorMsg := map[string]string{"error": "Error creating notification payload."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	notificationURL := "https://notifications-server:8089/api/notifications/create"

	timeout = 2000 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err = s.HTTPSperformAuthorizationRequestWithContextAndBody(ctx, token, notificationURL, "POST", notificationPayloadJSON)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Error creating notification request."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		errorMsg := map[string]string{"error": "Notification service is not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		errorMsg := map[string]string{"error": "Error creating notification."}
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

func (s *ReservationsHandler) GetAllReservations(rw http.ResponseWriter, h *http.Request) {
	token := h.Header.Get("Authorization")

	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Authorization service not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		errorMsg := map[string]string{"error": "Authorization service not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		errorMsg := map[string]string{"error": "Unauthorized."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(resp.Body)

	var response struct {
		LoggedInUser struct {
			ID       string        `json:"id"`
			UserRole data.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}
	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}
	guestID := response.LoggedInUser.ID
	userRole := response.LoggedInUser.UserRole

	if userRole != data.Guest {
		errorMsg := map[string]string{"error": "Permission denied. Only guests can get reservations"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusForbidden)
		return
	}

	reservations, err := s.repo.GetAllReservations(guestID)
	if err != nil {
		s.logger.Print("Error getting reservations: ", err)
		error2.ReturnJSONError(rw, err, http.StatusBadRequest)
		return
	}
	if len(reservations) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := reservations.ToJSON(rw); err != nil {
		s.logger.Println("Error encoding JSON:", err)
	}
}

func (s *ReservationsHandler) CancelReservation(rw http.ResponseWriter, h *http.Request) {
	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Authorization service not available. Try again later"}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		errorMsg := map[string]string{"error": "Authorization service not available.  Try again later"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		errorMsg := map[string]string{"error": "Unauthorized."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(resp.Body)

	var response struct {
		LoggedInUser struct {
			ID       string        `json:"id"`
			UserRole data.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}
	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}
	guestID := response.LoggedInUser.ID
	userRole := response.LoggedInUser.UserRole

	if userRole != data.Guest {
		errorMsg := map[string]string{"error": "Permission denied. Only guests can delete reservations"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusForbidden)
		return
	}
	vars := mux.Vars(h)
	reservationIDString := vars["id"]

	fmt.Println("rsv ID")
	fmt.Println(reservationIDString)

	accommodationID, err := s.repo.GetReservationAccommodationID(reservationIDString, guestID)
	if err != nil {
		errorMsg := map[string]string{"error": err.Error()}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	if err := s.repo.CancelReservationByID(guestID, reservationIDString); err != nil {
		s.logger.Println("Error canceling reservation:", err)
		if strings.Contains(err.Error(), "Cannot cancel reservation, check-in date has already started") {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"error":"Cannot cancel reservation, check-in date has already started"}`))
			return
		}
		errorMsg := map[string]string{"error": err.Error()}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	fmt.Println(accommodationID)
	fmt.Println("ACCOMMODATION ID")

	urlAccommodationCheck := "https://acc-server:8083/api/accommodations/get/" + accommodationID

	resp, err = s.HTTPSperformAuthorizationRequestWithContext(ctx, token, urlAccommodationCheck)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Accommodation service is not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		errorMsg := map[string]string{"error": "Accommodation service is not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCodeAccommodation := resp.StatusCode
	fmt.Println(statusCodeAccommodation)
	if statusCodeAccommodation != 200 {
		errorMsg := map[string]string{"error": "Accommodation with that id does not exist."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	var responseAccommodation struct {
		AccommodationName      string `json:"accommodation_name"`
		AccommodationLocation  string `json:"accommodation_location"`
		AccommodationHostId    string `json:"host_id"`
		AccommodationMinGuests int    `json:"accommodation_min_guests"`
		AccommodationMaxGuests int    `json:"accommodation_max_guests"`
	}
	decoder = json.NewDecoder(resp.Body)

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseAccommodation); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			errorMsg := map[string]string{"error": "Invalid date format."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		fmt.Println("Acommodaiton errored")
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	urlHostCheck := "https://auth-server:8080/api/users/getById/" + responseAccommodation.AccommodationHostId

	resp, err = s.HTTPSperformAuthorizationRequestWithContext(ctx, token, urlHostCheck)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Authorization service is not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		errorMsg := map[string]string{"error": "Authorization service is not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCodeHostCheck := resp.StatusCode
	fmt.Println(statusCodeHostCheck)
	if statusCodeHostCheck != 200 {
		errorMsg := map[string]string{"error": "Host with that id does not exist."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	decoder = json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var responseHost struct {
		Host struct {
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"user"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseHost); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			errorMsg := map[string]string{"error": "Invalid date format."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		fmt.Println("User has errored")
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	notificationPayload := map[string]interface{}{
		"host_id":    responseAccommodation.AccommodationHostId,
		"host_email": responseHost.Host.Email,
		"notification_text": "Dear " + responseHost.Host.Username + "\n your reservation has been cancelled! Your " +
			responseAccommodation.AccommodationName + " has been cancelled.",
	}

	notificationPayloadJSON, err := json.Marshal(notificationPayload)
	if err != nil {
		errorMsg := map[string]string{"error": "Error creating notification payload."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	notificationURL := "https://notifications-server:8089/api/notifications/create"

	timeout = 2000 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err = s.HTTPSperformAuthorizationRequestWithContextAndBody(ctx, token, notificationURL, "POST", notificationPayloadJSON)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Error creating notification request."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		errorMsg := map[string]string{"error": "Notification service is not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		errorMsg := map[string]string{"error": "Error creating notification."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (s *ReservationsHandler) GetReservationByAccommodationIdAndCheckOut(rw http.ResponseWriter, h *http.Request) {
	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Authorization service not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		errorMsg := map[string]string{"error": "Authorization service not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		errorMsg := map[string]string{"error": "Unauthorized."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(resp.Body)

	var response struct {
		LoggedInUser struct {
			ID       string        `json:"id"`
			UserRole data.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}

	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}
	vars := mux.Vars(h)
	accIDString := vars["accId"]
	userRole := response.LoggedInUser.UserRole

	if userRole != data.Host {
		errorMsg := map[string]string{"error": "Permission denied. Only hosts can see reservations for their guests"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusForbidden)
		return
	}

	counter := s.repo.GetReservationByAccommodationIDAndCheckOut(accIDString)
	if counter == -1 {
		s.logger.Println("Error fetching reservations:", counter)
		error2.ReturnJSONError(rw, counter, http.StatusBadRequest)
		return
	}

	var Number struct {
		NumberRsv int `json:"number"`
	}

	Number.NumberRsv = counter

	responseJSON, err := json.Marshal(Number)
	if err != nil {
		error2.ReturnJSONError(rw, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(responseJSON)

}

func (s *ReservationsHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		if s.logger != nil {
			s.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)
		}

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func (s *ReservationsHandler) CheckAvailability(rw http.ResponseWriter, h *http.Request) {
	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Authorization service not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		errorMsg := map[string]string{"error": "Authorization service not available."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		errorMsg := map[string]string{"error": "Unauthorized."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(resp.Body)

	var response struct {
		LoggedInUser struct {
			ID       string        `json:"id"`
			UserRole data.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}

	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	userRole := response.LoggedInUser.UserRole

	if userRole != data.Guest {
		errorMsg := map[string]string{"error": "Permission denied. Only guests can check availability of accommodations."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusForbidden)
		return
	}

	vars := mux.Vars(h)
	accIDString, ok := vars["accId"]
	if !ok {
		errorMsg := map[string]string{"error": "Missing accommodationId in the URL"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	var checkAvailabilityRequest data.CheckAvailability
	if err := json.NewDecoder(h.Body).Decode(&checkAvailabilityRequest); err != nil {
		errorMsg := map[string]string{"error": "Invalid request body. Check the request format."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	accIDConvert, err := primitive.ObjectIDFromHex(accIDString)
	checkAvailabilityRequest.CheckInDate = time.Date(
		checkAvailabilityRequest.CheckInDate.Year(),
		checkAvailabilityRequest.CheckInDate.Month(),
		checkAvailabilityRequest.CheckInDate.Day(),
		0, 0, 0, 0,
		checkAvailabilityRequest.CheckInDate.Location())

	checkAvailabilityRequest.CheckOutDate = time.Date(
		checkAvailabilityRequest.CheckOutDate.Year(),
		checkAvailabilityRequest.CheckOutDate.Month(),
		checkAvailabilityRequest.CheckOutDate.Day(),
		0, 0, 0, 0,
		checkAvailabilityRequest.CheckOutDate.Location())

	isAvailable, err := s.serviceAv.IsAvailable(
		accIDConvert,
		checkAvailabilityRequest.CheckInDate,
		checkAvailabilityRequest.CheckOutDate,
	)

	if err != nil {
		fmt.Println(err)
		errorMsg := map[string]string{"error": "Accommodation is not available for the specified dates, try another ones."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	if !isAvailable {
		errorMsg := map[string]string{"error": "Accommodation is booked for the specified dates, try another ones."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	successMsg := map[string]string{"message": "Accommodation is available for the specified dates."}
	responseJSON, err := json.Marshal(successMsg)
	if err != nil {
		error2.ReturnJSONError(rw, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(responseJSON)
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

func (s *ReservationsHandler) HTTPSperformAuthorizationRequestWithContextAndBody(
	ctx context.Context, token string, url string, method string, requestBody []byte,
) (*http.Response, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
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
