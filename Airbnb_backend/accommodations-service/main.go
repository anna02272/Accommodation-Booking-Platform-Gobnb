package main

import (
	"accomodations-service/handlers"
	"accomodations-service/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	server      *gin.Engine
	ctx         context.Context
	mongoclient *mongo.Client

	accommodationService services.AccommodationService
	accommodationHandler handlers.AccommodationHandler

	accommodationCollection *mongo.Collection
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

	logger := log.New(os.Stdout, "[accommodation-api] ", log.LstdFlags)

	// Collections
	accommodationCollection = mongoclient.Database("Gobnb").Collection("accommodation")
	accommodationService = services.NewAccommodationServiceImpl(accommodationCollection)
	accommodationHandler = handlers.NewAccommodationHandler(logger, accommodationService)

	server = gin.Default()
}

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments
	//(for eg. when running multiple docker api's and want to override the default port)
	// port := os.Getenv("PORT")
	// if len(port) == 0 {
	// 	port = "8083"
	// }

	// // Initialize context
	// timeoutContext, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	// defer cancel()

	// //Initialize the logger we are going to use, with prefix and datetime for every log
	// logger := log.New(os.Stdout, "[accommodation-api] ", log.LstdFlags)
	// storeLogger := log.New(os.Stdout, "[accommodation-store] ", log.LstdFlags)

	// // NoSQL: Initialize Accommodation Repository store
	// store, err := domain.New(storeLogger)
	// if err != nil {
	// 	logger.Fatal(err)
	// }
	// defer store.CloseSession()
	// store.CreateTable()

	// accommodationHandler := handlers.NewAccommodationsHandler(logger, store)

	defer mongoclient.Disconnect(ctx)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"https://localhost:4200"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")

	server.Use(cors.New(corsConfig))

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(accommodationHandler.MiddlewareContentTypeSet)

	postAccommodation := router.Methods(http.MethodPost).Subrouter()
	postAccommodation.HandleFunc("/api/accommodations/create", accommodationHandler.CreateAccommodations)
	postAccommodation.Use(accommodationHandler.MiddlewareAccommodationDeserialization)

	getAccommodationById := router.Methods(http.MethodGet).Subrouter()
	getAccommodationById.HandleFunc("/api/accommodations/get/{id:[a-zA-Z0-9-]+}", accommodationHandler.GetAccommodationById)

	getAccommodations := router.Methods(http.MethodGet).Subrouter()
	getAccommodations.HandleFunc("/api/accommodations/get", accommodationHandler.GetAllAccommodations)

	// setAccommodationAvailabilty := router.Methods(http.MethodPost).Subrouter()
	// setAccommodationAvailabilty.HandleFunc("/api/accommodations/availability/{id:[a-zA-Z0-9-]+}", accommodationHandler.SetAccommodationAvailability)

	// setAccommodationPrice := router.Methods(http.MethodPost).Subrouter()
	// setAccommodationPrice.HandleFunc("/api/accommodations/price/{id:[a-zA-Z0-9-]+}", accommodationHandler.SetAccommodationPrice)

	// headersOk := gorillaHandlers.AllowedHeaders([]string{"X-Requested-With", "Authorization", "Content-Type"})
	// originsOk := gorillaHandlers.AllowedOrigins([]string{"https://localhost:4200"})
	// methodsOk := gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// handlerForHttp := gorillaHandlers.CORS(originsOk, headersOk, methodsOk)(router)

	// //Initialize the server
	// server := http.Server{
	// 	Addr:         ":" + port,
	// 	Handler:      handlerForHttp,
	// 	IdleTimeout:  1000 * time.Second,
	// 	ReadTimeout:  1000 * time.Second,
	// 	WriteTimeout: 1000 * time.Second,
	// }

	// logger.Println("Server listening on port", port)

	// err = server.ListenAndServeTLS("/app/accomm-service.crt", "/app/accomm_decrypted_key.pem")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// sigCh := make(chan os.Signal)
	// signal.Notify(sigCh, os.Interrupt)
	// signal.Notify(sigCh, os.Kill)

	// sig := <-sigCh
	// logger.Println("Received terminate, graceful shutdown", sig)

	// //Try to shutdown gracefully
	// if server.Shutdown(timeoutContext) != nil {
	// 	logger.Fatal("Cannot gracefully shutdown...")
	// }
	// logger.Println("Server stopped")

	err = server.RunTLS(":8083", "/app/accomm-service.crt", "/app/accomm_decrypted_key.pem")
	if err != nil {
		fmt.Println(err)
		return
	}

}
