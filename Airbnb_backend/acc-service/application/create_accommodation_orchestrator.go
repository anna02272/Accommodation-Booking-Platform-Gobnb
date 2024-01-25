package application

import (
	"acc-service/common/create_accommodation"
	"acc-service/common/saga"
	"acc-service/domain"
)

type CreateAccommodationOrchestrator struct {
	commandPublisher saga.Publisher
	replySubscriber  saga.Subscriber
}

func NewCreateAccommodationOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber) (*CreateAccommodationOrchestrator, error) {
	o := &CreateAccommodationOrchestrator{
		commandPublisher: publisher,
		replySubscriber:  subscriber,
	}
	err := o.replySubscriber.Subscribe(o.handle)
	if err != nil {
		return nil, err
	}
	return o, nil
}
func (o *CreateAccommodationOrchestrator) Start(accommodation *domain.AccommodationWithAvailability) error {
	event := &create_accommodation.CreateAccommodationCommand{
		Type: create_accommodation.AddAccommodation,
		Accommodation: create_accommodation.AccommodationWithAvailability{
			ID: accommodation.ID,
		},
	}

	return o.commandPublisher.Publish(event)
}

func (o *CreateAccommodationOrchestrator) handle(reply *create_accommodation.CreateAccommodationReply) {
	command := create_accommodation.CreateAccommodationCommand{Accommodation: reply.Accommodation}
	command.Type = o.nextCommandType(reply.Type)
	if command.Type != create_accommodation.UnknownCommand {
		_ = o.commandPublisher.Publish(command)
	}
}

//} Tu mas dodat

func (o *CreateAccommodationOrchestrator) nextCommandType(reply create_accommodation.CreateAccommodationReplyType) create_accommodation.CreateAccommodationCommandType {
	switch reply {
	case create_accommodation.AccommodationAdded:
		return create_accommodation.AddAvailability
	case create_accommodation.AccommodationNotAdded:
		return create_accommodation.CancelAvailability
	case create_accommodation.AccommodationRolledBack:
		return create_accommodation.CancelAvailability
	case create_accommodation.AvailabilityNotAdded:
		return create_accommodation.RollbackAccommodation
	default:
		return create_accommodation.UnknownCommand
	}
}
