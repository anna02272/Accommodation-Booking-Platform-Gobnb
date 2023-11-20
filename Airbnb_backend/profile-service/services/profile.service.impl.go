package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"profile-service/domain"
)

type UserServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewUserServiceImpl(collection *mongo.Collection) *UserServiceImpl {
	return &UserServiceImpl{collection: collection}
}

func (uc *UserServiceImpl) Registration(user *domain.User) error {

	newUser := *user

	// Treba koristiti uc.collection umesto uc.ctx
	res, err := uc.collection.InsertOne(uc.ctx, &newUser)
	if err != nil {
		return err
	}

	var responseUser *domain.UserResponse
	query := bson.M{"_id": res.InsertedID}

	err = uc.collection.FindOne(uc.ctx, query).Decode(&responseUser)
	if err != nil {
		return err
	}

	return nil
}
