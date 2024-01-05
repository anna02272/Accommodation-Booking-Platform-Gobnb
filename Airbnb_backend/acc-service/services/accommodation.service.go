package services

import (
	"acc-service/domain"
	"context"
	"net/http"
)

type AccommodationService interface {
	InsertAccommodation(rw http.ResponseWriter, accomm *domain.AccommodationWithAvailability, hostID string, ctx context.Context, token string) (*domain.Accommodation, string, error)
	GetAccommodationByID(accommodationID string, ctx context.Context) (*domain.Accommodation, error)
	GetAccommodationByHostIdAndAccId(hostId string, accId string, ctx context.Context) (*domain.Accommodation, error)
	GetAccommodationsByHostId(hostId string, ctx context.Context) ([]*domain.Accommodation, error)
	GetAllAccommodations(ctx context.Context) ([]*domain.Accommodation, error)
	DeleteAccommodation(accommodationID string, hostId string, ctx context.Context) error
	GetAccommodationBySearch(location string, guests string, amenities map[string]bool, amenitiesExist bool, ctx context.Context) ([]*domain.Accommodation, error)
}
