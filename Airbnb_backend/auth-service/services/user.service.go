package services

import (
	"auth-service/domain"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserService interface {
	FindUserById(string) (*domain.User, error)
	FindUserByEmail(string) (*domain.User, error)
	FindUserByUsername(string) (*domain.User, error)
	FindCredentialsByEmail(string) (*domain.Credentials, error)
	SendUserToProfileService(rw http.ResponseWriter, user *domain.User) error
	FindUserByVerifCode(ctx *gin.Context) (*domain.Credentials, error)
	FindUserByResetPassCode(ctx *gin.Context) (*domain.Credentials, error)
	UpdateUser(user *domain.User) error
	DeleteCredentials(user *domain.User) error
	HTTPSperformAuthorizationRequestWithContext(ctx context.Context, user *domain.User, url string) (*http.Response, error)
}
