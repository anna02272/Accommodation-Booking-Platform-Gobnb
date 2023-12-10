package services

import "rating-service/domain"

type HostRatingService interface {
	SaveRating(rating *domain.RateHost) error
}
