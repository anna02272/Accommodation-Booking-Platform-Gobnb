package services

import (
	"acc-service/application"
	"acc-service/domain"
	error2 "acc-service/error"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"go.mongodb.org/mongo-driver/mongo"
)

type AccommodationServiceImpl struct {
	collection   *mongo.Collection
	ctx          context.Context
	Tracer       trace.Tracer
	orchestrator *application.CreateAccommodationOrchestrator
}

func NewAccommodationServiceImpl(collection *mongo.Collection, ctx context.Context,
	tr trace.Tracer, orchestrator *application.CreateAccommodationOrchestrator) AccommodationService {
	return &AccommodationServiceImpl{collection, ctx, tr, orchestrator}
}

func (s *AccommodationServiceImpl) InsertAccommodation(accomm *domain.AccommodationWithAvailability, hostID string, ctx context.Context) (*domain.Accommodation, string, error) {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.InsertAccommodation")
	defer span.End()

	accomm.HostId = hostID
	accommodation := &domain.Accommodation{
		ID:        accomm.ID,
		HostId:    hostID,
		Name:      accomm.Name,
		Location:  accomm.Location,
		Amenities: accomm.Amenities,
		MinGuests: accomm.MinGuests,
		MaxGuests: accomm.MaxGuests,
		Active:    false,
	}

	result, err := s.collection.InsertOne(context.Background(), accommodation)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, "", err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		span.SetStatus(codes.Error, "failed to get inserted ID")
		return nil, "", errors.New("failed to get inserted ID")
	}

	insertedID = result.InsertedID.(primitive.ObjectID)
	accomm.ID = insertedID
	return accommodation, insertedID.Hex(), nil
}

func (s *AccommodationServiceImpl) CreateAvailabilityInReservationService(rw http.ResponseWriter, accomm *domain.AccommodationWithAvailability, ctx context.Context, token string) error {
	ctx, span := s.Tracer.Start(ctx, "AccommodationService.CreateAvailabilityInReservationService")
	defer span.End()

	url := "https://res-server:8082/api/availability/create/" + accomm.ID.Hex()
	timeout := 2000 * time.Second
	ctxx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	availability := &domain.AvailabilityPeriod{
		StartDate:        accomm.StartDate,
		EndDate:          accomm.EndDate,
		Price:            accomm.Price,
		PriceType:        accomm.PriceType,
		AvailabilityType: accomm.AvailabilityType,
	}

	resp, err := s.HTTPSperformAuthorizationRequestWithContext(ctx, availability, url, token)
	if err != nil {
		if ctxx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Reservation service not available")
			errorMsg := map[string]string{"error": "Reservation service not available"}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return nil
		}
		span.SetStatus(codes.Error, "Reservation service not available")
		errorMsg := map[string]string{"error": "Reservation service not available"}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return nil
	}

	defer resp.Body.Close()

	return nil
}

func (us *AccommodationServiceImpl) HTTPSperformAuthorizationRequestWithContext(ctx context.Context, availability *domain.AvailabilityPeriod, url string, token string) (*http.Response, error) {
	reqBody, err := json.Marshal(availability)
	if err != nil {
		return nil, fmt.Errorf("error marshaling user JSON: %v", err)
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
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

func (s *AccommodationServiceImpl) SendToRatingService(accommodation *domain.AccommodationWithAvailability, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "AccommodationService.SendToRating")
	defer span.End()

	var rw http.ResponseWriter
	url := "https://rating-server:8087/api/rating/createAccommodation"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctxx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContextt(ctx, accommodation, url)
	if err != nil {
		if ctxx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Rating service not available..")
			errorMsg := map[string]string{"error": "Rating service not available.."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return nil
		}
		span.SetStatus(codes.Error, "Rating service not available..")
		errorMsg := map[string]string{"error": "Rating service not available.."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return nil
	}

	defer resp.Body.Close()

	return nil
}
func (s *AccommodationServiceImpl) HTTPSperformAuthorizationRequestWithContextt(ctx context.Context, accommodation *domain.AccommodationWithAvailability, url string) (*http.Response, error) {
	reqBody, err := json.Marshal(accommodation)
	if err != nil {
		return nil, fmt.Errorf("error marshaling accommodation JSON: %v", err)
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

func (s *AccommodationServiceImpl) GetAllAccommodations(ctx context.Context) ([]*domain.Accommodation, error) {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.GetAllAccommodations")
	defer span.End()

	cursor, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer cursor.Close(context.Background())

	var accommodations []*domain.Accommodation
	for cursor.Next(context.Background()) {
		var acc domain.Accommodation
		if err := cursor.Decode(&acc); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		accommodations = append(accommodations, &acc)
	}

	if err := cursor.Err(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return accommodations, nil
}

func (s *AccommodationServiceImpl) GetAccommodationByID(accommodationID string, ctx context.Context) (*domain.Accommodation, error) {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.GetAccommodationByID")
	defer span.End()

	objID, err := primitive.ObjectIDFromHex(accommodationID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var accommodation domain.Accommodation
	err = s.collection.FindOne(s.ctx, bson.M{"_id": objID}).Decode(&accommodation)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return &accommodation, nil
}

func (s *AccommodationServiceImpl) GetAccommodationsByHostId(hostId string, ctx context.Context) ([]*domain.Accommodation, error) {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.GetAccommodationByHostID")
	defer span.End()

	filter := bson.M{"host_id": hostId}
	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var accommodations []*domain.Accommodation
	for cursor.Next(context.Background()) {
		var acc domain.Accommodation
		if err := cursor.Decode(&acc); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		accommodations = append(accommodations, &acc)
	}

	if err := cursor.Err(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return accommodations, nil
}

func (s *AccommodationServiceImpl) GetAccommodationByHostIdAndAccId(hostId string, accId string, ctx context.Context) (*domain.Accommodation, error) {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.GetAccommodationByHostIdAndAccId")
	defer span.End()

	objID, err := primitive.ObjectIDFromHex(accId)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	filter := bson.M{"host_id": hostId, "_id": objID}

	var accommodation domain.Accommodation
	err = s.collection.FindOne(context.Background(), filter).Decode(&accommodation)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &accommodation, nil
}
func (s *AccommodationServiceImpl) SendToDelete(rw http.ResponseWriter, accommodationId string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "AccService.SendToDelete")
	defer span.End()

	url := "https://rating-server:8087/api/rating/deleteAccommodation"

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctxx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := s.HTTPSperformAuthorizationRequestWithContexttt(ctx, accommodationId, url)
	if err != nil {
		if ctxx.Err() == context.DeadlineExceeded {
			span.SetStatus(codes.Error, "Rating service not available..")
			errorMsg := map[string]string{"error": "Rating service not available.."}
			error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
			return nil
		}
		span.SetStatus(codes.Error, "Rating service not available..")
		errorMsg := map[string]string{"error": "Rating service not available.."}
		error2.ReturnJSONError(rw, errorMsg, http.StatusBadRequest)
		return nil
	}

	defer resp.Body.Close()

	return nil
}
func (s *AccommodationServiceImpl) HTTPSperformAuthorizationRequestWithContexttt(ctx context.Context, accommodationId string, url string) (*http.Response, error) {
	reqBody, err := json.Marshal(accommodationId)
	if err != nil {
		return nil, fmt.Errorf("error marshaling accommodation JSON: %v", err)
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(reqBody))
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
func (s *AccommodationServiceImpl) DeleteAccommodation(accommodationID string, hostID string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.DeleteAccommodation")
	defer span.End()
	var rw http.ResponseWriter
	er := s.SendToDelete(rw, accommodationID, ctx)
	if er != nil {
		//span.SetStatus(codes.Error, er.Error())
		return nil
	}

	objID, err := primitive.ObjectIDFromHex(accommodationID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	filter := bson.M{"_id": objID, "host_id": hostID}

	_, err = s.collection.DeleteOne(context.Background(), filter)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	//span.SetStatus(codes.Error, err.Error())
	return nil
}

func (s *AccommodationServiceImpl) GetAccommodationBySearch(location string, guests string, amenities map[string]bool, amenitiesExist bool, ctx context.Context) ([]*domain.Accommodation, error) {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.GetAccommodationBySearch")
	defer span.End()
	filter := bson.M{}

	if location != "" {
		filter["accommodation_location"] = location
	}

	// if guests != "" {
	// 	guests, err := strconv.Atoi(guests)
	// 	if err != nil {
	// 		return nil, errors.New("failed to parse guests")
	// 	}

	// 	filter["accommodation_min_guests"] = bson.M{"$gte": guests}
	// }

	// if guests != "" {
	// 	guests, err := strconv.Atoi(guests)
	// 	if err != nil {
	// 		return nil, errors.New("failed to parse maxGuests")
	// 	}

	// 	filter["accommodation_max_guests"] = bson.M{"$lte": guests}
	// }

	if guests != "" {
		guests, err := strconv.Atoi(guests)
		if err != nil {
			span.SetStatus(codes.Error, "failed to parse guests")
			return nil, errors.New("failed to parse guests")
		}

		filter["accommodation_min_guests"] = bson.M{"$lte": guests}
		filter["accommodation_max_guests"] = bson.M{"$gte": guests}
	}

	if amenitiesExist {
		var tv = amenities["TV"]
		var wifi = amenities["WiFi"]
		var ac = amenities["AC"]
		fmt.Println("in service: ", tv, wifi, ac)
		if tv == true {
			filter["accommodation_amenities.TV"] = tv
		}

		if wifi == true {
			filter["accommodation_amenities.WiFi"] = wifi
		}
		if ac == true {
			filter["accommodation_amenities.AC"] = ac
		}
	}

	cursor, err := s.collection.Find(context.Background(), filter)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	defer cursor.Close(context.Background())

	var accommodations []*domain.Accommodation
	for cursor.Next(context.Background()) {
		var acc domain.Accommodation
		if err := cursor.Decode(&acc); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		accommodations = append(accommodations, &acc)
	}

	if err := cursor.Err(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return accommodations, nil

}

func (s *AccommodationServiceImpl) GetHostIdByAccommodationId(accID string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(accID)
	if err != nil {
		return "", err
	}

	var accommodation domain.Accommodation
	err = s.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&accommodation)
	if err != nil {
		return "", err
	}

	//accommodation.hostid to string
	//var hostid string
	//hostid = accommodation.HostId

	return accommodation.HostId, nil
}
