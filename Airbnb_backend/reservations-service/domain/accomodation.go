package domain

type Accommodation struct {
	Name          string
	Location      string
	Benefits      string
	MinGuests     int
	MaxGuests     int
	Pictures      []string
	PricePerNight float64
	IsFree        bool
}
