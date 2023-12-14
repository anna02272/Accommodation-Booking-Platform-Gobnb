package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Availability struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	AccommodationID  primitive.ObjectID `bson:"accommodation_id" json:"accommodation_id"`
	Date             primitive.DateTime `bson:"date" json:"date"`
	Price            float64            `bson:"price" json:"price"`
	PriceType        PriceType          `bson:"price_type" json:"price_type"`
	AvailabilityType AvailabilityType   `bson:"availability_type" json:"availability_type"`
}

type AvailabilityPeriod struct {
	// ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	// AccommodationID  primitive.ObjectID `bson:"accommodation_id" json:"accommodation_id"`
	StartDate        primitive.DateTime `bson:"start_date" json:"start_date"`
	EndDate          primitive.DateTime `bson:"end_date" json:"end_date"`
	Price            float64            `bson:"price" json:"price"`
	PriceType        PriceType          `bson:"price_type" json:"price_type"`
	AvailabilityType AvailabilityType   `bson:"availability_type" json:"availability_type"`
}

type CheckAvailability struct {
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
}

type AvailabilityType string

type PriceType string

const (
	Available   AvailabilityType = "Available"
	Unavailable AvailabilityType = "Unavailable"
	Booked      AvailabilityType = "Booked"
)

const (
	PerPerson PriceType = "PerPerson"
	PerDay    PriceType = "PerDay"
)

func (a1 *Availability) ToJSON(w1 io.Writer) error {
	e1 := json.NewEncoder(w1)
	return e1.Encode(a1)
}

func (a2 *Availability) FromJSON(r2 io.Reader) error {
	d2 := json.NewDecoder(r2)
	return d2.Decode(a2)
}

func (a3 *AvailabilityPeriod) ToJSON(w3 io.Writer) error {
	e3 := json.NewEncoder(w3)
	return e3.Encode(a3)
}

func (a4 *AvailabilityPeriod) FromJSON(r4 io.Reader) error {
	d4 := json.NewDecoder(r4)
	return d4.Decode(a4)
}
