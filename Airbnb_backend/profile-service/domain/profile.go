package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email" validate:"required,email"`
	Name     string             `bson:"name" json:"name"`
	Lastname string             `bson:"lastname" json:"lastname"`
	Address  Address            `bson:"address" json:"address"`
	Age      int                `bson:"age,omitempty" json:"age"`
	Gender   Gender             `bson:"gender,omitempty" json:"gender"`
	UserRole UserRole           `bson:"userRole" json:"userRole"`
}

type UserResponse struct {
	Username string   `bson:"username" json:"username"`
	Email    string   `bson:"email" json:"email" validate:"required,email"`
	Name     string   `bson:"name" json:"name"`
	Lastname string   `bson:"lastname" json:"lastname"`
	UserRole UserRole `bson:"userRole" json:"userRole"`
}

type Address struct {
	Street  string `bson:"street,omitempty" json:"street"`
	City    string `bson:"city,omitempty" json:"city"`
	Country string `bson:"country,omitempty" json:"country"`
}

type Gender string

const (
	Male   = "Male"
	Female = "Female"
	Other  = "Other"
)

type UserRole string

const (
	Guest = "Guest"
	Host  = "Host"
)
