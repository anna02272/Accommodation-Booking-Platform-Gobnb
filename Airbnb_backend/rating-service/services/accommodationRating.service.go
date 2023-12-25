package services

import "rating-service/domain"

type AccommodationRatingService interface {
	SaveRating(rating *domain.RateAccommodation) error
	DeleteRating(accommodationID, guestID string) error
	GetAllRatings() ([]*domain.RateAccommodation, float64, error)
	GetByAccommodationAndGuest(accommodationID, guestID string) ([]domain.RateAccommodation, error)
}
