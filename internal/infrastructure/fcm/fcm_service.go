package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FCMService interface {
	SendToDevice(ctx context.Context, token string, title, body string, data map[string]string) error
	SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error
	SendToCondition(ctx context.Context, condition, title, body string, data map[string]string) error
	SendMulticast(ctx context.Context, tokens []string, data map[string]string) ([]string, error)
}

type fcmService struct {
	client *messaging.Client
}

// NewFCMService initializes the FCM service with the Firebase Admin SDK.
func NewFCMService(serviceAccountPath string) (FCMService, error) {
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase Messaging client: %w", err)
	}

	return &fcmService{client: client}, nil
}

func (s *fcmService) SendToDevice(ctx context.Context, token string, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	response, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send message to device: %w", err)
	}

	fmt.Printf("Successfully sent message: %s\n", response)
	return nil
}

func (s *fcmService) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	response, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send message to topic: %w", err)
	}

	fmt.Printf("Successfully sent message to topic: %s\n", response)
	return nil
}

func (s *fcmService) SendToCondition(ctx context.Context, condition, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Condition: condition,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	response, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send message with condition: %w", err)
	}

	fmt.Printf("Successfully sent message with condition: %s\n", response)
	return nil
}

func (s *fcmService) SendMulticast(ctx context.Context, tokens []string, data map[string]string) ([]string, error) {
	message := &messaging.MulticastMessage{
		Data:   data,
		Tokens: tokens,
	}

	response, err := s.client.SendMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send multicast message: %w", err)
	}

	failedTokens := []string{}
	for idx, resp := range response.Responses {
		if !resp.Success {
			failedTokens = append(failedTokens, tokens[idx])
		}
	}

	fmt.Printf("Multicast success: %d, failures: %d\n", response.SuccessCount, response.FailureCount)
	return failedTokens, nil
}
