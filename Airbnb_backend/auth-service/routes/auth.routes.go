package routes

import (
	"auth-service/handlers"
	"github.com/gin-gonic/gin"
)

type AuthRouteHandler struct {
	authHandler handlers.AuthHandler
}

func NewAuthRouteHandler(authHandler handlers.AuthHandler) AuthRouteHandler {
	return AuthRouteHandler{authHandler}
}

func (rc *AuthRouteHandler) AuthRoute(rg *gin.RouterGroup) {
	router := rg.Group("/auth")

	router.POST("/login", rc.authHandler.Login)

}
