package services

import (
	"auth-service/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewUserServiceImpl(collection *mongo.Collection, ctx context.Context) UserService {
	return &UserServiceImpl{collection, ctx}
}

func (us *UserServiceImpl) FindUserById(id string) (*domain.User, error) {
	oid, _ := primitive.ObjectIDFromHex(id)

	var user *domain.User

	query := bson.M{"_id": oid}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.User{}, err
		}
		return nil, err
	}

	return user, nil
}

func (us *UserServiceImpl) FindUserByEmail(email string) (*domain.User, error) {
	var user *domain.User

	query := bson.M{"email": strings.ToLower(email)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.User{}, err
		}
		return nil, err
	}

	return user, nil
}
func (us *UserServiceImpl) FindUserByUsername(username string) (*domain.User, error) {
	var user *domain.User

	query := bson.M{"username": strings.ToLower(username)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.User{}, err
		}
		return nil, err
	}

	return user, nil
}

//	func (us *UserServiceImpl) SendUserToProfileService(user *domain.User) error {
//		// Slanje HTTP zahteva ka profile-servisu
//		url := "http://localhost:8084/api/profile/createUser"
//		reqBody, err := json.Marshal(user)
//		if err != nil {
//			return err
//		}
//
//		_, err = http.Post(url, "application/json", bytes.NewBuffer(reqBody))
//		if err != nil {
//			return err
//		}
//
//		return nil
//	}
func (us *UserServiceImpl) SendUserToProfileService(user *domain.User) error {
	// Slanje HTTP zahteva ka profile-servisu
	url := "http://profile-server:8084/api/profile/createUser"
	reqBody, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error marshaling user JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusServiceUnavailable {
			return fmt.Errorf("service unavailable: %s", resp.Status)
		}

		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response status: %s, body: %s", resp.Status, body)
	}

	return nil
}
