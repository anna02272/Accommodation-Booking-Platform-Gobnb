package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"rating-service/domain"
)

type HostRatingServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewHostRatingServiceImpl(collection *mongo.Collection, ctx context.Context) HostRatingService {
	return &HostRatingServiceImpl{collection, ctx}
}
func (s *HostRatingServiceImpl) SaveRating(rating *domain.RateHost) error {
	_, err := s.collection.InsertOne(context.Background(), rating)
	if err != nil {
		return err
	}
	return nil
}
func (s *HostRatingServiceImpl) InsertAccommodation(accomm *domain.Accommodation, hostID string) (*domain.Accommodation, string, error) {
	accomm.HostId = hostID

	result, err := s.collection.InsertOne(context.Background(), accomm)
	if err != nil {
		return nil, "", err
	}

	insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	return accomm, insertedID, nil
}
