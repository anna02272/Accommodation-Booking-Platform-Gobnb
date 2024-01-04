package services

import (
	"context"
	"reservations-service/data"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AvailabilityService interface {
	InsertAvailability(availability *data.Availability, ctx context.Context) (*data.Availability, error)
	InsertMulitipleAvailability(availability data.AvailabilityPeriod, accId primitive.ObjectID, ctx context.Context) ([]*data.Availability, error)
	GetAvailabilityByAccommodationId(accommodationID primitive.ObjectID, ctx context.Context) ([]*data.Availability, error)
	IsAvailable(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, ctx context.Context) (bool, error)
	BookAccommodation(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, ctx context.Context) error
	MakeAccommodationAvailable(accommodationID string, startDate time.Time, endDate time.Time, ctx context.Context) error
	GetPrices(accID primitive.ObjectID, startDate time.Time, endDate time.Time, ctx context.Context) ([]*data.PriceResponse, error)
	// GetAvailabilityByID(availabilityID string, ctx context.Context) (*data.Availability, error)
	// GetAvailabilitysByHostId(hostId string, ctx context.Context) ([]*data.Availability, error)
	// GetAllAvailabilitys() ([]*data.Availability, error, ctx context.Context)
}
