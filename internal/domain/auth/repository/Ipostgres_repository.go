package repository

import (
	"context"
	"time"

	"github.com/dk5761/go-serv/internal/domain/auth/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateTokenVersion(ctx context.Context, userID uuid.UUID, tokenVersion int) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin time.Time) error
}
