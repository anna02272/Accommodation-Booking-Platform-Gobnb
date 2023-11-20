package services

import "profile-service/domain"

type ProfileService interface {
	Registration(user *domain.User) error
}
