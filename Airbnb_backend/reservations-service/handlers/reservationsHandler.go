package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reservations-service/data"
	error2 "reservations-service/error"
	"reservations-service/repository"
)

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
	fmt.Printf("Before url")
	url := "http://auth-server:8080/api/users/currentUser"
	fmt.Printf("After url")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		error2.ReturnJSONError(rw, "Error performing request", http.StatusBadRequest)
		return
	}
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error occurred")
		error2.ReturnJSONError(rw, "Error performing request", http.StatusBadRequest)
		return
	}

	statusCode := resp.StatusCode
	if statusCode != 200 {
		error2.ReturnJSONError(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read the response body
	// Create a JSON decoder for the response body
	decoder := json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var response struct {
		LoggedInUser struct {
			ID string `json:"id"`
		} `json:"Logged in user"`
		Message string `json:"message"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&response); err != nil {
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	// Access the 'id' from the decoded struct
	guestId := response.LoggedInUser.ID
	guestReservation := h.Context().Value(KeyProduct{}).(*data.ReservationByGuestCreate)
	errReservation := s.repo.InsertReservationByGuest(guestReservation, guestId)
	if errReservation != nil {
		s.logger.Print("Database exception: ", errReservation)
		error2.ReturnJSONError(rw, "Reservation with that guest_id, accommodation_id, and check-in date exists.",
			http.StatusBadRequest)
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
