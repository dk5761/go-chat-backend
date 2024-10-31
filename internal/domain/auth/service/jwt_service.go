package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService defines methods to generate and validate tokens.
type JWTService interface {
	GenerateToken(userID uuid.UUID, tokenTimeStamp time.Time) (string, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
}

type jwtService struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTService initializes a new JWTService with the given secret key and token duration.
func NewJWTService(secretKey string, tokenDurationMinutes int) JWTService {
	return &jwtService{
		secretKey:     secretKey,
		tokenDuration: time.Duration(tokenDurationMinutes) * time.Minute,
	}
}

// CustomClaims defines the custom JWT claims structure.
type CustomClaims struct {
	UserID  uuid.UUID `json:"user_id"`
	TokenTS int64     `json:"token_ts"` // Use int64 for Unix timestamp
	jwt.RegisteredClaims
}

// GenerateToken creates a JWT with user ID and token timestamp.
func (s *jwtService) GenerateToken(userID uuid.UUID, tokenTimeStamp time.Time) (string, error) {
	claims := &CustomClaims{
		UserID:  userID,
		TokenTS: tokenTimeStamp.Unix(), // Set the token timestamp as Unix time
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT, returning the custom claims.
func (s *jwtService) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
