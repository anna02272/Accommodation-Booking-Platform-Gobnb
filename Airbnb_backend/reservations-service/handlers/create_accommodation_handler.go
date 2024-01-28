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
		log.Println("availability", availability)
		log.Println("availability", availability.StartDate, availability.EndDate, availability.Price, availability.PriceType, availability.AvailabilityType)
		if availability.StartDate != primitive.DateTime(0) &&
			availability.EndDate != primitive.DateTime(0) &&
			availability.Price != 0.0 &&
			availability.PriceType != "" &&
			availability.AvailabilityType != "" {
			handler.availabilityService.InsertMulitipleAvailability(availability, objectID, context.Background())
			//handler.availabilityService.InsertMulitipleAvailability(availability, primitive.ObjectID{}, context.Background())
			//invalidID := "invalid_object_id"
			//_, err := primitive.ObjectIDFromHex(invalidID)
			//if err != nil {
			//	log.Printf("Error converting ID to ObjectID: %v", err)
			//}
			if err != nil {
				log.Printf("Error inserting availability: %s, Error: %v", err)
				log.Println("create_accommodation.AvailabilityNotAdded:")
				reply.Type = create_accommodation.AvailabilityNotAdded
				break
			} else {
				log.Println("create_accommodation.AvailabilityAdded:")
				reply.Type = create_accommodation.AvailabilityAdded
			}
		} else {
			log.Println("create_accommodation.AvailabilityAdded:")
			reply.Type = create_accommodation.AvailabilityAdded
		}

	case create_accommodation.RollbackAvailability:
		log.Println("create_accommodation.RollbackAvailability:")
		err = handler.availabilityService.DeleteAvailability(objectIDZero, startDate, endDate, context.Background())
		if err != nil {
			return
		}
		err = handler.availabilityService.DeleteAvailability(objectID, startDate, endDate, context.Background())
		if err != nil {
			return
		}
		reply.Type = create_accommodation.AvailabilityRolledBack
	default:
		log.Printf("Unknown command type: %v", command.Type)
		reply.Type = create_accommodation.UnknownReply
	}

	if reply.Type != create_accommodation.UnknownReply {
		_ = handler.replyPublisher.Publish(reply)
	}

}
