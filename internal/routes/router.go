package routes

import (
	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/auth"
	"github.com/dk5761/go-serv/internal/domain/chat"
	"github.com/dk5761/go-serv/internal/infrastructure/middlewares"
	"github.com/dk5761/go-serv/internal/infrastructure/storage"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitRoutes(
	router *gin.Engine,
	db *pgxpool.Pool,
	mongoDB *mongo.Database,
	cacheClient *redis.Client,
	storageService storage.StorageService,
	config *configs.Config,
) {
	// Initialize repositories
	userRepo := auth.NewPostgresUserRepository(db)
	messageRepo := chat.NewMongoMessageRepository(mongoDB)

	// Initialize services
	jwtService := auth.NewJWTService(config.JWT.SecretKey, config.JWT.TokenDuration)
	authService := auth.NewAuthService(userRepo, jwtService)
	chatService := chat.NewChatService(messageRepo, storageService)

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService)
	chatHandler := chat.NewChatHandler(chatService)

	// Apply middlewares
	router.Use(middlewares.ErrorHandler())

	// Public routes
	public := router.Group("/api")
	{
		public.POST("/signup", authHandler.SignUp)
		public.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middlewares.JWTAuthMiddleware(jwtService, userRepo))
	{
		protected.GET("/profile", authHandler.Profile)
		protected.GET("/chat/ws", chatHandler.HandleWebSocket)
		protected.POST("/chat/send", chatHandler.SendMessage)
	}
}
