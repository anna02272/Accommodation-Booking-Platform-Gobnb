package handlers

import (
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/create_accommodation"
	"github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024-common/common/saga"
	"log"
	"rating-service/domain"
	"rating-service/services"
)

type CreateAccommodationCommandHandler struct {
	recommService     services.RecommendationService
	replyPublisher    saga.Publisher
	commandSubscriber saga.Subscriber
}

func NewCreateAccommodationCommandHandler(recommService services.RecommendationService,
	publisher saga.Publisher, subscriber saga.Subscriber) (*CreateAccommodationCommandHandler, error) {
	o := &CreateAccommodationCommandHandler{
		recommService:     recommService,
		replyPublisher:    publisher,
		commandSubscriber: subscriber,
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

	accommodation := domain.AccommodationRec{
		ID:        command.Accommodation.ID,
		HostId:    command.Accommodation.HostId,
		Name:      command.Accommodation.Name,
		Location:  command.Accommodation.Location,
		MinGuests: command.Accommodation.MinGuests,
		MaxGuests: command.Accommodation.MaxGuests,
		Active:    command.Accommodation.Active,
	}

	switch command.Type {

	case create_accommodation.AddRecommendation:
		if err := handler.recommService.CreateAccommodation(&accommodation); err != nil {
			reply.Type = create_accommodation.RecommendationNotAdded
		} else {
			reply.Type = create_accommodation.RecommendationAdded
		}

	case create_accommodation.RollbackAccommodation:
		if err := handler.recommService.DeleteAccommodation(accommodation.ID); err != nil {
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
