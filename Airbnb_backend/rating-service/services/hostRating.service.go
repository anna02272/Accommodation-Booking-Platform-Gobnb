package services

import "rating-service/domain"

type HostRatingService interface {
	SaveRating(rating *domain.RateHost) error
	InsertAccommodation(accomm *domain.Accommodation, hostID string) (*domain.Accommodation, string, error)
}
