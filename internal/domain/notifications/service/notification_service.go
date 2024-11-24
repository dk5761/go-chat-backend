package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/dk5761/go-serv/internal/domain/notifications/models"
	"github.com/dk5761/go-serv/internal/domain/notifications/repository"
	"github.com/dk5761/go-serv/internal/infrastructure/fcm"
)

type NotificationService struct {
	repo repository.NotificationRepository
	fcm  *fcm.FCMService
}

func NewNotificationService(repo repository.NotificationRepository, fcm *fcm.FCMService) *NotificationService {
	return &NotificationService{
		repo: repo,
		fcm:  fcm,
	}
}

func (s *NotificationService) SendNotification(token, title, body string, data map[string]string) error {
	notification := &models.Notification{
		ID:        uuid.New().String(),
		Token:     token,
		Title:     title,
		Body:      body,
		Data:      data,
		CreatedAt: time.Now(),
		Status:    models.StatusPending,
	}

	err := s.fcm.SendNotification(notification)
	if err != nil {
		notification.Status = models.StatusFailed
		return s.repo.Save(notification)
	}

	notification.Status = models.StatusSent
	return s.repo.Save(notification)
}

func (s *NotificationService) RetryFailedNotifications() error {
	notifications, err := s.repo.FindPendingNotifications()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if notification.RetryCount >= 3 {
			notification.Status = models.StatusFailed
			s.repo.Save(notification)
			continue
		}

		err := s.fcm.SendNotification(notification)
		if err != nil {
			notification.RetryCount++
			s.repo.Save(notification)
			continue
		}

		notification.Status = models.StatusSent
		s.repo.Save(notification)
	}
	return nil
}

func (s *NotificationService) DeleteOldNotifications(olderThan time.Duration) error {
	return s.repo.DeleteOldNotifications(int(olderThan.Seconds()))
}
