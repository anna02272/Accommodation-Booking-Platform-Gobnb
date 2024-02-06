package handlers

import (
	"acc-service/domain"
	"acc-service/services"
	"context"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/create_accommodation"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
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

func (handler *CreateAccommodationCommandHandler) handle(command *create_accommodation.CreateAccommodationCommand) {
	id := command.Accommodation.ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return
	}

	accommodation := &domain.AccommodationWithAvailability{
		ID:               objectID,
		HostId:           command.Accommodation.HostId,
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
		_, _, err := handler.accommodationService.InsertAccommodation(accommodation, accommodation.HostId, context.Background())
		if err != nil {
			reply.Type = create_accommodation.AccommodationNotAdded
			return
		} else {
			log.Println(accommodation.StartDate, accommodation.EndDate, accommodation.Price)
			if accommodation.StartDate != primitive.DateTime(0) &&
				accommodation.EndDate != primitive.DateTime(0) &&
				accommodation.Price != 0.0 {
				reply.Type = create_accommodation.AccommodationAdded
			} else {

				reply.Type = create_accommodation.AvailabilityAdded
			}
		}

	case create_accommodation.RollbackAccommodation:
		err := handler.accommodationService.DeleteAccommodationForSaga(id, accommodation.HostId, context.Background())
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
