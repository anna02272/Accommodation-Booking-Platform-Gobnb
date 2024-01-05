package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reservations-service/data"
	error2 "reservations-service/error"
	"reservations-service/services"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// type KeyProduct struct{}
// type KeyProduct struct{}

type AvailabilityHandler struct {
	availabilityService services.AvailabilityService
	DB                  *mongo.Collection
	logger              *log.Logger
	Tracer              trace.Tracer
}

func NewAvailabilityHandler(availabilityService services.AvailabilityService, db *mongo.Collection, lg *log.Logger, tr trace.Tracer) AvailabilityHandler {
	return AvailabilityHandler{availabilityService, db, lg, tr}
}

func (s *AvailabilityHandler) CreateMultipleAvailability(rw http.ResponseWriter, h *http.Request) {
	ctx, span := s.Tracer.Start(h.Context(), "AvailabilityHandler.CreateMultipleAvailability")
	defer span.End()
	vars := mux.Vars(h)
	accIdParam := vars["id"]
	accId, err := primitive.ObjectIDFromHex(accIdParam)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		panic(err)
	}

	log.Printf("Received request for availability creation. Accommodation ID: %s", accId.Hex())

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctxx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctxx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}
		span.SetStatus(codes.Error, "Error performing authorization request")
		error2.ReturnJSONError(rw, "Error performing authorization request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode != 200 {
		span.SetStatus(codes.Error, "Unauthorized.")
		errorMsg := map[string]string{"error": "Unauthorized."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(resp.Body)

	var response struct {
		LoggedInUser struct {
			ID       string        `json:"id"`
			UserRole data.UserRole `json:"userRole"`
		} `json:"user"`
		Message string `json:"message"`
	}
	if err := decoder.Decode(&response); err != nil {
		if strings.Contains(err.Error(), "cannot parse") {
			span.SetStatus(codes.Error, "Invalid date format in the response")
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}
		span.SetStatus(codes.Error, "Error decoding JSON response: %v"+err.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole
	if userRole != data.Host {
		span.SetStatus(codes.Error, "Permission denied. Only hosts can create availabilities.")
		error2.ReturnJSONError(rw, "Permission denied. Only hosts can create availabilities.", http.StatusForbidden)
		return
	}

	var avail data.AvailabilityPeriod
	err5 := json.NewDecoder(h.Body).Decode(&avail)
	if err5 != nil {
		span.SetStatus(codes.Error, err5.Error())
		http.Error(rw, err5.Error(), http.StatusBadRequest)
		return
	}
	insertedAvail, err := s.availabilityService.InsertMulitipleAvailability(avail, accId, ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	jsonResponse, err1 := json.Marshal(insertedAvail)
	if err1 != nil {
		span.SetStatus(codes.Error, "Error marshaling JSON"+err1.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "Availability created")
	rw.Write(jsonResponse)
}

func (s *AvailabilityHandler) GetAvailabilityByAccommodationId(rw http.ResponseWriter, h *http.Request) {
	ctx, span := s.Tracer.Start(h.Context(), "AvailabilityHandler.GetAvailabilityByAccommodationId")
	defer span.End()

	vars := mux.Vars(h)
	accIdParam := vars["id"]
	accId, err := primitive.ObjectIDFromHex(accIdParam)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		panic(err)
	}

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctxx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctxx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}
		span.SetStatus(codes.Error, "Error performing authorization request")
		error2.ReturnJSONError(rw, "Error performing authorization request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 200 {
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
		span.SetStatus(codes.Error, "Error decoding JSON response"+err.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole

	if userRole != data.Host {
		span.SetStatus(codes.Error, "Permission denied. Only hosts can create availabilities.")
		error2.ReturnJSONError(rw, "Permission denied. Only hosts can create availabilities.", http.StatusForbidden)
		return
	}

	availabilities, err := s.availabilityService.GetAvailabilityByAccommodationId(accId, ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	jsonResponse, err1 := json.Marshal(availabilities)
	if err1 != nil {
		span.SetStatus(codes.Error, "Error marshaling JSON"+err1.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}
	span.SetStatus(codes.Ok, "Get availability by accommodation id successful")
	rw.Write(jsonResponse)
}

func (s *AvailabilityHandler) GetPrices(rw http.ResponseWriter, h *http.Request) {
	spanCtx, span := s.Tracer.Start(h.Context(), "AvailabilityHandler.GetPrices")
	defer span.End()
	// rw := c.Writer
	// h := c.Request
	vars := mux.Vars(h)
	accIdParam := vars["accId"]
	accId, err := primitive.ObjectIDFromHex(accIdParam)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		panic(err)
	}

	var checkPriceRequest data.CheckAvailability
	if err := json.NewDecoder(h.Body).Decode(&checkPriceRequest); err != nil {
		span.SetStatus(codes.Error, "Invalid request body. Check the request format.")
		errorMsg := map[string]string{"error": "Invalid request body. Check the request format."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	//accIDConvert, err := primitive.ObjectIDFromHex(accIDString)
	checkPriceRequest.CheckInDate = time.Date(
		checkPriceRequest.CheckInDate.Year(),
		checkPriceRequest.CheckInDate.Month(),
		checkPriceRequest.CheckInDate.Day(),
		0, 0, 0, 0,
		checkPriceRequest.CheckInDate.Location())

	checkPriceRequest.CheckOutDate = time.Date(
		checkPriceRequest.CheckOutDate.Year(),
		checkPriceRequest.CheckOutDate.Month(),
		checkPriceRequest.CheckOutDate.Day(),
		0, 0, 0, 0,
		checkPriceRequest.CheckOutDate.Location())

	prices, err := s.availabilityService.GetPrices(accId, checkPriceRequest.CheckInDate, checkPriceRequest.CheckOutDate, spanCtx)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	jsonResponse, err1 := json.Marshal(prices)
	if err1 != nil {
		span.SetStatus(codes.Error, "Error marshaling JSON"+err1.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}
	span.SetStatus(codes.Ok, "Get availability by accommodation id successful")
	rw.Write(jsonResponse)
}

func (s *AvailabilityHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
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
