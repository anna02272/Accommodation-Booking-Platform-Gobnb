package services

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"reservations-service/data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AvailabilityServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
	Tracer     trace.Tracer
}

func NewAvailabilityServiceImpl(collection *mongo.Collection, ctx context.Context, tr trace.Tracer) AvailabilityService {
	return &AvailabilityServiceImpl{collection, ctx, tr}
}

func (s *AvailabilityServiceImpl) InsertAvailability(accomm *data.Availability, ctx context.Context) (*data.Availability, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.InsertAvailability")
	defer span.End()
	//accomm.Date = primitive.NewDateTimeFromTime(accomm.Date)

	_, err := s.collection.InsertOne(context.Background(), accomm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	//insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	return accomm, nil
}

func (s *AvailabilityServiceImpl) InsertMulitipleAvailability(accomm data.AvailabilityPeriod, accId primitive.ObjectID, ctx context.Context) ([]*data.Availability, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.InsertMulitipleAvailability")
	defer span.End()

	var insertedAvailabilities []*data.Availability
	var startDate1 = accomm.StartDate
	startDate := time.Unix(int64(startDate1)/1000, (int64(startDate1)%1000)*1000000)
	var endDate1 = accomm.EndDate
	endDate := time.Unix(int64(endDate1)/1000, (int64(endDate1)%1000)*1000000)

	// _, err := s.IsAvailable(accId, startDate, endDate)
	// if err != nil {
	// 	if err.Error() == "Availability not defined for accommodation." {
	// 		fmt.Println("Zero availability")
	// 	} else {
	// 		return nil, err
	// 	}
	// }

	// if isAvailable {
	// 	if err.Error() == "Availability not defined for accommodation." {
	// 		fmt.Println("Zero availability")
	// 	} else {
	// 		return nil, errors.New("Accommodation is already defined and available for the given date range.")
	// 	}
	// }
	isBooked, err := s.IsBooked(accId, startDate, endDate, ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if isBooked {
		span.SetStatus(codes.Error, "Accommodation is already defined and booked for the given date range.")
		return nil, errors.New("Accommodation is already defined and booked for the given date range.")
	}

	// if !isAvailable {
	// 	return nil, errors.New("Accommodation is already defined and unavailable or booked for the given date range.")
	// }

	if startDate.After(endDate) {
		span.SetStatus(codes.Error, "Start date is after end date.")
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

		exists, err := s.Exists(accId, d, ctx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		if exists {
			//set update
			filter := bson.M{
				"accommodation_id": accId,
				"date":             dt,
			}
			update := bson.M{
				"$set": bson.M{
					"availability_type": accomm.AvailabilityType,
					"price":             accomm.Price,
					"price_type":        accomm.PriceType,
				},
			}
			_, err := s.collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				return nil, err
			}

			fmt.Print("Updated")

			continue
		}

		_, err2 := s.collection.InsertOne(context.Background(), newAccomm)
		if err2 != nil {
			span.SetStatus(codes.Error, err2.Error())
			return nil, err2
		}

		fmt.Print("Inserted")

		insertedAvailabilities = append(insertedAvailabilities, &newAccomm)
	}

	return insertedAvailabilities, nil
}

func (s *AvailabilityServiceImpl) GetAllAvailability(accommodationID primitive.ObjectID, ctx context.Context) ([]*data.Availability, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.GetAllAvailability")
	defer span.End()

	filter := bson.M{
		"accommodation_id": accommodationID,
	}
	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	var availabilities []*data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return availabilities, nil
}

func (s *AvailabilityServiceImpl) GetAvailabilityByID(availabilityID primitive.ObjectID, ctx context.Context) (*data.Availability, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.GetAvailabilityByID")
	defer span.End()

	filter := bson.M{
		"_id": availabilityID,
	}
	var availability data.Availability
	err := s.collection.FindOne(context.Background(), filter).Decode(&availability)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return &availability, nil
}

func (s *AvailabilityServiceImpl) EditAvailability(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, availabilityType data.AvailabilityType, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.EditAvailability")
	defer span.End()

	isAvailable, err := s.IsAvailable(accommodationID, startDate, endDate, ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if !isAvailable {
		span.SetStatus(codes.Error, "Accommodation is not available for the given date range.")
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
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	result = result
	return nil
}

func (s *AvailabilityServiceImpl) DeleteAvailability(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.DeleteAvailability")
	defer span.End()

	isAvailable, err := s.IsAvailable(accommodationID, startDate, endDate, ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if !isAvailable {
		span.SetStatus(codes.Error, "Accommodation is not available for the given date range.")
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

func (s *AvailabilityServiceImpl) IsAvailable(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, ctx context.Context) (bool, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.IsAvailable")
	defer span.End()

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
		span.SetStatus(codes.Error, err.Error())
		fmt.Println("Database Query Error:", err)
		return false, err
	}

	var availabilities []data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		span.SetStatus(codes.Error, err.Error())
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
			span.SetStatus(codes.Error, "Availability not defined for accommodation.")
			return false, errors.New("Availability not defined for accommodation.")
			fmt.Println("Zero availability in the database.")
		}
		span.SetStatus(codes.Error, "Not all dates are defined in the database.")
		return false, errors.New("Not all dates are defined in the database.")
	}

	return true, nil
}

func (s *AvailabilityServiceImpl) IsNotDefined(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, ctx context.Context) (bool, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.IsNotDefined")
	defer span.End()

	filter := bson.M{
		"accommodation_id": accommodationID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	var availabilities []data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	counter := len(availabilities)
	// difference between startDate and endDate in days
	//days := int(endDate.Sub(startDate).Hours()/24) + 1

	if counter == 0 {
		return true, nil
	}
	span.SetStatus(codes.Error, "Dates are already defined in the database.")
	return false, errors.New("Dates are already defined in the database.")
}

func (s *AvailabilityServiceImpl) IsBooked(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, ctx context.Context) (bool, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.IsBooked")
	defer span.End()

	filter := bson.M{
		"accommodation_id": accommodationID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
		"availability_type": data.Booked,
	}
	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	var availabilities []data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	counter := len(availabilities)
	// difference between startDate and endDate in days
	//days := int(endDate.Sub(startDate).Hours()/24) + 1

	// if counter == 0 {
	// 	return false, errors.New("Availability not defined for accommodation.")
	// }

	// for _, availability := range availabilities {
	// 	if availability.AvailabilityType != data.Booked {
	// 		fmt.Printf(availability.Date.String() + " is not booked.")
	// 		return false, nil
	// 	}
	// }

	// return true, nil
	if counter == 0 {
		return false, nil
	} else {
		return true, nil
	}

}

func (s *AvailabilityServiceImpl) Exists(accommodationID primitive.ObjectID, startDate time.Time, ctx context.Context) (bool, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.Exists")
	defer span.End()

	filter := bson.M{
		"accommodation_id": accommodationID,
		"date":             startDate,
	}

	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		return false, err
	}

	var availabilities []data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	counter := len(availabilities)
	// difference between startDate and endDate in days
	//days := int(endDate.Sub(startDate).Hours()/24) + 1

	if counter == 0 {
		return false, nil
	}

	fmt.Print("Exists")

	return true, nil
}

func (s *AvailabilityServiceImpl) BookAccommodation(accommodationID primitive.ObjectID, startDate time.Time, endDate time.Time, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.BookAccommodation")
	defer span.End()

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

func (s *AvailabilityServiceImpl) GetAvailabilityByAccommodationId(accommodationID primitive.ObjectID, ctx context.Context) ([]*data.Availability, error) {
	ctx, span := s.Tracer.Start(ctx, "AvailabilityService.GetAvailabilityByAccommodationId")
	defer span.End()

	filter := bson.M{
		"accommodation_id": accommodationID,
	}
	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	var availabilities []*data.Availability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return availabilities, nil
}
