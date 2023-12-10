package domain

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
)

type Accommodation struct {
	ID        primitive.ObjectID `bson:"id,omitempty" json:"id"`
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
