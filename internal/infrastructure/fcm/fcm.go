package fcm

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"

	"github.com/dk5761/go-serv/internal/domain/notifications/models"
)

type FCMService struct {
	client *messaging.Client
}

func NewFCMService(app *firebase.App) (*FCMService, error) {
	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, err
	}
	return &FCMService{client: client}, nil
}

func (s *FCMService) SendNotification(notification *models.Notification) error {
	msg := &messaging.Message{
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
		Data:  notification.Data,
		Token: notification.Token,
	}

	_, err := s.client.Send(context.Background(), msg)
	return err
}
