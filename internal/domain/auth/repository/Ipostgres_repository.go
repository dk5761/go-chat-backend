package repository

import (
	"context"
	"time"

	"github.com/dk5761/go-serv/internal/domain/auth/models"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user-related data operations.
type UserRepository interface {
	// CreateUser inserts a new user into the database.
	CreateUser(ctx context.Context, user *models.User) error

	// GetUserByEmail retrieves a user by their email address.
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// GetUserByID retrieves a user by their unique ID.
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)

	// UpdateLastLogin updates the last login time and last login token of a user.
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin time.Time, lastLoginToken time.Time) error

	// UpdateUserTimestamps updates the updated_at field for a user.
	UpdateUserTimestamps(ctx context.Context, userID uuid.UUID, updatedAt time.Time) error
}
