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

type Rating string

const (
	one   = "1"
	two   = "2"
	three = "3"
	four  = "4"
	five  = "5"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email" validate:"required,email"`
}
