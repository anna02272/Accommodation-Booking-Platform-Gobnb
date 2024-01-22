package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"rating-service/domain"
	"rating-service/services"
)

type RecommendationHandler struct {
	rec    services.RecommendationService
	driver neo4j.DriverWithContext
	Tracer trace.Tracer
	logger *log.Logger
}
type KeyProduct struct{}

func NewRecommendationHandler(recommendationService services.RecommendationService, driver neo4j.DriverWithContext, tr trace.Tracer, l *log.Logger) RecommendationHandler {
	return RecommendationHandler{recommendationService, driver, tr, l}
}
func (r *RecommendationHandler) CreateUser(c *gin.Context) {
	//	log.Println("ovde")
	var user domain.NeoUser
	if err := c.ShouldBindJSON(&user); err != nil {
		// Ako vežanje nije uspelo, vratite grešku
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//
	log.Println(&user)

	r.CreateUserNext(c.Writer, c.Request, &user)

}
func (r *RecommendationHandler) CreateUserNext(rw http.ResponseWriter, h *http.Request, user *domain.NeoUser) {
	log.Println("next")
	log.Println(user)

	if user == nil {
		// Handle the case when the value is not present or is not of the expected type
		log.Println("User not found in the context or not of type *domain.NeoUser2")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	err := r.rec.CreateUser(user)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		log.Println("usao sam u error")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)

	log.Println("ovde sam")
	//user, ok := h.Context().Value(KeyProduct{}).(*domain.User)
	//if !ok || user == nil {
	//	log.Println("ovde je greska")
	//	rw.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	//logger := log.New(os.Stdout, "rating-service ", log.LstdFlags|log.Lshortfile)
	//
	//// Inicijalizuj Neo4j drajver
	//neo4jUri := os.Getenv("NEO4J_DB")
	//neo4jUser := os.Getenv("NEO4J_USERNAME")
	//neo4jPass := os.Getenv("NEO4J_PASS")
	//neo4jAuth := neo4j.BasicAuth(neo4jUser, neo4jPass, "")
	//
	//neo4jDriver, err := neo4j.NewDriverWithContext(neo4jUri, neo4jAuth)
	//if err != nil {
	//	log.Panic(err)
	//}
	//// Inicijalizuj tracer
	//traceProvider := otel.GetTracerProvider()
	//tracer := traceProvider.Tracer("rating-service")
	//recommendationService := services.NewRecommendationServiceImpl(neo4jDriver, tracer, logger)
	//
	//if recommendationService == nil {
	//	log.Println("recommendationService je nil")
	//	rw.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//
	//// Provera da li CreateUser metoda postoji u recommendationService
	//if _, ok := interface{}(r.recommendationService).(interface {
	//	CreateUser(user *domain.User) error
	//}); !ok {
	//	log.Println("CreateUser metoda nije implementirana u recommendationService")
	//	rw.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//err := r.recommendationService.CreateUser(user)
	//if err != nil {
	//	r.logger.Print("Database exception: ", err)
	//	log.Println("ovde je greska,48")
	//	rw.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//rw.WriteHeader(http.StatusCreated)
}
