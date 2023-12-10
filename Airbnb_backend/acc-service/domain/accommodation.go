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
	Images    []string           `bson:"accommodation_images" json:"accommodation_images"`
	Active    bool               `bson:"accommodation_active" json:"accommodation_active"`
}

type Accommodations []*Accommodation

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

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email" validate:"required,email"`
}
