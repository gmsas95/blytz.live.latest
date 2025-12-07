# Blytz Live Auction Platform - AI Agent Guide

## Project Overview

Blytz is a production-ready microservices-based live auction platform built with Go and modern cloud-native technologies. The platform enables real-time bidding, chat functionality, and payment processing with WebRTC streaming capabilities.

**Key Architecture Decisions:**
- Microservices architecture with 9 core services (8 business services + 1 gateway)
- Self-hosted authentication system (Better Auth) for cost optimization
- Redis-backed real-time bidding system with atomic Lua scripts
- PostgreSQL for persistent data storage
- Docker containerization with multi-stage builds
- LiveKit integration for WebRTC streaming
- Go workspace configuration with 10 modules (9 services + shared)

## Technology Stack

### Backend Services
- **Language**: Go 1.25 with workspace configuration (go.work)
- **Web Framework**: Gin (HTTP router and middleware)
- **Database**: PostgreSQL 15 with GORM ORM
- **Cache**: Redis 7 for session management and real-time data
- **Authentication**: Better Auth (self-hosted JWT-based system)
- **Monitoring**: Prometheus metrics collection
- **Logging**: Uber Zap structured logging
- **Real-time**: LiveKit for WebRTC streaming

### Frontend Applications
- **Main Frontend**: Next.js 14.2.33 with React 18 (Port 3000)
- **Demo Frontend**: Viewer platform with LiveKit integration (Port 3001)
- **Seller Frontend**: Broadcaster platform for auction hosts (Port 3002)
- **UI Framework**: Radix UI components with Tailwind CSS
- **State Management**: TanStack Query for data fetching
- **Testing**: Jest, Playwright for E2E testing with accessibility support

### Infrastructure
- **Containerization**: Docker with multi-stage builds
- **Orchestration**: Docker Compose (9 different configurations)
- **Deployment**: Dokploy for VPS deployment
- **Reverse Proxy**: Nginx and Traefik configurations
- **Real-time Communication**: LiveKit for WebRTC streaming
- **SSL/TLS**: Let's Encrypt integration

## Project Structure

```
/home/sas/blytzmvp-clean/
├── services/                    # 9 microservices
│   ├── auth-service/           # Port 8084 - JWT authentication
│   ├── product-service/        # Port 8082 - Product catalog
│   ├── auction-service/        # Port 8083 - Auction management
│   ├── chat-service/           # Port 8088 - Real-time chat
│   ├── order-service/          # Port 8085 - Order processing
│   ├── payment-service/        # Port 8086 - Payment processing
│   ├── logistics-service/      # Port 8087 - Shipping management
│   ├── gateway/                # Port 8080 - API gateway
│   └── livekit-service/        # WebRTC streaming service
├── shared/                     # Shared packages (CRITICAL: /pkg/ structure)
│   └── pkg/
│       ├── auth/               # Authentication client
│       ├── errors/             # Error handling
│       ├── middleware/         # Common middleware
│       ├── utils/              # Utility functions
│       └── constants/          # Shared constants
├── frontend/                   # Main Next.js frontend (v2.0.1)
├── frontend-demo/              # Demo viewer frontend
├── frontend-seller/            # Seller broadcaster frontend
├── functions/                  # Firebase cloud functions
├── infra/                      # Infrastructure configurations
├── scripts/                    # 19 utility scripts
├── openapi/                    # API specifications
├── grafana/                    # Monitoring dashboards
└── docker-compose*.yml         # 9 deployment configurations
```

## Development Commands

### Local Development
```bash
# Start all services with Docker Compose
docker-compose up -d

# Test individual service
cd services/auth-service/cmd && go run main.go

# Run comprehensive tests
./test-microservices.sh

# Validate architecture
./validate-architecture.sh

# Use Makefile for CI/CD pipeline simulation
make ci-pipeline

# Setup Go environment
source setup-go-env.sh
```

### Build and Test
```bash
# Test specific service (from Makefile)
make test SERVICE=auth-service

# Test all services in parallel
make test-all

# Build Docker images
make build-all

# Health check
make health-check

# Load testing
make load-test

# Full CI pipeline
make ci-pipeline
```

### Frontend Development
```bash
# Main frontend
cd frontend
npm install
npm run dev        # Start development server on port 3000
npm run build      # Build for production
npm run lint       # Run ESLint
npm run format     # Format code with Prettier
npm run test:e2e   # Run Playwright E2E tests

# Demo frontend (port 3001)
cd frontend-demo
npm run dev

# Seller frontend (port 3002)
cd frontend-seller
npm run dev
```

## Code Style Guidelines

### Go Code Standards
- **Package Structure**: Each service follows `cmd/main.go` entry point pattern
- **Error Handling**: Use wrapped errors with context (`fmt.Errorf("failed to...: %w", err)`)
- **Logging**: Use Uber Zap for structured logging with contextual fields
- **Configuration**: Environment-based configuration with validation
- **HTTP Responses**: Consistent JSON response format with error codes
- **Import Paths**: Always use `/pkg/` prefix for shared imports (e.g., `github.com/gmsas95/blytz-mvp/shared/pkg/auth`)
- **Go Version**: Go 1.25 with workspace configuration

### Frontend Standards
- **Component Structure**: Use Radix UI primitives with Tailwind CSS
- **State Management**: TanStack Query for server state, React hooks for local state
- **TypeScript**: Strict mode enabled, comprehensive type definitions
- **Code Formatting**: Prettier with consistent configuration
- **Component Naming**: PascalCase for components, camelCase for functions/variables
- **Testing**: Jest for unit tests, Playwright for E2E with accessibility testing

### API Design
- **RESTful Principles**: Resource-based URLs with proper HTTP methods
- **Versioning**: `/api/v1/` prefix for all endpoints
- **Authentication**: JWT Bearer tokens in Authorization header
- **Response Format**: Standardized JSON with `data`, `error`, `message` fields
- **Error Codes**: Consistent HTTP status codes with detailed error messages

## Authentication System

The platform uses a self-hosted Better Auth system providing:
- **Cost Savings**: 97% reduction vs Firebase Auth ($480-585/month saved)
- **Performance**: Local queries (~5ms) vs external API calls (~100ms)
- **Security**: Complete data ownership, no vendor lock-in
- **Features**: JWT-based auth with registration, login, token refresh

**Integration Pattern:**
```go
import "github.com/gmsas95/blytz-mvp/shared/pkg/auth"

authClient := auth.NewAuthClient("http://auth-service:8084")
router.Use(auth.GinAuthMiddleware(authClient))
```

**Protected Endpoints:**
- `POST /api/v1/auctions` - Create auction
- `PUT /api/v1/auctions/:id` - Update auction
- `DELETE /api/v1/auctions/:id` - Delete auction
- `POST /api/v1/auctions/:id/bids` - Place bid

## Testing Strategy

### Unit Testing
- Go testing with race detection and coverage reporting
- Test files follow `*_test.go` naming convention
- Mock external dependencies and database calls
- Aim for >80% code coverage on critical paths

### Integration Testing
- Service-to-service communication testing
- Database integration tests with test containers
- API endpoint testing with realistic payloads
- Authentication flow validation

### Frontend Testing
- **Jest**: Unit testing with jsdom environment
- **Playwright**: E2E testing with multiple browsers
- **Accessibility**: Axe-core integration for a11y testing
- **Visual Regression**: Snapshot testing capabilities
- **Coverage**: Comprehensive coverage reporting

### Load Testing
- K6 framework for performance validation
- Concurrent user simulation for auction bidding
- Database query performance testing
- Memory and CPU usage monitoring

### Health Checks
All services expose `/health` endpoints for monitoring:
```bash
curl http://localhost:8084/health  # Auth service
curl http://localhost:8082/health  # Product service
curl http://localhost:8083/health  # Auction service
curl http://localhost:8085/health  # Order service
curl http://localhost:8086/health  # Payment service
curl http://localhost:8087/health  # Logistics service
curl http://localhost:8088/health  # Chat service
curl http://localhost:8080/health  # Gateway service
```

## Deployment Process

### Local Development
1. **Firebase Emulators**: Start local Firebase functions
2. **Docker Compose**: Orchestrate all services with dependencies
3. **Hot Reloading**: Frontend development with auto-refresh
4. **Health Monitoring**: Built-in service health checks

### Production Deployment
1. **Dokploy Configuration**: VPS deployment with SSL/TLS
2. **Database Setup**: PostgreSQL with proper backups
3. **Environment Variables**: Production-specific configurations
4. **SSL Certificates**: Let's Encrypt for HTTPS
5. **Monitoring Setup**: Prometheus and Grafana integration

### Multi-Environment Support
- **Development**: Local Docker Compose setup
- **Staging**: VPS with production-like configuration
- **Production**: Full production deployment with monitoring

### Docker Configurations
- **docker-compose.yml**: Main production configuration
- **docker-compose.dev.yml**: Development environment
- **docker-compose.prod.yml**: Production-specific setup
- **docker-compose.hostinger.yml**: VPS deployment
- **docker-compose.dokploy.yml**: Dokploy deployment
- **docker-compose.simple.yml**: Simplified setup
- **docker-compose.cloud-livekit.yml**: Cloud LiveKit integration
- **docker-compose-demo-seller.yml**: Demo seller setup
- **docker-compose-add-services.yml**: Additional services

## Security Considerations

### Authentication & Authorization
- JWT tokens with HS256 algorithm
- Token expiration and refresh mechanisms
- Protected endpoints with middleware validation
- User session management with Redis

### Data Protection
- Environment variables for sensitive configuration
- Database connection encryption
- Input validation and sanitization
- Rate limiting on API endpoints

### Infrastructure Security
- Docker container security best practices
- Network isolation between services
- SSL/TLS encryption for all communications
- Regular security updates and patches

## Common Development Patterns

### Service Implementation Pattern
1. **Configuration Loading**: Environment-based config with validation
2. **Database Connection**: GORM with connection pooling
3. **Router Setup**: Gin with middleware chain
4. **Handler Functions**: Consistent error handling and logging
5. **Graceful Shutdown**: Signal handling for clean service termination

### Error Handling Pattern
```go
if err != nil {
    logger.Error("Failed to process request", zap.Error(err))
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "internal_server_error",
        "message": "Failed to process request",
    })
    return
}
```

### Database Operations
- Use GORM for ORM functionality
- Implement repository pattern for data access
- Handle transactions for complex operations
- Use database migrations for schema changes

## Monitoring and Observability

### Metrics Collection
- Prometheus metrics for all services
- Custom business metrics for auctions and bids
- Database query performance metrics
- HTTP request/response metrics

### Logging Standards
- Structured logging with Uber Zap
- Contextual logging with request IDs
- Log levels: debug, info, warn, error
- Centralized log aggregation in production

### Health Monitoring
- Service health endpoints with detailed status
- Database connectivity checks
- External service dependency checks
- Automatic alerting for service failures

## Troubleshooting Guide

### Common Issues
1. **Service Won't Start**: Check Docker logs, database connectivity, environment variables
2. **Database Connection Issues**: Verify PostgreSQL is running, check connection strings
3. **Build Failures**: Ensure Go 1.25+ is installed, run `go mod tidy`, verify shared imports
4. **Authentication Issues**: Check auth service on port 8084, verify JWT secrets
5. **Frontend Build Issues**: Ensure Node.js 18+, check package.json dependencies

### Debug Commands
```bash
# Check service logs
docker-compose logs <service-name>

# Test service connectivity
curl http://localhost:<port>/health

# Check database connectivity
docker-compose exec postgres psql -U postgres -d auth_db

# Validate shared package structure
./verify-shared-migration.sh

# Check Go workspace
go work sync

# Test individual service
cd services/auth-service/cmd && go run main.go
```

## Performance Optimization

### Database Optimization
- Proper indexing on frequently queried columns
- Connection pooling configuration
- Query optimization and caching strategies
- Database maintenance and cleanup procedures

### Service Optimization
- Efficient JSON serialization/deserialization
- Connection reuse for external services
- Memory management and garbage collection tuning
- CPU profiling and bottleneck identification

### Frontend Optimization
- Code splitting and lazy loading
- Image optimization and CDN usage
- Caching strategies for API responses
- Bundle size optimization

## Shared Package Architecture (CRITICAL)

**All shared packages now live in `/shared/pkg/` - this is MANDATORY for consistency.**

```go
// ✅ CORRECT - Use pkg/ prefix
import "github.com/gmsas95/blytz-mvp/shared/pkg/auth"
import "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
import "github.com/gmsas95/blytz-mvp/shared/pkg/utils"

// ❌ INCORRECT - Will cause build failures
import "github.com/gmsas95/blytz-mvp/shared/auth"
import "github.com/gmsas95/blytz-mvp/shared/errors"
```

**Package Structure:**
- `shared/pkg/auth/` - Authentication client and middleware
- `shared/pkg/errors/` - Common error types and handling
- `shared/pkg/utils/` - Utility functions (logger, validation, responses)
- `shared/pkg/constants/` - Shared constants and configurations
- `shared/pkg/middleware/` - Common middleware functions

## Authentication Integration Pattern

**For any new microservice, follow this exact pattern:**

```go
import "github.com/gmsas95/blytz-mvp/shared/pkg/auth"

// In router setup
authClient := auth.NewAuthClient("http://auth-service:8084")

// Public routes (no auth required)
public := router.Group("/api/v1/public")
{
    public.GET("/health", healthHandler)
}

// Protected routes (auth required)
protected := router.Group("/api/v1/protected")
protected.Use(auth.GinAuthMiddleware(authClient))
{
    protected.POST("/create", createHandler)  // Requires auth
}

// In handlers
userID := c.GetString("userID")  // Get authenticated user
```

## Service Communication
- **Auth Service**: Port 8084 - JWT token validation
- **Service-to-Service**: Use shared auth client for internal calls
- **External APIs**: All protected endpoints require Bearer token
- **Gateway**: Port 8080 - Routes requests to appropriate services

## Environment Configuration

### Required Environment Files
- `.env` - Main environment configuration
- `.env.dev` - Development environment
- `.env.production` - Production environment
- `.env.staging` - Staging environment

### Key Environment Variables
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing secret
- `BETTER_AUTH_SECRET` - Better Auth secret
- `REDIS_URL` - Redis connection string
- `FIUU_*` - Payment gateway configuration

## Build and Deployment Scripts

### Core Scripts
- `test-microservices.sh` - Test service startup and compilation
- `validate-architecture.sh` - Validate overall architecture
- `verify-shared-migration.sh` - Ensure shared package consistency
- `setup-go-env.sh` - Setup Go environment
- `deploy-all-services.sh` - Deploy all services
- `start-emulator.sh` - Start Firebase emulators

### CI/CD Pipeline
- GitHub Actions workflow for automated testing
- Makefile for local CI/CD simulation
- Dokploy configuration for VPS deployment
- Multi-stage Docker builds for optimization

This guide serves as the comprehensive reference for AI agents working on the Blytz platform. Always refer to the actual codebase and existing patterns when implementing new features or fixing issues.