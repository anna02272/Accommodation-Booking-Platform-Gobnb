package services

import (
	"auth-service/domain"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthService interface {
	Login(loginInput *domain.LoginInput, ctx context.Context) (*domain.User, error)
	Registration(rw http.ResponseWriter, user *domain.User, ctx context.Context) (*domain.UserResponse, error)
	//Registration(http.ResponseWriter, *domain.User) (*domain.UserResponse, error)
	ResendVerificationEmail(ctx *gin.Context)
	SendVerificationEmail(credentials *domain.Credentials, ctx context.Context) error
	SendPasswordResetToken(credentials *domain.Credentials, ctx context.Context) error
}
