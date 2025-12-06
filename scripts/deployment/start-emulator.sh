#!/bin/bash

echo "🚀 Starting Blytz Emulator Environment..."

# Set environment variables
export DATABASE_URL="postgres://blytz:bJot0TOZtAvj7kATOT6U6WL252AMSPQgjui90CbJXGQ=@localhost:5432/blytz_prod"
export REDIS_URL="redis://localhost:6379"
export JWT_SECRET="p2wGaXBYVke45unnZbbCraCVoDWci7HE1B8wSQAO5xr7VjofhPc+J8IHRQeQ7HRtsHZO777QxN+OV26cJjv5DA=="
export BETTER_AUTH_SECRET="yElmlr9FT/4WDFniup7EK2G2IKydnaEeX4ifr80UE0WdfKnCaOzJ5tTSjS+OITLqhNbJg/v4cDQnS/g5JAc09Q=="
export NODE_ENV="development"
export AUTH_SERVICE_URL="http://localhost:8084"

# Start services in background
echo "🔧 Starting microservices..."
./bin/auth-service > logs/auth-service.log 2>&1 &
./bin/auction-service > logs/auction-service.log 2>&1 &
./bin/chat-service > logs/chat-service.log 2>&1 &
./bin/product-service > logs/product-service.log 2>&1 &
./bin/payment-service > logs/payment-service.log 2>&1 &
./bin/order-service > logs/order-service.log 2>&1 &
./bin/logistics-service > logs/logistics-service.log 2>&1 &

# Wait for services to start
echo "⏳ Waiting for services to start..."
sleep 10

# Test services
echo "🧪 Testing services..."
echo "Auth Service: $(curl -s http://localhost:8084/health || echo '❌ Failed')"
echo "Auction Service: $(curl -s http://localhost:8083/health || echo '❌ Failed')"
echo "Chat Service: $(curl -s http://localhost:8088/health || echo '❌ Failed')"
echo "Product Service: $(curl -s http://localhost:8082/health || echo '❌ Failed')"
echo "Order Service: $(curl -s http://localhost:8085/health || echo '❌ Failed')"
echo "Payment Service: $(curl -s http://localhost:8086/health || echo '❌ Failed')"
echo "Logistics Service: $(curl -s http://localhost:8087/health || echo '❌ Failed')"

echo ""
echo "✅ Blytz Emulator Environment Started!"
echo ""
echo "🌐 Access Points:"
echo "📱 Main Frontend: http://localhost:3000"
echo "📺 Demo Frontend: http://localhost:3001"
echo "🛍️  Seller Frontend: http://localhost:3002"
echo "🔥 Firebase Functions: http://localhost:5001"
echo "🔥 Firebase UI: http://localhost:4000"
echo ""
echo "🔧 API Endpoints:"
echo "🔐 Auth Service: http://localhost:8084"
echo "🏷️  Product Service: http://localhost:8082"
echo "🏪 Auction Service: http://localhost:8083"
echo "💬 Chat Service: http://localhost:8088"
echo "📦 Order Service: http://localhost:8085"
echo "💳 Payment Service: http://localhost:8086"
echo "🚚 Logistics Service: http://localhost:8087"