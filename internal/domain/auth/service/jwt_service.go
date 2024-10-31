package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService defines methods to generate, validate, and refresh tokens.
type JWTService interface {
	GenerateToken(userID uuid.UUID, tokenTimeStamp time.Time) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
	ValidateRefreshToken(refreshToken string) (*CustomClaims, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
}

type jwtService struct {
	secretKey       string
	tokenDuration   time.Duration
	refreshSecret   string        // Separate secret for refresh tokens
	refreshDuration time.Duration // Expiry duration for refresh tokens
}

// NewJWTService initializes a new JWTService with given secrets and durations.
func NewJWTService(secretKey string, tokenDurationMinutes, refreshDurationMinutes int, refreshSecret string) JWTService {
	return &jwtService{
		secretKey:       secretKey,
		tokenDuration:   time.Duration(tokenDurationMinutes) * time.Minute,
		refreshSecret:   refreshSecret,
		refreshDuration: time.Duration(refreshDurationMinutes) * time.Minute,
	}
}

// CustomClaims defines the custom JWT claims structure.
type CustomClaims struct {
	UserID  uuid.UUID `json:"user_id"`
	TokenTS int64     `json:"token_ts"` // Unix timestamp for validation
	jwt.RegisteredClaims
}

// GenerateToken creates an access JWT with user ID and token timestamp.
func (s *jwtService) GenerateToken(userID uuid.UUID, tokenTimeStamp time.Time) (string, error) {
	claims := &CustomClaims{
		UserID:  userID,
		TokenTS: tokenTimeStamp.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// GenerateRefreshToken creates a refresh token with a longer duration.
func (s *jwtService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := &CustomClaims{
		UserID:  userID,
		TokenTS: time.Now().Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return refreshToken.SignedString([]byte(s.refreshSecret))
}

// ValidateToken parses and validates an access JWT, returning the custom claims.
func (s *jwtService) ValidateToken(tokenString string) (*CustomClaims, error) {
	return s.validateTokenWithSecret(tokenString, s.secretKey)
}

// ValidateRefreshToken parses and validates a refresh token, returning the custom claims.
func (s *jwtService) ValidateRefreshToken(refreshToken string) (*CustomClaims, error) {
	return s.validateTokenWithSecret(refreshToken, s.refreshSecret)
}

// Helper method to parse and validate JWT with a specific secret.
func (s *jwtService) validateTokenWithSecret(tokenString, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func (s *jwtService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Validate the refresh token
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	// Generate a new access token using the user ID from the refresh token claims
	return s.GenerateToken(claims.UserID, time.Now())
}

func (s *authService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Check if the user exists
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Delete the user from the repository
	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		return err
	}
	return nil
}

// func (s *authService) UpdateUserProfile(ctx context.Context, userID uuid.UUID, updates models.User) (*models.User, error) {
// 	// Fetch the current user from the repository
// 	user, err := s.userRepo.GetUserByID(ctx, userID)
// 	if err != nil {
// 		return nil, errors.New("user not found")
// 	}

// 	// Update fields as needed
// 	if updates.Email != "" {
// 		user.Email = updates.Email
// 	}
// 	if updates.DisplayName != "" {
// 		user.DisplayName = updates.DisplayName
// 	}
// 	if updates.ProfilePictureURL != "" {
// 		user.ProfilePictureURL = updates.ProfilePictureURL
// 	}

// 	// Save updated user
// 	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
// 		return nil, err
// 	}
// 	return user, nil
// }
