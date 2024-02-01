package data

import (
	"encoding/json"
	"io"
	"time"
)

type MonthlyReport struct {
	ReportID         TimeUUID `json:"report_id_time_created"`
	AccommodationID  string   `json:"accommodation_id"`
	Year             int      `json:"year"`
	Month            int      `json:"month"`
	ReservationCount int      `json:"reservation_count"`
	RatingCount      int      `json:"rating_count"`
	PageVisits       int      `json:"page_visits"`
	AverageVisitTime float64  `json:"average_visit_time"`
}

type DailyReport struct {
	ReportID         TimeUUID  `json:"report_id_time_created"`
	AccommodationID  string    `json:"accommodation_id"`
	Date             time.Time `json:"date"`
	ReservationCount int       `json:"reservation_count"`
	RatingCount      int       `json:"rating_count"`
	PageVisits       int       `json:"page_visits"`
	AverageVisitTime float64   `json:"average_visit_time"`
}

func (o *DailyReport) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *MonthlyReport) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
