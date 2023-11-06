package services

import "auth-service/domain"

type UserService interface {
	FindUserById(string) (*domain.User, error)
	FindUserByEmail(string) (*domain.User, error)
	FindUserByUsername(string) (*domain.User, error)
}
