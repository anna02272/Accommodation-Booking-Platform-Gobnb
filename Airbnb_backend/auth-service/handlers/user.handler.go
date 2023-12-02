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
	//user.Name = html.EscapeString(user.Name)
	//user.Password = html.EscapeString(user.Password)
	//user.Email = html.EscapeString(user.Email)
	//user.Username = html.EscapeString(user.Username)
	//user.Lastname = html.EscapeString(user.Lastname)
	//user.Address.Country = html.EscapeString(user.Address.Country)
	//user.Address.City = html.EscapeString(user.Address.City)
	//user.Address.Street = html.EscapeString(user.Address.Street)

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

	//user.Name = html.EscapeString(user.Name)
	//user.Password = html.EscapeString(user.Password)
	//user.Email = html.EscapeString(user.Email)
	//user.Username = html.EscapeString(user.Username)
	//user.Lastname = html.EscapeString(user.Lastname)
	//user.Address.Country = html.EscapeString(user.Address.Country)
	//user.Address.City = html.EscapeString(user.Address.City)
	//user.Address.Street = html.EscapeString(user.Address.Street)

	return user, nil
}

func (ac *UserHandler) ChangePassword(ctx *gin.Context) {
	var updatePassword *domain.PasswordChangeRequest
	//updatePassword.CurrentPassword = html.EscapeString(updatePassword.CurrentPassword)
	//updatePassword.NewPassword = html.EscapeString(updatePassword.NewPassword)
	//updatePassword.ConfirmNewPassword = html.EscapeString(updatePassword.ConfirmNewPassword)

	tokenString := ctx.GetHeader("Authorization")
	//tokenString = html.EscapeString(tokenString)

	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService)
	//user.Name = html.EscapeString(user.Name)
	//user.Password = html.EscapeString(user.Password)
	//user.Email = html.EscapeString(user.Email)
	//user.Username = html.EscapeString(user.Username)
	//user.Lastname = html.EscapeString(user.Lastname)
	//user.Address.Country = html.EscapeString(user.Address.Country)
	//user.Address.City = html.EscapeString(user.Address.City)
	//user.Address.Street = html.EscapeString(user.Address.Street)

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
	tokenString := ctx.GetHeader("Authorization")
	tokenString = html.EscapeString(tokenString)

	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]
	user, err := GetUserFromToken(tokenString, ac.userService)

	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}
	urlProfile := "https://profile-server:8084/api/profile/delete/" + user.Email
	fmt.Println(urlProfile)

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := ac.HTTPSperformAuthorizationRequestWithContext(ctxRest, tokenString, urlProfile, "DELETE")
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
