package handlers

import (
	"context"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/create_accommodation"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"reservations-service/data"
	"reservations-service/services"
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

func (handler *CreateAccommodationCommandHandler) handle(rw http.ResponseWriter, command *create_accommodation.CreateAccommodationCommand, hostID string, ctx context.Context, token string) {
	log.Println("CreateAccommodationCommandHandler handle method started")

	reply := create_accommodation.CreateAccommodationReply{Accommodation: command.Accommodation}

	switch command.Type {
	case create_accommodation.AddAvailability:
		accommodationId := command.Accommodation.ID
		objectID, err := primitive.ObjectIDFromHex(accommodationId)
		if err != nil {
			log.Printf("Error converting ObjectID for ID: %s, Error: %v", accommodationId, err)
			return
		}
		startDate := command.Accommodation.StartDate
		endDate := command.Accommodation.EndDate
		price := command.Accommodation.Price
		priceType := command.Accommodation.PriceType
		availabilityType := command.Accommodation.AvailabilityType
		availability := data.AvailabilityPeriod{
			StartDate:        startDate,
			EndDate:          endDate,
			Price:            price,
			PriceType:        data.PriceType(priceType),
			AvailabilityType: data.AvailabilityType(availabilityType),
		}
		handler.availabilityService.InsertMulitipleAvailability(availability, objectID, context.Background())
		if err != nil {
			log.Printf("Error inserting availability for ID: %s, Error: %v", accommodationId, err)
			reply.Type = create_accommodation.AvailabilityNotAdded
			break
		}
		reply.Type = create_accommodation.AvailabilityAdded
	//case create_accommodation.CancelAvailability:
	//	log.Println("Handling RollbackAccommodation")
	//	reply.Type = create_accommodation.AccommodationRolledBack
	default:
		log.Printf("Unknown command type: %v", command.Type)
		reply.Type = create_accommodation.UnknownReply
	}

	if reply.Type != create_accommodation.UnknownReply {
		_ = handler.replyPublisher.Publish(reply)
	}

	log.Println("CreateAccommodationCommandHandler handle method completed")
}
