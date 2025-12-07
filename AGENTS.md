# AGENTS.md

This file provides guidance for AI agents (Cursor, Copilot, Claude, etc.) when working with code in this repository.

## Project Overview

Blytz Live Auction MVP - A real-time livestream commerce platform built with Go microservices, React frontend, and React Native mobile app. This is a comprehensive microservices e-commerce platform with 10+ services.

## Architecture

### Core Services (Go 1.25)
- **Auth Service** (Port 8085): Authentication with JWT tokens
- **Product Service** (Port 8086): Product catalog management
- **Auction Service** (Port 8087): Real-time auction engine
- **Order Service** (Port 8088): Order processing and management
- **Payment Service** (Port 8089): Payment processing with Stripe
- **Chat Service** (Port 8090): Real-time messaging
- **Logistics Service** (Port 8091): Shipping and tracking
- **Gateway Service** (Port 8092): API gateway and routing
- **LiveKit Service** (Port 8093): Video streaming integration
- **Notification Service** (Port 8094): Email/push notifications

### Frontend Applications
- **Frontend** (Next.js): React web application
- **Frontend Mobile** (React Native): Mobile application

### Shared Infrastructure
- **Shared Package**: Common utilities, errors, auth client
- **PostgreSQL**: Primary database
- **Redis**: Caching and session storage

## Build, Test, and Development Commands

### Go Services Commands

#### Running Individual Services
```bash
# Run any service from its directory
cd services/auth-service
go run main.go

# Or from main-simple.go (if exists)
go run main-simple.go

# Build service
go build -o service-name .

# Run with specific port
PORT=8085 go run main.go
```

#### Testing Commands
```bash
# Run all tests in a service
cd services/auth-service
go test ./... -v

# Run tests with coverage
go test ./... -v -race -coverprofile=coverage.out
go tool cover -func=coverage.out

# Run specific test
go test -v ./internal/services -run TestAuthService

# Run integration tests
go test -v ./tests/integration/...

# Test all services (from root)
./scripts/test.sh
./scripts/test.sh auth-service
```

#### Build and Lint Commands
```bash
# Build all services
cd services/auth-service && go build ./...
cd services/product-service && go build ./...
# ... repeat for each service

# Format Go code
go fmt ./...

# Vet Go code
go vet ./...

# Tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Sync workspace (from root)
go work sync
```

### Frontend Commands (Next.js)

#### Development
```bash
cd frontend
npm run dev              # Start development server
npm run build            # Build for production
npm run start            # Start production server
npm run type-check       # TypeScript type checking
```

#### Testing
```bash
npm run test             # Run Jest tests
npm run test:watch       # Run tests in watch mode
npm run test:coverage    # Run tests with coverage
npm run test:e2e         # Run Playwright E2E tests
npm run test:e2e:ui      # Run E2E tests with UI
npm run test:e2e:debug   # Debug E2E tests
```

#### Code Quality
```bash
npm run lint             # Run ESLint
npm run lint:fix         # Fix ESLint issues
npm run format           # Format with Prettier
npm run format:check     # Check formatting
```

### Mobile App Commands (React Native)
```bash
cd frontend-mobile-rn
npm install              # Install dependencies
npm start                # Start Metro bundler
npx react-native run-ios    # Run on iOS
npx react-native run-android # Run on Android
```

## Code Style and Patterns

### Go Code Style

#### Import Organization
```go
import (
    // Standard library
    "context"
    "fmt"
    "net/http"
    "time"

    // External dependencies
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.uber.org/zap"

    // Internal packages
    "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
    "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
    "github.com/gmsas95/blytz-mvp/services/auth-service/internal/models"
)
```

#### Naming Conventions
- **Package names**: lowercase, single word when possible (`auth`, `models`, `handlers`)
- **Struct names**: PascalCase (`User`, `AuthService`, `LoginRequest`)
- **Interface names**: PascalCase, often ending with `er` (`Authenticator`, `Validator`)
- **Function names**: PascalCase for exported, camelCase for unexported
- **Constants**: UPPER_SNAKE_CASE for exported, camelCase for unexported
- **Variables**: camelCase, short names for local scope (`u`, `ctx`, `req`)

#### Error Handling Pattern
```go
// Use shared error types
import "github.com/gmsas95/blytz-mvp/shared/pkg/errors"

// Create errors with proper types
if err != nil {
    return errors.ValidationError("INVALID_EMAIL", "Invalid email format")
}

// Handle errors in handlers
if err := h.authService.RegisterUser(&user); err != nil {
    utils.SendErrorResponse(c, err)
    return
}
```

#### Response Pattern
```go
// Use shared response utilities
import "github.com/gmsas95/blytz-mvp/shared/pkg/utils"

// Success response
utils.SendSuccessResponse(c, http.StatusOK, userData)

// Error response
utils.SendErrorResponse(c, errors.ErrInvalidRequestBody)

// Custom response
utils.SendJSON(c, http.StatusCreated, &utils.Response{
    Success: true,
    Message: "User created successfully",
    Data: userData,
})
```

#### Service Structure Pattern
```
services/service-name/
├── main.go                 # Entry point
├── go.mod                  # Go module
├── go.sum                  # Dependencies
├── Dockerfile              # Docker configuration
├── internal/               # Private application code
│   ├── api/
│   │   ├── handlers/       # HTTP handlers
│   │   ├── routes/         # Route definitions
│   │   └── middleware/     # HTTP middleware
│   ├── services/           # Business logic
│   ├── models/             # Data models
│   ├── config/             # Configuration
│   └── database/           # Database setup
├── pkg/                    # Public library code
├── tests/                  # Integration tests
└── scripts/                # Utility scripts
```

### TypeScript/React Code Style

#### Import Organization
```typescript
// React and Next.js
import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import type { NextPage } from 'next';

// External libraries
import axios from 'axios';
import { zodResolver } from '@hookform/resolvers/zod';

// Internal modules (alias @/)
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/use-auth';
import { apiClient } from '@/lib/api-client';
import type { User } from '@/types/user';
```

#### Component Pattern
```typescript
// Use functional components with TypeScript
interface UserProfileProps {
  userId: string;
  onUpdate?: (user: User) => void;
}

const UserProfile: React.FC<UserProfileProps> = ({ userId, onUpdate }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // Component logic here

  return (
    <div className="user-profile">
      {/* JSX content */}
    </div>
  );
};

export default UserProfile;
```

## Testing Patterns

### Go Testing
```go
// Use testify for assertions
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestAuthService_RegisterUser(t *testing.T) {
    // Arrange
    service := NewAuthService(testDB, testConfig)
    user := &models.User{
        Email: "test@example.com",
        Password: "password123",
    }

    // Act
    err := service.RegisterUser(user)

    // Assert
    assert.NoError(t, err)
    
    // Verify user was created
    createdUser, err := service.GetUserByEmail(user.Email)
    assert.NoError(t, err)
    assert.Equal(t, user.Email, createdUser.Email)
}
```

### React Testing
```typescript
// Use React Testing Library
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { AuthProvider } from '@/contexts/auth-context';
import { LoginForm } from '@/components/auth/login-form';

test('should submit login form', async () => {
    render(
        <AuthProvider>
            <LoginForm />
        </AuthProvider>
    );

    fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'test@example.com' }
    });

    fireEvent.click(screen.getByRole('button', { name: /login/i }));

    await waitFor(() => {
        expect(screen.getByText(/welcome/i)).toBeInTheDocument();
    });
});
```

## Configuration Management

### Environment Variables
```bash
# Go services use .env files
DATABASE_URL=postgresql://user:pass@localhost:5432/db
JWT_SECRET=your-secret-key
REDIS_URL=redis://localhost:6379
PORT=8085

# Frontend uses .env.local
NEXT_PUBLIC_API_URL=http://localhost:8085
NEXT_PUBLIC_WS_URL=ws://localhost:8085
```

### Go Configuration Pattern
```go
type Config struct {
    DatabaseURL string `env:"DATABASE_URL" envDefault:"localhost:5432"`
    JWTSecret   string `env:"JWT_SECRET" envDefault:"secret"`
    Port        string `env:"PORT" envDefault:"8085"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

## Database Patterns

### GORM Models
```go
type User struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    Email     string    `gorm:"uniqueIndex;not null" json:"email"`
    Password  string    `gorm:"not null" json:"-"`
    Name      string    `gorm:"not null" json:"name"`
    Role      string    `gorm:"default:user" json:"role"`
    IsActive  bool      `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Auto-migration
db.AutoMigrate(&User{})
```

## API Design Patterns

### REST API Structure
```go
// Standard route patterns
GET    /api/v1/users           // List users
GET    /api/v1/users/:id       // Get user
POST   /api/v1/users           // Create user
PUT    /api/v1/users/:id       // Update user
DELETE /api/v1/users/:id       // Delete user

// Nested resources
GET    /api/v1/users/:id/orders    // Get user's orders
POST   /api/v1/users/:id/orders    // Create order for user
```

### Request/Response Patterns
```go
// Request DTOs
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// Response DTOs
type UserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}
```

## Security Patterns

### Authentication Middleware
```go
// Use shared auth middleware
import "github.com/gmsas95/blytz-mvp/shared/pkg/auth"

// Protected routes
protected := router.Group("/api/v1")
protected.Use(auth.GinAuthMiddleware(authClient))
{
    protected.GET("/profile", h.GetProfile)
    protected.PUT("/profile", h.UpdateProfile)
}
```

### Input Validation
```go
// Use Gin binding for validation
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,containsany=!@#$%^&*"`
}

// Custom validation
func ValidatePassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    return len(password) >= 8 && containsUpper(password) && containsNumber(password)
}
```

## Performance Patterns

### Database Queries
```go
// Use prepared statements
stmt, err := db.Prepare("SELECT * FROM users WHERE email = $1")
if err != nil {
    return err
}
defer stmt.Close()

// Use transactions
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// Batch operations
var users []User
db.FindInBatches(&users, 100, func(tx *gorm.DB, batch int) error {
    // Process batch
    return nil
})
```

### Caching Pattern
```go
// Redis caching
func (s *UserService) GetUserByID(id string) (*User, error) {
    // Try cache first
    cached, err := s.redis.Get(ctx, "user:"+id).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // Fallback to database
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // Cache the result
    data, _ := json.Marshal(user)
    s.redis.Set(ctx, "user:"+id, data, time.Hour)

    return user, nil
}
```

## Development Workflow

### Git Workflow
```bash
# Feature branch workflow
git checkout -b feature/user-authentication
git add .
git commit -m "feat: implement JWT authentication"
git push origin feature/user-authentication

# Commit message format
feat: new feature
fix: bug fix
docs: documentation
style: formatting
refactor: code refactoring
test: adding tests
chore: maintenance
```

### Local Development Setup
```bash
# Start all services
docker-compose up -d

# Start individual service
cd services/auth-service
go run main.go

# Run tests
./scripts/test.sh

# Check service health
curl http://localhost:8085/health
```

## Common Issues and Solutions

### Go Module Issues
```bash
# Fix module dependencies
go mod tidy
go mod download

# Fix workspace issues
go work sync

# Clear module cache
go clean -modcache
```

### Database Connection Issues
```bash
# Check database connection
psql $DATABASE_URL

# Reset database
go run scripts/reset-db.go

# Run migrations
go run scripts/migrate.go up
```

### Frontend Build Issues
```bash
# Clear Next.js cache
rm -rf .next

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install

# Fix TypeScript issues
npm run type-check
```

## Monitoring and Debugging

### Logging Pattern
```go
// Use structured logging with zap
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
logger.Info("User login",
    zap.String("user_id", userID),
    zap.String("email", email),
    zap.Duration("duration", time.Since(start)),
)
```

### Health Checks
```go
// Standard health endpoint
func (h *HealthHandler) Check(c *gin.Context) {
    status := utils.NewHealthStatus("ok")
    status.AddService("database", h.checkDatabase())
    status.AddService("redis", h.checkRedis())
    
    utils.SendSuccessResponse(c, http.StatusOK, status)
}
```

## Deployment

### Docker Pattern
```dockerfile
# Multi-stage builds
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/service .
EXPOSE 8085
CMD ["./service"]
```

### Environment-Specific Config
```go
// Load config based on environment
func LoadConfig() (*Config, error) {
    env := os.Getenv("ENV")
    if env == "production" {
        return loadProductionConfig()
    }
    return loadDevelopmentConfig()
}
```

---

## Quick Reference Commands

### Essential Commands
```bash
# Go services
go run main.go                    # Run service
go test ./... -v                 # Run tests
go build .                       # Build service
go mod tidy                      # Clean dependencies

# Frontend
npm run dev                      # Start dev server
npm run build                    # Build for production
npm run test                     # Run tests
npm run lint                     # Check code quality

# Testing
./scripts/test.sh                # Test all services
npm run test:e2e                 # E2E tests

# Docker
docker-compose up -d             # Start all services
docker-compose logs -f service    # View logs
docker-compose restart service   # Restart service
```

### Service Ports
- Auth Service: 8085
- Product Service: 8086
- Auction Service: 8087
- Order Service: 8088
- Payment Service: 8089
- Chat Service: 8090
- Logistics Service: 8091
- Gateway Service: 8092
- LiveKit Service: 8093
- Notification Service: 8094
- Frontend: 3000

This guide should help AI agents understand the codebase structure, patterns, and workflows for effective development assistance.
