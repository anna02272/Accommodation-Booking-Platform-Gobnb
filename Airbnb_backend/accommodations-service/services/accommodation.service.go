package services

import (
	"accomodations-service/domain"
)

type AccommodationService interface {
	InsertAccommodation(accommodation *domain.Accommodation, hostId string) (*domain.Accommodation, error)
	GetAccommodations(id string) (*domain.Accommodation, error)
	GetAllAccommodations() (*domain.Accommodations, error)
}
