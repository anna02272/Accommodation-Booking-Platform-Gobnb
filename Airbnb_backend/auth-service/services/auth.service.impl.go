package services

import (
	"auth-service/config"
	"auth-service/domain"
	error2 "auth-service/error"
	"auth-service/utils"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"strings"
	"time"
)

type AuthServiceImpl struct {
	collection  *mongo.Collection
	ctx         context.Context
	userService UserService
	Tracer      trace.Tracer
}

func NewAuthService(collection *mongo.Collection, ctx context.Context, userService UserService, tr trace.Tracer) AuthService {
	return &AuthServiceImpl{collection, ctx, userService, tr}
}

func (uc *AuthServiceImpl) Login(loginInput *domain.LoginInput, ctx context.Context) (*domain.User, error) {
	ctx, span := uc.Tracer.Start(ctx, "AuthService.Login")
	defer span.End()

	return nil, nil
}

func (uc *AuthServiceImpl) Registration(rw http.ResponseWriter, user *domain.User, ctx context.Context) (*domain.UserResponse, error) {
	ctx, span := uc.Tracer.Start(ctx, "AuthService.Registration")
	defer span.End()

	hashedPassword, _ := utils.HashPassword(user.Password)
	user.Password = hashedPassword
	code := randstr.String(20)
	verificationCode := utils.Encode(code)

	credentials := &domain.Credentials{
		ID:               primitive.NewObjectID(),
		Username:         user.Username,
		Password:         hashedPassword,
		UserRole:         user.UserRole,
		Email:            user.Email,
		Verified:         false,
		VerificationCode: verificationCode,
		VerifyAt:         time.Now().Add(time.Minute * 10),
	}
	res, err := uc.collection.InsertOne(ctx, credentials)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	_ = uc.SendToRatingService(credentials, ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, nil
	}

	err = uc.userService.SendUserToProfileService(rw, user, ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Send Verification Email
	if err := uc.SendVerificationEmail(credentials, ctx); err != nil {
		span.SetStatus(codes.Error, err.Error())
		log.Printf("Error sending verification email: %v", err)
		return nil, err
	}

	var newUser *domain.UserResponse
	query := bson.M{"_id": res.InsertedID}

	err = uc.collection.FindOne(ctx, query).Decode(&newUser)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return newUser, nil

}
func (uc *AuthServiceImpl) SendToRatingService(user *domain.Credentials, ctx context.Context) error {
	ctx, span := uc.Tracer.Start(ctx, "AuthService.SendToRating")
	defer span.End()

	var rw http.ResponseWriter
	url := "https://rating-server:8087/api/rating/createUser"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctxx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := uc.HTTPSperformAuthorizationRequest(ctx, user, url)
	if err != nil {
		if ctxx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Rating service not available..")
			errorMsg := map[string]string{"error": "Rating service not available.."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return nil
		}
		span.SetStatus(codes.Error, "Rating service not available..")
		errorMsg := map[string]string{"error": "Rating service not available.."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return nil
	}

	defer resp.Body.Close()

	return nil
}
func (uc *AuthServiceImpl) HTTPSperformAuthorizationRequest(ctx context.Context, user *domain.Credentials, url string) (*http.Response, error) {
	reqBody, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("error marshaling reservation JSON: %v", err)
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
func (uc *AuthServiceImpl) SendVerificationEmail(credentials *domain.Credentials, ctx context.Context) error {
	ctx, span := uc.Tracer.Start(ctx, "AuthService.SendVerificationEmail")
	defer span.End()

	var username = credentials.Username
	if strings.Contains(username, " ") {
		username = strings.Split(username, " ")[1]
	}

	emailData := utils.EmailData{
		URL:      credentials.VerificationCode,
		Username: username,
		Subject:  "Your account verification code",
	}
	config := config.LoadConfig()
	return utils.SendEmail(credentials, &emailData, config)
}

func (uc *AuthServiceImpl) SendPasswordResetToken(credentials *domain.Credentials, ctx context.Context) error {
	ctx, span := uc.Tracer.Start(ctx, "AuthService.SendPasswordResetToken")
	defer span.End()

	var username = credentials.Username
	if strings.Contains(username, " ") {
		username = strings.Split(username, " ")[1]
	}

	emailData := utils.EmailData{
		URL:      credentials.PasswordResetToken,
		Username: username,
		Subject:  "Your account password reset code (valid for 10min)",
	}
	config := config.LoadConfig()
	return utils.SendEmail(credentials, &emailData, config)
}

func (uc *AuthServiceImpl) ResendVerificationEmail(ctx *gin.Context) {
	spanCtx, span := uc.Tracer.Start(ctx.Request.Context(), "AuthService.ResendVerificationEmail")
	defer span.End()
	ctx.Request = ctx.Request.WithContext(spanCtx)

	email := ctx.Param("email")

	user, err := uc.userService.FindCredentialsByEmail(email, ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Error, "User not found")
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User not found"})
			return
		}
		span.SetStatus(codes.Error, "Internal Server Error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}

	// Generate a new verification code
	code := randstr.String(20)
	verificationCode := utils.Encode(code)
	verifyAt := time.Now().Add(time.Minute * 10)
	// Update the user in the database with the new verification code
	_, err = uc.collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"verificationCode": verificationCode, "verifyAt": verifyAt, "verified": false}},
	)
	if err != nil {
		span.SetStatus(codes.Error, "Internal Server Error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}

	user.VerificationCode = verificationCode

	// Send the verification email
	if err := uc.SendVerificationEmail(user, ctx); err != nil {
		span.SetStatus(codes.Error, "Error sending verification email")
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error sending verification email"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Verification email resent successfully"})
}
