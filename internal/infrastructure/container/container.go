package container

import (
	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/auth"
	"github.com/dk5761/go-serv/internal/domain/chat"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	AuthHandler *auth.AuthHandler
	ChatHandler *chat.ChatHandler
}

func NewContainer(
	db *pgxpool.Pool,
	mongoDB *mongo.Database,
	cacheClient *redis.Client,
	storageService storage.StorageService,
	config *configs.Config,
) *Container {
	// Initialize Repositories
	userRepo := auth.NewPostgresUserRepository(db)
	messageRepo := chat.NewMongoMessageRepository(mongoDB)

	// Initialize Services
	jwtService := auth.NewJWTService(config.JWT.SecretKey, config.JWT.TokenDuration)
	authService := auth.NewAuthService(userRepo, jwtService)
	chatService := chat.NewChatService(messageRepo, storageService)

	// Initialize Handlers
	authHandler := auth.NewAuthHandler(authService, jwtService, userRepo)
	chatHandler := chat.NewChatHandler(chatService)

	return &Container{
		AuthHandler: authHandler,
		ChatHandler: chatHandler,
	}
}
