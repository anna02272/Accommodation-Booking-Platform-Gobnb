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

	router.POST("/rateHost/:hostId", rc.hostRatingHandler.RateHost)
}
