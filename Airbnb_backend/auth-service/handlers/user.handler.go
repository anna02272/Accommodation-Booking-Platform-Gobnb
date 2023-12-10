package handlers

import (
	"auth-service/domain"
	"auth-service/services"
	"auth-service/utils"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"html"
	"net/http"
	"time"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) UserHandler {
	return UserHandler{userService}
}

func (ac *UserHandler) CurrentUser(ctx *gin.Context) {
	tokenString := ctx.GetHeader("Authorization")
	tokenString = html.EscapeString(tokenString)

	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Token is valid", "user": user})
}
func GetUserFromToken(tokenString string, userService services.UserService) (*domain.User, error) {
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
	userID := ctx.Param("userId")

	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	user, err := uh.userService.FindUserById(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (ac *UserHandler) ChangePassword(ctx *gin.Context) {
	var updatePassword *domain.PasswordChangeRequest

	tokenString := ctx.GetHeader("Authorization")

	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	if err := ctx.ShouldBindJSON(&updatePassword); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if !utils.ValidatePassword(updatePassword.NewPassword) {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid password format"})
		return
	}

	if updatePassword.NewPassword != updatePassword.ConfirmNewPassword {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Passwords do not match"})
		return
	}
	if err := utils.VerifyPassword(user.Password, updatePassword.CurrentPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Current password and new password do not match"})
		return
	}
	hashedNewPassword, err := utils.HashPassword(updatePassword.NewPassword)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash new password"})
		return
	}

	user.Password = hashedNewPassword

	if err := ac.userService.UpdateUser(user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (ac *UserHandler) DeleteUser(ctx *gin.Context) {
	tokenStringHeader := ctx.GetHeader("Authorization")
	tokenString := html.EscapeString(tokenStringHeader)

	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]
	user, err := GetUserFromToken(tokenString, ac.userService)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	fmt.Println(user.UserRole)
	if user.UserRole == "Guest" {
		fmt.Println("here")

		urlCheckReservations := "https://res-server:8082/api/reservations/getAll"
		fmt.Println(urlCheckReservations)

		timeout := 2000 * time.Second // Adjust the timeout duration as needed
		ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		respRes, errRes := ac.HTTPSperformAuthorizationRequestWithContext(ctxRest, tokenStringHeader, urlCheckReservations, "GET")
		if errRes != nil {
			fmt.Println(err)
			if ctx.Err() == context.DeadlineExceeded {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch user reservations"})
				return
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch user reservations"})
			return
		}
		defer respRes.Body.Close()

		fmt.Println(respRes.StatusCode)
		if respRes.StatusCode != 404 {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "You cannot delete your profile, you have active reservations"})
			return
		}
	}

	if user.UserRole == "Host" {
		fmt.Println("here")

		userIDString := user.ID.String()
		urlCheckReservations := "https://acc-server:8083/api/accommodations/get/" + userIDString
		fmt.Println(urlCheckReservations)

		timeout := 2000 * time.Second // Adjust the timeout duration as needed
		ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		respRes, errRes := ac.HTTPSperformAuthorizationRequestWithContext(ctxRest, tokenStringHeader, urlCheckReservations, "GET")
		if errRes != nil {
			fmt.Println(err)
			if ctx.Err() == context.DeadlineExceeded {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch host accommodations"})
				return
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch host accommodations"})
			return
		}
		defer respRes.Body.Close()

		fmt.Println(respRes.StatusCode)
		if respRes.StatusCode != 404 {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "You cannot delete your profile, you have created accommodations"})
			return
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
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}

	err = ac.userService.DeleteCredentials(user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (s *UserHandler) HTTPSperformAuthorizationRequestWithContext(ctx context.Context, token string, url string, method string) (*http.Response, error) {
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
