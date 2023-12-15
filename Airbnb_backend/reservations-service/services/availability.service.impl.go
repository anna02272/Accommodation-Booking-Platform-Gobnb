package services

import (
	"context"
	"errors"
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

func (s *AvailabilityServiceImpl) InsertMulitipleAvailability(accomm data.AvailabilityPeriod, accId primitive.ObjectID) ([]*data.Availability, error) {
	var insertedAvailabilities []*data.Availability
	var startDate1 = accomm.StartDate
	startDate := time.Unix(int64(startDate1)/1000, (int64(startDate1)%1000)*1000000)
	var endDate1 = accomm.EndDate
	endDate := time.Unix(int64(endDate1)/1000, (int64(endDate1)%1000)*1000000)

	isAvailable, err := s.IsAvailable(accId, startDate, endDate)
	if err != nil {
		if err.Error() == "Availability not defined for accommodation." {
			fmt.Println("Zero availability")
		} else {
			isNotDefined, err := s.IsNotDefined(accId, startDate, endDate)
			if err != nil {
				return nil, errors.New("Some of the dates are already defined in the database")
			}

			if !isNotDefined {
				return nil, errors.New("Accommodation is already defined and available for the given date range")
			}
		}
	}

	if isAvailable {
		if err.Error() == "Availability not defined for accommodation." {
			fmt.Println("Zero availability")
		} else {
			return nil, errors.New("Accommodation is already defined and available for the given date range.")
		}
	}

	if startDate.After(endDate) {
		return nil, errors.New("Start date is after end date.")
	}

	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		var newAccomm data.Availability
		newAccomm.AccommodationID = accId
		newAccomm.AvailabilityType = accomm.AvailabilityType
		newAccomm.Price = accomm.Price
		newAccomm.PriceType = accomm.PriceType
		dt := primitive.DateTime(d.UnixNano() / 1000000)
		newAccomm.Date = dt

		_, err := s.collection.InsertOne(context.Background(), newAccomm)
		if err != nil {
			return nil, err
		}

		insertedAvailabilities = append(insertedAvailabilities, &newAccomm)
	}

	return insertedAvailabilities, nil
}

func (s *AvailabilityServiceImpl) GetAllAvailability(accommodationID primitive.ObjectID) ([]*data.Availability, error) {
	filter := bson.M{
		"accommodation_id": accommodationID,
	}
	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	var availabilities []*data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		return nil, err
	}
	return availabilities, nil
}

func (s *AvailabilityServiceImpl) GetAvailabilityByID(availabilityID primitive.ObjectID) (*data.Availability, error) {
	filter := bson.M{
		"_id": availabilityID,
	}
	var availability data.Availability
	err := s.collection.FindOne(context.Background(), filter).Decode(&availability)
	if err != nil {
		return nil, err
	}
	return &availability, nil
}

func (s *AvailabilityServiceImpl) EditAvailability(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, availabilityType data.AvailabilityType) error {

	isAvailable, err := s.IsAvailable(accommodationID, startDate, endDate)
	if err != nil {
		return err
	}
	if !isAvailable {
		return errors.New("Accommodation is not available for the given date range.")
	}

	filter := bson.M{
		"accommodation_id": accommodationID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"availability_type": availabilityType,
		},
	}
	result, err := s.collection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return err
	}
	result = result
	return nil
}

func (s *AvailabilityServiceImpl) DeleteAvailability(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time) error {

	isAvailable, err := s.IsAvailable(accommodationID, startDate, endDate)
	if err != nil {
		return err
	}
	if !isAvailable {
		return errors.New("Accommodation is not available for the given date range.")
	}

	filter := bson.M{
		"accommodation_id": accommodationID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	_, err = s.collection.DeleteMany(context.Background(), filter)
	return err
}

func (s *AvailabilityServiceImpl) IsAvailable(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time) (bool, error) {
	fmt.Println(startDate.String())
	fmt.Println(endDate.String())
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

	counter := len(availabilities)
	// difference between startDate and endDate in days
	days := int(endDate.Sub(startDate).Hours()/24) + 1

	for _, availability := range availabilities {
		fmt.Println(availability.ID)
		fmt.Println("here")
		fmt.Println(availability.AvailabilityType)
		fmt.Println(availability.Date)
		if availability.AvailabilityType != data.Available {
			return false, nil
		}
	}

	fmt.Println("counter: ", counter)
	fmt.Println("days: ", days)

	if counter != days {
		if counter == 0 {
			return false, errors.New("Availability not defined for accommodation.")
			fmt.Println("Zero availability in the database.")
		}
		return false, errors.New("Not all dates are defined in the database.")
	}

	return true, nil
}

func (s *AvailabilityServiceImpl) IsNotDefined(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time) (bool, error) {
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

	counter := len(availabilities)
	// difference between startDate and endDate in days
	//days := int(endDate.Sub(startDate).Hours()/24) + 1

	if counter == 0 {
		return true, nil
	}

	return false, errors.New("Dates are already defined in the database.")
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

func (s *AvailabilityServiceImpl) GetAvailabilityByAccommodationId(accommodationID primitive.ObjectID) ([]*data.Availability, error) {
	filter := bson.M{
		"accommodation_id": accommodationID,
	}
	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	var availabilities []*data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		return nil, err
	}
	return availabilities, nil
}
