package main

import (
	"auth-service/config"
	"auth-service/handlers"
	"auth-service/routes"
	"auth-service/services"
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var (
	server      *gin.Engine
	ctx         context.Context
	mongoclient *mongo.Client

	userService      services.UserService
	UserHandler      handlers.UserHandler
	UserRouteHandler routes.UserRouteHandler

	authCollection   *mongo.Collection
	authService      services.AuthService
	AuthHandler      handlers.AuthHandler
	AuthRouteHandler routes.AuthRouteHandler
)

func init() {
	ctx = context.TODO()

	//logging
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	lumberjackLog := &lumberjack.Logger{
		Filename:  "/auth-service/logs/logfile.log",
		MaxSize:   1,
		LocalTime: true,
	}
	logger.SetOutput(lumberjackLog)
	defer func() {
		if err := lumberjackLog.Close(); err != nil {
			logger.WithFields(logrus.Fields{"path": "auth/main"}).Error("Error closing log file:", err)
		}
	}()
	logger.WithFields(logrus.Fields{"path": "auth/main"}).Info("This is an info message, finaly")
	logger.WithFields(logrus.Fields{"path": "auth/main"}).Error("This is an error message")
	//logging

	mongoconn := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI"))
	mongoclient, err := mongo.Connect(ctx, mongoconn)

	if err != nil {
		panic(err)
	}

	if err := mongoclient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	logger.Info("MongoDB successfully connected...")

	cfg := config.LoadConfig()
	tracerProvider, err := NewTracerProvider(cfg.ServiceName, cfg.JaegerAddress)
	if err != nil {
		logger.Fatalf("JaegerTraceProvider failed to Initialize. Error :%s", err)
	}
	tracer := tracerProvider.Tracer(cfg.ServiceName)

	// Collections
	authCollection = mongoclient.Database("Gobnb").Collection("auth")
	userService = services.NewUserServiceImpl(authCollection, ctx, tracer)
	authService = services.NewAuthService(authCollection, ctx, userService, tracer)
	AuthHandler = handlers.NewAuthHandler(authService, userService, authCollection, tracer, logger)
	AuthRouteHandler = routes.NewAuthRouteHandler(AuthHandler, authService)
	UserHandler = handlers.NewUserHandler(userService, tracer, logger)
	UserRouteHandler = routes.NewRouteUserHandler(UserHandler)

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

	AuthRouteHandler.AuthRoute(router)
	UserRouteHandler.UserRoute(router)

	err := server.RunTLS(":8080", "/app/auth-service.crt", "/app/auth-service.key")
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
