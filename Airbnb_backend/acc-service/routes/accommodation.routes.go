package routes

import (
	"acc-service/domain"
	"acc-service/handlers"
	"acc-service/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccommodationRouteHandler struct {
	accommodationHandler handlers.AccommodationHandler
	//accommodationService services.AccommodationService
}

func NewAccommodationRouteHandler(accommodationHandler handlers.AccommodationHandler, accommodationService services.AccommodationService) AccommodationRouteHandler {
	return AccommodationRouteHandler{accommodationHandler}
}

func (rc *AccommodationRouteHandler) AccommodationRoute(rg *gin.RouterGroup) {
	router := rg.Group("/accommodations")
	router.Use(MiddlewareContentTypeSet)
	router.Use(handlers.ExtractTraceInfoMiddleware())
	router.POST("/create", MiddlewareAccommodationDeserialization, rc.accommodationHandler.CreateAccommodations)
	router.GET("/get/:id", rc.accommodationHandler.GetAccommodationByID)
	router.GET("/get", rc.accommodationHandler.GetAllAccommodations)
	router.GET("/get/host/:hostId", rc.accommodationHandler.GetAccommodationsByHostId)
	router.DELETE("/delete/:accId", rc.accommodationHandler.DeleteAccommodation)
	router.POST("/upload/images/:accId", rc.accommodationHandler.CacheAndStoreImages)
	router.GET("/images/:accId", rc.accommodationHandler.GetAccommodationImages)
	router.GET("/get/hostid/:accId", rc.accommodationHandler.GetHostIdByAccommodationId)
}

func MiddlewareContentTypeSet(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Next()
}

func MiddlewareAccommodationDeserialization(c *gin.Context) {
	var accommodation domain.AccommodationWithAvailability

	if err := c.ShouldBindJSON(&accommodation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to decode JSON"})
		c.Abort()
		return
	}

	c.Set("accommodation", accommodation)
	c.Next()
}
