package routes

import (
	"github.com/gin-gonic/gin"
	"acc-service/handlers"
	"acc-service/services"
)

type AccommodationRouteHandler struct {
	accommodationHandler handlers.AccommodationHandler
	accommodationService services.AccommodationService
}

func NewAccommodationRouteHandler(accommodationHandler handlers.AccommodationHandler, accommodationService services.AccommodationService) AccommodationRouteHandler {
	return AccommodationRouteHandler{accommodationHandler, accommodationService}
}

func (rc *AccommodationRouteHandler) AccommodationRoute(rg *gin.RouterGroup) {
	router := rg.Group("/accommodation")

	router.POST("/create", rc.accommodationHandler.AddAccommodation)
}
