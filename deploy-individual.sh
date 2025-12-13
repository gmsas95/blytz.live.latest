#!/bin/bash

# Individual service deployment for 2C8GB VPS
# Deploys services one by one to avoid overwhelming the system

set -e

echo "🚀 INDIVIDUAL SERVICE DEPLOYMENT FOR 2C8GB VPS"
echo "=================================================="

# Function to deploy single service
deploy_single_service() {
    local service_name=$1
    local image_tag=$2
    local port=$3
    local memory_mb=$4
    
    echo "📦 Deploying $service_name ($memory_mb MB RAM)..."
    
    # Stop if already running
    docker stop $service_name 2>/dev/null || true
    
    # Remove if exists
    docker rm $service_name 2>/dev/null || true
    
    # Deploy with memory limit
    docker run -d \
        --name $service_name \
        --memory="${memory_mb}m" \
        --cpus="0.25" \
        -p $port:$port \
        --restart unless-stopped \
        $image_tag
    
    if [ $? -eq 0 ]; then
        echo "✅ $service_name deployed successfully"
        echo "🌐 $service_name URL: http://localhost:$port"
    else
        echo "❌ $service_name deployment failed"
    fi
    
    echo ""
}

# Check system resources
echo "💾 Checking system resources..."
echo "Available memory: $(free -h | grep '^Mem:' | awk '{print $7}')"
echo "Available disk: $(df -h / | grep '^/dev/' | awk '{print $4}')"

echo ""
echo "🎯 Available services to deploy:"
echo "1. postgres    (256MB, port 5434)"
echo "2. frontend     (256MB, port 3000)"
echo "3. auth-service (128MB, port 8085)"
echo "4. product-service (128MB, port 8086)"
echo "5. auction-service (256MB, port 8087)"
echo ""
echo "💡 Usage: deploy_individual_service [service-name] [image-tag] [port] [memory-mb]"
echo ""
echo "🔧 Example: deploy_individual_service postgres gmsas95/blytz-live-services:postgres 5434 256"
echo ""
echo "⚠️  Start with essential services first (postgres, frontend)"
echo "📊 Monitor with: docker stats --no-stream"