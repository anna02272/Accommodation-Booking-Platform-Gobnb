package handlers

import (
	"acc-service/common/create_accommodation"
	"acc-service/common/saga"
	"acc-service/domain"
	"acc-service/services"
	"context"
)

type CreateAccommodationCommandHandler struct {
	accommodationService services.AccommodationService
	replyPublisher       saga.Publisher
	commandSubscriber    saga.Subscriber
}

func NewCreateAccommodationCommandHandler(accommodationService services.AccommodationService, publisher saga.Publisher, subscriber saga.Subscriber) (*CreateOrderCommandHandler, error) {
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
	accommodation := &domain.Accommodation{ID: id}

	reply := create_accommodation.CreateAccommodationReply{Accommodation: command.Accommodation}

	switch command.Type {
	case create_accommodation.AddAvailability:
		err := handler.accommodationService.InsertAccommodation(accommodation, context.Background())
		if err != nil {
			return
		}
		reply.Type = create_accommodation.AccommodationAdded
	case create_accommodation.RollbackAccommodation:
		err := handler.accommodationService.DeleteAccommodation(accommodation.ID, accommodation.HostId, context.Background())
		if err != nil {
			return
		}
		reply.Type = create_accommodation.AccommodationNotAdded
	default:
		reply.Type = create_accommodation.UnknownReply
	}

	if reply.Type != create_accommodation.UnknownReply {
		_ = handler.replyPublisher.Publish(reply)
	}
}
