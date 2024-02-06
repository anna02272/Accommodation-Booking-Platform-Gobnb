package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"profile-service/domain"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type UserServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
	Tracer     trace.Tracer
}

func NewUserServiceImpl(collection *mongo.Collection, tr trace.Tracer) *UserServiceImpl {
	return &UserServiceImpl{collection: collection, Tracer: tr}
}

func (uc *UserServiceImpl) Registration(user *domain.User, ctx context.Context) error {
	ctx, span := uc.Tracer.Start(ctx, "ProfileService.Registration")
	defer span.End()
	newUser := *user

	// Treba koristiti uc.collection umesto uc.ctx
	res, err := uc.collection.InsertOne(ctx, &newUser)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	var responseUser *domain.UserResponse
	query := bson.M{"_id": res.InsertedID}

	err = uc.collection.FindOne(uc.ctx, query).Decode(&responseUser)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	// if user is a host, set it to featured
	if user.UserRole == "Host" {
		err := uc.SetUnfeatured(res.InsertedID.(string))
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}

	return nil
}

func (uc *UserServiceImpl) DeleteUserProfile(email string, ctx context.Context) error {
	ctx, span := uc.Tracer.Start(ctx, "ProfileService.DeleteUserProfile")
	defer span.End()

	filter := bson.M{"email": email}
	_, err := uc.collection.DeleteOne(ctx, filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("error deleting user: %v", err)
	}

	return nil
}

func (us *UserServiceImpl) FindUserByEmail(email string, ctx context.Context) error {
	ctx, span := us.Tracer.Start(ctx, "ProfileService.FindUserByEmail")
	defer span.End()

	var user *domain.User

	// Improved email format validation using regular expression
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		span.SetStatus(codes.Error, "Invalid email format")
		return errors.New("Invalid email format")
	}

	query := bson.M{"email": strings.ToLower(email)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Error, "User does not exist")
			return errors.New("User does not exist") // No user found, return nil user and nil error
		}
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (us *UserServiceImpl) FindProfileByEmail(email string, ctx context.Context) (*domain.User, error) {
	ctx, span := us.Tracer.Start(ctx, "UserService.FindProfileByEmail")
	defer span.End()

	var user *domain.User

	// Improved email format validation using regular expression
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		span.SetStatus(codes.Error, "Invalid email format")
		return nil, errors.New("Invalid email format")
	}

	query := bson.M{"email": strings.ToLower(email)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Error, "User does not exist")
			return nil, errors.New("User does not exist") // No user found, return nil user and nil error
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	err = us.SendUserToAuthService(user, ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return user, nil
}

func (us *UserServiceImpl) SendUserToAuthService(user *domain.User, ctx context.Context) error {
	ctx, span := us.Tracer.Start(ctx, "ProfileService.SendUserToAuthService")
	defer span.End()

	url := "https://auth-server:8080/api/users/currentProfile"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, _ := us.HTTPSperformAuthorizationRequestWithContext(ctx, user, url)

	defer resp.Body.Close()

	return nil
}

func (uc *UserServiceImpl) UpdateUser(user *domain.User, ctx context.Context) error {
	ctx, span := uc.Tracer.Start(ctx, "ProfileService.UpdateUser")
	defer span.End()

	filter := bson.M{"email": user.Email}

	// Prvo provjeravamo postoji li korisnik s datim emailom
	existingUser := uc.collection.FindOne(uc.ctx, filter)
	if existingUser.Err() != nil {
		span.SetStatus(codes.Error, existingUser.Err().Error())
		return existingUser.Err()
	}
	log.Println(existingUser)

	// Ako korisnik s datim emailom postoji, vršimo ažuriranje
	update := bson.M{"$set": user}
	log.Println(update)
	_, err := uc.collection.UpdateOne(uc.ctx, filter, update)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// func (uc *UserServiceImpl) CheckIsFeatured(user *domain.User, ctx context.Context) (bool, error) {

// 	return false, nil

// }

func (uc *UserServiceImpl) IsFeatured(hostID string) (bool, error) {

	var user *domain.User
	fmt.Println("hostId in service: ", hostID)
	oid, _ := primitive.ObjectIDFromHex(hostID)
	//var responseUser *domain.UserResponse
	query := bson.M{"_id": oid}

	var err = uc.collection.FindOne(uc.ctx, query).Decode(&user)
	if err != nil {
		//span.SetStatus(codes.Error, err.Error())

		return false, err
	}

	return user.Featured, nil

}

func (uc *UserServiceImpl) SetFeatured(hostID string) error {

	id, _ := primitive.ObjectIDFromHex(hostID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"featured": true}}
	_, err := uc.collection.UpdateOne(uc.ctx, filter, update)
	if err != nil {
		return err
	}

	return nil

}

func (uc *UserServiceImpl) SetUnfeatured(hostID string) error {

	id, _ := primitive.ObjectIDFromHex(hostID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"featured": false}}
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
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	// Perform the request with the provided context
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
