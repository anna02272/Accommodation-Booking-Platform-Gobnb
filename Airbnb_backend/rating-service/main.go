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
	"os"
	"rating-service/handlers"
	"rating-service/routes"
	"rating-service/services"
)

var (
	server      *gin.Engine
	ctx         context.Context
	mongoclient *mongo.Client

	hostRatingCollection          *mongo.Collection
	hostRatingService             services.HostRatingService
	HostRatingHandler             handlers.HostRatingHandler
	accommodationRatingCollection *mongo.Collection
	accommodationRatingService    services.AccommodationRatingService
	AccommodationRatingHandler    handlers.AccommodationRatingHandler
	RatingRouteHandler            routes.RatingRouteHandler
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
	hostRatingCollection = mongoclient.Database("Gobnb").Collection("host-rating")
	hostRatingService = services.NewHostRatingServiceImpl(hostRatingCollection, ctx)
	HostRatingHandler = handlers.NewHostRatingHandler(hostRatingService, hostRatingCollection)
	accommodationRatingCollection = mongoclient.Database("Gobnb").Collection("accommodation-rating")
	accommodationRatingService = services.NewAccommodationRatingServiceImpl(accommodationRatingCollection, ctx)
	AccommodationRatingHandler = handlers.NewAccommodationRatingHandler(accommodationRatingService, accommodationRatingCollection)

	RatingRouteHandler = routes.NewRatingRouteHandler(HostRatingHandler, hostRatingService, AccommodationRatingHandler, accommodationRatingService)

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

	RatingRouteHandler.RatingRoute(router)

	err := server.RunTLS(":8087", "/app/rating-service.crt", "/app/rating-service.key")
	if err != nil {
		fmt.Println(err)
		return
	}
}
