package domain

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"io"
)

type Accommodation struct {
	Name            string
	Location        string
	AccommodationId gocql.UUID
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
