package main

import (
	"accomodations-service/domain"
	"accomodations-service/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments
	//(for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8083"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[accommodation-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[accommodation-store] ", log.LstdFlags)

	// NoSQL: Initialize Accommodation Repository store
	store, err := domain.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()
	store.CreateTable()

	accommodationsHandler := handlers.NewAccommodationsHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(accommodationsHandler.MiddlewareContentTypeSet)

	postAccommodation := router.Methods(http.MethodPost).Subrouter()
	postAccommodation.HandleFunc("/api/accommodations/create", accommodationsHandler.CreateAccommodations)
	postAccommodation.Use(accommodationsHandler.MiddlewareAccommodationDeserialization)

	getAccommodationById := router.Methods(http.MethodGet).Subrouter()
	getAccommodationById.HandleFunc("/api/accommodations/get/{id:[a-zA-Z0-9-]+}", accommodationsHandler.GetAccommodationById)

	setAccommodationAvailabilty := router.Methods(http.MethodPost).Subrouter()
	setAccommodationAvailabilty.HandleFunc("/api/accommodations/availability/{id:[a-zA-Z0-9-]+}", accommodationsHandler.SetAccommodationAvailability)

	setAccommodationPrice := router.Methods(http.MethodPost).Subrouter()
	setAccommodationPrice.HandleFunc("/api/accommodations/price/{id:[a-zA-Z0-9-]+}", accommodationsHandler.SetAccommodationPrice)

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	logger.Println("Server listening on port", port)
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

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
