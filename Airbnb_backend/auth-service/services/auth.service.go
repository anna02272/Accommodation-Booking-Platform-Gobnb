package services

import (
	"auth-service/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthService interface {
	Login(*domain.LoginInput) (*domain.User, error)
	Registration(http.ResponseWriter, *domain.User) (*domain.UserResponse, error)
	ResendVerificationEmail(ctx *gin.Context)
	SendVerificationEmail(credentials *domain.Credentials) error
	SendPasswordResetToken(credentials *domain.Credentials) error
}
