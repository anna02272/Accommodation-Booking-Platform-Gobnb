package services

import (
	"acc-service/domain"
	"context"
	"errors"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"go.mongodb.org/mongo-driver/mongo"
)

type AccommodationServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
	Tracer     trace.Tracer
}

func NewAccommodationServiceImpl(collection *mongo.Collection, ctx context.Context, tr trace.Tracer) AccommodationService {
	return &AccommodationServiceImpl{collection, ctx, tr}
}

func (s *AccommodationServiceImpl) InsertAccommodation(accomm *domain.Accommodation, hostID string, ctx context.Context) (*domain.Accommodation, string, error) {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.InsertAccommodation")
	defer span.End()

	accomm.HostId = hostID

	result, err := s.collection.InsertOne(context.Background(), accomm)
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

	return accomm, insertedID.Hex(), nil
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

func (s *AccommodationServiceImpl) DeleteAccommodation(accommodationID string, hostID string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(s.ctx, "AccommodationService.DeleteAccommodation")
	defer span.End()

	objID, err := primitive.ObjectIDFromHex(accommodationID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	filter := bson.M{"_id": objID, "host_id": hostID}

	_, err = s.collection.DeleteOne(context.Background(), filter)
	span.SetStatus(codes.Error, err.Error())
	return err
}

func (s *AccommodationServiceImpl) GetAccommodationBySearch(location string, guests string, amenities map[string]bool, amenitiesExist bool) ([]*domain.Accommodation, error) {

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
			return nil, errors.New("failed to parse guests")
		}

		filter["accommodation_min_guests"] = bson.M{"$lte": guests}
		filter["accommodation_max_guests"] = bson.M{"$gte": guests}
	}

	// if amenitiesExist {
	// 	var tv = amenities["tv"]
	// 	var wifi = amenities["wifi"]
	// 	var ac = amenities["ac"]
	// 	filter["accommodation_amenities"] = bson.M{
	// 		"TV":   tv,
	// 		"Wifi": wifi,
	// 		"AC":   ac,
	// 	}
	// }

	cursor, err := s.collection.Find(context.Background(), filter)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	var accommodations []*domain.Accommodation
	for cursor.Next(context.Background()) {
		var acc domain.Accommodation
		if err := cursor.Decode(&acc); err != nil {
			return nil, err
		}

		accommodations = append(accommodations, &acc)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return accommodations, nil

}
