package services

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"notification-service/config"
	"notification-service/domain"
	"notification-service/utils"
	"time"
)

type NotificationServiceImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewNotificationServiceImpl(collection *mongo.Collection, ctx context.Context) NotificationService {
	return &NotificationServiceImpl{collection, ctx}
}

func (s *NotificationServiceImpl) InsertNotification(notif *domain.NotificationCreate) (*domain.Notification, string, error) {
	var notification domain.Notification
	notification.HostId = notif.HostId
	notification.HostEmail = notif.HostEmail
	notification.DateAndTime = primitive.NewDateTimeFromTime(time.Now())
	notification.Text = notif.Text
	notification.ID = primitive.NewObjectID()

	result, err := s.collection.InsertOne(context.Background(), notification)
	if err != nil {
		return nil, "", err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, "", errors.New("failed to get inserted ID")
	}
	if err := s.SendNotificationEmail(notification.Text, "New notification", notification.HostEmail); err != nil {
		log.Printf("Error sending verification email: %v", err)
		return nil, "", err
	}

	insertedID = result.InsertedID.(primitive.ObjectID)

	return &notification, insertedID.Hex(), nil
}

func (s *NotificationServiceImpl) SendNotificationEmail(text string, subject string, email string) error {

	emailData := utils.EmailData{
		Subject: subject,
		Text:    text,
		Email:   email,
	}
	config := config.LoadConfig()
	return utils.SendEmail(&emailData, config)
}

func (s *NotificationServiceImpl) GetNotificationsByHostId(hostId string) ([]*domain.Notification, error) {
	filter := bson.M{"host_id": hostId}
	options := options.Find().SetSort(bson.D{{"date_and_time", 1}}) // 1 for ascending order

	cursor, err := s.collection.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var notifications []*domain.Notification
	for cursor.Next(context.Background()) {
		var notif domain.Notification
		if err := cursor.Decode(&notif); err != nil {
			return nil, err
		}
		notifications = append(notifications, &notif)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return notifications, nil
}
