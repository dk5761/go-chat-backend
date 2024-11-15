package container

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/auth"
	authHandler "github.com/dk5761/go-serv/internal/domain/auth/handler"
	userRepo "github.com/dk5761/go-serv/internal/domain/auth/repository"
	"github.com/dk5761/go-serv/internal/domain/chat"
	chatHandler "github.com/dk5761/go-serv/internal/domain/chat/handler"
	"github.com/dk5761/go-serv/internal/domain/chat/repository"
	"github.com/dk5761/go-serv/internal/domain/chat/websocket"
	notificationRepo "github.com/dk5761/go-serv/internal/domain/notifications/repository"
	"github.com/dk5761/go-serv/internal/domain/notifications/service"
	"github.com/dk5761/go-serv/internal/infrastructure/fcm"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
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

	chatRepo := repository.NewMongoMessageRepository(mongoDB)
	userRepo := userRepo.NewPostgresUserRepository(db)

	fcms, err := fcm.NewFCMService("")
	if err != nil {
		fmt.Println("failed fcm Initialize")
	}

	notificationRepo := notificationRepo.NewMongoNotificationRepository(mongoDB)
	notificationService := service.NewNotificationService(fcms, notificationRepo)

	wsManager := websocket.NewWebSocketManager(chatRepo, userRepo, notificationService)

	// Initialize Repositories
	authHandlerInit := auth.NewAuthHandler(db, config)
	chatHandlerInit := chat.NewChatHandler(mongoDB, config, wsManager)

	return &Container{
		AuthHandler: authHandlerInit,
		ChatHandler: chatHandlerInit,
	}
}
