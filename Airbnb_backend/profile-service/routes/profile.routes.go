package routes

import (
	"profile-service/handlers"

	"github.com/gin-gonic/gin"
)

type ProfileRouteHandler struct {
	profileHandler handlers.ProfileHandler
}

func NewRouteProfileHandler(profileHandler handlers.ProfileHandler) ProfileRouteHandler {
	return ProfileRouteHandler{profileHandler}
}

func (rc *ProfileRouteHandler) ProfileRoute(rg *gin.RouterGroup) {
	router := rg.Group("/profile")
	router.Use(handlers.ExtractTraceInfoMiddleware())
	router.POST("/createUser", rc.profileHandler.CreateProfile)
	router.DELETE("/delete/:email", rc.profileHandler.DeleteProfile)
	router.POST("/updateUser", rc.profileHandler.UpdateUser)
	router.GET("/getUser/:email", rc.profileHandler.FindUserByEmail)
	router.GET("/isFeatured/:hostId", rc.profileHandler.IsFeatured)
	router.POST("/setFeatured/:hostId", rc.profileHandler.SetFeatured)
	router.POST("/setUnfeatured/:hostId", rc.profileHandler.SetUnfeatured)
}
