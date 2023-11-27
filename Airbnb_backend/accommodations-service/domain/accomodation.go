package domain

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gocql/gocql"
)

type Accommodation struct {
	Name            string               `json:"accommodation_name"`
	Location        string               `json:"accommodation_location"`
	AccommodationId gocql.UUID           `json:"accommodation_id"`
	Amenities       string               `json:"accommodation_amenities"`
	MinGuests       int                  `json:"accommodation_min_guests"`
	MaxGuests       int                  `json:"accommodation_max_guests"`
	ImageUrl        string               `json:"accommodation_image_url"`
	Availability    map[time.Time]bool   `json:"accommodation_availability"`
	Prices          map[time.Time]string `json:"accommodation_prices"`
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

/*
func isAccommodationAvailable(accommodation *Accommodation, checkIn, checkOut time.Time) bool {
	for t := checkIn; t.Before(checkOut); t = t.AddDate(0, 0, 1) {
		if !accommodation.Availability[t] {
			return false
		}
	}
	return true
}

func getAccommodationPrice(accommodation *Accommodation, date time.Time) float64 {
	return accommodation.Prices[date]
}

func updateAccommodationAvailability(accommodation *Accommodation, newAvailability map[time.Time]bool) {
	for date, available := range newAvailability {
		if !available {
			if !isDateRangeAvailable(accommodation, date, date.AddDate(0, 0, 1)) {
				return
			}
		}
	}

	for date, available := range newAvailability {
		accommodation.Availability[date] = available
	}

}

func isDateRangeAvailable(accommodation *Accommodation, checkIn, checkOut time.Time) bool {
	for t := checkIn; t.Before(checkOut); t = t.AddDate(0, 0, 1) {
		if !accommodation.Availability[t] {
			return false
		}
	}
	return true
}
*/
