package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/dk5761/go-serv/internal/domain/notifications/service"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
)

type NotificationWorker struct {
	notificationService *service.NotificationService
	cleanupInterval     time.Duration
	retryInterval       time.Duration
	wg                  sync.WaitGroup
	stopCh              chan struct{}
}

func NewNotificationWorker(
	service *service.NotificationService,
	cleanupInterval time.Duration,
	retryInterval time.Duration,
) *NotificationWorker {
	return &NotificationWorker{
		notificationService: service,
		cleanupInterval:     cleanupInterval,
		retryInterval:       retryInterval,
		stopCh:              make(chan struct{}),
	}
}

func (w *NotificationWorker) Start() {
	w.wg.Add(2)
	go w.runCleanupWorker()
	go w.runRetryWorker()
}

func (w *NotificationWorker) Stop(ctx context.Context) error {
	close(w.stopCh)
	done := make(chan struct{})

	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *NotificationWorker) runCleanupWorker() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.notificationService.DeleteOldNotifications(2); err != nil {
				logging.Logger.Error("Failed to cleanup notifications", zap.Error(err))
			}
		case <-w.stopCh:
			return
		}
	}
}

func (w *NotificationWorker) runRetryWorker() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.retryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.notificationService.RetryFailedNotifications(); err != nil {
				logging.Logger.Error("Failed to retry notifications", zap.Error(err))
			}
		case <-w.stopCh:
			return
		}
	}
}
