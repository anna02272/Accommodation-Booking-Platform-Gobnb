package domain

import (
	"encoding/json"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Accommodation struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	HostId   string             `bson:"hostId" json:"host_id"`
	Name     string             `bson:"name" json:"accommodation_name"`
	Location string             `bson:"location" json:"accommodation_location"`
	//AccommodationId gocql.UUID         `bson:"accommodationId" json:"accommodation_id"`
	Amenities map[string]bool `bson:"amenities" json:"accommodation_amenities"`
	MinGuests int             `bson:"minGuests" json:"accommodation_min_guests"`
	MaxGuests int             `bson:"maxGuests" json:"accommodation_max_guests"`
	Images    []string        `bson:"images" json:"accommodation_images"`
	Active    bool            `bson:"active" json:"accommodation_active"`
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
