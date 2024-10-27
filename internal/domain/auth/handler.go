package auth

import (
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService}
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	// Implementation
}

func (h *AuthHandler) Login(c *gin.Context) {
	// Implementation
}

func (h *AuthHandler) Profile(c *gin.Context) {
	// Implementation
}
