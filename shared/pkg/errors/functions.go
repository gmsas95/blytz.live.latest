package errors

// ValidationError creates a new validation error
func ValidationError(code, message string) *AppError {
	return &AppError{
		Type:    "VALIDATION_ERROR",
		Code:    code,
		Message: message,
		HTTPStatus: 400,
	}
}

// NewValidationError creates a new validation error (alias for ValidationError)
func NewValidationError(code, message string) *AppError {
	return ValidationError(code, message)
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

// NewAuthenticationError creates a new authentication error (alias for AuthenticationError)
func NewAuthenticationError(code, message string) *AppError {
	return AuthenticationError(code, message)
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

// NewAuthorizationError creates a new authorization error (alias for AuthorizationError)
func NewAuthorizationError(code, message string) *AppError {
	return AuthorizationError(code, message)
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

// NewNotFoundError creates a new not found error (alias for NotFoundError)
func NewNotFoundError(code, message string) *AppError {
	return NotFoundError(code, message)
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

// NewConflictError creates a new conflict error (alias for ConflictError)
func NewConflictError(code, message string) *AppError {
	return ConflictError(code, message)
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

// NewBusinessError creates a new business logic error (alias for BusinessError)
func NewBusinessError(code, message string) *AppError {
	return BusinessError(code, message)
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

// NewExternalServiceError creates a new external service error (alias for ExternalServiceError)
func NewExternalServiceError(code, message, details string) *AppError {
	return ExternalServiceError(code, message, details)
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

// NewDatabaseError creates a new database error (alias for DatabaseError)
func NewDatabaseError(code, message string) *AppError {
	return DatabaseError(code, message)
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

// NewNetworkError creates a new network error (alias for NetworkError)
func NewNetworkError(code, message string) *AppError {
	return NetworkError(code, message)
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

// NewInternalError creates a new internal error (alias for InternalError)
func NewInternalError(code, message string) *AppError {
	return InternalError(code, message)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}