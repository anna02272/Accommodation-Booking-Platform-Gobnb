package domain

import (
	"encoding/json"
	"io"

	"github.com/gocql/gocql"
)

type Accommodation struct {
	Name            string     `json:"accommodation_name"`
	Location        string     `json:"accommodation_location"`
	AccommodationId gocql.UUID `json:"accommodation_id"`
	Amenities       string     `json:"accommodation_amenities"`
	MinGuests       int        `json:"accommodation_min_guests"`
	MaxGuests       int        `json:"accommodation_max_guests"`
	ImageUrl        string     `json:"accommodation_image_url"`
}

type Accommodations []*Accommodation

func (o *Accommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Accommodation) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
