package services

import (
	"acc-service/domain"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
)

type AccommodationServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewAccommodationServiceImpl(collection *mongo.Collection, ctx context.Context) AccommodationService {
	return &AccommodationServiceImpl{collection, ctx}
}

func (s *AccommodationServiceImpl) InsertAccommodation(accomm *domain.Accommodation, hostID string) (*domain.Accommodation, string, error) {
	accomm.HostId = hostID
	accomm.ID = primitive.NilObjectID

	result, err := s.collection.InsertOne(context.Background(), accomm)
	if err != nil {
		return nil, "", err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, "", errors.New("failed to get inserted ID")
	}

	insertedID = result.InsertedID.(primitive.ObjectID)

	return accomm, insertedID.Hex(), nil
}

func (s *AccommodationServiceImpl) GetAllAccommodations() ([]*domain.Accommodation, error) {
	cursor, err := s.collection.Find(context.Background(), bson.M{})
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

func (s *AccommodationServiceImpl) GetAccommodationByID(accommodationID string) (*domain.Accommodation, error) {
	objID, err := primitive.ObjectIDFromHex(accommodationID)
	if err != nil {
		return nil, err
	}

	var accommodation domain.Accommodation
	err = s.collection.FindOne(s.ctx, bson.M{"_id": objID}).Decode(&accommodation)
	if err != nil {
		return nil, err
	}
	return &accommodation, nil
}

func (s *AccommodationServiceImpl) GetAccommodationsByHostId(hostId string) ([]*domain.Accommodation, error) {
	var accommodations []*domain.Accommodation
	cursor, err := s.collection.Find(context.Background(), domain.Accommodation{HostId: hostId})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &accommodations); err != nil {
		return nil, err
	}
	return accommodations, nil
}
