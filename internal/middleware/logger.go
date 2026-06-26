// Package middleware provides Gin middlewares for security, context, and observability.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/pkg/logger"
)

// Logging Incoming http Requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		traceID := c.GetHeader("X-Trace-ID")
		if _, err := uuid.Parse(traceID); err != nil {
			traceID = uuid.New().String()
		}

		c.Set(traceIDKey, traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)

		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
		
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		ip := c.ClientIP()

		if path == "/health" {
			return
		}

		logger.Info(c.Request.Context(), "HTTP Request",
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip", ip,
		)
	}
}
