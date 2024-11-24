package container

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/auth"
	authHandler "github.com/dk5761/go-serv/internal/domain/auth/handler"
	"github.com/dk5761/go-serv/internal/domain/chat"
	chatHandler "github.com/dk5761/go-serv/internal/domain/chat/handler"
	"github.com/dk5761/go-serv/internal/domain/chat/repository"
	"github.com/dk5761/go-serv/internal/domain/chat/websocket"
	nRepo "github.com/dk5761/go-serv/internal/domain/notifications/repository"
	"github.com/dk5761/go-serv/internal/domain/notifications/service"
	"github.com/dk5761/go-serv/internal/infrastructure/fcm"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/dk5761/go-serv/pkg/worker"
)

type Container struct {
	AuthHandler *authHandler.AuthHandler
	ChatHandler *chatHandler.ChatHandler
}

func NewContainer(
	db *pgxpool.Pool,
	mongoDB *mongo.Database,
	cacheClient *redis.Client,
	storageService storage.StorageService,
	config *configs.Config,
) *Container {

	// Initialize Repositories
	chatRepo := repository.NewMongoMessageRepository(mongoDB)
	notificationRepo := nRepo.NewMongoMessageRepository(mongoDB)
	authHandlerInit := auth.NewAuthHandler(db, config)

	wsManager := websocket.NewWebSocketManager(chatRepo)

	chatHandlerInit := chat.NewChatHandler(mongoDB, config, wsManager)

	// Initialize FCM Config
	fcmConfig := &fcm.FCMConfig{
		CredentialsFile: "/home/dk/D/Code/go-chat-backend/configs/go-chat-pk.json",
	}

	// Initialize FCM Service
	fcmService, err := fcm.InitializeFCM(fcmConfig)
	if err != nil {
		logging.Logger.Fatal("Failed to initialize FCM service", zap.Error(err))
	}

	// Initialize Notification Service
	notificationService := service.NewNotificationService(
		notificationRepo,
		fcmService,
	)

	// Initialize Notification Worker
	worker := worker.NewNotificationWorker(
		notificationService, 2*time.Hour, 5*time.Minute,
	)

	// Start the worker
	worker.Start()

	// Graceful shutdown
	if err := worker.Stop(context.Background()); err != nil {
		logging.Logger.Error("Failed to stop notification worker", zap.Error(err))
	}

	return &Container{
		AuthHandler: authHandlerInit,
		ChatHandler: chatHandlerInit,
	}
}
