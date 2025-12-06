package utils

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/gmsas95/blytz-mvp/shared/pkg/errors"
)

// Logger is a wrapper around zap logger with additional functionality
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger() (*Logger, error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &Logger{Logger: zapLogger}, nil
}

// NewDevelopmentLogger creates a development logger
func NewDevelopmentLogger() (*Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &Logger{Logger: zapLogger}, nil
}

// WithRequest adds request context to the logger
func (l *Logger) WithRequest(c *gin.Context) *zap.Logger {
	return l.With(
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo represents error information in responses
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta represents metadata in responses
type Meta struct {
	Timestamp int64 `json:"timestamp"`
	Version   string `json:"version"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Timestamp: time.Now().Unix(),
			Version:   "1.0.0",
		},
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code, message string, details interface{}) *Response {
	return &Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: &Meta{
			Timestamp: time.Now().Unix(),
			Version:   "1.0.0",
		},
	}
}

// NewMessageResponse creates a message-only response
func NewMessageResponse(message string) *Response {
	return &Response{
		Success: true,
		Message: message,
		Meta: &Meta{
			Timestamp: time.Now().Unix(),
			Version:   "1.0.0",
		},
	}
}

// SendJSON sends a JSON response with standard format
func SendJSON(c *gin.Context, statusCode int, response *Response) {
	c.JSON(statusCode, response)
}

// SendSuccess sends a success response
func SendSuccess(c *gin.Context, data interface{}) {
	SendJSON(c, http.StatusOK, NewSuccessResponse(data))
}

// SendCreated sends a created response
func SendCreated(c *gin.Context, data interface{}) {
	SendJSON(c, http.StatusCreated, NewSuccessResponse(data))
}

// SendError sends an error response
func SendError(c *gin.Context, statusCode int, code, message string, details interface{}) {
	SendJSON(c, statusCode, NewErrorResponse(code, message, details))
}

// SendBadRequest sends a bad request response
func SendBadRequest(c *gin.Context, message string) {
	SendError(c, http.StatusBadRequest, "BAD_REQUEST", message, nil)
}

// SendUnauthorized sends an unauthorized response
func SendUnauthorized(c *gin.Context, message string) {
	SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// SendForbidden sends a forbidden response
func SendForbidden(c *gin.Context, message string) {
	SendError(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// SendNotFound sends a not found response
func SendNotFound(c *gin.Context, message string) {
	SendError(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

// SendConflict sends a conflict response
func SendConflict(c *gin.Context, message string) {
	SendError(c, http.StatusConflict, "CONFLICT", message, nil)
}

// SendInternalError sends an internal server error response
func SendInternalError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, nil)
}

// SendErrorResponse sends an error response using shared error types
func SendErrorResponse(c *gin.Context, err error) {
	// Check if it's an AppError from shared errors package
	if appErr, ok := err.(*errors.AppError); ok {
		SendError(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		return
	}
	
	// Fallback for other error types
	SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
}

// SendSuccessResponse sends a success response with status code and data
func SendSuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	SendJSON(c, statusCode, NewSuccessResponse(data))
}

// Environment utility functions
func GetEnv(key, defaultValue string) string {
	// Use os.Getenv instead of gin.Env (which doesn't exist in newer Gin versions)
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validation utilities
func IsValidUUID(uuid string) bool {
	// Simple UUID validation - in production, use a proper UUID library
	return len(uuid) == 36 && uuid[8] == '-' && uuid[13] == '-' && uuid[18] == '-' && uuid[23] == '-'
}

// Health check utilities
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp int64             `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// NewHealthStatus creates a new health status
func NewHealthStatus(status string) *HealthStatus {
	return &HealthStatus{
		Status:    status,
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]string),
		Details:   make(map[string]interface{}),
	}
}

// AddService adds a service health check
func (h *HealthStatus) AddService(name, status string) {
	h.Services[name] = status
}

// AddDetail adds a detail to the health check
func (h *HealthStatus) AddDetail(key string, value interface{}) {
	h.Details[key] = value
}

// IsHealthy checks if all services are healthy
func (h *HealthStatus) IsHealthy() bool {
	if h.Status != "ok" {
		return false
	}
	for _, status := range h.Services {
		if status != "ok" {
			return false
		}
	}
	return true
}