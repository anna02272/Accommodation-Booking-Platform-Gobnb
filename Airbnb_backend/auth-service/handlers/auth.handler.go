package handlers

import (
	"auth-service/domain"
	"auth-service/services"
	"auth-service/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type AuthHandler struct {
	authService services.AuthService
	userService services.UserService
}

func NewAuthHandler(authService services.AuthService, userService services.UserService) AuthHandler {
	return AuthHandler{authService, userService}
}

func (ac *AuthHandler) Login(ctx *gin.Context) {
	var credentials *domain.LoginInput

	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	user, err := ac.userService.FindUserByEmail(credentials.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if err := utils.VerifyNotHashedPassword(user.Password, credentials.Password); err != nil {
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
