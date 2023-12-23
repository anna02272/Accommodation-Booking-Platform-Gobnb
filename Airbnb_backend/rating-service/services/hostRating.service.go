package services

import (
	"context"
	"rating-service/domain"
)

type HostRatingService interface {
	SaveRating(rating *domain.RateHost, ctx context.Context) error
	DeleteRating(hostID, guestID string, ctx context.Context) error
	GetAllRatings(ctx context.Context) ([]*domain.RateHost, float64, error)
	GetByHostAndGuest(hostID, guestID string, ctx context.Context) ([]domain.RateHost, error)
}
