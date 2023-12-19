package services

import "profile-service/domain"

type ProfileService interface {
	Registration(user *domain.User) error
	DeleteUserProfile(email string) error
	FindUserByEmail(email string) error
	UpdateUser(user *domain.User) error
}
