package main

import (
	"context"
	"fmt"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reservations-service/handlers"
	"reservations-service/repository"
	"time"
)

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments
	//(for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[reservation-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[reservation-store] ", log.LstdFlags)

	// NoSQL: Initialize Reservation Repository store
	store, err := repository.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()
	store.CreateTable()

	reservationsHandler := handlers.NewReservationsHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(reservationsHandler.MiddlewareContentTypeSet)

	//getReservationsByGuest := router.Methods(http.MethodGet).Subrouter()
	//getReservationsByGuest.HandleFunc("/guests/{id}", reservationsHandler.GetReservationsByGuest)

	postReservationForGuest := router.Methods(http.MethodPost).Subrouter()
	postReservationForGuest.HandleFunc("/api/reservations/create", reservationsHandler.CreateReservationForGuest)
	postReservationForGuest.Use(reservationsHandler.MiddlewareReservationForGuestDeserialization)

	headersOk := gorillaHandlers.AllowedHeaders([]string{"X-Requested-With", "Authorization", "Content-Type"})
	originsOk := gorillaHandlers.AllowedOrigins([]string{"http://localhost:4200",
		"http://localhost:4200/"})
	methodsOk := gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	hanlderForHttp := gorillaHandlers.CORS(originsOk, headersOk, methodsOk)(router)

	// Serve over HTTPS
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      hanlderForHttp,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	logger.Println("Server listening on port", port)

	err = server.ListenAndServeTLS("/app/reservations.crt", "/app/decrypted_key.pem")
	if err != nil {
		fmt.Println(err)
		return
	}

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
