#!/bin/bash

# Pre-build script for 2C8GB VPS
# Prepares services for deployment without memory-intensive builds

set -e

echo "🔧 PREPARING SERVICES FOR 2C8GB VPS DEPLOYMENT"
echo "======================================"

# Check what services need building
echo "🔍 Checking service status..."

# Function to build service with memory limit
build_service() {
    local service_name=$1
    local service_path=$2
    local memory_limit=$3
    
    echo "Building $service_name with $memory_limit memory limit..."
    
    # Build with extreme memory optimization
    DOCKER_BUILDKIT=1 BUILDKIT_PROGRESS=plain \
    docker build \
        --memory=$memory_limit \
        --no-cache \
        -f "./services/$service_path/Dockerfile" \
        -t "blytz-$service_name:vps-ready" \
        "./services/$service_path"
    
    if [ $? -eq 0 ]; then
        echo "✅ $service_name built successfully"
        
        # Push to your local registry for VPS deployment
        echo "📦 Pushing $service_name to local registry..."
        docker tag "blytz-$service_name:vps-ready" "localhost:5000/blytz-$service_name:latest"
        docker push "localhost:5000/blytz-$service_name:latest"
        
        if [ $? -eq 0 ]; then
            echo "✅ $service_name pushed to registry"
        else
            echo "❌ Failed to push $service_name"
        fi
    else
        echo "❌ $service_name build failed"
    fi
}

# Kill any existing builds to free memory
echo "🧹 Cleaning up any existing builds..."
docker system prune -af
docker image prune -af

# Build core services first (lower memory)
echo "📦 Building core services..."

build_service "auth-service" "auth-service" "512m"
build_service "product-service" "product-service" "512m"

# Clean up between builds
echo "🧹 Intermediate cleanup..."
docker system prune -f

# Build additional services
echo "📦 Building additional services..."

build_service "auction-service" "auction-service" "768m"
build_service "order-service" "order-service" "512m"
build_service "payment-service" "payment-service" "512m"
build_service "chat-service" "chat-service" "768m"
build_service "logistics-service" "logistics-service" "512m"
build_service "search-service" "search-service" "768m"
build_service "notification-service" "notification-service" "512m"
build_service "gateway" "gateway-service" "512m"

# Build frontend last (highest memory)
echo "📦 Building frontend..."
build_service "frontend" "frontend" "1g"

# Final cleanup
echo "🧹 Final cleanup..."
docker system prune -f

echo ""
echo "✅ Service preparation completed!"
echo ""
echo "🌐 All services built and pushed to localhost:5000"
echo ""
echo "📋 Next steps for VPS deployment:"
echo "   1. SSH into your VPS"
echo "   2. Pull services: docker pull localhost:5000/blytz-[service-name]:latest"
echo "   3. Deploy: docker compose -f docker-compose.production.yml up -d"
echo ""
echo "💡 Memory usage per service:"
echo "   - Auth/Product/Order/Payment/Chat/Logistics/Search/Notification: 512MB"
echo "   - Auction/Gateway: 768MB"
echo "   - Frontend: 1GB"
echo ""
echo "🔧 This bypasses VPS build memory issues!"