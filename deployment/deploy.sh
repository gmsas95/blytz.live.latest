#!/bin/bash

# 🚀 Blytz Platform Production Deployment Script
echo "🚀 Deploying Blytz Platform to Production..."

# Load environment variables
if [ -f ".env.production" ]; then
    source .env.production
else
    echo "❌ .env.production not found. Run ./scripts/generate-secrets.sh first"
    exit 1
fi

# Stop existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.production.yml down

# Build and start services
echo "🏗️ Building and starting services..."
docker-compose -f docker-compose.production.yml build --no-cache
docker-compose -f docker-compose.production.yml up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 30

# Check service health
echo "🏥 Checking service health..."
services=("gateway:8080" "auth-service:8085" "auction-service:8087" "product-service:8086" "chat-service:8090" "order-service:8088" "payment-service:8089" "notification-service:8094" "livekit-service:8093")

for service in "${services[@]}"; do
    IFS=':' read -r name port <<< "$service"
    if curl -f "http://localhost:$port/health" >/dev/null 2>&1; then
        echo "✅ $name is healthy"
    else
        echo "❌ $name is not healthy"
        docker-compose -f docker-compose.production.yml logs "$name"
    fi
done

# Run security tests
echo "🔒 Running security tests..."
if ./scripts/security-test.sh; then
    echo "✅ All security tests passed"
else
    echo "⚠️  Some security tests failed - check logs"
fi

echo "🎉 Deployment completed!"
echo ""
echo "🌐 Application URLs:"
echo "   Main App: https://blytz.app"
echo "   API: https://blytz.app/api/v1"
echo "   Health: https://blytz.app/health"
echo ""
echo "📊 Service Status:"
docker-compose -f docker-compose.production.yml ps