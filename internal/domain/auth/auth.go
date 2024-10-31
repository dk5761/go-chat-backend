package auth

import (
	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/auth/handler"
	"github.com/dk5761/go-serv/internal/domain/auth/repository"
	"github.com/dk5761/go-serv/internal/domain/auth/service"
	"github.com/jackc/pgx/v4/pgxpool"
)

// NewAuthHandler initializes and returns an AuthHandler with all dependencies injected.
func NewAuthHandler(db *pgxpool.Pool, config *configs.Config) *handler.AuthHandler {
	// Initialize repository with the provided database connection
	userRepo := repository.NewPostgresUserRepository(db)

	// Initialize JWT service with configurations from config
	jwtService := service.NewJWTService(config.JWT.SecretKey, config.JWT.TokenDuration)

	// Initialize auth service with the repository and JWT service
	authService := service.NewAuthService(userRepo, jwtService)

	// Return a new handler with all dependencies set up
	return handler.NewAuthHandler(authService, jwtService, userRepo)
}
