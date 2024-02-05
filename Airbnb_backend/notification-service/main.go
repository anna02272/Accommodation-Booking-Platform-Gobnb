package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"net/http"
	"notification-service/config"
	"notification-service/handlers"
	"notification-service/routes"
	"notification-service/services"
	"os"
)

var (
	server                   *gin.Engine
	ctx                      context.Context
	mongoclient              *mongo.Client
	notificationCollection   *mongo.Collection
	notificationService      services.NotificationService
	NotificationHandler      handlers.NotificationHandler
	NotificationRouteHandler routes.NotificationRouteHandler
)

func init() {
	ctx = context.TODO()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	lumberjackLog := &lumberjack.Logger{
		Filename:  "/notification-service/logs/logfile.log",
		MaxSize:   1,
		LocalTime: true,
	}
	logger.SetOutput(lumberjackLog)
	defer func() {
		if err := lumberjackLog.Close(); err != nil {
			logger.WithFields(logrus.Fields{"path": "notification/main"}).Error("Error closing log file:", err)
		}
	}()

	logger.WithFields(logrus.Fields{"path": "notification/main"}).Info("This is an info message, finaly")
	logger.WithFields(logrus.Fields{"path": "notification/main"}).Error("This is an error message")
	mongoconn := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI"))
	mongoclient, err := mongo.Connect(ctx, mongoconn)

	if err != nil {
		panic(err)
	}

	if err := mongoclient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("MongoDB successfully connected...")
	cfg := config.LoadConfig()
	tracerProvider, err := NewTracerProvider(cfg.ServiceName, cfg.JaegerAddress)
	if err != nil {
		log.Fatal("JaegerTraceProvider failed to Initialize", err)
	}
	tracer := tracerProvider.Tracer(cfg.ServiceName)

	//circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
	//	Name: "HTTPSRequest",
	//	OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
	//		// Optionally, you can log state changes.
	//		fmt.Printf("Circuit Breaker state changed from %s to %s\n", from, to)
	//	},
	//})

	notificationCollection = mongoclient.Database("Gobnb").Collection("notification")
	notificationService = services.NewNotificationServiceImpl(notificationCollection, ctx, tracer)
	NotificationHandler = handlers.NewNotificationHandler(notificationService, notificationCollection, tracer, logger)
	NotificationRouteHandler = routes.NewNotificationRouteHandler(NotificationHandler, notificationService)

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

	NotificationRouteHandler.NotificationRoute(router)

	err := server.RunTLS(":8089", "/app/notifications-service.crt", "/app/notifications_decrypted_key.pem")
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
