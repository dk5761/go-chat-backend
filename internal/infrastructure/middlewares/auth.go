package middlewares

import (
	"net/http"
	"strings"

	authRepo "github.com/dk5761/go-serv/internal/domain/auth/repository"
	authService "github.com/dk5761/go-serv/internal/domain/auth/service"
	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func JWTAuthMiddleware(jwtService authService.JWTService, userRepo authRepo.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logging.Logger.Error("Invalid token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check token version
		user, err := userRepo.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		if claims.TokenVersion != user.TokenVersion {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has been invalidated"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
