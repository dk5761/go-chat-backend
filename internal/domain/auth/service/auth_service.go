package service

import (
	"context"
	"errors"
	"time"

	"github.com/dk5761/go-serv/internal/domain/auth/models"
	"github.com/dk5761/go-serv/internal/domain/auth/repository"
	"github.com/dk5761/go-serv/internal/domain/common/helpers"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type authService struct {
	userRepo   repository.UserRepository
	jwtService JWTService
}

func NewAuthService(userRepo repository.UserRepository, jwtService JWTService) AuthService {
	return &authService{userRepo, jwtService}
}

func (s *authService) SignUp(ctx context.Context, email, password string) error {
	// Check if user already exists
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		return err
	}

	user := &models.User{
		ID:             uuid.New(),
		Email:          email,
		PasswordHash:   hashedPassword,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastLogin:      time.Now(),
		LastLoginToken: time.Now(),
	}

	return s.userRepo.CreateUser(ctx, user)
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !helpers.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	// Update last login time
	err = s.userRepo.UpdateLastLogin(ctx, user.ID, time.Now(), time.Now())
	if err != nil {
		logging.Logger.Error("Failed to update last login", zap.Error(err))
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID, user.LastLoginToken)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	// Implement a method to get user by ID

	return &models.User{}, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	// Retrieve the user to confirm existence
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Set a new timestamp for LastLoginToken to invalidate existing tokens
	newTokenTimestamp := time.Now()

	// Update the user's LastLoginToken to the new timestamp
	return s.userRepo.UpdateLastLogin(ctx, userID, user.LastLogin, newTokenTimestamp)
}