package handlers

import (
	"acc-service/application"
	"acc-service/cache"
	"acc-service/domain"
	error2 "acc-service/error"
	hdfs_store "acc-service/hdfs-store"
	"acc-service/services"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type AccommodationHandler struct {
	accommodationService services.AccommodationService
	DB                   *mongo.Collection
	hdfs                 *hdfs_store.FileStorage
	imageCache           *cache.ImageCache
	Tracer               trace.Tracer
	CircuitBreaker       *gobreaker.CircuitBreaker
	orchestrator         *application.CreateAccommodationOrchestrator
	logger               *logrus.Logger
}

func NewAccommodationHandler(accommodationService services.AccommodationService, imageCache *cache.ImageCache,
	hdfs *hdfs_store.FileStorage, db *mongo.Collection, tr trace.Tracer, orchestrator *application.CreateAccommodationOrchestrator, logger *logrus.Logger) AccommodationHandler {
	circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "HTTPSRequest",
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit Breaker state changed from %s to %s\n", from, to)
		},
	})

	return AccommodationHandler{
		accommodationService: accommodationService,
		DB:                   db,
		hdfs:                 hdfs,
		imageCache:           imageCache,
		Tracer:               tr,
		CircuitBreaker:       circuitBreaker,
		orchestrator:         orchestrator,
		logger:               logger,
	}

}

func (s *AccommodationHandler) CreateAccommodations(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationHandler.CreateAccommodations")
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

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole

	if userRole != domain.Host {
		s.logger.Error("Permission denied. Only host can create accommodation.")
		span.SetStatus(codes.Error, "Permission denied. Only hosts can create accommodations.")
		error2.ReturnJSONError(rw, "Permission denied. Only hosts can create accommodations.", http.StatusForbidden)
		return
	}

	id := primitive.NewObjectID()
	accommodation, exists := c.Get("accommodation")
	if !exists {
		s.logger.Error("Accommodation not found in context")
		span.SetStatus(codes.Error, "Accommodation not found in context")
		error2.ReturnJSONError(rw, "Accommodation not found in context", http.StatusBadRequest)
		return
	}
	acc, ok := accommodation.(domain.AccommodationWithAvailability)
	if !ok {
		s.logger.Error("Invalid type for Accommodation")
		span.SetStatus(codes.Error, "Invalid type for Accommodation")
		error2.ReturnJSONError(rw, "Invalid type for Accommodation", http.StatusBadRequest)
		return
	}
	acc.ID = id
	acc.HostId = response.LoggedInUser.ID

	err = s.orchestrator.Start(spanCtx, &acc)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	//	insertedAcc, _, err := s.accommodationService.InsertAccommodation(&acc, response.LoggedInUser.ID, spanCtx)
	//	if err != nil {
	//		span.SetStatus(codes.Error, err.Error())
	//		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
	//		return
	//	}
	rw.WriteHeader(http.StatusCreated)
	jsonResponse, err1 := json.Marshal(acc)
	if err1 != nil {
		span.SetStatus(codes.Error, err1.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}
	s.logger.Info("Successfully created accommodation")
	span.SetStatus(codes.Ok, "Successfully created accommodation")
	rw.Write(jsonResponse)
}

func (s *AccommodationHandler) GetAllAccommodations(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationHandler.GetAllAccommodations")
	defer span.End()

	location := c.Query("location")
	guests := c.Query("guests")
	tv := c.Query("tv")
	wifi := c.Query("wifi")
	ac := c.Query("ac")

	var amenitiesExist bool = false

	amenities := make(map[string]bool)
	if tv != "" || wifi != "" || ac != "" {
		amenitiesExist = true
		amenities["TV"], _ = strconv.ParseBool(tv)
		amenities["WiFi"], _ = strconv.ParseBool(wifi)
		amenities["AC"], _ = strconv.ParseBool(ac)
	}
	if location != "" || guests != "" || amenitiesExist {
		accommodations, err := s.accommodationService.GetAccommodationBySearch(location, guests, amenities, amenitiesExist, spanCtx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		s.logger.Info("Search success")
		span.SetStatus(codes.Ok, "Search success")
		c.JSON(http.StatusOK, accommodations)
		return
	}

	accommodations, err := s.accommodationService.GetAllAccommodations(spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(accommodations)
	s.logger.Info("Get all success")
	span.SetStatus(codes.Ok, "Get all success")
	c.JSON(http.StatusOK, accommodations)
}

func (s *AccommodationHandler) GetAccommodationByID(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationHandler.GetAccommodationByID")
	defer span.End()

	accommodationID := c.Param("id")

	accommodation, err := s.accommodationService.GetAccommodationByID(accommodationID, spanCtx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			s.logger.Error("Accommodation not found")
			span.SetStatus(codes.Error, "Accommodation not found")
			error2.ReturnJSONError(c.Writer, "Accommodation not found", http.StatusNotFound)
		} else {
			span.SetStatus(codes.Error, err.Error())
			error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	s.logger.Info("Get accommodation by id successfully")
	span.SetStatus(codes.Ok, "Got accommodation by id successfully")
	c.JSON(http.StatusOK, accommodation)
}

func (s *AccommodationHandler) GetAccommodationsByHostId(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationHandler.GetAccommodationsByHostId")
	defer span.End()

	hostID := c.Param("hostId")

	accs, err := s.accommodationService.GetAccommodationsByHostId(hostID, spanCtx)
	if err != nil {
		s.logger.Error("Failed to get accommodations")
		span.SetStatus(codes.Error, "Failed to get accommodations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get accommodations"})
		return
	}

	if len(accs) == 0 {
		s.logger.Error("No accommodations found for this host")
		span.SetStatus(codes.Error, "No accommodations found for this host")
		c.JSON(http.StatusOK, gin.H{"message": "No accommodations found for this host", "accommodations": []interface{}{}})
		return
	}
	s.logger.Info("Got accommodation by host id successfully")
	span.SetStatus(codes.Ok, "Got accommodation by host id successfully")
	c.JSON(http.StatusOK, accs)
}

func (s *AccommodationHandler) DeleteAccommodation(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationHandler.DeleteAccommodation")
	defer span.End()

	accId := c.Param("accId")

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
			s.logger.Error("Circuit is open. Authorization service is not available")
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
		s.logger.Info("Error performing authorization request")
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

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole
	userId := response.LoggedInUser.ID

	if userRole != domain.Host {
		s.logger.Error("Permission denied. Only host can delete accommodation")
		span.SetStatus(codes.Error, "Permission denied. Only hosts can delete accommodations.")
		error2.ReturnJSONError(rw, "Permission denied. Only hosts can delete accommodations.", http.StatusUnauthorized)
		return
	}

	accommodation, err := s.accommodationService.GetAccommodationByID(accId, spanCtx)
	if err != nil {
		s.logger.Error("Error fetching accommodation details")
		span.SetStatus(codes.Error, "Error fetching accommodation details.")
		errorMsg := map[string]string{"error": "Error fetching accommodation details."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	if accommodation.HostId != userId {
		s.logger.Error("Permission denied. You are not creator of this accommodation")
		span.SetStatus(codes.Error, "Permission denied. You are not the creator of this accommodation.")
		errorMsg := map[string]string{"error": "Permission denied. You are not the creator of this accommodation."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	url = "https://res-server:8082/api/reservations/get/" + accId

	resp, err = s.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, token, url)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			s.logger.Error("Circuit is open. Authorization service is not available")
			span.SetStatus(codes.Error, "Circuit is open. Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}
		if ctx.Err() == context.DeadlineExceeded {
			s.logger.Error("Reservation service is not available")
			span.SetStatus(codes.Error, "Reservation service is not available.")
			error2.ReturnJSONError(rw, "Reservation service is not available.", http.StatusBadRequest)
			return
		}

		s.logger.Error("Error performing reservation request")
		span.SetStatus(codes.Error, "Error performing reservation request")
		error2.ReturnJSONError(rw, "Error performing reservation request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode
	if statusCode != 200 {
		s.logger.Error("Reservation service error")
		span.SetStatus(codes.Error, "Reservation service error.")
		errorMsg := map[string]string{"error": "Reservation service error."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	// Read the response body
	// Create a JSON decoder for the response body
	decoder = json.NewDecoder(resp.Body)

	// Define a struct to represent the JSON structure
	var ReservationNumber struct {
		Number int `json:"number"`
	}

	// Decode the JSON response into the struct
	if err := decoder.Decode(&ReservationNumber); err != nil {
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

	counter := ReservationNumber.Number

	if counter != 0 {
		s.logger.Error("Cannot delete accommodation thas has reservations in the future or reservation in active")
		span.SetStatus(codes.Error, "Cannot delete accommodation that has reservations in future or reservation is active.")
		errorMsg := map[string]string{"error": "Cannot delete accommodation that has reservations in future or reservation is active."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return

	}

	err = s.accommodationService.DeleteAccommodation(accId, userId, spanCtx)
	if err != nil {
		fmt.Println(err)
		s.logger.Error("Failed to delete accommodation")
		span.SetStatus(codes.Error, "Failed to delete accommodation.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete accommodation."})
		return
	}
	s.logger.Info("Accommodation successfully deleted")
	span.SetStatus(codes.Ok, "Accommodation successfully deleted.")
	c.JSON(http.StatusOK, gin.H{"message": "Accommodation successfully deleted."})
	return
}

// get hostid by accommodation id
func (s *AccommodationHandler) GetHostIdByAccommodationId(c *gin.Context) {
	accID := c.Param("accId")

	hostid, err := s.accommodationService.GetHostIdByAccommodationId(accID)
	if err != nil {
		errorMsg := map[string]string{"error": "HostId not found."}
		error2.ReturnJSONError(c.Writer, errorMsg, http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, gin.H{"hostId": hostid})

}

func (s *AccommodationHandler) CacheAndStoreImages(c *gin.Context) {
	spanCtx, span := s.Tracer.Start(c.Request.Context(), "AccommodationHandler.CacheAndStoreImages")
	defer span.End()

	accommodationID := c.Param("accId")
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
			span.SetStatus(codes.Error, "Circuit is open. Authorization service is not available.")
			error2.ReturnJSONError(rw, "Authorization service is not available.", http.StatusBadRequest)
			return
		}

		if ctx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Authorization service is not available.")
			errorMsg := map[string]string{"error": "Authorization service is not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		span.SetStatus(codes.Error, "Error performing authorization request.")
		errorMsg := map[string]string{"error": "Error performing authorization request."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
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
			s.logger.Error("Invalid date format in the response")
			span.SetStatus(codes.Error, "Invalid date format in the response")
			errorMsg := map[string]string{"error": "Invalid date format in the response"}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		span.SetStatus(codes.Error, "Error decoding JSON response:"+err.Error())
		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole
	userId := response.LoggedInUser.ID

	if userRole != domain.Host {
		span.SetStatus(codes.Error, "Permission denied. Only hosts can change accommodations.")
		errorMsg := map[string]string{"error": "Permission denied. Only hosts can change accommodations."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	accommodation, err := s.accommodationService.GetAccommodationByID(accommodationID, spanCtx)
	if err != nil {
		s.logger.Error("Accommodation not found")
		span.SetStatus(codes.Error, "Accommodation not found.")
		errorMsg := map[string]string{"error": "Accommodation not found."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	if accommodation.HostId != userId {
		span.SetStatus(codes.Error, "Permission denied. You are not the creator of this accommodation.")
		errorMsg := map[string]string{"error": "Permission denied. You are not the creator of this accommodation."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		span.SetStatus(codes.Error, "Error parsing multipart form.")
		errorMsg := map[string]string{"error": "Error parsing multipart form."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	files := c.Request.MultipartForm.File["image"]

	// Loop through each file
	for _, fileHeader := range files {
		// Open the uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			s.logger.Error("Error opening one of the upload files")
			span.SetStatus(codes.Error, "Error opening one of the uploaded files.")
			errorMsg := map[string]string{"error": "Error opening one of the uploaded files."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the file content into a byte slice
		imageData, err := ioutil.ReadAll(file)
		if err != nil {
			span.SetStatus(codes.Error, "Error reading file content.")
			errorMsg := map[string]string{"error": "Error reading file content."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		// Cache the image in Redis
		imageID := cache.GenerateUniqueImageID()
		fmt.Println(imageID)
		fmt.Println("imageID HERE")
		accID := accommodationID
		if err := s.imageCache.PostImage(imageID, accID, imageData, spanCtx); err != nil {
			span.SetStatus(codes.Error, "Error caching image data.")
			errorMsg := map[string]string{"error": "Error caching image data."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		filename := fmt.Sprintf("%s/%s-image-%d", accID, imageID, len(files))
		err = s.hdfs.WriteFileBytes(imageData, filename, accID, spanCtx)
		if err != nil {
			span.SetStatus(codes.Error, "Error storing image in HDFS.")
			errorMsg := map[string]string{"error": "Error storing image in HDFS."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
	}
	s.logger.Info("Images cached in Redis and stored in HDFS")
	span.SetStatus(codes.Ok, "Images cached in Redis and stored in HDFS")
	c.JSON(http.StatusOK, gin.H{
		"message": "Images cached in Redis and stored in HDFS",
	})
}

//	func (ah *AccommodationHandler) CreateAccommodationImages(c *gin.Context) {
//		var images cache.Images
//		var accID string
//		if err := c.BindJSON(&images); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode request body"})
//			return
//		}
//
//		for _, image := range images {
//			fmt.Println("loop")
//			ah.hdfs.WriteFileBytes(image.Data, image.AccID+"-image-"+image.ID)
//			fmt.Println(image)
//			fmt.Println("HERE")
//			accID = image.AccID
//		}
//		ah.imageCache.PostAll(accID, images)
//
//		c.JSON(http.StatusCreated, gin.H{"message": "Images created successfully"})
//	}
func (ah *AccommodationHandler) GetAccommodationImages(c *gin.Context) {
	spanCtx, span := ah.Tracer.Start(c.Request.Context(), "AccommodationHandler.GetAccommodationImages")
	defer span.End()

	accID := c.Param("accId")

	var images []*cache.Image
	var root = "/hdfs/created/"
	dirName := fmt.Sprintf("%s%s", root, accID)

	files, err := ah.hdfs.Client.ReadDir(dirName)
	if err != nil {
		ah.logger.Error("Unable to read dir.")
		span.SetStatus(codes.Error, "Unable to read dir.")
		errorMsg := map[string]string{"error": "Unable to read dir."}
		error2.ReturnJSONError(c.Writer, errorMsg, http.StatusBadRequest)
		return

	}

	for _, filename := range files {
		parts := strings.Split(filename.Name(), "-")

		data, err := ah.hdfs.ReadFileBytes(dirName+"/"+filename.Name(), spanCtx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			break
		}

		if len(parts) < 1 {
			ah.logger.Error("Unable to parse name of the file")
			span.SetStatus(codes.Error, "Unable to parse name of the file.")
			errorMsg := map[string]string{"error": "Unable to parse name of the file."}
			error2.ReturnJSONError(c.Writer, errorMsg, http.StatusBadRequest)
			return
		}
		parts = strings.Split(parts[0], "_")
		// Convert the first part to an integer
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			ah.logger.Error("Unable to parse name of the files")
			span.SetStatus(codes.Error, "Unable to parse name of the file.")
			errorMsg := map[string]string{"error": "Unable to parse name of the file."}
			error2.ReturnJSONError(c.Writer, errorMsg, http.StatusBadRequest)
			return
		}

		image := &cache.Image{
			ID:    strconv.Itoa(id),
			Data:  data,
			AccID: accID,
		}
		images = append(images, image)
	}

	if len(images) > 0 {
		err := ah.imageCache.PostAll(accID, images, spanCtx)
		if err != nil {
			ah.logger.Error("Unable to write to cache")
			span.SetStatus(codes.Error, "Unable to write to cache.")
			errorMsg := map[string]string{"error": "Unable to write to cache."}
			error2.ReturnJSONError(c.Writer, errorMsg, http.StatusBadRequest)
			return
		}
	}
	ah.logger.Info("Got accommodation images successfully")
	span.SetStatus(codes.Ok, "Got accommodation images successfully")
	c.JSON(http.StatusOK, images)
}

func (s *AccommodationHandler) HTTPSperformAuthorizationRequestWithCircuitBreaker(ctx context.Context, token string, url string) (*http.Response, error) {
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
