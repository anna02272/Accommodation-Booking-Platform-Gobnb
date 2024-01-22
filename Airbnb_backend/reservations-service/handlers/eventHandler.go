package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gocql/gocql"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"reservations-service/data"
	error2 "reservations-service/error"
	"reservations-service/repository"
	"strings"
	"time"
)

var validateFieldsEvent = validator.New()

type KeyProductEvent struct{}

type EventHandler struct {
	logger         *log.Logger
	Repo           *repository.EventRepo
	Tracer         trace.Tracer
	CircuitBreaker *gobreaker.CircuitBreaker
}

func NewEventHandler(l *log.Logger, r *repository.EventRepo, tracer trace.Tracer) EventHandler {
	circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "HTTPSRequest",
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit Breaker state changed from %s to %s\n", from, to)
		},
	})
	return EventHandler{
		logger:         l,
		Repo:           r,
		Tracer:         tracer,
		CircuitBreaker: circuitBreaker,
	}
}

func (s *EventHandler) InsertEventIntoEventStore(rw http.ResponseWriter, h *http.Request) {
	ctx, span := s.Tracer.Start(h.Context(), "EventHandler.InsertEventIntoEventStore")
	defer span.End()

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctxx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithCircuitBreakerEvent(ctx, token, url)
	if err != nil {
		if ctxx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Authorization service not available")
			errorMsg := map[string]string{"error": "Authorization service not available.."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusInternalServerError)
			return
		}
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			span.SetStatus(codes.Error, "Circuit is open. Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}

	}

	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		span.SetStatus(codes.Error, "Unauthorized")
		errorMsg := map[string]string{"error": "Unauthorized."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	// Read the response body
	// Create a JSON decoder for the response body
	decoder := json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var response struct {
		LoggedInUser struct {
			ID       string        `json:"id"`
			UserRole data.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			span.SetStatus(codes.Error, "Invalid date format in the response")
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}
		span.SetStatus(codes.Error, "Error decoding JSON response")
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	eventInsert := h.Context().Value(KeyProductEvent{}).(*data.EventJson)

	eventID := gocql.TimeUUID()

	event := &data.AccommodationEvent{
		Event:              eventInsert.Event,
		GuestID:            response.LoggedInUser.ID,
		EventIdTimeCreated: data.TimeUUID(eventID),
		AccommodationID:    eventInsert.AccommodationID,
	}

	fmt.Println(event.AccommodationID)
	errEvent := s.Repo.InsertEvent(ctx, event)
	if errEvent != nil {
		span.SetStatus(codes.Error, "Error storing event")
		s.logger.Print("Database exception: ", errEvent)
		errorMsg := map[string]string{"error": "Error storing event"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusCreated)

}

func (s *EventHandler) HTTPSperformAuthorizationRequestWithContextEvent(ctx context.Context, token string, url string) (*http.Response, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	// Perform the request with the provided context
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *EventHandler) HTTPSperformAuthorizationRequestWithCircuitBreakerEvent(ctx context.Context, token string, url string) (*http.Response, error) {
	requestFunc := func() (interface{}, error) {
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

		return resp, nil
	}

	result, err := s.CircuitBreaker.Execute(requestFunc)
	fmt.Println("here circuit breaker")
	if err != nil {
		return nil, err
	}

	fmt.Println(result)
	fmt.Println("circuit breaker result")
	resp, ok := result.(*http.Response)
	if !ok {
		return nil, errors.New("unexpected response type from Circuit Breaker")
	}

	return resp, nil
}

func (s *EventHandler) MiddlewareReservationForEventDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		patient := &data.EventJson{}
		err := patient.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			s.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProductEvent{}, patient)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}
