#!/bin/bash

# Optimized build script for 2C8GB VPS
# Builds services sequentially to reduce memory usage

set -e

echo "🚀 Starting optimized Docker build for 2C8GB VPS..."

# Free up memory before starting
echo "🧹 Cleaning up Docker system..."
docker system prune -f
docker image prune -f

# Set memory limits for build
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# Build core services first (smaller memory footprint)
echo "📦 Building core services..."

echo "Building auth-service..."
docker compose build --no-cache auth-service --memory=1g

echo "Building product-service..."
docker compose build --no-cache product-service --memory=1g

echo "Building payment-service..."
docker compose build --no-cache payment-service --memory=1g

# Clean up between builds
echo "🧹 Intermediate cleanup..."
docker system prune -f

# Build medium services
echo "📦 Building medium services..."

echo "Building order-service..."
docker compose build --no-cache order-service --memory=1g

echo "Building chat-service..."
docker compose build --no-cache chat-service --memory=1g

echo "Building logistics-service..."
docker compose build --no-cache logistics-service --memory=1g

# Clean up again
echo "🧹 Intermediate cleanup..."
docker system prune -f

# Build larger services
echo "📦 Building larger services..."

echo "Building auction-service..."
docker compose build --no-cache auction-service --memory=1.5g

echo "Building search-service..."
docker compose build --no-cache search-service --memory=1.5g

echo "Building notification-service..."
docker compose build --no-cache notification-service --memory=1g

# Clean up again
echo "🧹 Intermediate cleanup..."
docker system prune -f

# Build gateway and frontend last
echo "📦 Building gateway and frontend..."

echo "Building gateway..."
docker compose build --no-cache gateway --memory=1g

echo "Building frontend..."
docker compose build --no-cache frontend --memory=1.5g

# Final cleanup
echo "🧹 Final cleanup..."
docker system prune -f

echo "✅ Build completed successfully!"
echo "💡 Tip: Run 'docker compose up -d' to start services"