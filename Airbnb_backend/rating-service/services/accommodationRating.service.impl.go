package services

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"rating-service/domain"
)

type AccommodationRatingServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewAccommodationRatingServiceImpl(collection *mongo.Collection, ctx context.Context) AccommodationRatingService {
	return &AccommodationRatingServiceImpl{collection, ctx}
}
func (s *AccommodationRatingServiceImpl) SaveRating(rating *domain.RateAccommodation) error {
	filter := bson.M{
		"accommodationID": rating.Accommodation,
		"guest._id":       rating.Guest.ID,
	}

	existingRating := &domain.RateAccommodation{}
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

func (s *AccommodationRatingServiceImpl) DeleteRating(accommodationID, guestID string) error {
	guestObjectID, err := primitive.ObjectIDFromHex(guestID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"accommodationID": accommodationID,
		"guest._id":       guestObjectID,
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

func (s *AccommodationRatingServiceImpl) GetAllRatings() ([]*domain.RateAccommodation, float64, error) {
	cursor, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	var ratings []*domain.RateAccommodation
	totalRating := 0

	for cursor.Next(context.Background()) {
		var rating domain.RateAccommodation
		if err := cursor.Decode(&rating); err != nil {
			return nil, 0, err
		}
		ratings = append(ratings, &rating)
		totalRating += int(rating.Rating)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	averageRating := 0.0
	if len(ratings) > 0 {
		averageRating = float64(totalRating) / float64(len(ratings))
	}

	return ratings, averageRating, nil
}

func (s *AccommodationRatingServiceImpl) GetByAccommodationAndGuest(accommodationID, guestID string) ([]domain.RateAccommodation, error) {

	guestObjectID, err := primitive.ObjectIDFromHex(guestID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"accommodationID": accommodationID,
		"guest._id":       guestObjectID,
	}

	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var ratings []domain.RateAccommodation
	if err := cursor.All(context.Background(), &ratings); err != nil {
		return nil, err
	}

	return ratings, nil
}
