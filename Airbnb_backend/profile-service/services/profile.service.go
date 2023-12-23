package services

import (
	"context"
	"profile-service/domain"
)

type ProfileService interface {
	Registration(user *domain.User, ctx context.Context) error
	DeleteUserProfile(email string, ctx context.Context) error
	FindUserByEmail(email string, ctx context.Context) error
	FindProfileByEmail(email string, ctx context.Context) (*domain.User, error)
	UpdateUser(user *domain.User, ctx context.Context) error
	SendUserToAuthService(user *domain.User, ctx context.Context) error
}
