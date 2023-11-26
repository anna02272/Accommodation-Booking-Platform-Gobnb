package handlers

import (
	"auth-service/domain"
	"auth-service/services"
	"auth-service/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"html"
	"net/http"
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
	user.Name = html.EscapeString(user.Name)
	user.Password = html.EscapeString(user.Password)
	user.Email = html.EscapeString(user.Email)
	user.Username = html.EscapeString(user.Username)
	user.Lastname = html.EscapeString(user.Lastname)
	user.Address.Country = html.EscapeString(user.Address.Country)
	user.Address.City = html.EscapeString(user.Address.City)
	user.Address.Street = html.EscapeString(user.Address.Street)

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

	user.Name = html.EscapeString(user.Name)
	user.Password = html.EscapeString(user.Password)
	user.Email = html.EscapeString(user.Email)
	user.Username = html.EscapeString(user.Username)
	user.Lastname = html.EscapeString(user.Lastname)
	user.Address.Country = html.EscapeString(user.Address.Country)
	user.Address.City = html.EscapeString(user.Address.City)
	user.Address.Street = html.EscapeString(user.Address.Street)

	return user, nil
}

func (ac *UserHandler) ChangePassword(ctx *gin.Context) {
	var updatePassword *domain.PasswordChangeRequest
	updatePassword.CurrentPassword = html.EscapeString(updatePassword.CurrentPassword)
	updatePassword.NewPassword = html.EscapeString(updatePassword.NewPassword)
	updatePassword.ConfirmNewPassword = html.EscapeString(updatePassword.ConfirmNewPassword)

	tokenString := ctx.GetHeader("Authorization")
	tokenString = html.EscapeString(tokenString)

	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService)
	user.Name = html.EscapeString(user.Name)
	user.Password = html.EscapeString(user.Password)
	user.Email = html.EscapeString(user.Email)
	user.Username = html.EscapeString(user.Username)
	user.Lastname = html.EscapeString(user.Lastname)
	user.Address.Country = html.EscapeString(user.Address.Country)
	user.Address.City = html.EscapeString(user.Address.City)
	user.Address.Street = html.EscapeString(user.Address.Street)

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
