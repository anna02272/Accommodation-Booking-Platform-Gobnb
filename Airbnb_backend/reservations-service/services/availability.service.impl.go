package services

import (
	"context"
	"fmt"
	"reservations-service/data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AvailabilityServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewAvailabilityServiceImpl(collection *mongo.Collection, ctx context.Context) AvailabilityService {
	return &AvailabilityServiceImpl{collection, ctx}
}

func (s *AvailabilityServiceImpl) InsertAvailability(accomm *data.Availability) (*data.Availability, error) {

	//accomm.Date = primitive.NewDateTimeFromTime(accomm.Date)

	_, err := s.collection.InsertOne(context.Background(), accomm)
	if err != nil {
		return nil, err
	}

	//insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	return accomm, nil
}

func (s *AvailabilityServiceImpl) IsAvailable(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time) (bool, error) {
	filter := bson.M{
		"accommodation_id": accommodationID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		return false, err
	}
	var availabilities []data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		return false, err
	}
	if len(availabilities) == 0 {
		return false, nil
	}
	fmt.Println(len(availabilities))
	for _, availability := range availabilities {
		//fmt.Println(availability.ID)
		fmt.Println("here")
		fmt.Println(availability.AvailabilityType)
		fmt.Println(availability.Date)
		if availability.AvailabilityType == data.Booked {
			return false, nil
		}
	}
	return true, nil
}

func (s *AvailabilityServiceImpl) BookAccommodation(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time) error {
	filter := bson.M{
		"accommodation_id": accommodationID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"availability_type": data.Booked,
		},
	}
	_, err := s.collection.UpdateMany(context.Background(), filter, update)
	return err
}

// func (s *AvailabilityServiceImpl) GetAllAvailability() ([]*data.Availability, error) {
// 	cursor, err := s.collection.Find(context.Background(), bson.M{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(context.Background())

// 	var availability []*data.Availability
// 	for cursor.Next(context.Background()) {
// 		var acc data.Availability
// 		if err := cursor.Decode(&acc); err != nil {
// 			return nil, err
// 		}
// 		availability = append(availability, &acc)
// 	}

// 	if err := cursor.Err(); err != nil {
// 		return nil, err
// 	}

// 	return availability, nil
// }

// func (s *AvailabilityServiceImpl) GetAvailabilityByID(availabilityID string) (*data.Availability, error) {
// 	objID, err := primitive.ObjectIDFromHex(availabilityID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var availability data.Availability
// 	err = s.collection.FindOne(s.ctx, bson.M{"_id": objID}).Decode(&availability)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &availability, nil
// }

// func (s *AvailabilityServiceImpl) GetAvailabilityByHostId(hostId string) ([]*data.Availability, error) {
// 	var availability []*data.Availability
// 	cursor, err := s.collection.Find(context.Background(), data.Availability{HostId: hostId})
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err = cursor.All(context.Background(), &availability); err != nil {
// 		return nil, err
// 	}
// 	return availability, nil
//}
