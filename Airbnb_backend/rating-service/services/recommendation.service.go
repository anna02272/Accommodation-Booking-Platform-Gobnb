package services

import (
	"context"
	"rating-service/domain"
)

type RecommendationService interface {
	CheckConnection()
	CloseDriverConnection(ctx context.Context)
	CreateUser(user *domain.NeoUser) error
	CreateReservation(reservation *domain.ReservationByGuest) error
	CreateAccommodation(accommodation *domain.AccommodationRec) error
	DeleteAccommodation(accommodationID string) error
}
