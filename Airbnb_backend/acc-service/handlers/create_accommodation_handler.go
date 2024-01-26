package handlers

import (
	"acc-service/domain"
	"acc-service/services"
	"context"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/create_accommodation"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

type CreateAccommodationCommandHandler struct {
	accommodationService services.AccommodationService
	replyPublisher       saga.Publisher
	commandSubscriber    saga.Subscriber
}

func NewCreateAccommodationCommandHandler(accommodationService services.AccommodationService,
	publisher saga.Publisher, subscriber saga.Subscriber) (*CreateAccommodationCommandHandler, error) {
	o := &CreateAccommodationCommandHandler{
		accommodationService: accommodationService,
		replyPublisher:       publisher,
		commandSubscriber:    subscriber,
	}
	err := o.commandSubscriber.Subscribe(o.handle)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (handler *CreateAccommodationCommandHandler) handle(rw http.ResponseWriter, command *create_accommodation.CreateAccommodationCommand, hostID string, ctx context.Context, token string) {
	log.Printf("CreateAccommodationCommandHandler handle method started for ID: %s", command.Accommodation.ID)

	id := command.Accommodation.ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Error converting ObjectID for ID: %s, Error: %v", id, err)
		return
	}
	accommodation := &domain.AccommodationWithAvailability{
		ID:               objectID,
		HostId:           hostID,
		Name:             command.Accommodation.Name,
		Location:         command.Accommodation.Location,
		Amenities:        command.Accommodation.Amenities,
		MinGuests:        command.Accommodation.MinGuests,
		MaxGuests:        command.Accommodation.MaxGuests,
		Active:           command.Accommodation.Active,
		StartDate:        command.Accommodation.StartDate,
		EndDate:          command.Accommodation.EndDate,
		Price:            command.Accommodation.Price,
		PriceType:        domain.PriceType(command.Accommodation.PriceType),
		AvailabilityType: domain.AvailabilityType(command.Accommodation.AvailabilityType),
	}

	reply := create_accommodation.CreateAccommodationReply{Accommodation: command.Accommodation}

	switch command.Type {
	case create_accommodation.AddAccommodation:
		err, _, _ := handler.accommodationService.InsertAccommodation(rw, accommodation, accommodation.HostId, ctx, token)
		if err != nil {
			log.Printf("Error inserting accommodation for ID: %s, Error: %v", id, err)
			reply.Type = create_accommodation.AccommodationNotAdded
			return
		}
		reply.Type = create_accommodation.AccommodationAdded

	case create_accommodation.RollbackAccommodation:
		err := handler.accommodationService.DeleteAccommodation(id, accommodation.HostId, ctx)
		if err != nil {
			log.Printf("Error deleting accommodation for ID: %s, Error: %v", id, err)
			return
		}
		reply.Type = create_accommodation.AccommodationNotAdded
	default:
		reply.Type = create_accommodation.UnknownReply
	}

	if reply.Type != create_accommodation.UnknownReply {
		_ = handler.replyPublisher.Publish(reply)

	}

	log.Printf("CreateAccommodationCommandHandler handle method completed for ID: %s", command.Accommodation.ID)
}
