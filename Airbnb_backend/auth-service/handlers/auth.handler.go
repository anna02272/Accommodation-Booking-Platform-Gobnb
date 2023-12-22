package handlers

import (
	"auth-service/domain"
	"auth-service/services"
	"auth-service/utils"
	"context"
	"errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	Tracer      trace.Tracer
}

func NewAuthHandler(authService services.AuthService, userService services.UserService, db *mongo.Collection, tr trace.Tracer) AuthHandler {
	return AuthHandler{authService, userService, db, tr}
}

func (ac *AuthHandler) Login(ctx *gin.Context) {
	//_, span := ac.Tracer.Start(ctx.Request.Context(), "AuthHandler.Login")
	//defer span.End()

	var credentials *domain.LoginInput
	var userVerif *domain.Credentials

	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		//span.SetStatus(codes.Error, err.Error())
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
			//span.SetStatus(codes.Error, "Invalid email")
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email"})
			return
		} else if err == errors.New("invalid email format") {
			//span.SetStatus(codes.Error, "Invalid email format")
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email format"})
			return
		} else {
			//span.SetStatus(codes.Error, err.Error())
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			return
		}
	}
	userVerif, err = ac.userService.FindCredentialsByEmail(credentials.Email)
	if err != nil {
		//span.SetStatus(codes.Error, "Wrong credentials")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Wrong credentials"})
		return
	}

	if !userVerif.Verified {
		//span.SetStatus(codes.Error, "Please verify your email")
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "Please verify your email"})
		return
	}

	if err := utils.VerifyPassword(user.Password, credentials.Password); err != nil {
		//span.SetStatus(codes.Error, "Invalid password")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid password"})
		return
	}

	accessToken, err := utils.CreateToken(user.Username)
	if err != nil {
		//span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	//span.SetStatus(codes.Ok, "Login successful")
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "accessToken": accessToken})
}

func (ac *AuthHandler) Registration(ctx *gin.Context) {
	_, span := ac.Tracer.Start(ctx.Request.Context(), "AuthHandler.Registration")
	defer span.End()

	var user *domain.User
	rw := ctx.Writer

	if err := ctx.ShouldBindJSON(&user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	existingUser, err := ac.userService.FindUserByUsername(user.Username)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}
	if existingUser != nil {
		span.SetStatus(codes.Error, "Username already exists")
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Username already exists"})
		return
	}

	existingUser1, err := ac.userService.FindUserByEmail(user.Email)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Internal Server Error"})
		return
	}
	if existingUser1 != nil {
		span.SetStatus(codes.Error, "Email already exists")
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Email already exists"})
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
		span.SetStatus(codes.Error, "Invalid password format")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid password format"})
		return
	}

	passwordExistsBlackList, err := utils.CheckBlackList(user.Password, "blacklist.txt")

	if passwordExistsBlackList {
		span.SetStatus(codes.Error, "Password is in blacklist!")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Password is in blacklist!"})
		return
	}
	newUser, err := ac.authService.Registration(rw, user, ctx)

	if err != nil {
		if strings.Contains(err.Error(), "email already exist") {
			span.SetStatus(codes.Error, err.Error())
			ctx.JSON(http.StatusConflict, gin.H{"status": "error", "message": err.Error()})
			return
		}
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	message := "We sent an email with a verification code to " + newUser.Email
	span.SetStatus(codes.Ok, "Registration successful")
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
	if updatedUser.VerifyAt.Before(time.Now()) {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "The verify token has expired "})
		return
	}

	if updatedUser.Verified {
		log.Printf("User already verified: %s", updatedUser.Email)
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "User already verified"})
		return
	}

	updatedUser.VerificationCode = ""
	updatedUser.Verified = true
	updatedUser.VerifyAt = time.Time{}

	_, err = ac.DB.UpdateOne(context.TODO(),
		bson.M{"_id": updatedUser.ID},
		bson.M{"$set": bson.M{"verificationCode": "", "verifyAt": time.Time{}, "verified": true}})
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
	var user *domain.Credentials
	var updatedUser *domain.Credentials

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

	passwordExistsBlackList, err := utils.CheckBlackList(payload.Password, "blacklist.txt")

	if passwordExistsBlackList {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Password is in blacklist!"})
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

func (ac *AuthHandler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func ExtractTraceInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
