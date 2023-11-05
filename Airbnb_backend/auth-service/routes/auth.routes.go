package routes

import (
	"auth-service/handlers"
	"auth-service/services"
	"github.com/gin-gonic/gin"
)

type AuthRouteHandler struct {
	authHandler handlers.AuthHandler
}

func NewAuthRouteHandler(authHandler handlers.AuthHandler) AuthRouteHandler {
	return AuthRouteHandler{authHandler}
}

func (rc *AuthRouteHandler) AuthRoute(rg *gin.RouterGroup, userService services.UserService) {
	router := rg.Group("/auth")

	router.POST("/login", rc.authHandler.Login)
	//router.GET("/logout", utils.DeserializeUser(userService), rc.authHandler.LogoutUser)
}
