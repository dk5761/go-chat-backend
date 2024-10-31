package chat

import (
	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/chat/handler"
	"github.com/dk5761/go-serv/internal/domain/chat/repository"
	"github.com/dk5761/go-serv/internal/domain/chat/service"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewAuthHandler initializes and returns an AuthHandler with all dependencies injected.
func NewAuthHandler(db *mongo.Database, config *configs.Config) *handler.ChatHandler {
	// Initialize repository with the provided database connection
	chatRepo := repository.NewMongoMessageRepository(db)

	storageService := storage.NewS3StorageService(config.Storage.S3Config)

	// Initialize auth service with the repository and JWT service
	chatService := service.NewChatService(chatRepo, storageService)

	// Return a new handler with all dependencies set up
	return handler.NewChatHandler(chatService)
}
