#!/bin/bash

# ===========================================
# Blytz Live Auction Platform - Secret Generator
# ===========================================
# This script generates secure secrets for production use
# ===========================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to generate secure random string
generate_secret() {
    local length=${1:-32}
    openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length
}

# Function to generate JWT secret (URL-safe base64)
generate_jwt_secret() {
    openssl rand -base64 32 | tr -d '+/' | tr -d '='
}

# Function to generate hex secret
generate_hex_secret() {
    local length=${1:-64}
    openssl rand -hex $length
}

print_info "🔐 Blytz Platform - Secret Generator"
print_info "===================================="
echo

# Check if OpenSSL is available
if ! command -v openssl &> /dev/null; then
    print_error "OpenSSL is required but not installed. Please install OpenSSL."
    exit 1
fi

# Create .env.production file
ENV_FILE=".env.production"
print_info "Creating production environment file: $ENV_FILE"

cat > $ENV_FILE << EOF
# ===========================================
# Blytz Live Auction Platform - Production Secrets
# ===========================================
# Generated on: $(date)
# WARNING: These are production secrets. Keep them secure!
# ===========================================

# ===========================================
# AUTHENTICATION SECRETS (SECURE - GENERATED)
# ===========================================
JWT_SECRET=$(generate_jwt_secret)
BETTER_AUTH_SECRET=$(generate_jwt_secret)
SESSION_SECRET=$(generate_jwt_secret)
ENCRYPTION_KEY=$(generate_hex_secret 32)

# ===========================================
# DATABASE SECRETS
# ===========================================
POSTGRES_PASSWORD=$(generate_secret 24)
DATABASE_URL=postgres://blytz:\${POSTGRES_PASSWORD}@postgres:5432/blytz_prod?sslmode=require

# ===========================================
# REDIS SECRETS
# ===========================================
REDIS_PASSWORD=$(generate_secret 32)
REDIS_URL=redis://:\${REDIS_PASSWORD}@redis:6379

# ===========================================
# PAYMENT GATEWAY SECRETS
# ===========================================
FIUU_MERCHANT_ID=your_fiuu_merchant_id_here
FIUU_VERIFY_KEY=your_fiuu_verify_key_here

# ===========================================
# LIVEKIT SECRETS
# ===========================================
LIVEKIT_API_KEY=$(generate_secret 16)
LIVEKIT_API_SECRET=$(generate_secret 32)

# ===========================================
# NINJAVAN SECRETS
# ===========================================
NINJAVAN_CLIENT_ID=your_ninjavan_client_id_here
NINJAVAN_CLIENT_SECRET=your_ninjavan_client_secret_here

# ===========================================
# SECURITY SETTINGS
# ===========================================
CORS_ORIGINS=https://blytz.app,https://www.blytz.app,https://seller.blytz.app,https://demo.blytz.app
RATE_LIMIT_WINDOW_MS=900000
RATE_LIMIT_MAX_REQUESTS=100
REQUEST_SIZE_LIMIT=10485760
SECURITY_HEADERS_ENABLED=true
ENABLE_CSRF_PROTECTION=true

# ===========================================
# PRODUCTION SETTINGS
# ===========================================
NODE_ENV=production
ENVIRONMENT=production
PORT=8080
MODE=remote

# ===========================================
# LOGGING
# ===========================================
LOG_LEVEL=info
DEBUG=false
ENABLE_SWAGGER=false
ENABLE_PROFILER=false
EOF

print_success "Production secrets generated successfully!"
print_warning "⚠️  IMPORTANT SECURITY NOTES:"
echo
echo "1. 🔒 Keep this file secure and never commit it to version control"
echo "2. 🔐 Add $ENV_FILE to your .gitignore file"
echo "3. 🚀 Update the placeholder values (FIUU_*, NINJAVAN_*) with real values"
echo "4. 📋 Store these secrets in a secure location (password manager, vault, etc.)"
echo "5. 🔄 Rotate these secrets regularly (recommended every 90 days)"
echo

# Create .gitignore entry if not exists
if [ -f ".gitignore" ]; then
    if ! grep -q ".env.production" .gitignore; then
        echo "" >> .gitignore
        echo "# Production secrets" >> .gitignore
        echo ".env.production" >> .gitignore
        print_success "Added .env.production to .gitignore"
    fi
else
    echo "# Production secrets" > .gitignore
    echo ".env.production" >> .gitignore
    print_success "Created .gitignore with .env.production entry"
fi

# Display generated secrets (without sensitive values)
print_info "📋 Generated Configuration Summary:"
echo
echo "✅ JWT Secret: [GENERATED - 32+ characters]"
echo "✅ Better Auth Secret: [GENERATED - 32+ characters]"
echo "✅ Session Secret: [GENERATED - 32+ characters]"
echo "✅ Encryption Key: [GENERATED - 64 hex characters]"
echo "✅ PostgreSQL Password: [GENERATED - 24 characters]"
echo "✅ Redis Password: [GENERATED - 32 characters]"
echo "✅ LiveKit API Key: [GENERATED - 16 characters]"
echo "✅ LiveKit API Secret: [GENERATED - 32 characters]"
echo

# Create Docker Compose override file
DOCKER_OVERRIDE="docker-compose.override.yml"
print_info "Creating Docker Compose override for development: $DOCKER_OVERRIDE"

cat > $DOCKER_OVERRIDE << EOF
# ===========================================
# Docker Compose Override - Development
# ===========================================
# This file overrides production settings for development
# ===========================================

version: '3.8'

services:
  redis:
    command: redis-server --requirepass dev_redis_password_32_characters_minimum --maxmemory 512mb --maxmemory-policy allkeys-lru --maxclients 10000

  auth-service:
    environment:
      - JWT_SECRET=dev_jwt_secret_key_32_characters_minimum
      - BETTER_AUTH_SECRET=dev_better_auth_secret_key_32_characters_minimum
      - REDIS_PASSWORD=dev_redis_password_32_characters_minimum

  auction-service:
    environment:
      - JWT_SECRET=dev_jwt_secret_key_32_characters_minimum
      - REDIS_PASSWORD=dev_redis_password_32_characters_minimum

  chat-service:
    environment:
      - REDIS_PASSWORD=dev_redis_password_32_characters_minimum

  search-service:
    environment:
      - REDIS_PASSWORD=dev_redis_password_32_characters_minimum

  notification-service:
    environment:
      - REDIS_PASSWORD=dev_redis_password_32_characters_minimum
EOF

print_success "Docker Compose override created for development!"

echo
print_success "🎉 Secret generation completed!"
print_info "Next steps:"
echo "1. Review and update placeholder values in $ENV_FILE"
echo "2. Deploy using: docker-compose --env-file $ENV_FILE up -d"
echo "3. For development, use: docker-compose -f docker-compose.yml -f $DOCKER_OVERRIDE up -d"
echo