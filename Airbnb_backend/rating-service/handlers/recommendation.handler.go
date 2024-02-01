package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/otel/codes"
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
	var user domain.NeoUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r.CreateUserNext(c.Writer, c.Request, &user)

}
func (r *RecommendationHandler) CreateUserNext(rw http.ResponseWriter, h *http.Request, user *domain.NeoUser) {

	if user == nil {
		log.Println("User not found in the context or not of type *domain.NeoUser2")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	err := r.rec.CreateUser(user)
	if err != nil {
		r.logger.Print("Database exception: ", err)
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
	r.CreateReservationNext(c.Writer, c.Request, &reservation)

}
func (r *RecommendationHandler) CreateReservationNext(rw http.ResponseWriter, h *http.Request, reservation *domain.ReservationByGuest) {

	if reservation == nil {
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

	r.CreateAccommodationNext(c.Writer, c.Request, &accommodation)

}
func (r *RecommendationHandler) CreateAccommodationNext(rw http.ResponseWriter, h *http.Request, accommodation *domain.AccommodationRec) {

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
func (r *RecommendationHandler) CreateRecomRate(c *gin.Context) {

	var rate domain.RateAccommodationRec
	if err := c.ShouldBindJSON(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r.CreateRecomRateNext(c.Writer, c.Request, &rate)

}
func (r *RecommendationHandler) CreateRecomRateNext(rw http.ResponseWriter, h *http.Request, rate *domain.RateAccommodationRec) {

	if rate == nil {
		log.Println("Rate not found in the context or not of type *domain.RateAccommodationRec")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	err := r.rec.CreateRate(rate)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)

}
func (r *RecommendationHandler) GetRecommendation(ctx *gin.Context) {
	_, span := r.Tracer.Start(ctx.Request.Context(), "RecommendationHandler.GetRecommendation")
	defer span.End()
	id := ctx.Param("id")
	if id == "" {
		span.SetStatus(codes.Error, "Id is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Id is required"})
		return
	}

	acc, result := r.rec.GetRecommendation(id)
	if result != nil {
		span.SetStatus(codes.Error, "Accommodation not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Accommodation not found"})
		return
	}
	span.SetStatus(codes.Ok, "Found accommodation by id successfully")
	ctx.JSON(http.StatusOK, acc)
}
func (r *RecommendationHandler) DeleteReservation(c *gin.Context) {
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	guestId, exists := requestData["guestId"].(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "guestId is missing or not a string"})
		return
	}
	accommodationId, existss := requestData["accommodationId"].(string)
	if !existss {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accommodationId is missing or not a string"})
		return
	}
	err := r.rec.DeleteReservation(accommodationId, guestId)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting reservation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation deleted successfully"})
}

func (r *RecommendationHandler) DeleteRate(c *gin.Context) {
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	guestId, exists := requestData["guestId"].(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "guestId is missing or not a string"})
		return
	}
	accommodation, existss := requestData["accommodation"].(string)
	if !existss {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accommodation is missing or not a string"})
		return
	}
	err := r.rec.DeleteRate(accommodation, guestId)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting rate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rate deleted successfully"})
}

func (r *RecommendationHandler) DeleteUser(c *gin.Context) {
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, exists := requestData["userId"].(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is missing or not a string"})
		return
	}
	err := r.rec.DeleteUser(userId)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
func (r *RecommendationHandler) DeleteAccommodation(c *gin.Context) {
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accommodationId, exists := requestData["accommodationId"].(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accommodationId is missing or not a string"})
		return
	}
	err := r.rec.DeleteAccommodation(accommodationId)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting accommodation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Accommodation deleted successfully"})
}
