package errors

import (
	"fmt"
)

// ValidationError creates a new validation error
func ValidationError(message string) *AppError {
	return &AppError{
		Type:    "VALIDATION_ERROR",
		Code:    "VALIDATION_ERROR",
		Message: message,
		HTTPStatus: 400,
	}
}

// AuthenticationError creates a new authentication error
func AuthenticationError(code, message string) *AppError {
	return &AppError{
		Type:    "AUTHENTICATION_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 401,
	}
}

// AuthorizationError creates a new authorization error
func AuthorizationError(code, message string) *AppError {
	return &AppError{
		Type:    "AUTHORIZATION_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 403,
	}
}

// NotFoundError creates a new not found error
func NotFoundError(code, message string) *AppError {
	return &AppError{
		Type:    "NOT_FOUND_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 404,
	}
}

// ConflictError creates a new conflict error
func ConflictError(code, message string) *AppError {
	return &AppError{
		Type:    "CONFLICT_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 409,
	}
}

// BusinessError creates a new business logic error
func BusinessError(code, message string) *AppError {
	return &AppError{
		Type:    "BUSINESS_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 400,
	}
}

// ExternalServiceError creates a new external service error
func ExternalServiceError(code, message, details string) *AppError {
	return &AppError{
		Type:    "EXTERNAL_SERVICE_ERROR",
		Code:    code,
		Message: message,
		Details: details,
		HTTPStatus: 502,
	}
}

// DatabaseError creates a new database error
func DatabaseError(code, message string) *AppError {
	return &AppError{
		Type:    "DATABASE_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 500,
	}
}

// NetworkError creates a new network error
func NetworkError(code, message string) *AppError {
	return &AppError{
		Type:    "NETWORK_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 503,
	}
}

// InternalError creates a new internal error
func InternalError(code, message string) *AppError {
	return &AppError{
		Type:    "INTERNAL_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 500,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}