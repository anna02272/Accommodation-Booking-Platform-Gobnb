package services

import (
	"auth-service/config"
	"auth-service/domain"
	"auth-service/utils"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strings"
	"time"
)

type AuthServiceImpl struct {
	collection  *mongo.Collection
	ctx         context.Context
	userService UserService
}

func NewAuthService(collection *mongo.Collection, ctx context.Context, userService UserService) AuthService {
	return &AuthServiceImpl{collection, ctx, userService}
}

func (uc *AuthServiceImpl) Login(*domain.LoginInput) (*domain.User, error) {
	return nil, nil
}

func (uc *AuthServiceImpl) Registration(rw http.ResponseWriter, user *domain.User) (*domain.UserResponse, error) {
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
	res, err := uc.collection.InsertOne(uc.ctx, credentials)
	if err != nil {
		return nil, err
	}

	err = uc.userService.SendUserToProfileService(rw, user)
	if err != nil {
		return nil, err
	}

	// Send Verification Email
	if err := uc.SendVerificationEmail(credentials); err != nil {
		log.Printf("Error sending verification email: %v", err)
		return nil, err
	}

	var newUser *domain.UserResponse
	query := bson.M{"_id": res.InsertedID}

	err = uc.collection.FindOne(uc.ctx, query).Decode(&newUser)
	if err != nil {
		return nil, err
	}
	return newUser, nil

}

func (uc *AuthServiceImpl) SendVerificationEmail(credentials *domain.Credentials) error {
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

func (uc *AuthServiceImpl) SendPasswordResetToken(credentials *domain.Credentials) error {
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
	email := ctx.Param("email")

	user, err := uc.userService.FindCredentialsByEmail(email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User not found"})
			return
		}
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}

	user.VerificationCode = verificationCode

	// Send the verification email
	if err := uc.SendVerificationEmail(user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error sending verification email"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Verification email resent successfully"})
}
