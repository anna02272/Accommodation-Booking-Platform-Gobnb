package main

import (
	"acc-service/application"
	"acc-service/cache"
	"acc-service/config"
	"acc-service/handlers"
	hdfs_store "acc-service/hdfs-store"
	"acc-service/routes"
	"acc-service/services"
	"context"
	"fmt"
	"github.com/colinmarc/hdfs/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	saga "github.com/tamararankovic/microservices_demo/common/saga/messaging"
	"github.com/tamararankovic/microservices_demo/common/saga/messaging/nats"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
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
	imageCache                *cache.ImageCache
)

const (
	QueueGroup = "accommodation_service"
)

func init() {
	ctx = context.TODO()

	mongoconn := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI"))
	mongoclient, err := mongo.Connect(ctx, mongoconn)

	cfg := config.GetConfig()
	tracerProvider, err := NewTracerProvider(cfg.ServiceName, cfg.JaegerAddress)
	if err != nil {
		log.Fatal("JaegerTraceProvider failed to Initialize", err)
	}
	tracer := tracerProvider.Tracer(cfg.ServiceName)

	hdfsLogger := log.New(os.Stdout, "HDFS: ", log.LstdFlags)
	fileStorage, err := hdfs_store.New(hdfsLogger, tracer)
	if err != nil {
		panic(err)
	}

	redisLogger := log.New(os.Stdout, "REDIS CACHE: ", log.LstdFlags)
	imageCache = cache.New(redisLogger, tracer)
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
	accommodationService = services.NewAccommodationServiceImpl(accommodationCollection, ctx, tracer)
	AccommodationHandler = handlers.NewAccommodationHandler(accommodationService, imageCache, fileStorage, accommodationCollection, tracer)
	AccommodationRouteHandler = routes.NewAccommodationRouteHandler(AccommodationHandler, accommodationService)

	//commandPublisher := InitPublisher(cfg.CreateAccommodationCommandSubject)
	//replySubscriber := InitSubscriber(cfg.CreateAccommodationReplySubject, QueueGroup)
	//createOrderOrchestrator := InitCreateAccommodationOrchestrator(commandPublisher, replySubscriber)
	//commandSubscriber := InitSubscriber(cfg.CreateAccommodationCommandSubject, QueueGroup)
	//replyPublisher := InitPublisher(cfg.CreateAccommodationReplySubject)
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
func InitPublisher(subject string) saga.Publisher {
	cfg := config.GetConfig()
	publisher, err := nats.NewNATSPublisher(
		cfg.NatsHost, cfg.NatsPort,
		cfg.NatsUser, cfg.NatsPass, subject)
	if err != nil {
		log.Fatal(err)
	}
	return publisher
}

func InitSubscriber(subject, queueGroup string) saga.Subscriber {
	cfg := config.GetConfig()
	subscriber, err := nats.NewNATSSubscriber(
		cfg.NatsHost, cfg.NatsPort,
		cfg.NatsUser, cfg.NatsPass, subject, queueGroup)
	if err != nil {
		log.Fatal(err)
	}
	return subscriber
}

func InitCreateAccommodationOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber) *application.CreateAccommodationOrchestrator {
	orchestrator, err := application.NewCreateAccommodationOrchestrator(publisher, subscriber)
	if err != nil {
		log.Fatal(err)
	}
	return orchestrator
}
