package handlers

import (
	"auth-service/services"
	"auth-service/utils"
	"github.com/gin-gonic/gin"
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
	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := utils.GetUserFromToken(tokenString, ac.userService)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Token is valid", "Logged in user": user})
}
