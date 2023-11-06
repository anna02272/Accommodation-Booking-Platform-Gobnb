package services

import (
	"auth-service/domain"
)

type AuthService interface {
	Login(*domain.LoginInput) (*domain.User, error)
	Registration(*domain.User) (*domain.UserResponse, error)
}
