package domain

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"io"
	"time"
)

type ReservationByGuest struct {
	ReservationId gocql.UUID
	GuestId       gocql.UUID
	Accommodation Accommodation
	CheckInDate   time.Time
	CheckOutDate  time.Time
}

type ReservationsByGuest []*ReservationByGuest

func (o *ReservationsByGuest) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *ReservationByGuest) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
