package services

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"profile-service/domain"
	"regexp"
	"strings"
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

func (uc *UserServiceImpl) DeleteUserProfile(email string) error {
	filter := bson.M{"email": email}
	_, err := uc.collection.DeleteOne(uc.ctx, filter)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	return nil
}

func (us *UserServiceImpl) FindUserByEmail(email string) error {
	var user *domain.User

	// Improved email format validation using regular expression
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("Invalid email format")
	}

	query := bson.M{"email": strings.ToLower(email)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("User does not exist") // No user found, return nil user and nil error
		}
		return err
	}

	return nil
}

func (uc *UserServiceImpl) UpdateUser(user *domain.User) error {
	filter := bson.M{"email": user.Email}

	// Prvo provjeravamo postoji li korisnik s datim emailom
	existingUser := uc.collection.FindOne(uc.ctx, filter)
	if existingUser.Err() != nil {
		return existingUser.Err()
	}
	log.Println(existingUser)

	// Ako korisnik s datim emailom postoji, vršimo ažuriranje
	update := bson.M{"$set": user}
	log.Println(update)
	_, err := uc.collection.UpdateOne(uc.ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
