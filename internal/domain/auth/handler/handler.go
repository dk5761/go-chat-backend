package handler

import (
	"fmt"
	"net/http"

	"github.com/dk5761/go-serv/internal/domain/auth/dto"
	"github.com/dk5761/go-serv/internal/domain/auth/repository"
	"github.com/dk5761/go-serv/internal/domain/auth/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	AuthService service.AuthService
	JwtService  service.JWTService
	UserRepo    repository.UserRepository
}

func NewAuthHandler(authService service.AuthService, jwtService service.JWTService, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
		JwtService:  jwtService,
		UserRepo:    userRepo,
	}
}

// SignUp handles user registration requests.
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req dto.SignUpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	err := h.AuthService.SignUp(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login handles user authentication requests.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	token, err := h.AuthService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{Token: token})
}

// Profile retrieves the authenticated user's profile.
func (h *AuthHandler) Profile(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.AuthService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user profile"})
		return
	}

	fmt.Print(user)

	// Use ProfileResponse DTO to send the user's profile information
	c.JSON(http.StatusOK, dto.ProfileResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		LastLogin: user.LastLogin,
	})
}
