# Shared Package

This package contains shared utilities, error types, and common functionality used across all Blytz MVP services.

## Structure

```
shared/
├── pkg/
│   ├── errors/          # Custom error types and constants
│   └── utils/           # Utility functions and middleware
├── go.mod              # Go module definition
└── README.md           # This file
```

## Installation

```bash
go get github.com/gmsas95/blytz-mvp/shared
```

## Usage

### Errors Package

The errors package provides custom error types and predefined error constants for consistent error handling across services.

```go
import "github.com/gmsas95/blytz-mvp/shared/pkg/errors"

// Create a custom validation error
err := errors.NewValidationError("INVALID_EMAIL", "Invalid email format")

// Use predefined errors
if userNotFound {
    return errors.ErrUserNotFoundError
}

// Wrap an existing error
wrappedErr := errors.WrapError(originalErr, errors.DatabaseError, "QUERY_FAILED", "Failed to execute query")
```

### Utils Package

The utils package provides various utilities for common operations.

#### Response Utilities

```go
import "github.com/gmsas95/blytz-mvp/shared/pkg/utils"

// Send success response
utils.SendSuccessResponse(c, http.StatusOK, data)

// Send error response
utils.SendErrorResponse(c, err)

// Send paginated response
pagination := utils.CalculatePagination(page, perPage, total)
utils.SendPaginatedResponse(c, http.StatusOK, data, pagination)
```

#### CORS Middleware

```go
// Use default CORS configuration
router.Use(utils.DefaultCORSMiddleware())

// Use custom CORS configuration
config := &utils.CORSConfig{
    AllowedOrigins: []string{"https://example.com"},
    AllowedMethods: []string{"GET", "POST"},
}
router.Use(utils.CORSMiddleware(config))
```

#### Database Utilities

```go
import "github.com/gmsas95/blytz-mvp/shared/pkg/utils"

// Create database connection
config := utils.DefaultDatabaseConfig()
config.Host = "localhost"
config.DBName = "myapp"

db, err := utils.NewDatabase(ctx, config, logger)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Execute in transaction
err = db.ExecuteInTransaction(ctx, func(tx pgx.Tx) error {
    // Database operations here
    return nil
})
```

#### JWT Utilities

```go
// Create JWT manager
config := utils.DefaultJWTConfig()
config.SecretKey = "your-secret-key"
jwtManager := utils.NewJWTManager(config, logger)

// Generate token pair
tokenPair, err := jwtManager.GenerateTokenPair(userID, email, role, scopes)

// Validate token
claims, err := jwtManager.ValidateToken(tokenString)
```

#### Validation

```go
// Create validator
validator := utils.NewValidator()

// Validate fields
validator.ValidateRequired("email", email)
validator.ValidateEmail("email", email)
validator.ValidatePassword("password", password)

if validator.HasErrors() {
    errors := validator.GetErrorMap()
    // Handle validation errors
}

// Use convenience functions
if !utils.IsValidEmail(email) {
    // Handle invalid email
}
```

#### Logging

```go
// Create logger
config := utils.DefaultLoggingConfig()
config.Level = utils.DebugLevel
logger, err := utils.NewLogger(config)
if err != nil {
    log.Fatal(err)
}

// Log with context
logger.WithRequestID(requestID).
    WithUserID(userID).
    Info("User action completed")

// Use global logger
utils.InitGlobalLogger(config)
utils.GlobalLogInfo("Application started")
```

#### Authentication Middleware

```go
// Create auth client
jwtManager := utils.NewJWTManager(jwtConfig, logger)
authClient := utils.NewDefaultAuthClient(jwtManager)

// Apply authentication middleware
router.Use(utils.GinAuthMiddleware(authClient))

// Apply role-based middleware
adminOnly := router.Group("/admin")
adminOnly.Use(utils.AdminAuthMiddleware())
{
    adminOnly.GET("/users", handler.GetUsers)
}

// Apply scope-based middleware
api := router.Group("/api")
api.Use(utils.ScopeAuthMiddleware("read:users"))
{
    api.GET("/users", handler.GetUsers)
}
```

#### General Middleware

```go
// Request ID middleware
router.Use(utils.RequestIDMiddleware())

// Logging middleware
router.Use(utils.LoggingMiddleware(logger))

// Error handling middleware
router.Use(utils.ErrorHandlingMiddleware(logger))

// Security headers middleware
router.Use(utils.SecurityHeadersMiddleware())

// Rate limiting middleware
router.Use(utils.RateLimitMiddleware(100, time.Hour))

// Timeout middleware
router.Use(utils.TimeoutMiddleware(30 * time.Second))
```

## Configuration

### Database Configuration

```go
config := &utils.DatabaseConfig{
    Host:            "localhost",
    Port:            5432,
    User:            "postgres",
    Password:        "password",
    DBName:          "blytz_mvp",
    SSLMode:         "disable",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 1 * time.Minute,
}
```

### Redis Configuration

```go
config := &utils.RedisConfig{
    Host:         "localhost",
    Port:         6379,
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 3,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolTimeout:  4 * time.Second,
}
```

### JWT Configuration

```go
config := &utils.JWTConfig{
    SecretKey:          "your-secret-key",
    AccessTokenExpiry:   15 * time.Minute,
    RefreshTokenExpiry:  7 * 24 * time.Hour,
    Issuer:             "blytz-mvp",
    Audience:           "blytz-mvp-users",
    RefreshTokenLength:  32,
}
```

### Logging Configuration

```go
config := &utils.LoggingConfig{
    Level:         utils.InfoLevel,
    Format:        "json",
    Output:        "stdout",
    ServiceName:   "my-service",
    Environment:   "production",
    EnableCaller:  false,
    EnableStacktrace: true,
}
```

## Error Types

The errors package defines several error types:

- `ValidationError` - Input validation errors
- `AuthenticationError` - Authentication failures
- `AuthorizationError` - Authorization failures
- `NotFoundError` - Resource not found errors
- `ConflictError` - Resource conflicts
- `BusinessError` - Business logic violations
- `ExternalServiceError` - Third-party service errors
- `DatabaseError` - Database operation errors
- `NetworkError` - Network-related errors
- `TimeoutError` - Operation timeout errors
- `InternalError` - Internal server errors

## Best Practices

1. **Error Handling**: Always use the custom error types for consistent error responses
2. **Logging**: Use structured logging with request IDs for better traceability
3. **Validation**: Validate all input using the validation utilities
4. **Authentication**: Use the provided authentication middleware for secure endpoints
5. **Database**: Use transactions for multi-step operations
6. **CORS**: Configure CORS appropriately for your environment

## Examples

See the individual service implementations for more examples of how to use the shared package.

## Dependencies

- github.com/gin-gonic/gin v1.10.0
- github.com/golang-jwt/jwt/v5 v5.2.1
- github.com/google/uuid v1.6.0
- github.com/jackc/pgx/v5 v5.6.0
- github.com/redis/go-redis/v9 v9.7.0
- go.uber.org/zap v1.27.0