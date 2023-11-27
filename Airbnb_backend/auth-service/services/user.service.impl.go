package services

import (
	"auth-service/domain"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
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

	// Improved email format validation using regular expression
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, errors.New("Invalid email format")
	}

	query := bson.M{"email": strings.ToLower(email)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No user found, return nil user and nil error
		}
		return nil, err // Return other errors
	}

	return user, nil
}

func (us *UserServiceImpl) FindUserByUsername(username string) (*domain.User, error) {
	//var user *domain.User
	//
	//query := bson.M{"username": strings.ToLower(username)}
	//err := us.collection.FindOne(us.ctx, query).Decode(&user)
	//
	//if err != nil {
	//	if err == mongo.ErrNoDocuments {
	//		return &domain.User{}, err
	//	}
	//	return nil, err
	//}
	//
	//return user, nil
	var user *domain.User

	query := bson.M{"username": strings.ToLower(username)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No user found, return nil user and nil error
		}
		return nil, err // Return other errors
	}

	return user, nil
}
func (us *UserServiceImpl) FindCredentialsByEmail(email string) (*domain.Credentials, error) {
	var user *domain.Credentials

	query := bson.M{"email": strings.ToLower(email)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.Credentials{}, err
		}
		return nil, err
	}

	return user, nil
}
func (us *UserServiceImpl) SendUserToProfileService(user *domain.User) error {
	url := "https://profile-server:8084/api/profile/createUser"

	reqBody, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error marshaling user JSON: %v", err)
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := &http.Client{Transport: tr}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error making HTTPS request: %v", err)
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

func (us *UserServiceImpl) FindUserByVerifCode(ctx *gin.Context) (*domain.Credentials, error) {
	verificationCode := ctx.Params.ByName("verificationCode")

	var user *domain.Credentials
	query := bson.M{"verificationCode": verificationCode}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.Credentials{}, err
		}
		return nil, err
	}

	return user, nil
}
func (us *UserServiceImpl) FindUserByResetPassCode(ctx *gin.Context) (*domain.Credentials, error) {
	passwordResetToken := ctx.Params.ByName("passwordResetToken")

	log.Printf("Received password reset code: %s", passwordResetToken)

	var user *domain.Credentials
	query := bson.M{"passwordResetToken": passwordResetToken}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.Credentials{}, err
		}
		return nil, err
	}
	return user, nil
}

func (us *UserServiceImpl) UpdateUser(user *domain.User) error {
	if user.ID.IsZero() {
		return errors.New("invalid user ID")
	}

	filter := bson.M{"_id": user.ID}

	update := bson.M{
		"$set": bson.M{
			"password": user.Password,
		},
	}

	_, err := us.collection.UpdateOne(us.ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	return nil
}
