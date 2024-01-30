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
	reply := create_accommodation.CreateAccommodationReply{Accommodation: command.Accommodation}

	availability := data.AvailabilityPeriod{
		StartDate:        command.Accommodation.StartDate,
		EndDate:          command.Accommodation.EndDate,
		Price:            command.Accommodation.Price,
		PriceType:        data.PriceType(command.Accommodation.PriceType),
		AvailabilityType: data.AvailabilityType(command.Accommodation.AvailabilityType),
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
		if availability.StartDate != primitive.DateTime(0) &&
			availability.EndDate != primitive.DateTime(0) &&
			availability.Price != 0.0 {
			handler.availabilityService.InsertMulitipleAvailability(availability, objectID, context.Background())
			if err != nil {
				reply.Type = create_accommodation.AvailabilityNotAdded
			} else {
				reply.Type = create_accommodation.AvailabilityAdded
			}
		} else {
			reply.Type = create_accommodation.AvailabilityAdded
		}

	//case create_accommodation.RollbackAvailability:
	//	err = handler.availabilityService.DeleteAvailability(objectID, startDate, endDate, context.Background())
	//	if err != nil {
	//		return
	//	}
	//	reply.Type = create_accommodation.AvailabilityNotAdded

	case create_accommodation.RollbackAccommodation:
		err = handler.availabilityService.DeleteAvailability(objectID, startDate, endDate, context.Background())
		if err != nil {
			return
		}
		reply.Type = create_accommodation.UnknownReply

	default:
		reply.Type = create_accommodation.UnknownReply
	}

	if reply.Type != create_accommodation.UnknownReply {
		_ = handler.replyPublisher.Publish(reply)
	}

}
