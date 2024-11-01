package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dk5761/go-serv/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type JSONResponse struct {
	Data interface{} `json:"data"`

	TraceID string `json:"trace_id,omitempty"`
}

// StartTracingMiddleware starts a new span for each request and propagates the context
func StartTracingMiddleware() gin.HandlerFunc {
	tracer := otel.Tracer("github.com/dk5761/go-serv")

	return func(c *gin.Context) {

		if websocket.IsWebSocketUpgrade(c.Request) {
			c.Next()
			return
		}
		// Start a new span using the existing request context to propagate any trace information
		ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path)
		defer span.End() // Ensure the span ends after the request completes

		// Set the span context to Gin context, allowing later middlewares to access it
		c.Request = c.Request.WithContext(ctx)
		c.Next() // Continue to the next handler
	}
}

// TracingMiddleware extracts and sets the trace ID in the context for logging
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if websocket.IsWebSocketUpgrade(c.Request) {
			c.Next()
			return
		}
		// Retrieve trace ID from OpenTelemetry
		spanCtx := trace.SpanFromContext(c.Request.Context()).SpanContext()

		if spanCtx.HasTraceID() {
			traceID := spanCtx.TraceID().String()
			// Add trace ID to context for access in request lifecycle
			c.Set("trace_id", traceID)
		}

		c.Next()
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// WriteString captures the response body without sending it to the client immediately
func (w *responseWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

// TraceIDResponseMiddleware is the middleware that wraps JSON responses with a trace ID
func TraceIDResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if websocket.IsWebSocketUpgrade(c.Request) {
			c.Next()
			return
		}

		// Retrieve trace ID from context (set by another middleware, e.g., TracingMiddleware)
		traceID, exists := c.Get("trace_id")
		if !exists {
			traceID = "unknown" // Default if trace ID is missing
		}

		// Use our custom response writer to capture the response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer

		// Process the request
		c.Next()

		// Check if Content-Type header includes "application/json" and status is successful
		contentType := c.Writer.Header().Get("Content-Type")
		isJSON := false
		for _, part := range strings.Split(contentType, ";") {
			if strings.TrimSpace(part) == "application/json" {
				isJSON = true
				break
			}
		}

		if isJSON && c.Writer.Status() < http.StatusBadRequest {

			// Capture the original response body and unmarshal it
			var originalResponse interface{}
			if err := json.Unmarshal(writer.body.Bytes(), &originalResponse); err != nil {
				// Send an error response if unmarshalling fails
				c.Writer = writer.ResponseWriter // Restore original writer
				c.Writer.Header().Set("Content-Type", "application/json")
				c.Writer.WriteHeader(http.StatusInternalServerError)

				errorResponse := JSONResponse{

					TraceID: fmt.Sprintf("%v", traceID),
				}

				if err := json.NewEncoder(c.Writer).Encode(errorResponse); err != nil {
					logging.Logger.Error("Failed to write error response",
						zap.String("error", err.Error()),
						zap.String("trace_id", fmt.Sprintf("%v", traceID)),
					)
				}
				return
			}

			// Wrap the response in JSONResponse with trace ID
			wrappedResponse := JSONResponse{
				Data:    originalResponse,
				TraceID: fmt.Sprintf("%v", traceID),
			}

			// Marshal the wrapped response
			wrappedResponseBytes, err := json.Marshal(wrappedResponse)
			if err != nil {
				// Handle marshalling error
				c.Writer = writer.ResponseWriter // Restore original writer
				c.Writer.Header().Set("Content-Type", "application/json")
				c.Writer.WriteHeader(http.StatusInternalServerError)

				errorResponse := JSONResponse{
					TraceID: fmt.Sprintf("%v", traceID),
				}

				if err := json.NewEncoder(c.Writer).Encode(errorResponse); err != nil {
					logging.Logger.Error("Failed to write error response",
						zap.String("error", err.Error()),
						zap.String("trace_id", fmt.Sprintf("%v", traceID)),
					)
				}
				return
			}

			// Restore the original writer before writing the modified response
			c.Writer = writer.ResponseWriter

			// Set headers and write only the wrapped response
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(wrappedResponseBytes)))
			c.Writer.WriteHeaderNow() // Commit headers

			if _, err := c.Writer.Write(wrappedResponseBytes); err != nil {
				logging.Logger.Error("Failed to write error response",
					zap.String("error", err.Error()),
					zap.String("trace_id", fmt.Sprintf("%v", traceID)),
				)
			}
		} else {
			fmt.Println("Processing non-JSON or error response")

			// For non-JSON or error responses, write the captured original body directly
			c.Writer = writer.ResponseWriter
			// Ensure that the original headers and status code are preserved
			c.Writer.WriteHeaderNow()
			if _, err := c.Writer.Write(writer.body.Bytes()); err != nil {
				logging.Logger.Error("Failed to write error response",
					zap.String("error", err.Error()),
					zap.String("trace_id", fmt.Sprintf("%v", traceID)),
				)
			}
		}
	}
}
