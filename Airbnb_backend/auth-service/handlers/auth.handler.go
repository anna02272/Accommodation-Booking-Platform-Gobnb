package handlers

import (
	"auth-service/domain"
	"auth-service/services"
	"auth-service/utils"
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	authService services.AuthService
	userService services.UserService
	DB          *mongo.Collection
}

func NewAuthHandler(authService services.AuthService, userService services.UserService, db *mongo.Collection) AuthHandler {
	return AuthHandler{authService, userService, db}
}

func (ac *AuthHandler) Login(ctx *gin.Context) {
	var credentials *domain.LoginInput
	//credentials.Email = html.EscapeString(credentials.Email)
	//credentials.Password = html.EscapeString(credentials.Password)

	var userVerif *domain.Credentials

	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	credentials.Email = strings.ReplaceAll(credentials.Email, "<", "")
	credentials.Email = strings.ReplaceAll(credentials.Email, ">", "")
	credentials.Password = strings.ReplaceAll(credentials.Password, "<", "")
	credentials.Password = strings.ReplaceAll(credentials.Password, ">", "")

	user, err := ac.userService.FindUserByEmail(credentials.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email"})
			return
		} else if err == errors.New("invalid email format") {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email format"})
			return
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			return
		}
	}
	userVerif, err = ac.userService.FindCredentialsByEmail(credentials.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}

	if !userVerif.Verified {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "Please verify your email"})
		return
	}

	if err := utils.VerifyPassword(user.Password, credentials.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid password"})
		return
	}

	accessToken, err := utils.CreateToken(user.Username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "accessToken": accessToken})
}

func (ac *AuthHandler) Registration(ctx *gin.Context) {
	var user *domain.User
	//user.Name = html.EscapeString(user.Name)
	//user.Password = html.EscapeString(user.Password)
	//user.Email = html.EscapeString(user.Email)
	//user.Username = html.EscapeString(user.Username)
	//user.Lastname = html.EscapeString(user.Lastname)
	//user.Address.Country = html.EscapeString(user.Address.Country)
	//user.Address.City = html.EscapeString(user.Address.City)
	//user.Address.Street = html.EscapeString(user.Address.Street)
	//user.Name = strings.ReplaceAll(user.Name, "<", "")
	//user.Name = strings.ReplaceAll(user.Name, ">", "")
	//user.Name = strings.ReplaceAll(user.Name, "/>", "")

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	user.Name = strings.ReplaceAll(user.Name, "<", "")
	user.Name = strings.ReplaceAll(user.Name, ">", "")
	user.Password = strings.ReplaceAll(user.Password, "<", "")
	user.Password = strings.ReplaceAll(user.Password, ">", "")
	user.Email = strings.ReplaceAll(user.Email, "<", "")
	user.Email = strings.ReplaceAll(user.Email, ">", "")
	user.Lastname = strings.ReplaceAll(user.Lastname, "<", "")
	user.Lastname = strings.ReplaceAll(user.Lastname, ">", "")
	user.Address.Country = strings.ReplaceAll(user.Address.Country, "<", "")
	user.Address.Country = strings.ReplaceAll(user.Address.Country, ">", "")
	user.Address.City = strings.ReplaceAll(user.Address.City, "<", "")
	user.Address.City = strings.ReplaceAll(user.Address.City, ">", "")
	user.Address.Street = strings.ReplaceAll(user.Address.Street, "<", "")
	user.Address.Street = strings.ReplaceAll(user.Address.Street, ">", "")

	if !utils.ValidatePassword(user.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid password format"})
		return
	}

	passwordExistsBlackList, err := utils.CheckBlackList(user.Password, "blacklist.txt")

	if passwordExistsBlackList {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Password is in blacklist!"})
		return
	}
	newUser, err := ac.authService.Registration(user)

	if err != nil {
		if strings.Contains(err.Error(), "email already exist") {
			ctx.JSON(http.StatusConflict, gin.H{"status": "error", "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	message := "We sent an email with a verification code to " + newUser.Email

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "message": message})
}

func (ac *AuthHandler) VerifyEmail(ctx *gin.Context) {
	updatedUser, err := ac.userService.FindUserByVerifCode(ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Invalid verification code or user doesn't exist: %s", updatedUser.VerificationCode)
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid verification code or user doesn't exist"})
		} else {
			log.Printf("Error during verification: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		}
		return
	}

	if updatedUser.Verified {
		log.Printf("User already verified: %s", updatedUser.Email)
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "User already verified"})
		return
	}

	updatedUser.VerificationCode = ""
	updatedUser.Verified = true

	_, err = ac.DB.UpdateOne(context.TODO(),
		bson.M{"_id": updatedUser.ID},
		bson.M{"$set": bson.M{"verificationCode": "", "verified": true}})
	if err != nil {
		log.Printf("Error updating user record: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}

	log.Printf("Email verified successfully for user: %s", updatedUser.Email)
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Email verified successfully"})
}

func (ac *AuthHandler) ForgotPassword(ctx *gin.Context) {
	var payload *domain.ForgotPasswordInput
	//payload.Email = html.EscapeString(payload.Email)
	//
	var user *domain.Credentials
	//user.Username = html.EscapeString(user.Username)
	//user.Password = html.EscapeString(user.Password)
	//user.Email = html.EscapeString(user.Email)
	//user.VerificationCode = html.EscapeString(user.VerificationCode)
	//user.PasswordResetToken = html.EscapeString(user.PasswordResetToken)

	var updatedUser *domain.Credentials
	//updatedUser.Username = html.EscapeString(updatedUser.Username)
	//updatedUser.Password = html.EscapeString(updatedUser.Password)
	//updatedUser.Email = html.EscapeString(updatedUser.Email)
	//updatedUser.VerificationCode = html.EscapeString(updatedUser.VerificationCode)
	//updatedUser.PasswordResetToken = html.EscapeString(updatedUser.PasswordResetToken)

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	payload.Email = strings.ReplaceAll(payload.Email, "<", "")
	payload.Email = strings.ReplaceAll(payload.Email, ">", "")

	message := "You will receive a reset email."

	err := ac.DB.FindOne(context.TODO(), bson.M{"email": strings.ToLower(payload.Email)}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		}
		return
	}
	if !user.Verified {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Account not verified"})
		return
	}
	// Generate Reset Code
	resetToken := randstr.String(20)
	passwordResetToken := utils.Encode(resetToken)
	passwordResetAt := time.Now().Add(time.Minute * 10)

	_, err = ac.DB.UpdateOne(context.TODO(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"passwordResetToken": passwordResetToken, "passwordResetAt": passwordResetAt}})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}
	err = ac.DB.FindOne(context.TODO(), bson.M{"_id": user.ID}).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User doesnt exist"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		}
		return
	}
	// Send Password reset Email
	if err := ac.authService.SendPasswordResetToken(updatedUser); err != nil {
		log.Printf("Error sending password reset email: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error sending password reset email"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
}

func (ac *AuthHandler) ResetPassword(ctx *gin.Context) {
	var payload *domain.ResetPasswordInput
	//payload.Password = html.EscapeString(payload.Password)
	//payload.PasswordConfirm = html.EscapeString(payload.PasswordConfirm)

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	payload.Password = strings.ReplaceAll(payload.Password, "<", "")
	payload.Password = strings.ReplaceAll(payload.Password, ">", "")
	payload.PasswordConfirm = strings.ReplaceAll(payload.PasswordConfirm, "<", "")
	payload.PasswordConfirm = strings.ReplaceAll(payload.PasswordConfirm, ">", "")

	if payload.Password != payload.PasswordConfirm {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Passwords do not match"})
		return
	}

	if !utils.ValidatePassword(payload.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid password format"})
		return
	}

	hashedPassword, _ := utils.HashPassword(payload.Password)

	updatedUser, err := ac.userService.FindUserByResetPassCode(ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "The reset token is invalid "})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		}
		return
	}
	if updatedUser.PasswordResetAt.Before(time.Now()) {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "The reset token has expired "})
		return
	}
	updatedUser.Password = hashedPassword
	updatedUser.PasswordResetToken = ""
	updatedUser.PasswordResetAt = time.Time{}

	_, err = ac.DB.UpdateOne(context.TODO(),
		bson.M{"_id": updatedUser.ID},
		bson.M{"$set": bson.M{
			"password":           hashedPassword,
			"passwordResetToken": "",
			"passwordResetAt":    time.Time{}}})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Password data updated successfully"})
}
