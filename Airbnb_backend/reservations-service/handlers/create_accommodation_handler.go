package handlers

import (
	"context"
	"fmt"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/create_accommodation"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"reservations-service/data"
	"reservations-service/services"
	"time"
)

type CreateAccommodationCommandHandler struct {
	availabilityService services.AvailabilityService
	replyPublisher      saga.Publisher
	commandSubscriber   saga.Subscriber
}

func NewCreateAccommodationCommandHandler(availabilityService services.AvailabilityService,
	publisher saga.Publisher, subscriber saga.Subscriber) (*CreateAccommodationCommandHandler, error) {
	o := &CreateAccommodationCommandHandler{
		availabilityService: availabilityService,
		replyPublisher:      publisher,
		commandSubscriber:   subscriber,
	}
	err := o.commandSubscriber.Subscribe(o.handle)
	if err != nil {
		log.Printf("Error subscribing to the command: %v", err)
		return nil, err
	}
	return o, nil
}
func (handler *CreateAccommodationCommandHandler) handle(command *create_accommodation.CreateAccommodationCommand) {
	log.Println("CreateAccommodationCommandHandler handle method started")
	reply := create_accommodation.CreateAccommodationReply{Accommodation: command.Accommodation}

	//objectID, err := primitive.ObjectIDFromHex(command.Accommodation.ID)
	//if err != nil {
	//	return
	//}
	invalidID := "invalid_object_id"
	availability := data.AvailabilityPeriod{
		StartDate:        command.Accommodation.StartDate,
		EndDate:          command.Accommodation.EndDate,
		Price:            command.Accommodation.Price,
		PriceType:        data.PriceType(command.Accommodation.PriceType),
		AvailabilityType: data.AvailabilityType(command.Accommodation.AvailabilityType),
	}

	idZero := "000000000000000000000000"
	objectIDZero, err := primitive.ObjectIDFromHex(idZero)
	if err != nil {
		fmt.Printf("Error converting ID to ObjectID: %v\n", err)
		return
	}
	id := command.Accommodation.ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Printf("Error converting ID to ObjectID: %v\n", err)
		return
	}
	startDate := time.Unix(int64(availability.StartDate)/1000, 0)
	endDate := time.Unix(int64(availability.EndDate)/1000, 0)

	switch command.Type {
	case create_accommodation.AddAvailability:
		log.Println("create_accommodation.AddAvailability:")

		handler.availabilityService.InsertMulitipleAvailability(availability, primitive.ObjectID{}, context.Background())
		_, err := primitive.ObjectIDFromHex(invalidID)
		if err != nil {
			log.Printf("Error converting ID to ObjectID: %v", err)
		}
		if err != nil {
			log.Printf("Error inserting availability: %s, Error: %v", err)
			log.Println("create_accommodation.AvailabilityNotAdded:")
			reply.Type = create_accommodation.AvailabilityNotAdded
			break
		}
		reply.Type = create_accommodation.AvailabilityAdded

	case create_accommodation.CancelAvailability:
		err = handler.availabilityService.DeleteAvailability(objectIDZero, startDate, endDate, context.Background())
		if err != nil {
			return
		}
		err = handler.availabilityService.DeleteAvailability(objectID, startDate, endDate, context.Background())
		if err != nil {
			return
		}

	default:
		log.Printf("Unknown command type: %v", command.Type)
		reply.Type = create_accommodation.UnknownReply
	}

	if reply.Type != create_accommodation.UnknownReply {
		_ = handler.replyPublisher.Publish(reply)
	}

	log.Println("CreateAccommodationCommandHandler handle method completed")
}
