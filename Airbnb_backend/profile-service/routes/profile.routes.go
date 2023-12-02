package routes

import (
	"github.com/gin-gonic/gin"
	"profile-service/handlers"
)

type ProfileRouteHandler struct {
	profileHandler handlers.ProfileHandler
}

func NewRouteProfileHandler(profileHandler handlers.ProfileHandler) ProfileRouteHandler {
	return ProfileRouteHandler{profileHandler}
}

func (rc *ProfileRouteHandler) ProfileRoute(rg *gin.RouterGroup) {
	router := rg.Group("/profile")
	router.POST("/createUser", rc.profileHandler.CreateProfile)
	router.DELETE("/delete/:email", rc.profileHandler.DeleteProfile)

}
