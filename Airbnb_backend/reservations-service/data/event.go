package data

import (
	"encoding/json"
	"io"
)

// CQRS Event Store
type AccommodationEvent struct {
	EventIdTimeCreated TimeUUID
	Event              string
	GuestID            string
	AccommodationID    string
}

type EventJson struct {
	Event           string `json:"event"`
	AccommodationID string `json:"accommodation_id"`
}

func (o *EventJson) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
