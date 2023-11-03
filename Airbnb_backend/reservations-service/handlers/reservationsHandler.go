package handlers

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"reservations-service/domain"
)

type KeyProduct struct{}

type ReservationsHandler struct {
	logger *log.Logger
	// NoSQL: injecting student repository
	repo *domain.ReservationRepo
}

// Injecting the logger makes this code much more testable.
func NewReservationsHandler(l *log.Logger, r *domain.ReservationRepo) *ReservationsHandler {
	return &ReservationsHandler{l, r}
}

func (s *ReservationsHandler) GetAllGuestIds(rw http.ResponseWriter, h *http.Request) {
	guestIds, err := s.repo.GetDistinctIds("guest_id", "reservations_by_guest")
	if err != nil {
		s.logger.Print("Database exception: ", err)
	}

	if guestIds == nil {
		return
	}

	s.logger.Println(guestIds)

	e := json.NewEncoder(rw)
	err = e.Encode(guestIds)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		s.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (s *ReservationsHandler) CraeteReservationForGuest(rw http.ResponseWriter, h *http.Request) {
	guestReservation := h.Context().Value(KeyProduct{}).(*domain.ReservationByGuest)
	err := s.repo.InsertReservationByGuest(guestReservation)
	if err != nil {
		s.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (s *ReservationsHandler) GetReservationsByGuest(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	guestId := vars["id"]

	reservationsByGuest, err := s.repo.GetReservationsByGuest(guestId)
	if err != nil {
		s.logger.Print("Database exception: ", err)
	}

	if reservationsByGuest == nil {
		return
	}

	err = reservationsByGuest.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		s.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (s *ReservationsHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		s.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func (s *ReservationsHandler) MiddlewareReservationForGuestDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		patient := &domain.ReservationByGuest{}
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
