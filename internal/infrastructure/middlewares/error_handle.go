package middlewares

import (
	"errors"
	"net/http"

	"github.com/dk5761/go-serv/internal/domain/common"
	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that maps custom errors to HTTP status codes.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var statusCode int

			// Map errors to HTTP status codes
			switch {
			case errors.Is(err, common.ErrNotFound):
				statusCode = http.StatusNotFound
			case errors.Is(err, common.ErrUnauthorized):
				statusCode = http.StatusUnauthorized
			case errors.Is(err, common.ErrForbidden):
				statusCode = http.StatusForbidden
			case errors.Is(err, common.ErrInvalidInput):
				statusCode = http.StatusBadRequest
			case errors.Is(err, common.ErrConflict):
				statusCode = http.StatusConflict
			case errors.Is(err, common.ErrInternalServer):
				statusCode = http.StatusInternalServerError
			default:
				statusCode = http.StatusInternalServerError
			}

			// Return JSON error response
			c.JSON(statusCode, gin.H{"error": err.Error()})
		}
	}
}
