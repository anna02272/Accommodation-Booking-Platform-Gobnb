package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"profile-service/domain"
	"regexp"
	"strings"
	"time"
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
func (us *UserServiceImpl) FindProfileByEmail(email string) (*domain.User, error) {
	var user *domain.User

	// Improved email format validation using regular expression
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, errors.New("Invalid email format")
	}

	query := bson.M{"email": strings.ToLower(email)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("User does not exist") // No user found, return nil user and nil error
		}
		return nil, err
	}
	err = us.SendUserToAuthService(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserServiceImpl) SendUserToAuthService(user *domain.User) error {
	url := "https://auth-server:8080/api/users/currentProfile"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	println("ovde")
	println(user.Name)
	println(url)
	resp, _ := us.HTTPSperformAuthorizationRequestWithContext(ctx, user, url)

	defer resp.Body.Close()

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

func (us *UserServiceImpl) HTTPSperformAuthorizationRequestWithContext(ctx context.Context, user *domain.User, url string) (*http.Response, error) {
	reqBody, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("error marshaling user JSON: %v", err)
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Perform the request with the provided context
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
