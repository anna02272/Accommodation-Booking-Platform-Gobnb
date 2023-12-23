package main

import (
	"context"
	"fmt"
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
	"profile-service/config"
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
	cfg := config.GetConfig()
	tracerProvider, err := NewTracerProvider(cfg.ServiceName, cfg.JaegerAddress)
	if err != nil {
		log.Fatal("JaegerTraceProvider failed to Initialize", err)
	}
	tracer := tracerProvider.Tracer(cfg.ServiceName)
	// Collections
	profileCollection = mongoclient.Database("Gobnb").Collection("profile")
	profileService = services.NewUserServiceImpl(profileCollection, tracer)
	ProfileHandler = handlers.NewProfileHandler(profileService, tracer)
	ProfileRouteHandler = routes.NewRouteProfileHandler(ProfileHandler)

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

	router.GET("/profile/delete/:email", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "OK", "message": "message"})
	})

	router.GET("/profile/createUser", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "hej", "hej": "hej"})
	})

	ProfileRouteHandler.ProfileRoute(router)

	err = server.RunTLS(":8084", "/app/profile-service.crt", "/app/profile-service.key")
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
