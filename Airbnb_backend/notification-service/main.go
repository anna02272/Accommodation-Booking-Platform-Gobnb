package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"notification-service/handlers"
	"notification-service/routes"
	"notification-service/services"
	"os"
)

var (
	server                   *gin.Engine
	ctx                      context.Context
	mongoclient              *mongo.Client
	notificationCollection   *mongo.Collection
	notificationService      services.NotificationService
	NotificationHandler      handlers.NotificationHandler
	NotificationRouteHandler routes.NotificationRouteHandler
)

func init() {
	ctx = context.TODO()

	mongoconn := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI"))
	mongoclient, err := mongo.Connect(ctx, mongoconn)

	if err != nil {
		panic(err)
	}

	if err := mongoclient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("MongoDB successfully connected...")

	notificationCollection = mongoclient.Database("Gobnb").Collection("notification")
	notificationService = services.NewNotificationServiceImpl(notificationCollection, ctx)
	NotificationHandler = handlers.NewNotificationHandler(notificationService, notificationCollection)
	NotificationRouteHandler = routes.NewNotificationRouteHandler(NotificationHandler, notificationService)

	server = gin.Default()
}

func main() {
	defer mongoclient.Disconnect(ctx)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"https://localhost:4200"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Message"})
	})

	NotificationRouteHandler.NotificationRoute(router)

	err := server.RunTLS(":8089", "/app/notifications-service.crt", "/app/notifications_decrypted_key.pem")
	if err != nil {
		fmt.Println(err)
		return
	}
}
