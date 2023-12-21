package services

import (
	"profile-service/domain"
)

type ProfileService interface {
	Registration(user *domain.User) error
	DeleteUserProfile(email string) error
	FindUserByEmail(email string) error
	FindProfileByEmail(email string) (*domain.User, error)
	UpdateUser(user *domain.User) error
	SendUserToAuthService(user *domain.User) error
}
