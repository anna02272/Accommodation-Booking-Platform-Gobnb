package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"rating-service/domain"
	"rating-service/services"
	"strconv"
	"strings"
	"time"
)

type AccommodationRatingHandler struct {
	accommodationRatingService services.AccommodationRatingService
	recommendationService      services.RecommendationService
	DB                         *mongo.Collection
	Tracer                     trace.Tracer
	CircuitBreaker             *gobreaker.CircuitBreaker
	logger                     *logger.Logger
}

func NewAccommodationRatingHandler(accommodationRatingService services.AccommodationRatingService, recommendationService services.RecommendationService, db *mongo.Collection, tr trace.Tracer, logger *logger.Logger) AccommodationRatingHandler {
	circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "HTTPSRequest",
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit Breaker state changed from %s to %s\n", from, to)
		},
	})
	return AccommodationRatingHandler{
		accommodationRatingService: accommodationRatingService,
		recommendationService:      recommendationService,
		DB:                         db,
		Tracer:                     tr,
		CircuitBreaker:             circuitBreaker,
		logger:                     logger,
	}
}

func (s *AccommodationRatingHandler) RateAccommodation(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.RateAccommodation")
	defer span.End()
	//rw := c.Writer
	//h := c.Request
	accommodationID := c.Param("accommodationId")

	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain current user information. Try again later")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information. Try again later"})
		return
	}
	println(accommodationID)
	accommodation, err := s.getAccommodationByIDFromAccommodationService(accommodationID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain accommodation information.Try again later.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain accommodation information.Try again later."})
		return
	}

	urlCheckReservations := "https://res-server:8082/api/reservations/getAll"

	timeout := 2000 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	respRes, errRes := s.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, token, urlCheckReservations)
	if errRes != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			fmt.Println(errRes)
			span.SetStatus(codes.Error, "Circuit is open. Authorization service is not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get reservations. Try again later."})
			return
		}

		span.SetStatus(codes.Error, "Failed to get reservations. Try again later.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get reservations. Try again later."})
		return
	}

	defer respRes.Body.Close()

	if respRes.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "You cannot rate this accommodation. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot rate this accommodation. You don't have reservations from him"})
		return
	}

	decoder := json.NewDecoder(respRes.Body)
	var reservations []domain.ReservationByGuest
	if err := decoder.Decode(&reservations); err != nil {
		fmt.Println(err)
		span.SetStatus(codes.Error, "Failed to decode reservations")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode reservations"})
		return
	}

	if len(reservations) == 0 {
		span.SetStatus(codes.Error, "You cannot rate this accommodation. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot rate this accommodation. You don't have reservations from him"})
		return
	}

	canRate := false
	for _, reservation := range reservations {
		if reservation.AccommodationId == accommodationID {
			canRate = true
			break
		}
	}

	if !canRate {
		span.SetStatus(codes.Error, "You cannot rate this accommodation. You don't have reservations from him")
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot rate this accommodation. You don't have reservations from him"})
		return
	}

	var requestBody struct {
		Rating int `json:"rating"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		span.SetStatus(codes.Error, "Invalid JSON request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request"})
		return
	}

	currentDateTime := primitive.NewDateTimeFromTime(time.Now())
	id := primitive.NewObjectID()
	println(accommodation)
	newRateAccommodation := &domain.RateAccommodation{
		ID:            id,
		Accommodation: accommodationID,
		Guest:         currentUser,
		DateAndTime:   currentDateTime,
		Rating:        domain.Rating(requestBody.Rating),
	}
	log.Println(newRateAccommodation.Rating)
	err, update := s.accommodationRatingService.SaveRating(newRateAccommodation, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to save rating")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to save rating"})
		return
	}
	log.Println("proba")
	log.Println(requestBody.Rating)
	a := domain.Rating(requestBody.Rating)
	log.Println(string(a))
	log.Println(string(requestBody.Rating))

	new := domain.RateAccommodationRec{
		ID:            id.Hex(),
		Accommodation: accommodationID,
		Guest:         currentUser.ID.Hex(),
		Rating:        requestBody.Rating,
	}
	log.Println("start")
	log.Println(new.ID)
	log.Println(new.Accommodation)
	log.Println(new.Guest)
	log.Println(new.Rating)
	var ctxx context.Context
	err = s.CreateRate(&new, ctxx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}
	//result := s.recommendationService.CreateRate(&new)
	//if result == nil {
	//	println("GRESKA JE U PUTANJI")
	//}
	//log.Println(result)
	urlAccommodationCheck := "https://acc-server:8083/api/accommodations/get/" + accommodationID

	timeout = 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, token, urlAccommodationCheck)
	if errRes != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			span.SetStatus(codes.Error, "Circuit is open. Accommodation service is not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Accommodation service not available."})
			return
		}
		if ctx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Accommodation service is not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Accommodation service not available."})
			return
		}

		span.SetStatus(codes.Error, "Accommodation service is not available.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Accommodation service not available."})
		return
	}
	defer resp.Body.Close()

	statusCodeAccommodation := resp.StatusCode
	fmt.Println(statusCodeAccommodation)

	if statusCodeAccommodation != 200 {
		span.SetStatus(codes.Error, "Accommodation with that id does not exist.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Accommodation with that id does not exist."})
		return
	}

	var responseAccommodation struct {
		AccommodationName      string `json:"accommodation_name"`
		AccommodationLocation  string `json:"accommodation_location"`
		AccommodationHostId    string `json:"host_id"`
		AccommodationMinGuests int    `json:"accommodation_min_guests"`
		AccommodationMaxGuests int    `json:"accommodation_max_guests"`
		AccommodationId        string `json:"_id"`
	}
	decoder = json.NewDecoder(resp.Body)

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseAccommodation); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			span.SetStatus(codes.Error, "Invalid date format.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format."})
			return
		}
		span.SetStatus(codes.Error, "Error decoding JSON response"+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding json response."})
		return
	}

	urlHostCheck := "https://auth-server:8080/api/users/getById/" + responseAccommodation.AccommodationHostId

	resp, err = s.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, token, urlHostCheck)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			span.SetStatus(codes.Error, "Circuit is open. Authorization service is not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization service not available."})
			return
		}

		if ctx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Authorization service is not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization service not available."})
			return
		}

		span.SetStatus(codes.Error, "Authorization service is not available.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization service not available."})
		return
	}
	defer resp.Body.Close()

	statusCodeHostCheck := resp.StatusCode
	if statusCodeHostCheck != 200 {
		span.SetStatus(codes.Error, "Host with that id does not exist.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host with that id does not exist."})
		return
	}

	decoder = json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var responseHost struct {
		Host struct {
			Email    string `json:"email"`
			Username string `json:"username"`
			HostID   string `json:"id"`
		} `json:"user"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&responseHost); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			span.SetStatus(codes.Error, "Invalid date format.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format."})
			return
		}
		span.SetStatus(codes.Error, "Error decoding JSON response"+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding json response."})
		return
	}

	notificationPayload := map[string]interface{}{
		"host_id":           responseAccommodation.AccommodationHostId,
		"host_email":        responseHost.Host.Email,
		"notification_text": "Dear " + responseHost.Host.Username + ", \n your accommodation " + responseAccommodation.AccommodationName + " has been rated. It got " + strconv.Itoa(requestBody.Rating) + " stars by " + currentUser.Username + "!",
	}

	notificationPayloadJSON, err := json.Marshal(notificationPayload)
	if err != nil {
		span.SetStatus(codes.Error, "Error creating notification payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification payload"})
		return
	}

	notificationURL := "https://notifications-server:8089/api/notifications/create"

	timeout = 2000 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err = s.HTTPSperformAuthorizationRequestWithContextAndBodyAccCircuitBreaker(spanCtx, token, notificationURL, "POST", notificationPayloadJSON)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			span.SetStatus(codes.Error, "Circuit is open. Notification service not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Notification service not available. Try again later."})
			return
		}
		span.SetStatus(codes.Error, "Error creating notification request")
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification request"})
			return
		}
		span.SetStatus(codes.Error, "Notification service not available.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification service not available."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		span.SetStatus(codes.Error, "Error creating notification")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification"})
		return
	}

	span.SetStatus(codes.Ok, "Rating successfully saved")

	if update {
		eventURL := "https://res-server:8082/api/event/store"

		timeout = 2000 * time.Second
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()

		eventPayload := map[string]interface{}{
			"event":            "Accommodation rating",
			"accommodation_id": responseAccommodation.AccommodationId,
		}

		eventPayloadJSON, err := json.Marshal(eventPayload)
		if err != nil {
			span.SetStatus(codes.Error, "Error creating notification payload")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification payload"})
			return
		}

		resp, err = s.HTTPSperformAuthorizationRequestWithContextAndBodyAccCircuitBreaker(spanCtx, token, eventURL, "POST", eventPayloadJSON)
		if err != nil {
			if errors.Is(err, gobreaker.ErrOpenState) {
				span.SetStatus(codes.Error, "Circuit is open. Error creating event request.")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event request."})
				return
			}
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event request"})
				return
			}

			span.SetStatus(codes.Error, "Error creating event request.")
			fmt.Println(err)
			fmt.Println("here")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event request."})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			span.SetStatus(codes.Error, "Error creating event")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rating successfully saved", "rating": newRateAccommodation})
}
func (s *AccommodationRatingHandler) CreateRate(rate *domain.RateAccommodationRec, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "AccommodationService.CreateRate")
	defer span.End()

	url := "https://rating-server:8087/api/rating/createRecomRate"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctxx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequest(ctx, rate, url)
	if err != nil {
		if ctxx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Rating service not available..")
			return nil
		}
		span.SetStatus(codes.Error, "Rating service not available..")
		return nil
	}

	defer resp.Body.Close()

	return nil
}
func (s *AccommodationRatingHandler) HTTPSperformAuthorizationRequest(ctx context.Context, rate *domain.RateAccommodationRec, url string) (*http.Response, error) {
	reqBody, err := json.Marshal(rate)
	if err != nil {
		return nil, fmt.Errorf("error marshaling reservation JSON: %v", err)
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	// Perform the request with the provided context
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
func (s *AccommodationRatingHandler) DeleteRatingAccommodation(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.DeleteRatingAccommodation")
	defer span.End()

	accommodationID := c.Param("accommodationId")
	println(accommodationID)
	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to obtain current user information.  Try again later.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information.  Try again later."})
		return
	}
	guestID := currentUser.ID.Hex()

	err = s.accommodationRatingService.DeleteRating(accommodationID, guestID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Rating successfully deleted")

	eventURL := "https://res-server:8082/api/event/store"

	timeout := 2000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	eventPayload := map[string]interface{}{
		"event":            "Accommodation rating delete",
		"accommodation_id": accommodationID,
	}

	eventPayloadJSON, err := json.Marshal(eventPayload)
	if err != nil {
		span.SetStatus(codes.Error, "Error creating notification payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating notification payload"})
		return
	}

	resp, err := s.HTTPSperformAuthorizationRequestWithContextAndBodyAccCircuitBreaker(spanCtx, token, eventURL, "POST", eventPayloadJSON)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			span.SetStatus(codes.Error, "Circuit is open. Reservation service not available.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event request"})
			return
		}
		span.SetStatus(codes.Error, "Error creating event request")
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event request"})
			return
		}
		span.SetStatus(codes.Error, "Reservation service not available.")
		fmt.Println(err)
		fmt.Println("here")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation service while event handling not available."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		span.SetStatus(codes.Error, "Error creating event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating successfully deleted"})
}

func (s *AccommodationRatingHandler) GetByAccommodationAndGuest(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.GetByAccommodationAndGuest")
	defer span.End()
	token := c.GetHeader("Authorization")
	currentUser, err := s.getCurrentUserFromAuthService(token, spanCtx)
	if err != nil {
		span.SetStatus(
			codes.Error, "Failed to obtain current user information")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to obtain current user information"})
		return
	}
	guestID := currentUser.ID.Hex()

	accommodationID := c.Param("accommodationId")

	ratings, err := s.accommodationRatingService.GetByAccommodationAndGuest(accommodationID, guestID, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Got ratings by accommodation and guest successfully")
	c.JSON(http.StatusOK, gin.H{"ratings": ratings})
}

func (s *AccommodationRatingHandler) getAccommodationByIDFromAccommodationService(accommodationID string, c context.Context) (*domain.Accommodation, error) {
	spanCtx, span := s.Tracer.Start(c, "AccommodationRatingHandler.getAccommodationByIDFromAccommodationService")
	defer span.End()
	url := "https://acc-server:8083/api/accommodations/get/" + accommodationID

	timeout := 2000 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := s.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, "", url)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			span.SetStatus(codes.Error, "Circuit is open. Accommodation service is not available.")
			return nil, errors.New("Accommodation service is not available")
		}

		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "Accommodation not found")
		return nil, errors.New("Accommodation not found")
	}

	var accommodationResponse domain.AccommodationResponse
	if err := json.NewDecoder(resp.Body).Decode(&accommodationResponse); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "Got accommodation by id from auth service")
	accommodation := domain.ConvertToDomainAccommodation(accommodationResponse)

	return &accommodation, nil
}

func (s *AccommodationRatingHandler) getCurrentUserFromAuthService(token string, c context.Context) (*domain.User, error) {
	spanCtx, span := s.Tracer.Start(c, "AccommodationRatingHandler.getCurrentUserFromAuthService")
	defer span.End()
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, token, url)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			span.SetStatus(codes.Error, "Circuit is open. Auth service is not available.")
			return nil, errors.New("Auth service is not available")
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "Unauthorized")
		return nil, errors.New("Unauthorized")
	}

	var userResponse domain.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "Got current user from auth service")
	user := domain.ConvertToDomainUser(userResponse)

	return &user, nil
}

func (s *AccommodationRatingHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	_, span := s.Tracer.Start(ctx, "AccommodationRatingHandler.HTTPSPerformAuthorizationRequestWithContext")
	defer span.End()
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *AccommodationRatingHandler) HTTPSperformAuthorizationRequestWithCircuitBreaker(ctx context.Context, token string, url string) (*http.Response, error) {
	maxRetries := 3
	type retryOperationFunc func() (interface{}, error)

	retryOperation := retryOperationFunc(func() (interface{}, error) {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", token)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

		client := &http.Client{Transport: tr}
		resp, err := client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		return resp, nil // Return the response as the first value
	})

	// Use an anonymous function to convert the result to the expected type
	result, err := s.CircuitBreaker.Execute(func() (interface{}, error) {
		return retryOperationWithExponentialBackoff(ctx, maxRetries, retryOperation)
	})
	if err != nil {
		// Handle or return the error
		return nil, err
	}

	resp, ok := result.(*http.Response)
	if !ok {
		return nil, errors.New("unexpected response type from Circuit Breaker")
	}

	return resp, nil
}

func (s *AccommodationRatingHandler) HTTPSperformAuthorizationRequestWithContextAndBodyAcc(
	ctx context.Context, token string, url string, method string, requestBody []byte,
) (*http.Response, error) {
	_, span := s.Tracer.Start(ctx, "AccommodationRatingHandler.HTTPSperformAuthorizationRequestWithContextAndBody")
	defer span.End()
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", token)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *AccommodationRatingHandler) HTTPSperformAuthorizationRequestWithContextAndBodyAccCircuitBreaker(
	ctx context.Context, token string, url string, method string, requestBody []byte,
) (*http.Response, error) {
	_, span := s.Tracer.Start(ctx, "AccommodationRatingHandler.HTTPSperformAuthorizationRequestWithContextAndBodyAccCircuitBreaker")
	maxRetries := 3

	// Define a retry operation function
	retryOperation := func() (interface{}, error) {
		// Use the Circuit Breaker to execute the request function
		result, err := s.CircuitBreaker.Execute(func() (interface{}, error) {
			tr := http.DefaultTransport.(*http.Transport).Clone()
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", token)
			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
			client := &http.Client{Transport: tr}
			resp, err := client.Do(req.WithContext(ctx))
			if err != nil {
				return nil, err
			}

			return resp, nil
		})

		// If there is an error, propagate it
		if err != nil {
			return nil, err
		}

		// Check the type of the result
		resp, ok := result.(*http.Response)
		if !ok {
			return nil, errors.New("unexpected response type from Circuit Breaker")
		}

		return resp, nil
	}

	// Use the retry mechanism
	result, err := s.CircuitBreaker.Execute(func() (interface{}, error) {
		return retryOperationWithExponentialBackoff(ctx, maxRetries, retryOperation)
	})
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*http.Response)
	if !ok {
		err := errors.New("unexpected response type from retry operation")
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return resp, nil

}

func (s *AccommodationRatingHandler) GetAllRatingsAccommodation(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationRatingHandler.GetAllRatingsAccommodation")
	defer span.End()

	ratings, averageRating, err := s.accommodationRatingService.GetAllRatingsAccommodation(spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"ratings":       ratings,
		"averageRating": averageRating,
	}
	span.SetStatus(codes.Ok, "Got all ratings successfully")
	c.JSON(http.StatusOK, response)
}

func retryOperationWithExponentialBackoff(ctx context.Context, maxRetries int, operation func() (interface{}, error)) (interface{}, error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Println("attempt loop: ")
		fmt.Println(attempt)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := operation()
		fmt.Println(result)
		if err == nil {
			fmt.Println("out of loop here")
			return result, nil
		}
		fmt.Printf("Attempt %d failed: %s\n", attempt, err.Error())
		backoff := time.Duration(attempt*attempt) * time.Second
		time.Sleep(backoff)
	}
	return nil, fmt.Errorf("max retries exceeded")
}
