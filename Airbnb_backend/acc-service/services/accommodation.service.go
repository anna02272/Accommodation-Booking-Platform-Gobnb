package services

import "acc-service/domain"

type AccommodationService interface {
	SaveAccommodation(accommodation *domain.Accommodation) error
}
