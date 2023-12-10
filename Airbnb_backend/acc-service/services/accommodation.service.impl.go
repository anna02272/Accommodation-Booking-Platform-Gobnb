package services

import (
	"acc-service/domain"
	"context"
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

func (s *AccommodationServiceImpl) SaveAccommodation(accommodation *domain.Accommodation) error {
	_, err := s.collection.InsertOne(context.Background(), accommodation)
	if err != nil {
		return err
	}
	return nil
}

func (s *AccommodationServiceImpl) GetAccommodationById(id primitive.ObjectID) (*domain.Accommodation, error) {
	var accommodation *domain.Accommodation
	err := s.collection.FindOne(context.Background(), domain.Accommodation{ID: id}).Decode(&accommodation)
	if err != nil {
		return nil, err
	}
	return accommodation, nil
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

func (s *AccommodationServiceImpl) GetAllAccommodations() ([]*domain.Accommodation, error) {
	var accommodations []*domain.Accommodation
	cursor, err := s.collection.Find(context.Background(), domain.Accommodation{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &accommodations); err != nil {
		return nil, err
	}
	return accommodations, nil
}
