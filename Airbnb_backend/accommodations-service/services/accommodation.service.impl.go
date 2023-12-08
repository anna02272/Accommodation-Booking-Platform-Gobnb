package services

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"accomodations-service/domain"
)

type AccommodationServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewAccommodationServiceImpl(collection *mongo.Collection) *AccommodationServiceImpl {
	return &AccommodationServiceImpl{collection: collection}
}

func (uc *AccommodationServiceImpl) InsertAccommodation(accommodation *domain.Accommodation, hostId string) (*domain.Accommodation, error) {

	newAccommodation := *accommodation

	// Treba koristiti uc.collection umesto uc.ctx
	res, err := uc.collection.InsertOne(uc.ctx, &newAccommodation)
	if err != nil {
		return nil, err
	}

	var responseAccommodation *domain.Accommodation
	query := bson.M{"_id": res.InsertedID}


	err = uc.collection.FindOne(uc.ctx, query).Decode(&responseAccommodation)
	if err != nil {
		return nil, err
	}

	return responseAccommodation, nil
}

func (us *AccommodationServiceImpl) GetAccommodations(id string) (*domain.Accommodation, error) {
	var acc *domain.Accommodation

	// Improved email format validation using regular expression
	// emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	// if !emailRegex.MatchString(email) {
	// 	return errors.New("Invalid email format")
	// }

	query := bson.M{"_id": id}
	err := us.collection.FindOne(us.ctx, query).Decode(&acc)

		//find accommodation from mongo database and put it in acc


	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("Accommodation does not exist") // No accommodations found, return nil user and nil error
		}
		return nil, err
	}

	return acc, nil
}

func (us *AccommodationServiceImpl) GetAllAccommodations() (*domain.Accommodations, error) {
	//find all accommodations from mongo database
	cursor, err := us.collection.Find(us.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var accommodations *domain.Accommodations
	if err = cursor.All(us.ctx, &accommodations); err != nil {
		return nil, err
	}

	return accommodations, nil
}