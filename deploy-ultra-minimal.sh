#!/bin/bash

# Ultra-minimal deployment for 2C8GB VPS
# PostgreSQL optimized for RAM usage (128MB total)

set -e

echo "🚨 ULTRA-MINIMAL DEPLOYMENT"
echo "Total expected RAM usage: ~320MB"
echo "=================================="

# Kill everything first
echo "🧹 Stopping all containers..."
docker compose -f docker-compose.yml down 2>/dev/null || true
docker compose -f docker-compose.production.yml down 2>/dev/null || true

# Aggressive cleanup
echo "🧹 Aggressive Docker cleanup..."
docker system prune -af
docker image prune -af
docker volume prune -f

# Create swap if needed
echo "💾 Checking and creating swap..."
if ! sudo swapon --show | grep -q "/swapfile"; then
    echo "Creating 1GB swap file..."
    sudo fallocate -l 1G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo "✅ Swap activated"
else
    echo "✅ Swap already exists"
fi

# Deploy ultra-minimal services
echo "📦 Deploying ultra-minimal services..."

echo "Starting PostgreSQL (128MB RAM limit)..."
docker compose -f docker-compose.minimal-postgres.yml up -d postgres

sleep 15

echo "Starting Frontend in mock mode (192MB RAM limit)..."
docker compose -f docker-compose.minimal-postgres.yml up -d frontend

sleep 10

# Check status
echo "📊 Service Status:"
docker compose -f docker-compose.minimal-postgres.yml ps

echo ""
echo "✅ Ultra-minimal deployment completed!"
echo ""
echo "🌐 Access URLs:"
echo "   Frontend: http://localhost:3000"
echo "   PostgreSQL: localhost:5434"
echo ""
echo "💾 Memory Usage:"
echo "   PostgreSQL: 128MB RAM"
echo "   Frontend: 192MB RAM"
echo "   Total: ~320MB RAM"
echo ""
echo "🔧 PostgreSQL Optimizations Applied:"
echo "   - Shared buffers: 16MB"
echo "   - Work memory: 1MB"
echo "   - Maintenance memory: 8MB"
echo "   - Effective cache: 32MB"
echo "   - WAL buffers: 4MB"
echo "   - Max connections: 10"
echo ""
echo "💡 To add backend services:"
echo "   docker compose -f docker-compose.minimal-postgres.yml up -d [service-name]"
echo ""
echo "📊 Monitor with:"
echo "   free -h              # System memory"
echo "   docker stats            # Container memory"
echo "   htop                  # Process memory"