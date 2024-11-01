package container

import (
	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/auth"
	authHandler "github.com/dk5761/go-serv/internal/domain/auth/handler"
	"github.com/dk5761/go-serv/internal/domain/chat"
	chatHandler "github.com/dk5761/go-serv/internal/domain/chat/handler"
	"github.com/dk5761/go-serv/internal/domain/chat/websocket"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
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
	authHandler := auth.NewAuthHandler(db, config)

	wsManager := websocket.NewWebSocketManager()
	chatHandler := chat.NewChatHandler(mongoDB, config, wsManager)

	return &Container{
		AuthHandler: authHandler,
		ChatHandler: chatHandler,
	}
}
