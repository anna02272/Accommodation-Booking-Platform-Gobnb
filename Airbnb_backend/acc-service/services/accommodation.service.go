package services

import "acc-service/domain"

type AccommodationService interface {
	SaveAccommodation(accommodation *domain.Accommodation) error
	GetAccommodationById(id string) (*domain.Accommodation, error)
	GetAccommodationsByHostId(hostId string) ([]*domain.Accommodation, error)
	GetAllAccommodations() ([]*domain.Accommodation, error)
}
