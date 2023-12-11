package routes

// import (
// 	"net/http"
// 	"reservations-service/data"
// 	"reservations-service/handlers"
// 	"reservations-service/services"

// 	"github.com/gin-gonic/gin"
// )

// type AvailabilityRouteHandler struct {
// 	availabilityHandler handlers.AvailabilityHandler
// 	availabilityService services.AvailabilityService
// }

// func NewAvailabilityRouteHandler(availabilityHandler handlers.AvailabilityHandler, availabilityService services.AvailabilityService) AvailabilityRouteHandler {
// 	return AvailabilityRouteHandler{availabilityHandler, availabilityService}
// }

// func (rc *AvailabilityRouteHandler) AvailabilityRoute(rg *gin.RouterGroup) {
// 	router := rg.Group("/availability")
// 	router.Use(MiddlewareContentTypeSet)
// 	router.POST("/create/:id", MiddlewareAvailabilityDeserialization, rc.availabilityHandler.CreateAvailability)
// 	// router.GET("/get/:id", rc.availabilityHandler.GetAvailabilityByID)
// 	// router.GET("/get", rc.availabilityHandler.GetAllAvailabilitys)
// 	// router.GET("/get/host/:hostId", rc.availabilityHandler.GetAvailabilitysByHostId)
// }

// func MiddlewareContentTypeSet(c *gin.Context) {
// 	c.Header("Content-Type", "application/json")
// 	c.Next()
// }

// func MiddlewareAvailabilityDeserialization(c *gin.Context) {
// 	var availability data.Availability

// 	if err := c.ShouldBindJSON(&availability); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to decode JSON"})
// 		c.Abort()
// 		return
// 	}

// 	c.Set("availability", availability)
// 	c.Next()
// }
