#!/bin/bash

# Blytz Platform Cleanup Script
# This script removes unnecessary files and reorganizes the project structure

set -e

echo "🧹 Starting Blytz Platform Cleanup..."

# Root directory to clean
ROOT_DIR="/home/sas/blytzmvp-clean"
cd "$ROOT_DIR"

# Create backup directory for important files
BACKUP_DIR="${ROOT_DIR}/cleanup_backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

echo "📦 Created backup directory: $BACKUP_DIR"

# Function to backup important files before deletion
backup_file() {
    local file="$1"
    if [[ -f "$file" ]]; then
        cp "$file" "$BACKUP_DIR/$(basename "$file")"
        echo "  Backed up: $(basename "$file")"
    fi
}

echo ""
echo "🗑️  Removing unnecessary markdown files..."

# Remove redundant documentation files but keep essential ones
REDUNDANT_MD=(
    "AGENTS_VERIFICATION_REPORT.md"
    "CLEANUP_SUMMARY.md"
    "CODEBASE_AUDIT_REPORT.md"
    "COMMIT_READY.md"
    "CURRENT_STATUS.md"
    "DEPLOYMENT_TRIGGER.md"
    "DEVELOPMENT_ROADMAP.md"
    "EXHIBITION_DAY_GUIDE.md"
    "FE Refactor.md"
    "FIREBASE_INTEGRATION_COMPLETE.md"
    "FLUTTER_STACK_GUIDE.md"
    "FRONTEND_API_UPDATE_RECOMMENDATIONS.md"
    "FRONTEND_BACKEND_API_MISMATCHES.md"
    "FRONTEND_BACKEND_INTEGRATION_FIX_SUMMARY.md"
    "FRONTEND_ENVIRONMENT_SETUP.md"
    "LOCAL_TESTING_GUIDE.md"
    "MVP_STATUS.md"
    "PHASE1_FIXES_SUMMARY.md"
    "PRE_COMMIT_CHECKLIST.md"
    "PROJECT_COMPLETION.md"
    "RELIABILITY_FIXES.md"
    "SHARED_PACKAGE_MIGRATION_FIX.md"
    "TECHNICAL_DEBT.md"
    "TEST_RESULTS_PHASE1.md"
    "auth_analysis.md"
)

for md_file in "${REDUNDANT_MD[@]}"; do
    if [[ -f "$md_file" ]]; then
        backup_file "$md_file"
        rm "$md_file"
        echo "  Removed: $md_file"
    fi
done

echo ""
echo "🗑️  Removing redundant test and deployment scripts..."

# Remove redundant scripts but keep essential ones
REDUNDANT_SCRIPTS=(
    "deploy-all-services.sh"
    "deploy-to-vps.sh"
    "deploy-nginx-config.sh"
    "diagnose-502-errors.sh"
    "fix-502-deployment-guide.sh"
    "fix-502-errors.sh"
    "fix-go-mods.sh"
    "setup-demo-seller.sh"
    "setup-hostinger-commands.sh"
    "setup-firebase-local.sh"
    "start-exhibition.sh"
    "dokploy-deploy.sh"
)

for script in "${REDUNDANT_SCRIPTS[@]}"; do
    if [[ -f "$script" ]]; then
        backup_file "$script"
        rm "$script"
        echo "  Removed: $script"
    fi
done

echo ""
echo "🗑️  Removing integration test scripts (keeping core tests)..."

# Remove integration test scripts but keep essential test infrastructure
find tests/integration -name "*.sh" -type f ! -name "final_test.sh" -exec rm {} \;
echo "  Removed redundant integration test scripts"

echo ""
echo "🗑️  Cleaning up service-specific build artifacts..."

# Remove build artifacts and temporary files
find services -name "*.out" -delete 2>/dev/null || true
find services -name "coverage.out" -delete 2>/dev/null || true
find services -name "main" -type f -delete 2>/dev/null || true
find services -name "auth-service" -type f -delete 2>/dev/null || true
find services -name "gateway" -type f -delete 2>/dev/null || true

echo ""
echo "🗑️  Cleaning up frontend build artifacts..."

# Remove frontend build directories and temporary files
rm -rf frontend/.next 2>/dev/null || true
rm -rf frontend-demo/.next 2>/dev/null || true
rm -rf frontend-seller/.next 2>/dev/null || true
find frontend -name "*.tsbuildinfo" -delete 2>/dev/null || true
find frontend-demo -name "*.tsbuildinfo" -delete 2>/dev/null || true
find frontend-seller -name "*.tsbuildinfo" -delete 2>/dev/null || true

echo ""
echo "🗑️  Cleaning up Flutter build artifacts..."

# Remove Flutter build artifacts
rm -rf blytz_flutter_app/.dart_tool 2>/dev/null || true
rm -rf blytz_flutter_app/build 2>/dev/null || true
rm -rf blytz_flutter_app/android/.gradle 2>/dev/null || true
rm -rf blytz_flutter_app/linux/flutter 2>/dev/null || true
find blytz_flutter_app -name "*.lock" -delete 2>/dev/null || true

echo ""
echo "🗑️  Removing Docker development files..."

# Remove redundant docker-compose files
REDUNDANT_DOCKER=(
    "docker-compose.dev.yml"
    "docker-compose.prod.yml"
    "docker-compose.dokploy.yml"
    "docker-compose.hostinger.yml"
    "docker-compose.simple.yml"
    "docker-compose.cloud-livekit.yml"
    "docker-compose-demo-seller.yml"
    "docker-compose.add-services.yml"
)

for docker_file in "${REDUNDANT_DOCKER[@]}"; do
    if [[ -f "$docker_file" ]]; then
        backup_file "$docker_file"
        rm "$docker_file"
        echo "  Removed: $docker_file"
    fi
done

echo ""
echo "🗑️  Cleaning up agent documentation..."

# Remove Claude agent documentation (keep AGENTS.md as reference)
if [[ -d ".claude" ]]; then
    backup_file ".claude"
    rm -rf ".claude"
    echo "  Removed: .claude directory"
fi

echo ""
echo "🗑️  Removing dynamic and temporary files..."

# Remove temporary and dynamic files
rm -rf dynamic/ 2>/dev/null || true
rm -f acme-temp.json 2>/dev/null || true

echo ""
echo "🗑️  Cleaning up planner directory..."

# Remove planner documentation
if [[ -d "planner" ]]; then
    backup_file "planner"
    rm -rf planner
    echo "  Removed: planner directory"
fi

echo ""
echo "🗑️  Removing redundant environment files..."

# Remove redundant environment files
REDUNDANT_ENV=(
    ".env.dev"
    ".env.production.example"
    ".env.staging.example"
)

for env_file in "${REDUNDANT_ENV[@]}"; do
    if [[ -f "$env_file" ]]; then
        backup_file "$env_file"
        rm "$env_file"
        echo "  Removed: $env_file"
    fi
done

echo ""
echo "🗑️  Cleaning up Git hooks..."

# Remove Git hooks (can be regenerated)
if [[ -d ".husky" ]]; then
    rm -rf .husky
    echo "  Removed: .husky directory"
fi

# Remove frontend Git hooks
find frontend frontend-demo frontend-seller -name ".husky" -type d -exec rm -rf {} \; 2>/dev/null || true

echo ""
echo "📁 Reorganizing project structure..."

# Create organized directory structure
mkdir -p docs/{api,deployment,development}
mkdir -p scripts/{deployment,testing,utilities}

# Move documentation to organized structure
if [[ -d "openapi" ]]; then
    mv openapi docs/api/
    echo "  Moved: openapi -> docs/api/"
fi

if [[ -d "grafana" ]]; then
    mv grafana docs/deployment/
    echo "  Moved: grafana -> docs/deployment/"
fi

# Move remaining scripts to organized structure
if [[ -d "scripts" ]]; then
    mv scripts scripts/utilities/
    echo "  Moved: scripts -> scripts/utilities/"
fi

if [[ -d "tests" ]]; then
    mv tests scripts/testing/
    echo "  Moved: tests -> scripts/testing/"
fi

# Move deployment scripts
for deploy_script in *.sh; do
    if [[ -f "$deploy_script" && "$deploy_script" == *deploy* ]]; then
        mv "$deploy_script" scripts/deployment/
        echo "  Moved: $deploy_script -> scripts/deployment/"
    fi
done

echo ""
echo "🔧 Creating simplified Docker Compose configuration..."

# Create a single, clean docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  # Database Services
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: blytz_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Backend Services
  auth-service:
    build: ./services/auth-service
    ports:
      - "8084:8084"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/auth_db
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=your-secret-key
      - BETTER_AUTH_SECRET=your-better-auth-secret
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  product-service:
    build: ./services/product-service
    ports:
      - "8082:8082"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/products_db
    depends_on:
      postgres:
        condition: service_healthy

  auction-service:
    build: ./services/auction-service
    ports:
      - "8083:8083"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/auction_db
      - REDIS_URL=redis://redis:6379
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  order-service:
    build: ./services/order-service
    ports:
      - "8085:8085"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/orders_db
    depends_on:
      postgres:
        condition: service_healthy

  payment-service:
    build: ./services/payment-service
    ports:
      - "8086:8086"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/payments_db
    depends_on:
      postgres:
        condition: service_healthy

  chat-service:
    build: ./services/chat-service
    ports:
      - "8088:8088"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/chat_db
    depends_on:
      postgres:
        condition: service_healthy

  logistics-service:
    build: ./services/logistics-service
    ports:
      - "8087:8087"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/logistics_db
    depends_on:
      postgres:
        condition: service_healthy

  # API Gateway
  gateway:
    build: ./services/gateway
    ports:
      - "8080:8080"
    depends_on:
      - auth-service
      - product-service
      - auction-service
      - order-service
      - payment-service
      - chat-service
      - logistics-service

  # Frontend
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    depends_on:
      - gateway

volumes:
  postgres_data:
  redis_data:
EOF

echo "  Created: simplified docker-compose.yml"

echo ""
echo "📝 Creating essential documentation..."

# Create concise development guide
cat > docs/development/SETUP.md << 'EOF'
# Development Setup

## Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- Flutter SDK (for mobile app)

## Quick Start
```bash
# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8080/health
```

## Individual Services
- Auth Service: http://localhost:8084
- Product Service: http://localhost:8082
- Auction Service: http://localhost:8083
- Order Service: http://localhost:8085
- Payment Service: http://localhost:8086
- Chat Service: http://localhost:8088
- Logistics Service: http://localhost:8087
- Gateway: http://localhost:8080
- Frontend: http://localhost:3000

## Local Development
See individual service directories for development instructions.
EOF

echo "  Created: docs/development/SETUP.md"

# Create deployment guide
cat > docs/deployment/DEPLOYMENT.md << 'EOF'
# Deployment Guide

## Production Deployment
```bash
# Set environment variables
export ENVIRONMENT=production
export JWT_SECRET=your-production-secret
export BETTER_AUTH_SECRET=your-production-auth-secret

# Deploy with production configuration
docker-compose -f docker-compose.yml up -d
```

## Environment Variables
- `ENVIRONMENT`: development/production
- `JWT_SECRET`: JWT signing secret
- `BETTER_AUTH_SECRET`: Better Auth secret
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string

## SSL/TLS
Configure SSL certificates using your preferred method (Let's Encrypt, etc.)
EOF

echo "  Created: docs/deployment/DEPLOYMENT.md"

echo ""
echo "📊 Cleanup Summary..."

# Count files before and after
BEFORE_FILES=$(find "$BACKUP_DIR" -type f | wc -l)
AFTER_FILES=$(find . -type f ! -path "./cleanup_backup_*" | wc -l)
DELETED_FILES=$((BEFORE_FILES - AFTER_FILES))

echo "  Files backed up: $BEFORE_FILES"
echo "  Files remaining: $AFTER_FILES"
echo "  Files removed: $DELETED_FILES"
echo ""
echo "✅ Cleanup completed successfully!"
echo "📦 Backup location: $BACKUP_DIR"
echo ""
echo "🚀 Your project is now clean and reorganized for the new repository!"
echo ""
echo "Next steps:"
echo "1. Review the simplified structure"
echo "2. Test docker-compose up -d"
echo "3. Update any remaining configuration files"
echo "4. Commit to your new repository"