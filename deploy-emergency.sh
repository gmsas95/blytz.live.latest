#!/bin/bash

# Emergency deployment for 2C8GB VPS with critical memory constraints
# ONLY runs essential services

set -e

echo "🚨 EMERGENCY DEPLOYMENT FOR 2C8GB VPS"
echo "======================================"

# Kill all existing containers to free memory
echo "🧹 Stopping all containers..."
docker compose -f docker-compose.yml down 2>/dev/null || true
docker compose -f docker-compose.production.yml down 2>/dev/null || true

# Force cleanup
echo "🧹 Force cleaning Docker..."
docker system prune -af
docker image prune -af
docker volume prune -f

# Create swap space if needed
echo "💾 Checking swap space..."
sudo swapon --show
if [ $? -ne 0 ]; then
    echo "⚠️  No swap detected. Creating 1GB swap file..."
    sudo fallocate -l 1G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo "✅ Swap file created and activated"
fi

# Deploy minimal services only
echo "📦 Deploying emergency configuration..."

echo "Starting PostgreSQL (256MB limit)..."
docker compose -f docker-compose.emergency.yml up -d postgres

echo "Starting Frontend in mock mode (256MB limit)..."
docker compose -f docker-compose.emergency.yml up -d frontend

echo "Starting Auth Service (128MB limit)..."
docker compose -f docker-compose.emergency.yml up -d auth-service

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 30

# Check status
echo "📊 Service Status:"
docker compose -f docker-compose.emergency.yml ps

echo ""
echo "✅ Emergency deployment completed!"
echo ""
echo "🌐 Access URLs:"
echo "   Frontend: http://localhost:3000"
echo "   Auth API: http://localhost:8085"
echo ""
echo "💡 Memory Usage:"
echo "   PostgreSQL: 256MB"
echo "   Frontend: 256MB" 
echo "   Auth Service: 128MB"
echo "   Total: ~640MB"
echo ""
echo "🔧 To add more services:"
echo "   1. Stop current: docker compose -f docker-compose.emergency.yml down"
echo "   2. Run individual service: docker compose -f docker-compose.emergency.yml up -d [service-name]"
echo ""
echo "⚠️  Monitor memory with: free -h and docker stats"