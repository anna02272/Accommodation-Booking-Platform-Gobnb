package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Credentials struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"password"`
	UserRole UserRole           `bson:"userRole" json:"userRole"`
	Email    string             `bson:"email" json:"email" validate:"required,email"`
}
