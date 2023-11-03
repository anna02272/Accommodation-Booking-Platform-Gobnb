package handlers

import (
	"auth-service/domain"
	"auth-service/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) UserHandler {
	return UserHandler{userService}
}

func (uc *UserHandler) GetCurrentUser(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(*domain.User)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": domain.FilteredResponse(currentUser)}})
}
