package data

import (
	"github.com/google/uuid"
)

type User struct {
	Id       uuid.UUID
	UserRole UserRole
}
type UserRole string

const (
	Guest UserRole = "Guest"
	Host  UserRole = "Host"
)

func (u User) Equals(user User) bool {
	return u.Id.String() == user.Id.String()
}
