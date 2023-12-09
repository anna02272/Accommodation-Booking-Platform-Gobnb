package services

import (
	"context"
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
