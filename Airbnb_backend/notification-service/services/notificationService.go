package services

import (
	"context"
	"notification-service/domain"
)

type NotificationService interface {
	InsertNotification(notif *domain.NotificationCreate, ctx context.Context) (*domain.Notification, string, error)
	SendNotificationEmail(text string, subject string, email string, ctx context.Context) error
	GetNotificationsByHostId(hostId string, ctx context.Context) ([]*domain.Notification, error)
}
