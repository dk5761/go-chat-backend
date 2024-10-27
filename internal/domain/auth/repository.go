package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateTokenVersion(ctx context.Context, userID uuid.UUID, tokenVersion int) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin time.Time) error
}
