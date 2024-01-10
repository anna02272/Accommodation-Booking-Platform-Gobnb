package application

import (
	"auth-service/common/create_user"
	"auth-service/common/saga"
	"auth-service/domain"
)

type CreateUserOrchestrator struct {
	commandPublisher saga.Publisher
	replySubscriber  saga.Subscriber
}

func NewCreateUserOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber) (*CreateUserOrchestrator, error) {
	o := &CreateUserOrchestrator{
		commandPublisher: publisher,
		replySubscriber:  subscriber,
	}
	err := o.replySubscriber.Subscribe(o.handle)
	if err != nil {
		return nil, err
	}
	return o, nil
}
func (o *CreateUserOrchestrator) Start(accommodation *domain.User) error {
	event := &create_user.CreateUserCommand{
		Type: create_user.AddUser,
		User: create_user.User{
			ID: accommodation.ID,
		},
	}

	return o.commandPublisher.Publish(event)
}

func (o *CreateUserOrchestrator) handle(reply *create_user.CreateUserReply) {
	command := create_user.CreateUserCommand{User: reply.User}
	command.Type = o.nextCommandType(reply.Type)
	if command.Type != create_user.UnknownCommand {
		_ = o.commandPublisher.Publish(command)
	}
}

func (o *CreateUserOrchestrator) nextCommandType(reply create_user.CreateUserReplyType) create_user.CreateUserCommandType {
	switch reply {
	case create_user.UserAdded:
		return create_user.AddProfile
	case create_user.UserNotAdded:
		return create_user.CancelProfile
	case create_user.ProfileAdded:
		return create_user.SendMail
	case create_user.ProfileNotAdded:
		return create_user.RollbackUser
	case create_user.MailSent:
		return create_user.UnknownCommand
	case create_user.MailFailed:
		return create_user.RollbackProfile
	default:
		return create_user.UnknownCommand
	}
}
