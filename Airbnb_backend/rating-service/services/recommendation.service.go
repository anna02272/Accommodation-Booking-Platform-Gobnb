package services

import (
	"context"
	"rating-service/domain"
)

type RecommendationService interface {
	CheckConnection()
	CloseDriverConnection(ctx context.Context)
	CreateUser(user *domain.NeoUser) error
}
