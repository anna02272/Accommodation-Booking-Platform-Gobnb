package services

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
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
	filter := bson.M{
		"host._id":  rating.Host.ID,
		"guest._id": rating.Guest.ID,
	}

	existingRating := &domain.RateHost{}
	err := s.collection.FindOne(context.Background(), filter).Decode(existingRating)

	if err == nil {
		update := bson.M{
			"$set": bson.M{"rating": rating.Rating, "date-and-time": rating.DateAndTime},
		}

		_, err := s.collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
		return nil
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	_, err = s.collection.InsertOne(context.Background(), rating)
	if err != nil {
		return err
	}

	return nil
}

func (s *HostRatingServiceImpl) DeleteRating(hostID, guestID string) error {
	hostObjectID, err := primitive.ObjectIDFromHex(hostID)
	if err != nil {
		return err
	}

	guestObjectID, err := primitive.ObjectIDFromHex(guestID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"host._id":  hostObjectID,
		"guest._id": guestObjectID,
	}

	result, err := s.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no rating found to delete")
	}

	return nil
}
