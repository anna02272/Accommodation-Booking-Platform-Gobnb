package data

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"io"
	"time"
)

type TimeUUID gocql.UUID

func (t TimeUUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(gocql.UUID(t).String())
}

type ReservationByGuest struct {
	ReservationIdTimeCreated TimeUUID
	GuestId                  string
	AccommodationId          string
	AccommodationName        string
	AccommodationLocation    string
	AccommodationHostId      string
	CheckInDate              time.Time
	CheckOutDate             time.Time
	NumberOfGuests           int
}

type ReservationByGuestCreate struct {
	AccommodationId string    `json:"accommodation_id"`
	CheckInDate     time.Time `json:"check_in_date"`
	CheckOutDate    time.Time `json:"check_out_date"`
	NumberOfGuests  int       `json:"number_of_guests"`
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
func (reservations ReservationsByGuest) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(reservations)
}
