package services

import (
	"reservations-service/data"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AvailabilityService interface {
	InsertAvailability(availability *data.Availability) (*data.Availability, error)
	IsAvailable(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time) (bool, error)
	BookAccommodation(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time) error
	// GetAvailabilityByID(availabilityID string) (*data.Availability, error)
	// GetAvailabilitysByHostId(hostId string) ([]*data.Availability, error)
	// GetAllAvailabilitys() ([]*data.Availability, error)
}
