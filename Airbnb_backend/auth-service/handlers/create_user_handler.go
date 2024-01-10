package handlers

import (
	"auth-service/common/create_user"
	"auth-service/common/saga"
	"auth-service/domain"
	"auth-service/services"
	"context"
	"go.opentelemetry.io/otel/trace"
)

type CreateUserCommandHandler struct {
	authService       services.AuthService
	userService       services.UserService
	replyPublisher    saga.Publisher
	commandSubscriber saga.Subscriber
	tracer            trace.Tracer
}

func NewCreateUserCommandHandler(authService services.AuthService, userService services.UserService, publisher saga.Publisher, subscriber saga.Subscriber, tracer trace.Tracer) (*CreateUserCommandHandler, error) {
	o := &CreateUserCommandHandler{
		authService:       authService,
		userService:       userService,
		replyPublisher:    publisher,
		commandSubscriber: subscriber,
		tracer:            tracer,
	}

	err := o.commandSubscriber.Subscribe(o.handle)
	if err != nil {
		return nil, err
	}
	return o, nil
}
func (handler *CreateUserCommandHandler) handle(command *create_user.CreateUserCommand) {
	user := &domain.User{ID: command.User.ID}
	reply := create_user.CreateUserReply{User: command.User}

	switch command.Type {
	//case create_user.AddUser:
	//	err := handler.authService.Registration(user, context.Background())
	//	if err != nil {
	//		return
	//	}
	//	reply.Type = create_user.UserAdded
	//case create_user.SendMail:
	//	err := handler.authService.SendVerificationEmail(user, context.Background())
	//	if err != nil {
	//		return
	//	}
	//	reply.Type = create_user.MailSent
	case create_user.RollbackUser:
		_ = handler.userService.DeleteCredentials(user, context.Background())
		reply.Type = create_user.UserRolledBack
	default:
		reply.Type = create_user.UnknownReply
	}

	if reply.Type != create_user.UnknownReply {
		_ = handler.replyPublisher.Publish(reply)
	}
}
