package services

import (
	"acc-service/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccommodationService interface {
	SaveAccommodation(accommodation *domain.Accommodation) error
	GetAccommodationById(id primitive.ObjectID) (*domain.Accommodation, error)
	GetAccommodationsByHostId(hostId string) ([]*domain.Accommodation, error)
	GetAllAccommodations() ([]*domain.Accommodation, error)
}
