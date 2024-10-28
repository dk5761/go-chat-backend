package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	AuthService AuthService
	JwtService  JWTService     // Changed to uppercase
	UserRepo    UserRepository // Changed to uppercase
}

func NewAuthHandler(authService AuthService, jwtService JWTService, userRepo UserRepository) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
		JwtService:  jwtService,
		UserRepo:    userRepo,
	}
}

// SignUp handles user registration requests.
func (h *AuthHandler) SignUp(c *gin.Context) {
	// Define a struct to bind the JSON request body
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	// Bind the JSON request to the struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Call the AuthService to register the user
	err := h.AuthService.SignUp(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	// Respond with success
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login handles user authentication requests.
func (h *AuthHandler) Login(c *gin.Context) {
	// Define a struct to bind the JSON request body
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// Bind the JSON request to the struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Call the AuthService to authenticate the user
	token, err := h.AuthService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Respond with the JWT token
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Profile retrieves the authenticated user's profile.
func (h *AuthHandler) Profile(c *gin.Context) {
	// Get the user ID from the context (set by JWT middleware)
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

	// Call the AuthService to get the user profile
	user, err := h.AuthService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user profile"})
		return
	}

	// Respond with the user's profile information
	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID.String(),
		"email":      user.Email,
		"last_login": user.LastLogin,
	})
}
