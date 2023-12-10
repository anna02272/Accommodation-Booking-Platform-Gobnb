package routes

import (
	"acc-service/handlers"
	"acc-service/services"

	"github.com/gin-gonic/gin"
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
	router.GET("/get/:id", rc.accommodationHandler.GetAccommodationById)
	router.GET("/get", rc.accommodationHandler.GetAllAccommodations)
	router.GET("/get/host/:hostId", rc.accommodationHandler.GetAccommodationsByHostId)
}
