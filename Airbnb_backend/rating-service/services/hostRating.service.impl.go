package services

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"rating-service/domain"
)

type HostRatingServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
	Tracer     trace.Tracer
}

func NewHostRatingServiceImpl(collection *mongo.Collection, ctx context.Context, tr trace.Tracer) HostRatingService {
	return &HostRatingServiceImpl{collection, ctx, tr}
}
func (s *HostRatingServiceImpl) SaveRating(rating *domain.RateHost, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "HostRatingService.SaveRating")
	defer span.End()

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
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		return nil
	} else if err != mongo.ErrNoDocuments {
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	_, err = s.collection.InsertOne(context.Background(), rating)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (s *HostRatingServiceImpl) DeleteRating(hostID, guestID string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "HostRatingService.DeleteRating")
	defer span.End()

	hostObjectID, err := primitive.ObjectIDFromHex(hostID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	guestObjectID, err := primitive.ObjectIDFromHex(guestID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	filter := bson.M{
		"host._id":  hostObjectID,
		"guest._id": guestObjectID,
	}

	result, err := s.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	if result.DeletedCount == 0 {
		span.SetStatus(codes.Error, "no rating found to delete")
		return errors.New("no rating found to delete")
	}

	return nil
}

func (s *HostRatingServiceImpl) GetAllRatings(ctx context.Context) ([]*domain.RateHost, float64, error) {
	ctx, span := s.Tracer.Start(ctx, "HostRatingService.GetAllRatings")
	defer span.End()

	cursor, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	var ratings []*domain.RateHost
	totalRating := 0

	for cursor.Next(context.Background()) {
		var rating domain.RateHost
		if err := cursor.Decode(&rating); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, 0, err
		}
		ratings = append(ratings, &rating)
		totalRating += int(rating.Rating)
	}

	if err := cursor.Err(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	averageRating := 0.0
	if len(ratings) > 0 {
		averageRating = float64(totalRating) / float64(len(ratings))
	}

	return ratings, averageRating, nil
}

func (s *HostRatingServiceImpl) GetByHostAndGuest(hostID, guestID string, ctx context.Context) ([]domain.RateHost, error) {
	ctx, span := s.Tracer.Start(ctx, "HostRatingService.GetByHostAndGuest")
	defer span.End()

	hostObjectID, err := primitive.ObjectIDFromHex(hostID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	guestObjectID, err := primitive.ObjectIDFromHex(guestID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	filter := bson.M{
		"host._id":  hostObjectID,
		"guest._id": guestObjectID,
	}

	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer cursor.Close(context.Background())

	var ratings []domain.RateHost
	if err := cursor.All(context.Background(), &ratings); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return ratings, nil
}
