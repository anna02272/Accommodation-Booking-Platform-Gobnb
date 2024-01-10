package create_accommodation

import "go.mongodb.org/mongo-driver/bson/primitive"

type AccommodationWithAvailability struct {
	ID               primitive.ObjectID
	HostId           string
	Name             string
	Location         string
	Amenities        map[string]bool
	MinGuests        int
	MaxGuests        int
	Active           bool
	StartDate        primitive.DateTime
	EndDate          primitive.DateTime
	Price            float64
	PriceType        PriceType
	AvailabilityType AvailabilityType
}

type AccommodationsWithAvailability []*AccommodationWithAvailability

type PriceType string

const (
	PerPerson PriceType = "PerPerson"
	PerDay    PriceType = "PerDay"
)

type AvailabilityType string

const (
	Available   AvailabilityType = "Available"
	Unavailable AvailabilityType = "Unavailable"
	Booked      AvailabilityType = "Booked"
)

type CreateAccommodationCommandType int8

const (
	AddAccommodation CreateAccommodationCommandType = iota
	RollbackAccommodation
	AddAvailability
	CancelAvailability
	UnknownCommand
)

type CreateAccommodationCommand struct {
	Accommodation AccommodationWithAvailability
	Type          CreateAccommodationCommandType
}

type CreateAccommodationReplyType int8

const (
	AccommodationAdded CreateAccommodationReplyType = iota
	AccommodationNotAdded
	AccommodationRolledBack
	AvailabilityAdded
	AvailabilityNotAdded
	UnknownReply
)

type CreateAccommodationReply struct {
	Accommodation AccommodationWithAvailability
	Type          CreateAccommodationReplyType
}
