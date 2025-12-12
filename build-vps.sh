#!/bin/bash

# VPS-optimized build script for 2C8GB VPS
# Uses optimized docker-compose.production.yml with reduced memory limits

set -e

echo "🚀 Starting VPS-optimized Docker build..."

# Free up memory before starting
echo "🧹 Cleaning up Docker system..."
docker system prune -f
docker image prune -f

# Set memory limits for build
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# Build in stages to manage memory usage
echo "📦 Stage 1: Building core services..."

echo "Building postgres and redis..."
docker compose -f docker-compose.production.yml build --no-cache postgres redis

echo "🧹 Cleanup after stage 1..."
docker system prune -f

echo "📦 Stage 2: Building essential services..."

echo "Building auth-service..."
docker compose -f docker-compose.production.yml build --no-cache auth-service --memory=512m

echo "Building product-service..."
docker compose -f docker-compose.production.yml build --no-cache product-service --memory=512m

echo "🧹 Cleanup after stage 2..."
docker system prune -f

echo "📦 Stage 3: Building additional services..."

echo "Building payment-service..."
docker compose -f docker-compose.production.yml build --no-cache payment-service --memory=512m

echo "Building order-service..."
docker compose -f docker-compose.production.yml build --no-cache order-service --memory=512m

echo "🧹 Cleanup after stage 3..."
docker system prune -f

echo "📦 Stage 4: Building remaining services..."

echo "Building auction-service..."
docker compose -f docker-compose.production.yml build --no-cache auction-service --memory=768m

echo "Building chat-service..."
docker compose -f docker-compose.production.yml build --no-cache chat-service --memory=768m

echo "Building logistics-service..."
docker compose -f docker-compose.production.yml build --no-cache logistics-service --memory=512m

echo "🧹 Cleanup after stage 4..."
docker system prune -f

echo "📦 Stage 5: Building gateway and frontend..."

echo "Building gateway..."
docker compose -f docker-compose.production.yml build --no-cache gateway --memory=512m

echo "Building frontend (with mock data)..."
docker compose -f docker-compose.production.yml build --no-cache frontend --memory=768m

# Final cleanup
echo "🧹 Final cleanup..."
docker system prune -f

echo "✅ VPS-optimized build completed successfully!"
echo ""
echo "💡 Next steps:"
echo "   1. Set up your environment variables in .env"
echo "   2. Run: docker compose -f docker-compose.production.yml up -d"
echo "   3. Check logs: docker compose -f docker-compose.production.yml logs -f"
echo ""
echo "🔧 Memory optimizations applied:"
echo "   - Reduced service memory limits to 192-384MB"
echo "   - Frontend set to MODE=mock to reduce backend dependencies"
echo "   - Sequential builds with cleanup between stages"
echo "   - PostgreSQL limited to 512MB"
echo "   - Redis limited to 256MB"