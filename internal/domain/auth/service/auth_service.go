package service

import (
	"context"
	"errors"
	"time"

	"github.com/dk5761/go-serv/internal/domain/auth/models"
	"github.com/dk5761/go-serv/internal/domain/auth/repository"
	"github.com/dk5761/go-serv/internal/domain/common"
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

	// Update last login and last login token timestamps
	newLoginTime := time.Now()
	err = s.userRepo.UpdateLastLogin(ctx, user.ID, newLoginTime, newLoginTime)
	if err != nil {
		logging.Logger.Error("Failed to update last login", zap.Error(err))
		return "", err // Return error since login state wasn't updated
	}

	// Retrieve the updated user to get the latest LastLoginToken
	updatedUser, err := s.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		logging.Logger.Error("Failed to retrieve updated user", zap.Error(err))
		return "", err
	}

	// Generate JWT token using the updated LastLoginToken
	token, err := s.jwtService.GenerateToken(updatedUser.ID, updatedUser.LastLoginToken)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	// Query the repository to get the user by ID
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
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
