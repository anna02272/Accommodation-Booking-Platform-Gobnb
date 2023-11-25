package main

import (
	"auth-service/handlers"
	"auth-service/routes"
	"auth-service/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	server      *gin.Engine
	ctx         context.Context
	mongoclient *mongo.Client

	userService      services.UserService
	UserHandler      handlers.UserHandler
	UserRouteHandler routes.UserRouteHandler

	authCollection   *mongo.Collection
	authService      services.AuthService
	AuthHandler      handlers.AuthHandler
	AuthRouteHandler routes.AuthRouteHandler
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

	// Collections
	authCollection = mongoclient.Database("Gobnb").Collection("auth")
	userService = services.NewUserServiceImpl(authCollection, ctx)
	authService = services.NewAuthService(authCollection, ctx, userService)
	AuthHandler = handlers.NewAuthHandler(authService, userService, authCollection)
	AuthRouteHandler = routes.NewAuthRouteHandler(AuthHandler, authService)
	UserHandler = handlers.NewUserHandler(userService)
	UserRouteHandler = routes.NewRouteUserHandler(UserHandler)

	server = gin.Default()
}

func main() {
	defer mongoclient.Disconnect(ctx)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	profileServerURL := "https://profile-server:8084/api/profile/createUser"

	resp, err := client.Get(profileServerURL)
	if err != nil {
		fmt.Println("Could not connect to profile-server:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status profile-server:", resp.Status)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:4200"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Message"})
	})

	AuthRouteHandler.AuthRoute(router)
	UserRouteHandler.UserRoute(router)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(server.Run(":" + port))

}
