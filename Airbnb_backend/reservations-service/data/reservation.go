package data

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"io"
	"time"
)

type ReservationByGuest struct {
	ReservationIdTimeCreated gocql.UUID
	GuestId                  string
	AccommodationId          gocql.UUID
	AccommodationName        string
	AccommodationLocation    string
	CheckInDate              time.Time
	CheckOutDate             time.Time
}

type ReservationByGuestCreate struct {
	ReservationIdTimeCreated gocql.UUID `json:"accommodation_id_time_created"`
	AccommodationId          gocql.UUID `json:"accommodation_id"`
	AccommodationName        string     `json:"accommodation_name"`
	AccommodationLocation    string     `json:"accommodation_location"`
	CheckInDate              time.Time  `json:"check_in_date"`
	CheckOutDate             time.Time  `json:"check_out_date"`
}

type ReservationsByGuest []*ReservationByGuest

func (o *ReservationByGuestCreate) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *ReservationByGuestCreate) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
