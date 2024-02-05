package services

import (
	"acc-service/domain"
	"context"
)

type AccommodationService interface {
	InsertAccommodation(accomm *domain.AccommodationWithAvailability, hostID string, ctx context.Context) (*domain.Accommodation, string, error)
	GetAccommodationByID(accommodationID string, ctx context.Context) (*domain.Accommodation, error)
	GetAccommodationByHostIdAndAccId(hostId string, accId string, ctx context.Context) (*domain.Accommodation, error)
	GetAccommodationsByHostId(hostId string, ctx context.Context) ([]*domain.Accommodation, error)
	GetAllAccommodations(ctx context.Context) ([]*domain.Accommodation, error)
	DeleteAccommodation(accommodationID string, hostId string, ctx context.Context) error
	GetAccommodationBySearch(location string, guests string, amenities map[string]bool, amenitiesExist bool, ctx context.Context) ([]*domain.Accommodation, error)
	GetHostIdByAccommodationId(accID string) (string, error)
}
