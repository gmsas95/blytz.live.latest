#!/bin/bash

# Git-only deployment for 2C8GB VPS
# Uses pre-built images from Docker registry
# No local compilation required

set -e

echo "🚀 GIT-ONLY DEPLOYMENT FOR 2C8GB VPS"
echo "=================================="

# Check if we can access Docker registry
echo "🔍 Checking Docker registry access..."
if ! docker info | grep -q "Registry"; then
    echo "❌ No Docker registry access configured"
    echo "💡 This script requires Docker registry access"
    exit 1
fi

# Stop all existing containers
echo "🧹 Stopping all containers..."
docker compose -f docker-compose.yml down 2>/dev/null || true
docker compose -f docker-compose.production.yml down 2>/dev/null || true

# Force cleanup
echo "🧹 Force cleaning Docker..."
docker system prune -af
docker image prune -af

# Pull latest images from registry
echo "📦 Pulling latest images..."
docker pull gmsas95/blytz-live-services:latest || {
    echo "❌ Failed to pull latest images"
    echo "💡 Falling back to local builds..."
    exec ./build-vps.sh
    exit 1
}

# Deploy using registry images
echo "🌐 Deploying from registry images..."

# Create minimal compose file for registry images
cat > docker-compose.registry.yml << EOF
version: '3.8'

services:
  postgres:
    image: gmsas95/blytz-live-services:postgres
    environment:
      POSTGRES_DB: blytz_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.25'
        reservations:
          memory: 128M
          cpus: '0.125'
    restart: unless-stopped

  frontend:
    image: gmsas95/blytz-live-services:frontend
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8085
      - NODE_ENV=production
      - MODE=mock
      - NEXT_PUBLIC_ENABLE_ANALYTICS=false
      - NEXT_PUBLIC_ENABLE_DEBUG=false
      - NODE_OPTIONS=--max-old-space-size=96
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.25'
        reservations:
          memory: 128M
          cpus: '0.125'
    restart: unless-stopped

  auth-service:
    image: gmsas95/blytz-live-services:auth-service
    ports:
      - "8085:8085"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/blytz_db?sslmode=disable&max_conns=5&max_idle_conns=2
      - ENVIRONMENT=production
      - LOG_LEVEL=error
      - PORT=8085
      - JWT_SECRET=git_deploy_jwt_secret_minimum_32_bytes
    depends_on:
      - postgres
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.2'
        reservations:
          memory: 64M
          cpus: '0.1'
    restart: unless-stopped

volumes:
  postgres_data:
EOF

# Deploy registry images
echo "🚀 Deploying registry images..."
docker compose -f docker-compose.registry.yml up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 20

# Check status
echo "📊 Service Status:"
docker compose -f docker-compose.registry.yml ps

echo ""
echo "✅ Git-only deployment completed!"
echo ""
echo "🌐 Access URLs:"
echo "   Frontend: http://localhost:3000"
echo "   Auth API: http://localhost:8085"
echo "   PostgreSQL: localhost:5434"
echo ""
echo "💾 Memory Usage:"
echo "   PostgreSQL: 256MB"
echo "   Frontend: 256MB"
echo "   Auth Service: 128MB"
echo "   Total: ~640MB"
echo ""
echo "🔧 Using pre-built Docker images from registry"
echo "📊 Monitor with: docker stats"