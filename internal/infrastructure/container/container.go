package container

import (
	"github.com/dk5761/go-serv/configs"
	"github.com/dk5761/go-serv/internal/domain/auth"
	authHandler "github.com/dk5761/go-serv/internal/domain/auth/handler"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Container struct {
	AuthHandler *authHandler.AuthHandler // Use the fully qualified handler type here
}

func NewContainer(db *pgxpool.Pool, config *configs.Config) *Container {
	// Initialize AuthHandler through the auth facade function
	authHandler := auth.NewAuthHandler(db, config)

	return &Container{
		AuthHandler: authHandler,
	}
}
