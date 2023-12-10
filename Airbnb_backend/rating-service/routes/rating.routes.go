package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rating-service/domain"
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
	router.Use(MiddlewareContentTypeSet)
	router.POST("/rateHost/:hostId", rc.hostRatingHandler.RateHost)
	router.POST("/create", MiddlewareAccommodationDeserialization, rc.hostRatingHandler.CreateAccommodations)
}
func MiddlewareContentTypeSet(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Next()
}

func MiddlewareAccommodationDeserialization(c *gin.Context) {
	var accommodation domain.Accommodation

	if err := c.ShouldBindJSON(&accommodation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to decode JSON"})
		c.Abort()
		return
	}

	c.Set("accommodation", accommodation)
	c.Next()
}
