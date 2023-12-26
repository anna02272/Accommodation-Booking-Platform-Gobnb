package services

import (
	"context"
	"rating-service/domain"
)

type AccommodationRatingService interface {
	SaveRating(rating *domain.RateAccommodation, ctx context.Context) error
	DeleteRating(accommodationID, guestID string, ctx context.Context) error
	GetAllRatingsAccommodation(ctx context.Context) ([]*domain.RateAccommodation, float64, error)
	GetByAccommodationAndGuest(accommodationID, guestID string, ctx context.Context) ([]domain.RateAccommodation, error)
}
