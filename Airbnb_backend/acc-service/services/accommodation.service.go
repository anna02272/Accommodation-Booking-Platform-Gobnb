package services

import (
	"acc-service/domain"
)

type AccommodationService interface {
	InsertAccommodation(accomm *domain.Accommodation, hostID string) (*domain.Accommodation, string, error)
	GetAccommodationByID(accommodationID string) (*domain.Accommodation, error)
	GetAccommodationByHostIdAndAccId(hostId string, accId string) (*domain.Accommodation, error)
	GetAccommodationsByHostId(hostId string) ([]*domain.Accommodation, error)
	GetAllAccommodations() ([]*domain.Accommodation, error)
	DeleteAccommodation(accommodationID string, hostId string) error
}
