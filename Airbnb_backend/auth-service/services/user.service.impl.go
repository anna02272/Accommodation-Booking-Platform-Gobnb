package services

import (
	"auth-service/domain"
	error2 "auth-service/error"
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
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	_ "io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type UserServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
	Tracer     trace.Tracer
}

func NewUserServiceImpl(collection *mongo.Collection, ctx context.Context, tr trace.Tracer) UserService {
	return &UserServiceImpl{collection, ctx, tr}
}

func (us *UserServiceImpl) FindUserById(id string, ctx context.Context) (*domain.User, error) {
	ctx, span := us.Tracer.Start(ctx, "UserService.FindUserById")
	defer span.End()

	oid, _ := primitive.ObjectIDFromHex(id)

	var user *domain.User

	query := bson.M{"_id": oid}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Error, err.Error())
			return &domain.User{}, err
		}
		return nil, err
	}

	return user, nil
}

func (us *UserServiceImpl) FindUserByEmail(email string, ctx context.Context) (*domain.User, error) {
	ctx, span := us.Tracer.Start(ctx, "UserService.FindUserByEmail")
	defer span.End()

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
			span.SetStatus(codes.Error, err.Error())
			return nil, nil // No user found, return nil user and nil error
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, err // Return other errors
	}

	return user, nil
}

func (us *UserServiceImpl) FindUserByUsername(username string) (*domain.User, error) {
	//ctx, span := us.Tracer.Start(ctx, "UserService.FindUserByUsername")
	//defer span.End()

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
			//span.SetStatus(codes.Error, err.Error())
			return nil, nil // No user found, return nil user and nil error
		}
		//span.SetStatus(codes.Error, err.Error())
		return nil, err // Return other errors
	}

	return user, nil
}

func (us *UserServiceImpl) FindCredentialsByEmail(email string, ctx context.Context) (*domain.Credentials, error) {
	ctx, span := us.Tracer.Start(ctx, "UserService.FindCredentialsByEmail")
	defer span.End()

	var user *domain.Credentials

	query := bson.M{"email": strings.ToLower(email)}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Error, err.Error())
			return &domain.Credentials{}, err
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return user, nil
}

func (us *UserServiceImpl) FindProfileInfoByEmail(ctx context.Context, email string) (*domain.CurrentUser, error) {
	ctx, span := us.Tracer.Start(ctx, "UserService.FindProfileInfoByEmail")
	defer span.End()

	url := "https://profile-server:8084/api/profile/getUser/" + email

	resp, err := us.HTTPSperformRequestWithContext(ctx, url)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "Failed to fetch user profile from profile service")
		return nil, errors.New("Failed to fetch user profile from profile service")
	}

	var profileUser *domain.CurrentUser
	if err := json.NewDecoder(resp.Body).Decode(&profileUser); err != nil {
		span.SetStatus(codes.Error, "Failed to decode response from profile service")
		return nil, errors.New("Failed to decode response from profile service")
	}

	return profileUser, nil
}

func (us *UserServiceImpl) SendUserToProfileService(rw http.ResponseWriter, user *domain.User, ctx context.Context) error {
	ctx, span := us.Tracer.Start(ctx, "UserService.SendUserToProfileService")
	defer span.End()

	url := "https://profile-server:8084/api/profile/createUser"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := us.HTTPSperformAuthorizationRequestWithContext(ctx, user, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Profile service not available..")
			errorMsg := map[string]string{"error": "Profile service not available.."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return nil
		}
		span.SetStatus(codes.Error, "Profile service not available..")
		errorMsg := map[string]string{"error": "Profile service not available.."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return nil
	}

	defer resp.Body.Close()

	return nil
}

func (us *UserServiceImpl) FindUserByVerifCode(ctx *gin.Context, ctxt context.Context) (*domain.Credentials, error) {
	ctxt, span := us.Tracer.Start(ctx, "UserService.FindUserByVerifCode")
	defer span.End()

	verificationCode := ctx.Params.ByName("verificationCode")

	var user *domain.Credentials
	query := bson.M{"verificationCode": verificationCode}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Error, err.Error())
			return &domain.Credentials{}, err
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return user, nil
}

func (us *UserServiceImpl) FindUserByResetPassCode(ctx *gin.Context, ctxt context.Context) (*domain.Credentials, error) {
	ctxt, span := us.Tracer.Start(ctx, "UserService.FindUserByResetPassCode")
	defer span.End()

	passwordResetToken := ctx.Params.ByName("passwordResetToken")

	log.Printf("Received password reset code: %s", passwordResetToken)

	var user *domain.Credentials
	query := bson.M{"passwordResetToken": passwordResetToken}
	err := us.collection.FindOne(us.ctx, query).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Error, err.Error())
			return &domain.Credentials{}, err
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return user, nil
}

func (us *UserServiceImpl) UpdateUser(user *domain.User, ctx context.Context) error {
	ctx, span := us.Tracer.Start(ctx, "UserService.UpdateUser")
	defer span.End()

	if user.ID.IsZero() {
		span.SetStatus(codes.Error, "invalid user ID")
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
		span.SetStatus(codes.Error, "error updating user:"+err.Error())
		return fmt.Errorf("error updating user: %v", err)
	}

	return nil
}

func (us *UserServiceImpl) DeleteCredentials(user *domain.User, ctx context.Context) error {
	ctx, span := us.Tracer.Start(ctx, "UserService.UpdateUser")
	defer span.End()

	fmt.Println(user.Email)
	if user.ID.IsZero() {
		span.SetStatus(codes.Error, "invalid user ID")
		return errors.New("invalid user ID")
	}
	filter := bson.M{"email": user.Email}
	fmt.Println(user.Email + "2")
	_, err := us.collection.DeleteOne(us.ctx, filter)
	if err != nil {
		span.SetStatus(codes.Error, "error deleting user credentials:"+err.Error())
		return fmt.Errorf("error deleting user credentials: %v", err)
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

func (us *UserServiceImpl) HTTPSperformRequestWithContext(ctx context.Context, url string) (*http.Response, error) {

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
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
