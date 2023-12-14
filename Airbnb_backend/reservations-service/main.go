package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reservations-service/handlers"
	"reservations-service/repository"
	"reservations-service/services"
	"time"

	"github.com/gin-gonic/gin"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	server2                *gin.Engine
	ctx                    context.Context
	mongoclient            *mongo.Client
	availabilityCollection *mongo.Collection
	availabilityService    services.AvailabilityService
	AvailabilityHandler    handlers.AvailabilityHandler
	//AvailabilityRouteHandler routes.AvailabilityRouteHandler
	logger2 *log.Logger
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
	availabilityCollection = mongoclient.Database("Gobnb").Collection("availability")
	availabilityService = services.NewAvailabilityServiceImpl(availabilityCollection, ctx)

	AvailabilityHandler = handlers.NewAvailabilityHandler(availabilityService, availabilityCollection, logger2)
	logger2 = log.New(os.Stdout, "[reservation-api] ", log.LstdFlags)
	//AvailabilityRouteHandler = routes.NewAvailabilityRouteHandler(AvailabilityHandler, availabilityService)

	server2 = gin.Default()
}

func main() {
	defer mongoclient.Disconnect(ctx)
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments
	//(for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[reservation-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[reservation-store] ", log.LstdFlags)

	// NoSQL: Initialize Reservation Repository store
	store, err := repository.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}

	//serviceAv, err := services.New(storeLogger)
	//if err != nil {
	//	logger.Fatal(err)
	//}

	defer store.CloseSession()
	store.CreateTable()

	//availabilityServiceImpl := services.NewAvailabilityServiceImpl(availabilityCollection, ctx)

	reservationsHandler := handlers.NewReservationsHandler(logger, availabilityService, store, availabilityCollection)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(reservationsHandler.MiddlewareContentTypeSet)

	postReservationForGuest := router.Methods(http.MethodPost).Subrouter()
	postReservationForGuest.HandleFunc("/api/reservations/create", reservationsHandler.CreateReservationForGuest)
	postReservationForGuest.Use(reservationsHandler.MiddlewareReservationForGuestDeserialization)

	getReservationForGuest := router.Methods(http.MethodGet).Subrouter()
	getReservationForGuest.HandleFunc("/api/reservations/getAll", reservationsHandler.GetAllReservations)

	getReservationByAccommodationIdAndCheckOut := router.Methods(http.MethodGet).Subrouter()
	getReservationByAccommodationIdAndCheckOut.HandleFunc("/api/reservations/get/{accId}", reservationsHandler.GetReservationByAccommodationIdAndCheckOut)

	cancelReservationForGuest := router.Methods(http.MethodDelete).Subrouter()
	cancelReservationForGuest.HandleFunc("/api/reservations/cancel/{id}", reservationsHandler.CancelReservation)

	createAvailability := router.Methods(http.MethodPost).Subrouter()
	createAvailability.HandleFunc("/api/availability/create/{id}", AvailabilityHandler.CreateAvailability)
	createAvailability.Use(AvailabilityHandler.MiddlewareAvailabilityDeserialization)

	headersOk := gorillaHandlers.AllowedHeaders([]string{"X-Requested-With", "Authorization", "Content-Type"})
	originsOk := gorillaHandlers.AllowedOrigins([]string{"https://localhost:4200",
		"https://localhost:4200/"})
	methodsOk := gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	hanlderForHttp := gorillaHandlers.CORS(originsOk, headersOk, methodsOk)(router)

	// Serve over HTTPS
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      hanlderForHttp,
		IdleTimeout:  1000 * time.Second,
		ReadTimeout:  1000 * time.Second,
		WriteTimeout: 1000 * time.Second,
	}

	logger.Println("Server listening on port", port)

	err = server.ListenAndServeTLS("/app/reservation-service.crt", "/app/reservation_decrypted_key.pem")
	if err != nil {
		fmt.Println(err)
		return
	}

	// defer mongoclient.Disconnect(ctx)

	// corsConfig := cors.DefaultConfig()
	// corsConfig.AllowOrigins = []string{"https://localhost:4200"}
	// corsConfig.AllowCredentials = true
	// corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")

	// server2.Use(cors.New(corsConfig))

	// router2 := server2.Group("/api")
	// router2.GET("/healthchecker", func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Message"})
	// })

	// AvailabilityRouteHandler.AvailabilityRoute(router2)

	// err2 := server2.RunTLS(":8082", "/app/reservation-service.crt", "/app/reservation-service.key")
	// if err2 != nil {
	// 	fmt.Println(err2)
	// 	return
	// }

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")

}
