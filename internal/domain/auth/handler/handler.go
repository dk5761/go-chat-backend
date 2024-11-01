package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dk5761/go-serv/internal/domain/auth/dto"
	"github.com/dk5761/go-serv/internal/domain/auth/models"
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

	err := h.AuthService.SignUp(c.Request.Context(), req.Email, req.Username, req.Password)
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
		Username:  user.Username,
		LastLogin: user.LastLogin,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Validate and refresh the token
	newToken, err := h.JwtService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{Token: newToken})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
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

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}
	updates := models.User{
		Email: req.Email,
	}

	updatedUser, err := h.AuthService.UpdateUserProfile(c.Request.Context(), userID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, dto.ProfileResponse{
		ID:        updatedUser.ID.String(),
		Email:     updatedUser.Email,
		LastLogin: updatedUser.LastLogin,
		// Include any additional updated fields
	})
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
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

	err := h.AuthService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func (h *AuthHandler) GetUserByUsername(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}

	updatedUser, err := h.AuthService.GetUserByUsername(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, dto.ProfileResponse{
		ID:        updatedUser.ID.String(),
		Email:     updatedUser.Email,
		Username:  updatedUser.Username,
		LastLogin: updatedUser.LastLogin,
		// Include any additional updated fields
	})
}

func (h *AuthHandler) GetUsers(c *gin.Context) {
	q := c.Query("q") // Retrieve the `q` query parameter for search
	if q == "" {
		c.JSON(http.StatusOK, gin.H{"message": "No query parameter provided"})
		return
	}

	// Retrieve and parse limit and offset with default values
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
		return
	}

	// Call the service with limit and offset
	users, err := h.AuthService.GetUsers(c.Request.Context(), q, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, users)
}
