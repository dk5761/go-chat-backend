package repository

import (
	"context"

	models "github.com/dk5761/go-serv/internal/domain/notifications/model"
)

type NotificationRepository interface {
	SaveNotification(ctx context.Context, notification *models.Notification) error
	GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
}
