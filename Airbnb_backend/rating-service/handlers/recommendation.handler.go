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

}
func (r *RecommendationHandler) CreateReservation(c *gin.Context) {
	var reservation domain.ReservationByGuest
	if err := c.ShouldBindJSON(&reservation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println(&reservation)

	r.CreateReservationNext(c.Writer, c.Request, &reservation)

}
func (r *RecommendationHandler) CreateReservationNext(rw http.ResponseWriter, h *http.Request, reservation *domain.ReservationByGuest) {
	log.Println("next")
	log.Println(reservation)

	if reservation == nil {
		// Handle the case when the value is not present or is not of the expected type
		log.Println("Resevation not found in the context or not of type *domain.Reservation")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	err := r.rec.CreateReservation(reservation)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)

}
func (r *RecommendationHandler) CreateAccommodation(c *gin.Context) {

	var accommodation domain.AccommodationRec
	if err := c.ShouldBindJSON(&accommodation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("OVDE SAAAAAAAAAAAMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM")
	log.Println(&accommodation)

	r.CreateAccommodationNext(c.Writer, c.Request, &accommodation)

}
func (r *RecommendationHandler) CreateAccommodationNext(rw http.ResponseWriter, h *http.Request, accommodation *domain.AccommodationRec) {
	log.Println("next")
	log.Println(accommodation)

	if accommodation == nil {
		log.Println("Accommodation not found in the context or not of type *domain.AccommodationRec")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	err := r.rec.CreateAccommodation(accommodation)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)

}
