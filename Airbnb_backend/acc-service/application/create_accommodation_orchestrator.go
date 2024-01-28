package application

import (
	"acc-service/domain"
	"fmt"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/create_accommodation"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"log"
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
	fmt.Println("handle")
	if err != nil {
		return nil, fmt.Errorf("error subscribing to reply: %v", err)
	}
	return o, nil
}

func (o *CreateAccommodationOrchestrator) Start(accommodation *domain.AccommodationWithAvailability) error {
	var priceType create_accommodation.PriceType
	var availabilityType create_accommodation.AvailabilityType
	if accommodation.PriceType == "PerPerson" {
		priceType = "PerPerson"
	} else {
		priceType = "PerDay"
	}

	if accommodation.AvailabilityType == "Available" {
		availabilityType = "Available"
	} else if accommodation.AvailabilityType == "Unavailable" {
		availabilityType = "Unavailable"
	} else {
		availabilityType = "Booked"
	}

	accomm := create_accommodation.AccommodationWithAvailability{
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
		PriceType:        priceType,
		AvailabilityType: availabilityType,
	}
	event := &create_accommodation.CreateAccommodationCommand{
		Accommodation: accomm,
		Type:          create_accommodation.AddAccommodation,
	}

	return o.commandPublisher.Publish(event)
}

func (o *CreateAccommodationOrchestrator) handle(reply *create_accommodation.CreateAccommodationReply) {
	command := create_accommodation.CreateAccommodationCommand{Accommodation: reply.Accommodation}
	command.Type = o.nextCommandType(*reply)
	if command.Type != create_accommodation.UnknownCommand {
		_ = o.commandPublisher.Publish(command)
	}
}

func (o *CreateAccommodationOrchestrator) nextCommandType(reply create_accommodation.CreateAccommodationReply) create_accommodation.CreateAccommodationCommandType {
	switch reply.Type {

	case create_accommodation.AccommodationAdded:
		log.Println("ACC ADDED")
		return create_accommodation.AddAvailability
	case create_accommodation.AccommodationNotAdded:
		log.Println("ACC NOT ADDED")
		return create_accommodation.CancelAvailability
	case create_accommodation.AccommodationRolledBack:
		log.Println("ACC ROLLEDBACK")
		return create_accommodation.RollbackAvailability

	case create_accommodation.AvailabilityAdded:
		log.Println("AVAILABILITY ADDED")
		return create_accommodation.AddRecommendation
	case create_accommodation.AvailabilityNotAdded:
		log.Println("AVAILABILITY NOT ADDED")
		return create_accommodation.RollbackAccommodation
	case create_accommodation.AvailabilityRolledBack:
		log.Println("AVAILABILITY ROLLEDBACK")
		return create_accommodation.RollbackRecommendation

	case create_accommodation.RecommendationAdded:
		log.Println("RECOMMENDATION ADDED")
		return create_accommodation.UnknownCommand
	case create_accommodation.RecommendationNotAdded:
		log.Println("RECOMMENDATION NOT ADDED")
		return create_accommodation.RollbackAccommodation
	case create_accommodation.RecommendationRolledBack:
		log.Println("RECOMMENDATION ROLLEDBACK")
		return create_accommodation.UnknownCommand

	default:
		log.Println("UNKNOWN")
		return create_accommodation.UnknownCommand
	}
}
