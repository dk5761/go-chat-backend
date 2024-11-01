package chat

import (
	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/chat/handler"
	"github.com/dk5761/go-serv/internal/domain/chat/repository"
	"github.com/dk5761/go-serv/internal/domain/chat/service"
	"github.com/dk5761/go-serv/internal/domain/chat/websocket"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewChatHandler initializes and returns a ChatHandler with all dependencies injected.
func NewChatHandler(db *mongo.Database, config *configs.Config, wsManager *websocket.WebSocketManager) *handler.ChatHandler {
	// Initialize repository with the provided database connection
	chatRepo := repository.NewMongoMessageRepository(db)

	// Initialize storage service with S3 configuration
	storageService := storage.NewS3StorageService(config.Storage.S3Config)

	// Initialize chat service with the repository, storage, and WebSocket manager
	chatService := service.NewChatService(chatRepo, storageService, wsManager)

	// Return a new handler with all dependencies set up
	return handler.NewChatHandler(chatService, wsManager)
}
