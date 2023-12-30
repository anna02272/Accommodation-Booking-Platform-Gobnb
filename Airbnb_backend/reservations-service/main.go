package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reservations-service/config"
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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var (
	server2                *gin.Engine
	ctx                    context.Context
	mongoclient            *mongo.Client
	availabilityCollection *mongo.Collection
	availabilityService    services.AvailabilityService
	AvailabilityHandler    handlers.AvailabilityHandler
	logger2                *log.Logger
)

func init() {
	ctx = context.TODO()

	cfg := config.GetConfig()
	tracerProvider, err := NewTracerProvider(cfg.ServiceName, cfg.JaegerAddress)
	if err != nil {
		log.Fatal("JaegerTraceProvider failed to Initialize", err)
	}
	tracer := tracerProvider.Tracer(cfg.ServiceName)

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
	logger2 = log.New(os.Stdout, "[reservation-api] ", log.LstdFlags)
	availabilityService = services.NewAvailabilityServiceImpl(availabilityCollection, ctx, tracer)
	AvailabilityHandler = handlers.NewAvailabilityHandler(availabilityService, availabilityCollection, logger2, tracer)

	server2 = gin.Default()
}

func main() {
	defer mongoclient.Disconnect(ctx)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[reservation-`ap`i] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[reservation-store] ", log.LstdFlags)

	cfg := config.GetConfig()
	cnt := context.Background()
	exp, err := newExporter(cfg.JaegerAddress)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}
	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(cnt) }()
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("reservations-service")
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// NoSQL: Initialize Reservation Repository store
	store, err := repository.New(storeLogger, tracer)
	if err != nil {
		logger.Fatal(err)
	}

	// NoSQL: Initialize Event Repository store
	eventStore, err := repository.NewEventRepo(storeLogger, tracer)
	if err != nil {
		logger.Fatal(err)
	}

	// NoSQL: Initialize Report Repository store
	reportStore, err := repository.NewReportRepo(storeLogger, tracer)
	if err != nil {
		logger.Fatal(err)
	}
	//serviceAv, err := services.New(storeLogger)
	//if err != nil {
	//	logger.Fatal(err)
	//}

	defer store.CloseSession()
	store.CreateTable()
	eventStore.CreateTableEventStore()
	reportStore.CreateTableDailyReport()
	reportStore.CreateTableMonthlyReport()
	reservationsHandler := handlers.NewReservationsHandler(logger, availabilityService, store, eventStore, availabilityCollection, tracer)
	eventHandler := handlers.NewEventHandler(logger, eventStore, tracer)
	reportHandler := handlers.NewReportHandler(logger, reportStore, eventStore, tracer)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(handlers.ExtractTraceInfoMiddleware)
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
	createAvailability.HandleFunc("/api/availability/create/{id}", AvailabilityHandler.CreateMultipleAvailability)

	getAvailabilityByAccId := router.Methods(http.MethodGet).Subrouter()
	getAvailabilityByAccId.HandleFunc("/api/availability/get/{id}", AvailabilityHandler.GetAvailabilityByAccommodationId)

	checkAvailability := router.Methods(http.MethodPost).Subrouter()
	checkAvailability.HandleFunc("/api/reservations/availability/{accId}", reservationsHandler.CheckAvailability)

	insertDailyReport := router.Methods(http.MethodPost).Subrouter()
	insertDailyReport.HandleFunc("/api/report/daily/{accId}", reportHandler.GenerateDailyReportForAccommodation)

	insertMonthlyReport := router.Methods(http.MethodPost).Subrouter()
	insertMonthlyReport.HandleFunc("/api/report/monthly/{accId}", reportHandler.GenerateMonthlyReportForAccommodation)

	insertEvent := router.Methods(http.MethodPost).Subrouter()
	insertEvent.HandleFunc("/api/event/store", eventHandler.InsertEventIntoEventStore)
	insertEvent.Use(eventHandler.MiddlewareReservationForEventDeserialization)

	getPrices := router.Methods(http.MethodPost).Subrouter()
	getPrices.HandleFunc("/api/reservations/prices/{accId}", AvailabilityHandler.GetPrices)

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

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")

}

func newExporter(address string) (*jaeger.Exporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(address)))
	if err != nil {
		return nil, err
	}
	return exp, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("reservations-service"),
		),
	)

	if err != nil {
		log.Printf("Error merging resources: %v", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}
func NewTracerProvider(serviceName, collectorEndpoint string) (*sdktrace.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(collectorEndpoint)))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize exporter due: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.DeploymentEnvironmentKey.String("development"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}
