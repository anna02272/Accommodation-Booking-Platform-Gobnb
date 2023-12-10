package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RateHost struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Host        *User              `bson:"host" json:"host"`
	Guest       *User              `bson:"guest" json:"guest"`
	DateAndTime primitive.DateTime `bson:"date-and-time" json:"date-and-time"`
	Rating      Rating             `bson:"rating" json:"rating"`
}

type Rating int

const (
	one   = 1
	two   = 2
	three = 3
	four  = 4
	five  = 5
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email" validate:"required,email"`
}

type UserResponse struct {
	Message string `json:"message"`
	User    struct {
		ID       primitive.ObjectID `json:"id"`
		Username string             `json:"username"`
		Email    string             `json:"email"`
		Name     string             `json:"name"`
		Lastname string             `json:"lastname"`
		Address  Address            `json:"address"`
		Age      int                `json:"age"`
		Gender   string             `json:"gender"`
		UserRole UserRole           `json:"userRole"`
	} `json:"user"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
}
type UserRole string

const (
	Guest = "Guest"
	Host  = "Host"
)

func ConvertToDomainUser(userResponse UserResponse) User {
	return User{
		ID:       userResponse.User.ID,
		Username: userResponse.User.Username,
		Email:    userResponse.User.Email,
	}
}
