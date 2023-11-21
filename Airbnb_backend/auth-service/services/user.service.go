package services

import (
	"auth-service/domain"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	FindUserById(string) (*domain.User, error)
	FindUserByEmail(string) (*domain.User, error)
	FindCredentialsByEmail(string) (*domain.Credentials, error)
	FindUserByUsername(string) (*domain.User, error)
	SendUserToProfileService(user *domain.User) error
	FindUserByVerifCode(ctx *gin.Context) (*domain.Credentials, error)
	FindUserByResetPassCode(ctx *gin.Context) (*domain.Credentials, error)
}
