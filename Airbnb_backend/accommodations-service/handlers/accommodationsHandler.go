package handlers

import (
	"accomodations-service/domain"
	"context"
	"log"
	"net/http"
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
	accommodation := h.Context().Value(KeyProduct{}).(*domain.Accommodation)
	err := s.repo.InsertAccommodation(accommodation)
	if err != nil {
		s.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
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
