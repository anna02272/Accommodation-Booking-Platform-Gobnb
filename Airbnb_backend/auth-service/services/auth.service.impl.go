package services

import (
	"auth-service/domain"
	"auth-service/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewAuthService(collection *mongo.Collection, ctx context.Context) AuthService {
	return &AuthServiceImpl{collection, ctx}
}

func (uc *AuthServiceImpl) Login(*domain.LoginInput) (*domain.User, error) {
	return nil, nil
}
func (uc *AuthServiceImpl) Registration(user *domain.User) (*domain.UserResponse, error) {
	user.Name = user.Name
	user.Lastname = user.Lastname
	user.Username = user.Username
	user.Email = user.Email
	user.UserRole = user.UserRole
	user.Address = user.Address
	user.Age = user.Age

	hashedPassword, _ := utils.HashPassword(user.Password)
	user.Password = hashedPassword

	user.Gender = user.Gender

	res, err := uc.collection.InsertOne(uc.ctx, &user)

	var newUser *domain.UserResponse
	query := bson.M{"_id": res.InsertedID}

	err = uc.collection.FindOne(uc.ctx, query).Decode(&newUser)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}
