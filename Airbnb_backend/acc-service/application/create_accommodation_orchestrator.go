package application

import (
	"acc-service/domain"
	"context"
	"fmt"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/create_accommodation"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"log"
	"net/http"
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
		return nil, fmt.Errorf("error subscribing to reply: %v", err)
	}
	return o, nil
}

func (o *CreateAccommodationOrchestrator) Start(rw http.ResponseWriter, accommodation *domain.AccommodationWithAvailability, ctx context.Context, token string) error {
	fmt.Println("ORCHESTRATOR STARTED INSIDE")
	event := &create_accommodation.CreateAccommodationCommand{
		Accommodation: create_accommodation.AccommodationWithAvailability{
			ID:               accommodation.ID.Hex(),
			HostId:           accommodation.HostId,
			Name:             accommodation.Name,
			Location:         accommodation.Location,
			Amenities:        accommodation.Amenities,
			MinGuests:        accommodation.MinGuests,
			MaxGuests:        accommodation.MaxGuests,
			Active:           accommodation.Active,
			StartDate:        accommodation.StartDate,
			EndDate:          accommodation.EndDate,
			Price:            accommodation.Price,
			PriceType:        create_accommodation.PriceType(accommodation.PriceType),
			AvailabilityType: create_accommodation.AvailabilityType(accommodation.AvailabilityType),
		},
		Type: create_accommodation.AddAccommodation,
	}
	return o.commandPublisher.Publish(event)
}

func (o *CreateAccommodationOrchestrator) handle(reply *create_accommodation.CreateAccommodationReply) {
	fmt.Println("ORCHESTRATOR HANDLE")
	command := create_accommodation.CreateAccommodationCommand{Accommodation: reply.Accommodation}
	command.Type = o.nextCommandType(reply.Type)
	if command.Type != create_accommodation.UnknownCommand {
		_ = o.commandPublisher.Publish(command)
	}
}

func (o *CreateAccommodationOrchestrator) nextCommandType(reply create_accommodation.CreateAccommodationReplyType) create_accommodation.CreateAccommodationCommandType {
	fmt.Println("ORCHESTRATOR nextCommandType")
	switch reply {
	case create_accommodation.AccommodationAdded:
		log.Println("ACC ADDED")
		return create_accommodation.AddAvailability
	case create_accommodation.AccommodationNotAdded:
		log.Println("ACC NOT ADDED")
		return create_accommodation.CancelAvailability
	//case create_accommodation.AccommodationRolledBack:
	//	log.Println("ACC ROLLEDBACK")
	//	return create_accommodation.CancelAvailability
	case create_accommodation.AvailabilityNotAdded:
		log.Println("ACC NOT ADDED")
		return create_accommodation.RollbackAccommodation
	default:
		log.Println("UNKNOWN")
		return create_accommodation.UnknownCommand
	}
}
