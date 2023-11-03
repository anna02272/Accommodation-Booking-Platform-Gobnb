package routes

import (
	"auth-service/handlers"
	"auth-service/services"
	"auth-service/utils"
	"github.com/gin-gonic/gin"
)

type UserRouteHandler struct {
	userHandler handlers.UserHandler
}

func NewRouteUserHandler(userHandler handlers.UserHandler) UserRouteHandler {
	return UserRouteHandler{userHandler}
}

func (uc *UserRouteHandler) UserRoute(rg *gin.RouterGroup, userService services.UserService) {

	router := rg.Group("users")
	router.Use(utils.DeserializeUser(userService))
	router.GET("/me", uc.userHandler.GetCurrentUser)
}
