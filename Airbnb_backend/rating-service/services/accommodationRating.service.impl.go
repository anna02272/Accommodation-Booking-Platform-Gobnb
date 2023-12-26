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

type AccommodationRatingServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
	Tracer     trace.Tracer
}

func NewAccommodationRatingServiceImpl(collection *mongo.Collection, ctx context.Context, tr trace.Tracer) AccommodationRatingService {
	return &AccommodationRatingServiceImpl{collection, ctx, tr}
}
func (s *AccommodationRatingServiceImpl) SaveRating(rating *domain.RateAccommodation, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "AccommodationRatingService.SaveRating")
	defer span.End()
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

func (s *AccommodationRatingServiceImpl) DeleteRating(accommodationID, guestID string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "AccommodationRatingService.DeleteRating")
	defer span.End()

	guestObjectID, err := primitive.ObjectIDFromHex(guestID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	filter := bson.M{
		"accommodationID": accommodationID,
		"guest._id":       guestObjectID,
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

func (s *AccommodationRatingServiceImpl) GetAllRatingsAccommodation(ctx context.Context) ([]*domain.RateAccommodation, float64, error) {
	ctx, span := s.Tracer.Start(ctx, "AccommodationRatingService.GetAllRatings")
	defer span.End()
	cursor, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	var ratings []*domain.RateAccommodation
	totalRating := 0

	for cursor.Next(context.Background()) {
		var rating domain.RateAccommodation
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

func (s *AccommodationRatingServiceImpl) GetByAccommodationAndGuest(accommodationID, guestID string, ctx context.Context) ([]domain.RateAccommodation, error) {
	ctx, span := s.Tracer.Start(ctx, "AccommodationRatingService.GetByAccommodationAndGuest")
	defer span.End()
	guestObjectID, err := primitive.ObjectIDFromHex(guestID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	filter := bson.M{
		"accommodationID": accommodationID,
		"guest._id":       guestObjectID,
	}

	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer cursor.Close(context.Background())

	var ratings []domain.RateAccommodation
	if err := cursor.All(context.Background(), &ratings); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return ratings, nil
}

//func (s *AccommodationRatingHandler) HTTPSPerformAuthorizationRequestWithContext(ctx context.Context, token string, url string) (*http.Response, error) {
//	tr := http.DefaultTransport.(*http.Transport).Clone()
//	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
//
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return nil, err
//	}
//	req.Header.Set("Authorization", token)
//
//	client := &http.Client{Transport: tr}
//	resp, err := client.Do(req.WithContext(ctx))
//	if err != nil {
//		return nil, err
//	}
//
//	return resp, nil
//}
