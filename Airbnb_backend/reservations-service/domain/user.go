package domain

import (
	"github.com/google/uuid"
)

// this needs to store user id,so we know who reserved the accomodation
type User struct {
	Id uuid.UUID
}

func (u User) Equals(user User) bool {
	return u.Id.String() == user.Id.String()
}