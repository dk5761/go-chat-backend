package middlewares

import (
	"time"

	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Proceed to the next handler
		c.Next()

		// Retrieve the trace ID from the context (set by the tracing middleware)
		traceID, exists := c.Get("trace_id")
		if !exists {
			traceID = "unknown" // Default if trace ID is missing
		}

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		errors := c.Errors.ByType(gin.ErrorTypePrivate).String()

		logging.Logger.Info("HTTP Request",
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Duration("latency", latency),
			zap.String("errors", errors),
			zap.String("trace_id", traceID.(string)),
		)
	}
}
