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
	ReservationIdTimeCreated TimeUUID  `json:"reservation_id_time_created"`
	GuestId                  string    `json:"guest_id"`
	AccommodationId          string    `json:"accommodation_id"`
	AccommodationName        string    `json:"accommodation_name"`
	AccommodationLocation    string    `json:"accommodation_location"`
	AccommodationHostId      string    `json:"accommodation_host_id"`
	CheckInDate              time.Time `json:"check_in_date"`
	CheckOutDate             time.Time `json:"check_out_date"`
	NumberOfGuests           int       `json:"number_of_guests"`
	IsCanceled               bool      `json:"is_canceled"`
}

//// CQRS Event Store
//type AccommodationEvent struct {
//	EventIdTimeCreated TimeUUID
//	Event              string
//	GuestID            string
//	AccommodationID    string
//}
//
//type EventJson struct {
//	Event           string `json:"event"`
//	AccommodationID string `json:"accommodation_id"`
//}

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

//	func (o *EventJson) FromJSON(r io.Reader) error {
//		d := json.NewDecoder(r)
//		return d.Decode(o)
//	}
func (reservations ReservationsByGuest) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(reservations)
}
