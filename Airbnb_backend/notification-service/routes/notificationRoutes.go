package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"notification-service/domain"
	"notification-service/handlers"
	"notification-service/services"
)

type NotificationRouteHandler struct {
	notificationHandler handlers.NotificationHandler
	notificationService services.NotificationService
}

func NewNotificationRouteHandler(notificationHandler handlers.NotificationHandler, notificationService services.NotificationService) NotificationRouteHandler {
	return NotificationRouteHandler{notificationHandler, notificationService}
}

func (nr *NotificationRouteHandler) NotificationRoute(rg *gin.RouterGroup) {
	router := rg.Group("/notifications")
	router.Use(MiddlewareContentTypeSet)
	router.Use(handlers.ExtractTraceInfoMiddleware())
	router.POST("/create", MiddlewareNotificationDeserialization, nr.notificationHandler.CreateNotification)
	router.GET("/host", nr.notificationHandler.GetNoitifcationsForHost)

}

func MiddlewareContentTypeSet(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Next()
}

func MiddlewareNotificationDeserialization(c *gin.Context) {
	var notification domain.NotificationCreate

	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to decode JSON"})
		c.Abort()
		return
	}

	c.Set("notification", notification)
	c.Next()
}
