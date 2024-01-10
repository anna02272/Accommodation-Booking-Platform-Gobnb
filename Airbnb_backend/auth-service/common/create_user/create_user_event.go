package create_user

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID
	Username string
	Password string
	Email    string
	Name     string
	Lastname string
	Address  Address
	Age      int
	Gender   Gender
	UserRole UserRole
}

type Address struct {
	Street  string
	City    string
	Country string
}

type Gender string

const (
	Male   = "Male"
	Female = "Female"
	Other  = "Other"
)

type UserRole string

const (
	Guest = "Guest"
	Host  = "Host"
)

type CreateUserCommandType int8

const (
	AddUser CreateUserCommandType = iota
	RollbackUser
	AddProfile
	CancelProfile
	RollbackProfile
	SendMail
	CancelMail
	UnknownCommand
)

type CreateUserCommand struct {
	User User
	Type CreateUserCommandType
}

type CreateUserReplyType int8

const (
	UserAdded CreateUserReplyType = iota
	UserNotAdded
	UserRolledBack
	ProfileAdded
	ProfileNotAdded
	MailSent
	MailFailed
	UnknownReply
)

type CreateUserReply struct {
	User User
	Type CreateUserReplyType
}
