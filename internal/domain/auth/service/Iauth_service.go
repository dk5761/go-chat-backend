package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/dk5761/go-serv/internal/domain/auth/models"
)

type AuthService interface {
	SignUp(ctx context.Context, email, username, password string) error
	Login(ctx context.Context, email, password string) (string, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, updates models.User) (*models.User, error)
	UpdateDeviceToken(ctx context.Context, userID uuid.UUID, updates models.User) (*models.User, error)
	GetUsers(ctx context.Context, q string, limit, offset int) ([]*models.User, int, error)
}
