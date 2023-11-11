package data

import "github.com/gocql/gocql"

type Accommodation struct {
	Name            string
	Location        string
	AccommodationId gocql.UUID
}
