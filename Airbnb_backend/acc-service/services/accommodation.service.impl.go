package services

import (
	"acc-service/domain"
	"context"

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
