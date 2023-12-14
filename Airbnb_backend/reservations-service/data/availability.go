package data

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"time"
)

type Availability struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	AccommodationID  primitive.ObjectID `bson:"accommodation_id" json:"accommodation_id"`
	Date             primitive.DateTime `bson:"date" json:"date"`
	Price            float64            `bson:"price" json:"price"`
	AvailabilityType AvailabilityType   `bson:"availability_type" json:"availability_type"`
}

type CheckAvailability struct {
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
}

type AvailabilityType string

const (
	Available   AvailabilityType = "Available"
	Unavailable AvailabilityType = "Unavailable"
	Booked      AvailabilityType = "Booked"
)

func (a *Availability) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(a)
}

func (a *Availability) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(a)
}
