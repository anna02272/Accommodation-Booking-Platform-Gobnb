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
	authCollection = mongoclient.Database("Airbnb").Collection("users")
	userService = services.NewUserServiceImpl(authCollection, ctx)
	authService = services.NewAuthService(authCollection, ctx, userService)
	AuthHandler = handlers.NewAuthHandler(authService, userService)
	AuthRouteHandler = routes.NewAuthRouteHandler(AuthHandler)
	UserHandler = handlers.NewUserHandler(userService)
	UserRouteHandler = routes.NewRouteUserHandler(UserHandler)

	server = gin.Default()
}

func main() {
	defer mongoclient.Disconnect(ctx)

	client := &http.Client{
		Timeout: 5 * time.Second, // Postavite odgovarajući timeout
	}

	// Adresa profile-server servisa
	profileServerURL := "http://profile-server:8084/api/profile/createUser" // Ili druga odgovarajuća ruta

	// Slanje GET zahteva na profile-server
	resp, err := client.Get(profileServerURL)
	if err != nil {
		fmt.Println("Nije moguće uspostaviti vezu sa profile-server:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status od profile-server:", resp.Status)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true

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
