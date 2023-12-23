package services

import "notification-service/domain"

type NotificationService interface {
	InsertNotification(notif *domain.NotificationCreate) (*domain.Notification, string, error)
	SendNotificationEmail(text string, subject string, email string) error
	GetNotificationsByHostId(hostId string) ([]*domain.Notification, error)
}
