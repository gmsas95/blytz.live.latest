package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	// Validation errors
	ValidationError ErrorType = "VALIDATION_ERROR"
	
	// Authentication errors
	AuthenticationError ErrorType = "AUTHENTICATION_ERROR"
	AuthorizationError   ErrorType = "AUTHORIZATION_ERROR"
	
	// Not found errors
	NotFoundError ErrorType = "NOT_FOUND_ERROR"
	
	// Conflict errors
	ConflictError ErrorType = "CONFLICT_ERROR"
	
	// Business logic errors
	BusinessError ErrorType = "BUSINESS_ERROR"
	
	// External service errors
	ExternalServiceError ErrorType = "EXTERNAL_SERVICE_ERROR"
	
	// Internal server errors
	InternalError ErrorType = "INTERNAL_ERROR"
	
	// Database errors
	DatabaseError ErrorType = "DATABASE_ERROR"
	
	// Network errors
	NetworkError ErrorType = "NETWORK_ERROR"
	
	// Timeout errors
	TimeoutError ErrorType = "TIMEOUT_ERROR"
)

// AppError represents a custom application error
type AppError struct {
	Type       ErrorType `json:"type"`
	Code       string    `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Cause      error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s - %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code, message, details string, httpStatus int) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(code, message string) *AppError {
	return &AppError{
		Type:       ValidationError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(code, message string) *AppError {
	return &AppError{
		Type:       AuthenticationError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(code, message string) *AppError {
	return &AppError{
		Type:       AuthorizationError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(code, message string) *AppError {
	return &AppError{
		Type:       NotFoundError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusNotFound,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(code, message string) *AppError {
	return &AppError{
		Type:       ConflictError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// NewBusinessError creates a new business logic error
func NewBusinessError(code, message string) *AppError {
	return &AppError{
		Type:       BusinessError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

// NewExternalServiceError creates a new external service error
func NewExternalServiceError(code, message, details string) *AppError {
	return &AppError{
		Type:       ExternalServiceError,
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusBadGateway,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(code, message string) *AppError {
	return &AppError{
		Type:       InternalError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(code, message string) *AppError {
	return &AppError{
		Type:       DatabaseError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(code, message string) *AppError {
	return &AppError{
		Type:       NetworkError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(code, message string) *AppError {
	return &AppError{
		Type:       TimeoutError,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusRequestTimeout,
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatusForType(errorType),
		Cause:      err,
	}
}

// getHTTPStatusForType returns the appropriate HTTP status code for an error type
func getHTTPStatusForType(errorType ErrorType) int {
	switch errorType {
	case ValidationError:
		return http.StatusBadRequest
	case AuthenticationError:
		return http.StatusUnauthorized
	case AuthorizationError:
		return http.StatusForbidden
	case NotFoundError:
		return http.StatusNotFound
	case ConflictError:
		return http.StatusConflict
	case BusinessError:
		return http.StatusUnprocessableEntity
	case ExternalServiceError:
		return http.StatusBadGateway
	case DatabaseError, InternalError:
		return http.StatusInternalServerError
	case NetworkError:
		return http.StatusServiceUnavailable
	case TimeoutError:
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr, ok := IsAppError(err); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}