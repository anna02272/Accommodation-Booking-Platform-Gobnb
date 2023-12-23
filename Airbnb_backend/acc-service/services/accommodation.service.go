package services

import (
	"acc-service/domain"
	"context"
)

type AccommodationService interface {
	InsertAccommodation(accomm *domain.Accommodation, hostID string, ctx context.Context) (*domain.Accommodation, string, error)
	GetAccommodationByID(accommodationID string, ctx context.Context) (*domain.Accommodation, error)
	GetAccommodationByHostIdAndAccId(hostId string, accId string, ctx context.Context) (*domain.Accommodation, error)
	GetAccommodationsByHostId(hostId string, ctx context.Context) ([]*domain.Accommodation, error)
	GetAllAccommodations(ctx context.Context) ([]*domain.Accommodation, error)
	DeleteAccommodation(accommodationID string, hostId string, ctx context.Context) error
}
