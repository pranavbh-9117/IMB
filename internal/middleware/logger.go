// Package middleware provides Gin middlewares for security, context, and observability.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/pkg/logger"
)

// RequestLogger is a Gin middleware that logs the incoming HTTP request details
// and injects a unique RequestID into the context for tracing.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate and inject Request ID
		reqID := uuid.New().String()
		c.Set("request_id", reqID)

		// Set in header for client-side tracing
		c.Writer.Header().Set("X-Request-ID", reqID)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Extract status code and method
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		ip := c.ClientIP()

		// Avoid logging noisy requests like health checks if they exist
		if path == "/health" {
			return
		}

		// Log structured data using the new logger
		logger.Info(c.Request.Context(), "HTTP Request",
			"request_id", reqID,
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip", ip,
		)
	}
}
