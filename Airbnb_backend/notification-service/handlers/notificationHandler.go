package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"notification-service/domain"
	error2 "notification-service/error"
	"notification-service/services"
	"strings"
	"time"
)

type NotificationHandler struct {
	notificationService services.NotificationService
	DB                  *mongo.Collection
	Tracer              trace.Tracer
	CircuitBreaker      *gobreaker.CircuitBreaker
	logger              *logrus.Logger
}

func NewNotificationHandler(notificationService services.NotificationService, db *mongo.Collection, tr trace.Tracer, logger *logrus.Logger) NotificationHandler {
	circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "HTTPSRequest",
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Optionally, you can log state changes.
			fmt.Printf("Circuit Breaker state changed from %s to %s\n", from, to)
		},
	})

	return NotificationHandler{
		notificationService: notificationService,
		DB:                  db,
		Tracer:              tr,
		CircuitBreaker:      circuitBreaker,
		logger:              logger,
	}
}

func (s *NotificationHandler) CreateNotification(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "NotificationHandler.CreateNotification")
	defer span.End()

	rw := c.Writer
	h := c.Request

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, token, url)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			s.logger.Error("Circuit is open. Authorization service is not available.")
			span.SetStatus(codes.Error, "Circuit is open. Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}

		if ctx.Err() == context.DeadlineExceeded {
			s.logger.Error("Authorization service is not available")
			span.SetStatus(codes.Error, "Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}
		s.logger.Error("Error performing authotization request")
		span.SetStatus(codes.Error, "Error performing authorization request")
		error2.ReturnJSONError(rw, "Error performing authorization request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		s.logger.Error("Unauthorized")
		span.SetStatus(codes.Error, "Unauthorized.")
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
			ID       string          `json:"id"`
			UserRole domain.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			s.logger.Error("Ivalid date format in the response")
			span.SetStatus(codes.Error, "Invalid date format in the response")
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}
		s.logger.Error("Error decoding JSON response")
		span.SetStatus(codes.Error, "Error decoding JSON response: "+err.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	notification, exists := c.Get("notification")
	if !exists {
		s.logger.Error("Notification not found in context")
		span.SetStatus(codes.Error, "Notification not found in context")
		error2.ReturnJSONError(rw, "Notification not found in context", http.StatusBadRequest)
		return
	}

	notif, ok := notification.(domain.NotificationCreate)
	if !ok {
		fmt.Println(notif)
		s.logger.Error("Invalid type for notification")
		span.SetStatus(codes.Error, "Invalid type for notification.")
		errorMsg := map[string]string{"error": "Invalid type for notification."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	insertedNotif, _, err := s.notificationService.InsertNotification(&notif, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	jsonResponse, err1 := json.Marshal(insertedNotif)
	if err1 != nil {
		span.SetStatus(codes.Error, "Error marshaling JSON: "+err1.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}
	s.logger.Info("Created notification successfully")
	span.SetStatus(codes.Ok, "Created notification successfully")
	rw.Write(jsonResponse)
}

func (s *NotificationHandler) GetNoitifcationsForHost(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "NotificationHandler.GetNotificationsforHost")
	defer span.End()

	rw := c.Writer
	h := c.Request

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, token, url)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			s.logger.Error("Circuit is open. AAuthorization service is not available")
			span.SetStatus(codes.Error, "Circuit is open. Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}

		if ctx.Err() == context.DeadlineExceeded {
			s.logger.Error("Authorization service is not available.")
			span.SetStatus(codes.Error, "Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}
		s.logger.Error("Error performing authorization request")
		span.SetStatus(codes.Error, "Error performing authorization request")
		error2.ReturnJSONError(rw, "Error performing authorization request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
		s.logger.Error("Unauthorized.")
		span.SetStatus(codes.Error, "Unauthorized.")
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
			ID       string          `json:"id"`
			UserRole domain.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			s.logger.Error("Invalid date format in the response")
			span.SetStatus(codes.Error, "Invalid date format in the response")
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}
		s.logger.Error("Error decoding JSON response")
		span.SetStatus(codes.Error, "Error decoding JSON response:"+err.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	hostID := response.LoggedInUser.ID
	userRole := response.LoggedInUser.UserRole

	if userRole != domain.Host {
		s.logger.Error("Permission denied. Only hosts have notifications")
		span.SetStatus(codes.Error, "Permission denied. Only hosts have notifications")
		errorMsg := map[string]string{"error": "Permission denied. Only hosts have notifications"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusForbidden)
		return
	}
	notifs, err := s.notificationService.GetNotificationsByHostId(hostID, spanCtx)
	if err != nil {
		s.logger.Error("Failed to get notifications")
		span.SetStatus(codes.Error, "Failed to get notifications")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	if len(notifs) == 0 {
		s.logger.Error("No notification for this host")
		span.SetStatus(codes.Error, "No notifications found for this host")
		c.JSON(http.StatusOK, gin.H{"message": "No notifications found for this host", "notifications": []interface{}{}})
		return
	}
	s.logger.Info("Got notification for host successfully")
	span.SetStatus(codes.Ok, "Got notification for host successfully")
	c.JSON(http.StatusOK, notifs)

}

func (s *NotificationHandler) HTTPSperformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
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

func (s *NotificationHandler) HTTPSperformAuthorizationRequestWithCircuitBreaker(ctx context.Context, token string, url string) (*http.Response, error) {
	maxRetries := 3

	retryOperation := func() (interface{}, error) {
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
		fmt.Println(resp)
		fmt.Println("resp here")
		return resp, nil // Return the response as the first value
	}

	//retryOpErr := retryOperationWithExponentialBackoff(ctx,3, retryOperation)
	//if (r)
	// Use an anonymous function to convert the result to the expected type
	result, err := s.CircuitBreaker.Execute(func() (interface{}, error) {
		return retryOperationWithExponentialBackoff(ctx, maxRetries, retryOperation)
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("result here")
	fmt.Println(result)
	resp, ok := result.(*http.Response)
	if !ok {
		fmt.Println(ok)
		fmt.Println("OK")
		return nil, errors.New("unexpected response type from Circuit Breaker")
	}
	return resp, nil
}

func (s *NotificationHandler) performAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	// Perform the request with the provided context
	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
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

func ExtractTraceInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
