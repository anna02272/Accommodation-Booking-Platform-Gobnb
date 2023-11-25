package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"os"
	"profile-service/handlers"
	"profile-service/routes"
	"profile-service/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	server      *gin.Engine
	ctx         context.Context
	mongoclient *mongo.Client

	profileService      services.ProfileService
	ProfileHandler      handlers.ProfileHandler
	ProfileRouteHandler routes.ProfileRouteHandler

	profileCollection *mongo.Collection
)
var err error

func init() {
	ctx = context.TODO()

	mongoconn := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI"))
	mongoclient, err = mongo.Connect(ctx, mongoconn)

	if err != nil {
		panic(err)
	}

	if err := mongoclient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("MongoDB successfully connected...")

	// Collections
	profileCollection = mongoclient.Database("Gobnb").Collection("profile")
	profileService = services.NewUserServiceImpl(profileCollection)
	ProfileHandler = handlers.NewProfileHandler(profileService)
	ProfileRouteHandler = routes.NewRouteProfileHandler(ProfileHandler)

	server = gin.Default()
}

func main() {

	defer mongoclient.Disconnect(ctx)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:4200"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Message"})
	})

	router.GET("/profile/createUser", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "hej", "hej": "hej"})
	})

	err = server.RunTLS(":8084", "/app/profile-service.crt", "/app/profile-service.key")
	if err != nil {
		fmt.Println(err)
		return
	}
}
