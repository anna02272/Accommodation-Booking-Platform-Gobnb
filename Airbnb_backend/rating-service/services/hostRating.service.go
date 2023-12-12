package services

import "rating-service/domain"

type HostRatingService interface {
	SaveRating(rating *domain.RateHost) error
	DeleteRating(hostID, guestID string) error
	GetAllRatings() ([]*domain.RateHost, float64, error)
	GetByHostAndGuest(hostID, guestID string) ([]domain.RateHost, error)
}
