package middlewares

import (
	"errors"
	"net/http"

	"github.com/dk5761/go-serv/internal/domain/common"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var statusCode int
			switch {
			case errors.Is(err, common.ErrNotFound):
				statusCode = http.StatusNotFound
			case errors.Is(err, common.ErrUnauthorized):
				statusCode = http.StatusUnauthorized
			default:
				statusCode = http.StatusInternalServerError
			}
			c.JSON(statusCode, gin.H{"error": err.Error()})
		}
	}
}
