package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	models "github.com/dk5761/go-serv/internal/domain/notifications/model"
	"github.com/dk5761/go-serv/internal/domain/notifications/repository"
	"github.com/dk5761/go-serv/internal/infrastructure/fcm"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
)

type NotificationService interface {
	SendNotification(ctx context.Context, notification *models.Notification, receiverId string) error
	SendMulticastNotification(ctx context.Context, notification *models.Notification, tokens []string) ([]string, error)
	GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
}

type notificationService struct {
	fcmService fcm.FCMService
	repo       repository.NotificationRepository
}

// NewNotificationService initializes the notification service with dependencies.
func NewNotificationService(fcmService fcm.FCMService, repo repository.NotificationRepository) NotificationService {
	return &notificationService{
		fcmService: fcmService,
		repo:       repo,
	}
}

func (s *notificationService) SendNotification(ctx context.Context, notification *models.Notification, receiverId string) error {

	if notification.DeviceToken == "" {
		return fmt.Errorf("no device token found for user")
	}

	// Store the notification in the database first
	if err := s.repo.SaveNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	// Send the notification using FCM
	err := s.fcmService.SendToDevice(ctx, notification.DeviceToken, notification.Title, notification.Body, notification.Data)
	if err != nil {

		logging.Logger.Error("Failed to send notification", zap.String("user_id", notification.UserID), zap.Error(err))
		return err
	}

	logging.Logger.Info("Notification sent successfully", zap.String("user_id", notification.UserID))
	return nil
}

func (s *notificationService) SendMulticastNotification(ctx context.Context, notification *models.Notification, tokens []string) ([]string, error) {
	// Send the notification using FCM multicast
	failedTokens, err := s.fcmService.SendMulticast(ctx, tokens, notification.Data)
	if err != nil {
		logging.Logger.Error("Failed to send multicast notification", zap.Error(err))
		return nil, err
	}

	logging.Logger.Info("Multicast notification sent successfully", zap.Int("failed_tokens_count", len(failedTokens)))
	return failedTokens, nil
}

func (s *notificationService) GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	// Retrieve notifications from the database
	notifications, err := s.repo.GetNotifications(ctx, userID, limit, offset)
	if err != nil {
		logging.Logger.Error("Failed to fetch notifications", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	return notifications, nil
}
