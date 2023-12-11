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

//
//Add this check that guest can rate it only if he had one reservation before from that host. This is get all : 	urlCheckReservations := "https://res-server:8082/api/reservations/getAllByHost"
//fmt.Println(urlCheckReservations)
//
//timeout := 2000 * time.Second // Adjust the timeout duration as needed
//ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
//defer cancel()
//
//respRes, errRes := ac.HTTPSperformAuthorizationRequestWithContext(ctxRest, tokenStringHeader, urlCheckReservations, "GET")
//if errRes != nil {
//fmt.Println(err)
//if ctx.Err() == context.DeadlineExceeded {
//ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch user reservations"})
//return
//}
//ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch user reservations"})
//return
//}
//defer respRes.Body.Close()
//fmt.Println(respRes.StatusCode)
//if respRes.StatusCode != 200 {
//ctx.JSON(http.StatusBadRequest, gin.H{"message": "You cannot rank host, you didn't have reservations from him before"})
//return
//}
//}

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
func (s *HostRatingServiceImpl) GetAllRatings() ([]*domain.RateHost, float64, error) {
	cursor, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	var ratings []*domain.RateHost
	totalRating := 0

	for cursor.Next(context.Background()) {
		var rating domain.RateHost
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
