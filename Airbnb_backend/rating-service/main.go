package main

import (
	"context"
	"fmt"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/nats"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
	"rating-service/config"
	"rating-service/handlers"
	"rating-service/routes"
	"rating-service/services"
)

var (
	server      *gin.Engine
	ctx         context.Context
	mongoclient *mongo.Client
	neo4jDriver neo4j.DriverWithContext
	driver      neo4j.Driver

	hostRatingCollection          *mongo.Collection
	hostRatingService             services.HostRatingService
	HostRatingHandler             handlers.HostRatingHandler
	accommodationRatingCollection *mongo.Collection
	accommodationRatingService    services.AccommodationRatingService
	AccommodationRatingHandler    handlers.AccommodationRatingHandler
	RecommendationHandler         handlers.RecommendationHandler
	recommendationService         services.RecommendationService
	RatingRouteHandler            routes.RatingRouteHandler
)

const (
	QueueGroup = "recommendation_service"
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
	//neo4jConnStr := "bolt://localhost:7687" // Promenite sa odgovarajućim podacima za vašu Neo4j bazu
	//neo4jDriver, err := neo4j.NewDriver(neo4jConnStr, neo4j.BasicAuth("neo4j", "password", ""))
	//if err != nil {
	//	panic(err)
	//}
	//defer neo4jDriver.Close()

	//fmt.Println("Neo4j successfully connected...")

	cfg := config.GetConfig()
	tracerProvider, err := NewTracerProvider(cfg.ServiceName, cfg.JaegerAddress)
	if err != nil {
		log.Fatal("JaegerTraceProvider failed to Initialize", err)
	}
	tracer := tracerProvider.Tracer(cfg.ServiceName)
	// Collections
	hostRatingCollection = mongoclient.Database("Gobnb").Collection("host-rating")
	accommodationRatingCollection = mongoclient.Database("Gobnb").Collection("accommodation-rating")

	logger := log.New(os.Stdout, "[rating-service] ", log.LstdFlags)

	neo4jConnStr := "bolt://neo4j:7687"
	neo4jDriver, err := neo4j.NewDriverWithContext(
		neo4jConnStr,
		neo4j.BasicAuth("neo4j", "password", ""),
		//neo4j.WithMaxConnLifetime(30*time.Second),
	)
	if err != nil {
		panic(err)
	}
	//defer neo4jDriver.Close()
	fmt.Println("Neo4j successfully connected...")

	commandSubscriber := InitSubscriber(cfg.CreateAccommodationCommandSubject, QueueGroup)
	replyPublisher := InitPublisher(cfg.CreateAccommodationReplySubject)

	hostRatingService = services.NewHostRatingServiceImpl(hostRatingCollection, ctx, tracer)
	HostRatingHandler = handlers.NewHostRatingHandler(hostRatingService, hostRatingCollection, tracer)
	accommodationRatingService = services.NewAccommodationRatingServiceImpl(accommodationRatingCollection, ctx, tracer)
	AccommodationRatingHandler = handlers.NewAccommodationRatingHandler(accommodationRatingService, recommendationService, accommodationRatingCollection, tracer)
	//AccommodationRatingHandler = handlers.NewAccommodationRatingHandler(accommodationRatingService, accommodationRatingCollection, tracer)
	recommendationService = services.NewRecommendationServiceImpl(neo4jDriver, tracer, logger)
	RecommendationHandler = handlers.NewRecommendationHandler(recommendationService, neo4jDriver, tracer, logger)
	RatingRouteHandler = routes.NewRatingRouteHandler(HostRatingHandler, hostRatingService, AccommodationRatingHandler, accommodationRatingService,
		RecommendationHandler, recommendationService)

	InitCreateAccommodationHandler(recommendationService, replyPublisher, commandSubscriber)

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
		"nats", cfg.NatsPort,
		cfg.NatsUser, cfg.NatsPass, subject)
	if err != nil {
		log.Fatal(err)
	}
	return publisher
}

func InitSubscriber(subject, queueGroup string) saga.Subscriber {
	cfg := config.GetConfig()
	subscriber, err := nats.NewNATSSubscriber(
		"nats", cfg.NatsPort,
		cfg.NatsUser, cfg.NatsPass, subject, queueGroup)
	if err != nil {
		log.Fatal(err)
	}
	return subscriber
}

func InitCreateAccommodationHandler(service services.RecommendationService, publisher saga.Publisher, subscriber saga.Subscriber) {
	_, err := handlers.NewCreateAccommodationCommandHandler(service, publisher, subscriber)
	if err != nil {
		log.Fatal(err)
	}
}
