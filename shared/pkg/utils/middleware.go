package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Logger interface for middleware logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// timeoutContext is a simple context wrapper for timeout
type timeoutContext struct {
	context.Context
	timeout time.Duration
}

// NewContextWithTimeout creates a new context with timeout
func NewContextWithTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, "timeout", timeout)
}

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = GenerateUUID()
		}
		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})
}

// LoggingMiddleware logs request information
func LoggingMiddleware(logger Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)
		return ""
	})
}

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(
			NewContextWithTimeout(c.Request.Context(), timeout),
		)
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// HealthCheckMiddleware provides a simple health check endpoint
func HealthCheckMiddleware(healthPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == healthPath {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"time":   time.Now().UTC(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}