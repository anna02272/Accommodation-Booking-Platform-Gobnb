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
