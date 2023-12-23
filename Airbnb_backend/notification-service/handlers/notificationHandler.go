package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
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
}

func NewNotificationHandler(notificationService services.NotificationService, db *mongo.Collection) NotificationHandler {
	return NotificationHandler{notificationService, db}
}

func (s *NotificationHandler) CreateNotification(c *gin.Context) {
	rw := c.Writer
	h := c.Request

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, "Error performing authorization request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
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
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	notification, exists := c.Get("notification")
	if !exists {
		error2.ReturnJSONError(rw, "Notification not found in context", http.StatusBadRequest)
		return
	}

	notif, ok := notification.(domain.NotificationCreate)
	if !ok {
		fmt.Println(notif)
		errorMsg := map[string]string{"error": "Invalid type for notification."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	insertedNotif, _, err := s.notificationService.InsertNotification(&notif)
	if err != nil {
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	jsonResponse, err1 := json.Marshal(insertedNotif)
	if err1 != nil {
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}
	rw.Write(jsonResponse)
}

func (s *NotificationHandler) GetNoitifcationsForHost(c *gin.Context) {
	rw := c.Writer
	h := c.Request

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, "Error performing authorization request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
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
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	hostID := response.LoggedInUser.ID
	userRole := response.LoggedInUser.UserRole

	if userRole != domain.Host {
		errorMsg := map[string]string{"error": "Permission denied. Only hosts have notifications"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusForbidden)
		return
	}
	notifs, err := s.notificationService.GetNotificationsByHostId(hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	if len(notifs) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No notifications found for this host", "notifications": []interface{}{}})
		return
	}

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

	// Perform the request with the provided context
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *NotificationHandler) performAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	// Perform the request with the provided context
	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
