package routes

import (
	"github.com/gin-gonic/gin"
	"rating-service/handlers"
	"rating-service/services"
)

type RatingRouteHandler struct {
	hostRatingHandler handlers.HostRatingHandler
	hostRatingService services.HostRatingService
}

func NewRatingRouteHandler(hostRatingHandler handlers.HostRatingHandler, hostRatingService services.HostRatingService) RatingRouteHandler {
	return RatingRouteHandler{hostRatingHandler, hostRatingService}
}

func (rc *RatingRouteHandler) RatingRoute(rg *gin.RouterGroup) {
	router := rg.Group("/rating")
	router.Use(handlers.ExtractTraceInfoMiddleware())
	router.POST("/rateHost/:hostId", rc.hostRatingHandler.RateHost)
	router.DELETE("/deleteRating/:hostId", rc.hostRatingHandler.DeleteRating)
	router.GET("/getAll", rc.hostRatingHandler.GetAllRatings)
	router.GET("/get/:hostId", rc.hostRatingHandler.GetByHostAndGuest)
}
