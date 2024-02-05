package main

import (
	"context"
	"fmt"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/nats"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
	//"log"
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
	// Create a new logger instance
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	lumberjackLog := &lumberjack.Logger{
		Filename:  "/rating-service/logs/logfile.log",
		MaxSize:   1,
		LocalTime: true,
	}
	logger.SetOutput(lumberjackLog)
	defer func() {
		if err := lumberjackLog.Close(); err != nil {
			logger.WithFields(logrus.Fields{"path": "rating/main"}).Error("Error closing log file:", err)
		}
	}()

	logger.WithFields(logrus.Fields{"path": "rating/main"}).Info("This is an info message, finaly")
	logger.WithFields(logrus.Fields{"path": "rating/main"}).Error("This is an error message")

	mongoconn := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI"))
	mongoclient, err := mongo.Connect(ctx, mongoconn)

	if err != nil {
		panic(err)
	}

	if err := mongoclient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	//for i := 0; i < 1500; i++ {
	logger.Infof("MongoDB successfully connected...")
	//}
	logger.Infof("MongoDB successfully connected...")

	cfg := config.GetConfig()
	tracerProvider, err := NewTracerProvider(cfg.ServiceName, cfg.JaegerAddress)
	if err != nil {
		logger.Fatal("JaegerTraceProvider failed to Initialize", err)
	}
	tracer := tracerProvider.Tracer(cfg.ServiceName)
	// Collections
	hostRatingCollection = mongoclient.Database("Gobnb").Collection("host-rating")
	accommodationRatingCollection = mongoclient.Database("Gobnb").Collection("accommodation-rating")

	neo4jConnStr := "bolt://neo4j:7687"
	neo4jDriver, err := neo4j.NewDriverWithContext(
		neo4jConnStr,
		neo4j.BasicAuth("neo4j", "password", ""),
	)
	if err != nil {
		logger.Fatal(err)
	}
	//for i := 0; i < 500; i++ {
	logger.Infof("Neo4j successfully connected...")

	//}

	commandSubscriber := InitSubscriber(cfg.CreateAccommodationCommandSubject, QueueGroup)
	replyPublisher := InitPublisher(cfg.CreateAccommodationReplySubject)

	hostRatingService = services.NewHostRatingServiceImpl(hostRatingCollection, ctx, tracer)
	HostRatingHandler = handlers.NewHostRatingHandler(hostRatingService, hostRatingCollection, tracer, logger)
	accommodationRatingService = services.NewAccommodationRatingServiceImpl(accommodationRatingCollection, ctx, tracer)
	AccommodationRatingHandler = handlers.NewAccommodationRatingHandler(accommodationRatingService, recommendationService, accommodationRatingCollection, tracer, logger)
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

//type Logger struct {
//	*logg.Logger
//}
//
//func (l Logger) Infof(message string) {
//	l.Infof(message)
//}
//func initLogger() *Logger {
//	baseLogger := logg.New()
//	baseLogger.SetLevel(logg.DebugLevel)
//	baseLogger.SetFormatter(&logg.JSONFormatter{})
//	logFile := &lumberjack.Logger{
//		Filename:   "./logs/log.log",
//		MaxSize:    1,
//		MaxBackups: 3,
//		MaxAge:     28,
//		Compress:   true,
//	}
//	multiWriter := io.MultiWriter(os.Stdout, logFile)
//	baseLogger.SetOutput(multiWriter)
//	return &Logger{
//		Logger: baseLogger,
//	}
//}
//func (l Logger) Info(message string, fields map[string]interface{}) {
//	l.WithFields(fields).Info(message)
//}

//func RotateLogs(file *os.File) {
//	currentTime := time.Now().Format("2006-01-02_15-04-05")
//	err := os.Rename("/app/rating-service/logs.log", "/app/rating-service/logs"+currentTime+".log")
//	if err != nil {
//		Logger.Error(err)
//	}
//	file.Close()
//
//	file, err = os.OpenFile("/app/rating-service/logs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
//	if err != nil {
//		Logger.Error(err)
//	}
//
//	Logger.SetOutput(file)
//}

//func initLogger() {
//	logFilePath := "logs.log"
//
//	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
//	if err != nil {
//		logrus.Fatal(err)
//	}
//
//	Logger.SetOutput(file)
//
//	rotationInterval := 24 * time.Hour
//	ticker := time.NewTicker(rotationInterval)
//	defer ticker.Stop()
//
//	go func() {
//		for range ticker.C {
//			rotateLogs(file)
//		}
//	}()
//}
//func rotateLogs(file *os.File) {
//	currentTime := time.Now().Format("2006-01-02_15-04-05")
//	err := os.Rename("logs.log", "logs"+currentTime+".log")
//	if err != nil {
//		Logger.Error(err)
//	}
//	file.Close()
//
//	file, err = os.OpenFile("logs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
//	if err != nil {
//		Logger.Error(err)
//	}
//
//	Logger.SetOutput(file)
//}
