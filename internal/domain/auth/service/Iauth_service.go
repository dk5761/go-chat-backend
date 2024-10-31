package service

import (
	"context"

	"github.com/dk5761/go-serv/internal/domain/auth/models"
	"github.com/google/uuid"
)

type AuthService interface {
	SignUp(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}
