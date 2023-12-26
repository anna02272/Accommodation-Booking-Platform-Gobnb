package services

import (
	"auth-service/domain"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserService interface {
	FindUserById(id string, ctx context.Context) (*domain.User, error)
	FindUserByEmail(email string, ctx context.Context) (*domain.User, error)
	FindUserByUsername(username string) (*domain.User, error)
	FindCredentialsByEmail(email string, ctx context.Context) (*domain.Credentials, error)
	SendUserToProfileService(rw http.ResponseWriter, user *domain.User, ctx context.Context) error
	FindUserByVerifCode(ctx *gin.Context, ctxt context.Context) (*domain.Credentials, error)
	FindUserByResetPassCode(ctx *gin.Context, ctxt context.Context) (*domain.Credentials, error)
	UpdateUser(user *domain.User, ctx context.Context) error
	DeleteCredentials(user *domain.User, ctx context.Context) error
	FindProfileInfoByEmail(ctx context.Context, email string) (*domain.CurrentUser, error)
	HTTPSperformAuthorizationRequestWithContext(ctx context.Context, user *domain.User, url string) (*http.Response, error)
}
