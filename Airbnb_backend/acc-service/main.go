package main

import (
	"acc-service/cache"
	"acc-service/handlers"
	"acc-service/routes"
	"acc-service/services"
	"context"
	"fmt"
	"github.com/colinmarc/hdfs/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
)

var (
	server                    *gin.Engine
	ctx                       context.Context
	mongoclient               *mongo.Client
	hdfsClient                *hdfs.Client
	accommodationCollection   *mongo.Collection
	accommodationService      services.AccommodationService
	AccommodationHandler      handlers.AccommodationHandler
	AccommodationRouteHandler routes.AccommodationRouteHandler
	redisCache                *cache.ImageCache
)

func init() {
	ctx = context.TODO()

	mongoconn := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI"))
	mongoclient, err := mongo.Connect(ctx, mongoconn)

	//hdfsLogger := log.New(os.Stdout, "HDFS: ", log.LstdFlags)
	//fileStorage, err := hdfs_store.New(hdfsLogger)
	//if err != nil {
	//	panic(err)
	//}

	redisLogger := log.New(os.Stdout, "REDIS CACHE: ", log.LstdFlags)
	imageCache := cache.New(redisLogger)
	imageCache.Ping()

	if err != nil {
		panic(err)
	}

	if err := mongoclient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("MongoDB successfully connected...")

	// Collections
	accommodationCollection = mongoclient.Database("Gobnb").Collection("accommodation")
	accommodationService = services.NewAccommodationServiceImpl(accommodationCollection, ctx)
	AccommodationHandler = handlers.NewAccommodationHandler(accommodationService, accommodationCollection)
	AccommodationRouteHandler = routes.NewAccommodationRouteHandler(AccommodationHandler, accommodationService)

	redisCache = cache.New(log.New(os.Stdout, "REDIS CACHE: ", log.LstdFlags))
	redisCache.Ping()

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

	AccommodationRouteHandler.AccommodationRoute(router)

	err := server.RunTLS(":8083", "/app/accomm-service.crt", "/app/accomm_decrypted_key.pem")
	if err != nil {
		fmt.Println(err)
		return
	}
}
