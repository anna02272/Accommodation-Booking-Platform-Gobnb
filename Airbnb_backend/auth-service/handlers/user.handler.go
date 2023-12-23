package handlers

import (
	"auth-service/domain"
	"auth-service/services"
	"auth-service/utils"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"html"
	"net/http"
	"time"
)

type UserHandler struct {
	userService services.UserService
	Tracer      trace.Tracer
}

func NewUserHandler(userService services.UserService, tr trace.Tracer) UserHandler {
	return UserHandler{userService, tr}
}

var currentProfileUser *domain.User

func (ac *UserHandler) CurrentUserProfile(ctx *gin.Context) {
	spanCtx, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.CurrentUserProfile")
	defer span.End()

	tokenString := ctx.GetHeader("Authorization")
	tokenString = html.EscapeString(tokenString)

	if tokenString == "" {
		span.SetStatus(codes.Error, "Missing authorization header")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService, spanCtx, ac.Tracer)
	if err != nil {
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}
	_, err = ac.userService.FindProfileInfoByEmail(spanCtx, user.Email)
	println(currentProfileUser)
	span.SetStatus(codes.Ok, "Current user retrieved successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Token is valid", "user": currentProfileUser})
}

func (ac *UserHandler) CurrentUser(ctx *gin.Context) {
	spanCtx, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.CurrentUser")
	defer span.End()

	tokenString := ctx.GetHeader("Authorization")
	tokenString = html.EscapeString(tokenString)

	if tokenString == "" {
		span.SetStatus(codes.Error, "Missing authorization header")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService, spanCtx, ac.Tracer)
	if err != nil {
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	span.SetStatus(codes.Ok, "Current user retrieved successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Token is valid", "user": user})
}

func GetUserFromToken(tokenString string, userService services.UserService, ctx context.Context, tracer trace.Tracer) (*domain.User, error) {
	_, span := tracer.Start(ctx, "UserHandler.CurrentUserProfile")
	defer span.End()

	tokenString = html.EscapeString(tokenString)

	if err := utils.VerifyToken(tokenString); err != nil {
		return nil, err
	}

	claims, err := utils.ParseTokenClaims(tokenString)
	if err != nil {
		return nil, err
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("invalid username in token")
	}

	user, err := userService.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uh *UserHandler) GetUserById(ctx *gin.Context) {
	spanCtx, span := uh.Tracer.Start(ctx.Request.Context(), "UserHandler.GetUserById")
	defer span.End()

	userID := ctx.Param("userId")

	if userID == "" {
		span.SetStatus(codes.Error, "User ID is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	user, err := uh.userService.FindUserById(userID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "User not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	span.SetStatus(codes.Error, "Get user by id successful")
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (ac *UserHandler) ChangePassword(ctx *gin.Context) {
	spanCtx, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.ChangePassword")
	defer span.End()

	var updatePassword *domain.PasswordChangeRequest

	tokenString := ctx.GetHeader("Authorization")

	if tokenString == "" {
		span.SetStatus(codes.Error, "Missing authorization header")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService, spanCtx, ac.Tracer)

	if err != nil {
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	if err := ctx.ShouldBindJSON(&updatePassword); err != nil {
		span.SetStatus(codes.Error, "Invalid request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	passwordExistsBlackList, err := utils.CheckBlackList(updatePassword.NewPassword, "blacklist.txt")

	if passwordExistsBlackList {
		span.SetStatus(codes.Error, "Password is in blacklist!")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Password is in blacklist!"})
		return
	}

	if !utils.ValidatePassword(updatePassword.NewPassword) {
		span.SetStatus(codes.Error, "Invalid password format")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid password format"})
		return
	}

	if updatePassword.NewPassword != updatePassword.ConfirmNewPassword {
		span.SetStatus(codes.Error, "Passwords do not match")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Passwords do not match"})
		return
	}
	if err := utils.VerifyPassword(user.Password, updatePassword.CurrentPassword); err != nil {
		span.SetStatus(codes.Error, "Current password and new password do not match")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Current password and new password do not match"})
		return
	}
	hashedNewPassword, err := utils.HashPassword(updatePassword.NewPassword)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to hash new password")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash new password"})
		return
	}

	user.Password = hashedNewPassword

	if err := ac.userService.UpdateUser(user, spanCtx); err != nil {
		span.SetStatus(codes.Error, "Failed to update password")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password"})
		return
	}
	span.SetStatus(codes.Ok, "Password updated successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (ac *UserHandler) DeleteUser(ctx *gin.Context) {
	spanCtx, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.DeleteUser")
	defer span.End()

	tokenStringHeader := ctx.GetHeader("Authorization")
	tokenString := html.EscapeString(tokenStringHeader)

	if tokenString == "" {
		span.SetStatus(codes.Error, "Missing authorization header")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]
	user, err := GetUserFromToken(tokenString, ac.userService, spanCtx, ac.Tracer)

	if err != nil {
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	if user == nil {
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	fmt.Println(user.UserRole)
	if user.UserRole == "Guest" {
		fmt.Println("here")

		urlCheckReservations := "https://res-server:8082/api/reservations/getAll"
		fmt.Println(urlCheckReservations)

		timeout := 2000 * time.Second // Adjust the timeout duration as needed
		ctxRest, cancel := context.WithTimeout(spanCtx, timeout)
		defer cancel()

		req, err := http.NewRequest(http.MethodGet, urlCheckReservations, nil)
		if err != nil {
			span.RecordError(err)
		}
		otel.GetTextMapPropagator().Inject(ctxRest, propagation.HeaderCarrier(req.Header))

		respRes, errRes := ac.HTTPSperformAuthorizationRequestWithContext(ctxRest, tokenStringHeader, urlCheckReservations, "GET")
		if errRes != nil {
			fmt.Println(err)
			if ctx.Err() == context.DeadlineExceeded {
				span.SetStatus(codes.Error, "Failed to fetch user reservations")
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch user reservations"})
				return
			}
			span.SetStatus(codes.Error, "Failed to fetch user reservations")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch user reservations"})
			return
		}
		defer respRes.Body.Close()
		fmt.Println(respRes.StatusCode)
		if respRes.StatusCode != 404 {
			span.SetStatus(codes.Error, "You cannot delete your profile, you have active reservations")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "You cannot delete your profile, you have active reservations"})
			return
		}
	}

	if user.UserRole == "Host" {
		fmt.Println("here")

		//userIDString := user.ID.String()
		userIDString := user.ID.Hex()
		fmt.Println(userIDString)
		urlCheckReservations := "https://acc-server:8083/api/accommodations/get/host/" + userIDString
		fmt.Println(urlCheckReservations)

		timeout := 2000 * time.Second // Adjust the timeout duration as needed
		ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		respRes, errRes := ac.HTTPSperformAuthorizationRequestWithContext(ctxRest, tokenStringHeader, urlCheckReservations, "GET")
		if errRes != nil {
			fmt.Println(err)
			if ctx.Err() == context.DeadlineExceeded {
				span.SetStatus(codes.Error, "Failed to fetch host accommodations")
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch host accommodations"})
				return
			}
			span.SetStatus(codes.Error, "Failed to fetch host accommodations")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch host accommodations"})
			return
		}
		defer respRes.Body.Close()

		fmt.Println("Resp res log")
		fmt.Println(respRes)
		fmt.Println(respRes.StatusCode)
		var response map[string]interface{}
		if err := json.NewDecoder(respRes.Body).Decode(&response); err != nil {
			span.SetStatus(codes.Error, "Failed to decode response")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to decode response"})
			return
		}
		fmt.Println("Reponse log")
		fmt.Println(response)

		if accommodations, ok := response["accommodations"].([]interface{}); ok {
			if len(accommodations) > 0 {
				span.SetStatus(codes.Error, "You cannot delete your profile, you have created accommodations")
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "You cannot delete your profile, you have created accommodations"})
				return
			}
		}
	}
	urlProfile := "https://profile-server:8084/api/profile/delete/" + user.Email
	fmt.Println(urlProfile)

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := ac.HTTPSperformAuthorizationRequestWithContext(ctxRest, tokenStringHeader, urlProfile, "DELETE")
	if err != nil {
		fmt.Println(err)
		if ctx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Failed to delete user credentials")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
			return
		}
		span.SetStatus(codes.Error, "Failed to delete user credentials")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		span.SetStatus(codes.Error, "Failed to delete user credentials")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}

	err = ac.userService.DeleteCredentials(user, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to delete user credentials")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}
	span.SetStatus(codes.Ok, "User deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (ac *UserHandler) HTTPSperformAuthorizationRequestWithContext(ctx context.Context, token string, url string, method string) (*http.Response, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	// Perform the request with the provided context
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (ac *UserHandler) CurrentProfile(ctx *gin.Context) {
	_, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.CurrentProfile")
	defer span.End()

	var user *domain.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	currentProfileUser = user

}
