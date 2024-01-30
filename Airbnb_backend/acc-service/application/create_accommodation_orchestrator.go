package application

import (
	"acc-service/domain"
	"context"
	"fmt"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/create_accommodation"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"
)

type CreateAccommodationOrchestrator struct {
	commandPublisher    saga.Publisher
	replySubscriber     saga.Subscriber
	tracer              trace.Tracer
	accommodationAdded  bool
	availabilityAdded   bool
	recommendationAdded bool
}

func NewCreateAccommodationOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber, tracer trace.Tracer) (*CreateAccommodationOrchestrator, error) {
	o := &CreateAccommodationOrchestrator{
		commandPublisher: publisher,
		replySubscriber:  subscriber,
		tracer:           tracer,
	}
	err := o.replySubscriber.Subscribe(o.handle)
	fmt.Println("handle")
	if err != nil {
		return nil, fmt.Errorf("error subscribing to reply: %v", err)
	}
	return o, nil
}

func (o *CreateAccommodationOrchestrator) Start(ctx context.Context, accommodation *domain.AccommodationWithAvailability) error {
	ctx, span := o.tracer.Start(ctx, "CreateAccommodationOrchestrator.Start")
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
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
	o.accommodationAdded = false
	o.availabilityAdded = false
	o.recommendationAdded = false
	if err := o.commandPublisher.Publish(event); err != nil {
		log.Println("Error publishing AddAccommodation command:", err)
	}

	select {
	case <-ctx.Done():
		if o.shouldRollback() {
			log.Println("shouldRollback")
			rollbackCommand := &create_accommodation.CreateAccommodationCommand{
				Accommodation: accomm,
				Type:          create_accommodation.RollbackAccommodation,
			}
			if err := o.commandPublisher.Publish(rollbackCommand); err != nil {
				log.Println("Error publishing RollbackRecommendation command:", err)
			}
		}
	}

	return nil
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
		o.accommodationAdded = true
		log.Println("ACC ADDED")
		return create_accommodation.AddAvailability
	case create_accommodation.AvailabilityAdded:
		o.availabilityAdded = true
		log.Println("AVAILABILITY ADDED")
		return create_accommodation.AddRecommendation
	case create_accommodation.RecommendationAdded:
		o.recommendationAdded = true
		log.Println("RECOMMENDATION ADDED")
		return create_accommodation.UnknownCommand

	case create_accommodation.RecommendationNotAdded:
		log.Println("RECOMMENDATION NOT ADDED")
		return create_accommodation.RollbackAccommodation
	case create_accommodation.AvailabilityNotAdded:
		log.Println("AVAILABILITY NOT ADDED")
		return create_accommodation.RollbackAccommodation

	default:
		log.Println("UNKNOWN")
		return create_accommodation.UnknownCommand
	}
}
func (o *CreateAccommodationOrchestrator) shouldRollback() bool {
	return !o.accommodationAdded || !o.availabilityAdded || !o.recommendationAdded
}
