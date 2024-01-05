package domain

import (
	"encoding/json"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Accommodation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	HostId    string             `bson:"host_id" json:"host_id"`
	Name      string             `bson:"accommodation_name" json:"accommodation_name"`
	Location  string             `bson:"accommodation_location" json:"accommodation_location"`
	Amenities map[string]bool    `bson:"accommodation_amenities" json:"accommodation_amenities"`
	MinGuests int                `bson:"accommodation_min_guests" json:"accommodation_min_guests"`
	MaxGuests int                `bson:"accommodation_max_guests" json:"accommodation_max_guests"`
	Active    bool               `bson:"accommodation_active" json:"accommodation_active"`
}

type Accommodations []*Accommodation

type AccommodationWithAvailability struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	HostId           string             `bson:"host_id" json:"host_id"`
	Name             string             `bson:"accommodation_name" json:"accommodation_name"`
	Location         string             `bson:"accommodation_location" json:"accommodation_location"`
	Amenities        map[string]bool    `bson:"accommodation_amenities" json:"accommodation_amenities"`
	MinGuests        int                `bson:"accommodation_min_guests" json:"accommodation_min_guests"`
	MaxGuests        int                `bson:"accommodation_max_guests" json:"accommodation_max_guests"`
	Active           bool               `bson:"accommodation_active" json:"accommodation_active"`
	StartDate        primitive.DateTime `bson:"start_date" json:"start_date"`
	EndDate          primitive.DateTime `bson:"end_date" json:"end_date"`
	Price            float64            `bson:"price" json:"price"`
	PriceType        PriceType          `bson:"price_type" json:"price_type"`
	AvailabilityType AvailabilityType   `bson:"availability_type" json:"availability_type"`
}

type AccommodationsWithAvailability []*AccommodationWithAvailability

type AvailabilityPeriod struct {
	StartDate        primitive.DateTime `bson:"start_date" json:"start_date"`
	EndDate          primitive.DateTime `bson:"end_date" json:"end_date"`
	Price            float64            `bson:"price" json:"price"`
	PriceType        PriceType          `bson:"price_type" json:"price_type"`
	AvailabilityType AvailabilityType   `bson:"availability_type" json:"availability_type"`
}
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

func (o *Accommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Accommodations) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Accommodation) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
