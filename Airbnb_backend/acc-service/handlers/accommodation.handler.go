package handlers

import (
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
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AccommodationHandler struct {
	accommodationService services.AccommodationService
	DB                   *mongo.Collection
	hdfs                 *hdfs_store.FileStorage
	imageCache           *cache.ImageCache
}

func NewAccommodationHandler(accommodationService services.AccommodationService, imageCache *cache.ImageCache,
	hdfs *hdfs_store.FileStorage, db *mongo.Collection) AccommodationHandler {
	return AccommodationHandler{accommodationService, db, hdfs, imageCache}
}
func (s *AccommodationHandler) CreateAccommodations(c *gin.Context) {
	rw := c.Writer
	h := c.Request

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, token, url)
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

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole

	if userRole != domain.Host {
		error2.ReturnJSONError(rw, "Permission denied. Only hosts can create accommodations.", http.StatusForbidden)
		return
	}

	id := primitive.NewObjectID()

	accommodation, exists := c.Get("accommodation")
	if !exists {
		error2.ReturnJSONError(rw, "Accommodation not found in context", http.StatusBadRequest)
		return
	}
	acc, ok := accommodation.(domain.Accommodation)
	if !ok {
		error2.ReturnJSONError(rw, "Invalid type for Accommodation", http.StatusBadRequest)
		return
	}
	acc.ID = id

	insertedAcc, _, err := s.accommodationService.InsertAccommodation(&acc, response.LoggedInUser.ID)
	if err != nil {
		error2.ReturnJSONError(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	jsonResponse, err1 := json.Marshal(insertedAcc)
	if err1 != nil {
		error2.ReturnJSONError(rw, fmt.Sprintf("Error marshaling JSON: %s", err1), http.StatusInternalServerError)
		return
	}
	rw.Write(jsonResponse)
}

func (s *AccommodationHandler) GetAllAccommodations(c *gin.Context) {
	accommodations, err := s.accommodationService.GetAllAccommodations()
	if err != nil {
		error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, accommodations)
}

func (s *AccommodationHandler) GetAccommodationByID(c *gin.Context) {
	accommodationID := c.Param("id")

	accommodation, err := s.accommodationService.GetAccommodationByID(accommodationID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			error2.ReturnJSONError(c.Writer, "Accommodation not found", http.StatusNotFound)
		} else {
			error2.ReturnJSONError(c.Writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, accommodation)
}

func (s *AccommodationHandler) GetAccommodationsByHostId(c *gin.Context) {
	hostID := c.Param("hostId")

	accs, err := s.accommodationService.GetAccommodationsByHostId(hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get accommodations"})
		return
	}

	if len(accs) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No accommodations found for this host", "accommodations": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, accs)
}

func (s *AccommodationHandler) DeleteAccommodation(c *gin.Context) {
	accId := c.Param("accId")

	rw := c.Writer
	h := c.Request

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, token, url)
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

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole
	userId := response.LoggedInUser.ID

	if userRole != domain.Host {
		error2.ReturnJSONError(rw, "Permission denied. Only hosts can delete accommodations.", http.StatusUnauthorized)
		return
	}

	accommodation, err := s.accommodationService.GetAccommodationByID(accId)
	if err != nil {
		errorMsg := map[string]string{"error": "Error fetching accommodation details."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	if accommodation.HostId != userId {
		errorMsg := map[string]string{"error": "Permission denied. You are not the creator of this accommodation."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	url = "https://res-server:8082/api/reservations/get/" + accId

	resp, err = s.HTTPSPerformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			error2.ReturnJSONError(rw, "Reservation service is not available.", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, "Error performing reservation request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode
	if statusCode != 200 {
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
			error2.ReturnJSONError(rw, "Invalid date format in the response", http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	counter := ReservationNumber.Number

	if counter != 0 {
		errorMsg := map[string]string{"error": "Cannot delete accommodation that has reservations in future or reservation is active."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return

	}

	err = s.accommodationService.DeleteAccommodation(accId, userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete accommodation."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Accommodation successfully deleted."})
	return
}

func (s *AccommodationHandler) CacheAndStoreImages(c *gin.Context) {
	accommodationID := c.Param("accId")
	fmt.Println(accommodationID)
	rw := c.Writer
	h := c.Request

	token := h.Header.Get("Authorization")
	url := "https://auth-server:8080/api/users/currentUser"

	timeout := 1000 * time.Second // Adjust the timeout duration as needed
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSPerformAuthorizationRequestWithContext(ctx, token, url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg := map[string]string{"error": "Authorization service is not available."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		errorMsg := map[string]string{"error": "Error performing authorization request."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
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
			errorMsg := map[string]string{"error": "Invalid date format in the response"}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		error2.ReturnJSONError(rw, fmt.Sprintf("Error decoding JSON response: %v", err), http.StatusBadRequest)
		return
	}

	// Access the 'id' from the decoded struct
	userRole := response.LoggedInUser.UserRole
	userId := response.LoggedInUser.ID

	if userRole != domain.Host {
		errorMsg := map[string]string{"error": "Permission denied. Only hosts can change accommodations."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	accommodation, err := s.accommodationService.GetAccommodationByID(accommodationID)
	if err != nil {
		errorMsg := map[string]string{"error": "Accommodation not found."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusUnauthorized)
		return
	}

	if accommodation.HostId != userId {
		errorMsg := map[string]string{"error": "Permission denied. You are not the creator of this accommodation."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		errorMsg := map[string]string{"error": "Error parsing multipart form."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return
	}
	fmt.Println("Files ready")
	files := c.Request.MultipartForm.File["image"]
	fmt.Println(files)
	fmt.Println("files here")

	// Loop through each file
	for _, fileHeader := range files {
		fmt.Println("Entered loop")
		// Open the uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			errorMsg := map[string]string{"error": "Error opening one of the uploaded files."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the file content into a byte slice
		imageData, err := ioutil.ReadAll(file)
		fmt.Println("imageData here")
		fmt.Println(imageData)
		if err != nil {
			errorMsg := map[string]string{"error": "Error reading file content."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		// Cache the image in Redis
		imageID := cache.GenerateUniqueImageID()
		fmt.Println(imageID)
		fmt.Println("imageID HERE")
		accID := accommodationID
		if err := s.imageCache.PostImage(imageID, accID, imageData); err != nil {
			errorMsg := map[string]string{"error": "Error caching image data."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}

		filename := fmt.Sprintf("%s/%s-image-%d", accID, imageID, len(files))
		err = s.hdfs.WriteFileBytes(imageData, filename, accID)
		fmt.Println(filename)
		fmt.Println("filename here")
		if err != nil {
			errorMsg := map[string]string{"error": "Error storing image in HDFS."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return
		}
	}

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
	accID := c.Param("accId")

	var images []*cache.Image
	var root = "/hdfs/created/"
	dirName := fmt.Sprintf("%s%s", root, accID)
	fmt.Println(dirName)
	fmt.Println("AFTER DIRNAME")

	files, err := ah.hdfs.Client.ReadDir(dirName)
	if err != nil {
		errorMsg := map[string]string{"error": "Unable to read dir."}
		error2.ReturnJSONError(c.Writer, errorMsg, http.StatusBadRequest)
		return

	}

	for _, filename := range files {
		parts := strings.Split(filename.Name(), "-")
		fmt.Println(parts)
		fmt.Println("here parts")

		data, err := ah.hdfs.ReadFileBytes(dirName + "/" + filename.Name())
		if err != nil {
			break
		}

		if len(parts) < 1 {
			errorMsg := map[string]string{"error": "Unable to parse name of the file."}
			error2.ReturnJSONError(c.Writer, errorMsg, http.StatusBadRequest)
			return
		}
		parts = strings.Split(parts[0], "_")
		// Convert the first part to an integer
		id, err := strconv.Atoi(parts[1])
		if err != nil {
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
		fmt.Println(images)
		fmt.Println("here images")
		err := ah.imageCache.PostAll(accID, images)
		if err != nil {
			errorMsg := map[string]string{"error": "Unable to write to cache."}
			error2.ReturnJSONError(c.Writer, errorMsg, http.StatusBadRequest)
			return
		}
	}

	c.JSON(http.StatusOK, images)
}

func (s *AccommodationHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
