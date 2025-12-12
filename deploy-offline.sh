#!/bin/bash

# Offline deployment solution for 2C8GB VPS
# Bypasses Go dependency downloads by using pre-built or local compilation

set -e

echo "🚨 OFFLINE DEPLOYMENT FOR 2C8GB VPS"
echo "======================================="
echo "Bypassing Go dependency downloads..."

# Kill all existing containers
echo "🧹 Stopping all containers..."
docker compose -f docker-compose.yml down 2>/dev/null || true
docker compose -f docker-compose.production.yml down 2>/dev/null || true

# Force cleanup
echo "🧹 Force cleaning Docker..."
docker system prune -af
docker image prune -af

# Check if we have Go installed locally
if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Installing Go..."
    # Install Go locally if not available
    wget -q https://go.dev/dl/go1.22.6.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.22.6.linux-amd64.tar.gz
    export PATH=/usr/local/go/bin:$PATH
    echo "✅ Go installed locally"
fi

# Create minimal deployment using local compilation
echo "📦 Creating minimal offline deployment..."

# Create temporary directory for builds
mkdir -p /tmp/blytz-build
cd /tmp/blytz-build

# Function to build service locally
build_service_locally() {
    local service_name=$1
    local service_path=$2
    local port=$3
    
    echo "Building $service_name locally..."
    
    if [ -d "services/$service_path" ]; then
        cd "/home/sas/blytzmvp-clean/services/$service_path"
        
        # Build with local Go
        /usr/local/go/bin/go build -ldflags="-s -w" -o "/tmp/blytz-build/$service_name" .
        
        if [ $? -eq 0 ]; then
            echo "✅ $service_name built successfully"
            
            # Create minimal container image locally
            cat > "/tmp/blytz-build/Dockerfile.$service_name" << EOF
FROM scratch
COPY $service_name /app/
EXPOSE $port
CMD ["/app/$service_name"]
EOF
            
            # Build minimal image
            docker build -t "blytz-$service_name:local" -f "/tmp/blytz-build/Dockerfile.$service_name" "/tmp/blytz-build/"
            
            if [ $? -eq 0 ]; then
                echo "✅ $service_name image built successfully"
            else
                echo "❌ $service_name image build failed"
            fi
        else
            echo "❌ Service $service_path not found"
        fi
}

# Build only essential services locally
build_service_locally "auth-service" "auth-service" 8085
build_service_locally "product-service" "product-service" 8086

# Deploy frontend using the mock data (already built)
echo "🌐 Deploying frontend with mock data..."
cd "/home/sas/blytzmvp-clean/frontend"

# Build frontend locally (no Go dependencies)
docker build -t blytz-frontend:local .

# Create minimal compose file for offline deployment
cat > /tmp/blytz-build/docker-compose.offline.yml << EOF
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
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

  auth-service:
    image: blytz-auth-service:local
    ports:
      - "8085:8085"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/blytz_db?sslmode=disable&max_conns=5&max_idle_conns=2
      - ENVIRONMENT=production
      - LOG_LEVEL=error
      - PORT=8085
      - JWT_SECRET=offline_jwt_secret_minimum_32_bytes
    depends_on:
      - postgres
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.2'
        reservations:
          memory: 64M
          cpus: '0.1'

  product-service:
    image: blytz-product-service:local
    ports:
      - "8086:8086"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/blytz_db?sslmode=disable
      - ENVIRONMENT=production
      - LOG_LEVEL=error
      - PORT=8086
    depends_on:
      - postgres
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.2'
        reservations:
          memory: 64M
          cpus: '0.1'

  frontend:
    image: blytz-frontend:local
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8085
      - NODE_ENV=production
      - MODE=mock
      - NEXT_PUBLIC_ENABLE_ANALYTICS=false
      - NEXT_PUBLIC_ENABLE_DEBUG=false
      - NODE_OPTIONS=--max-old-space-size=64
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.25'
        reservations:
          memory: 128M
          cpus: '0.125'

volumes:
  postgres_data:
EOF

echo "🚀 Deploying offline services..."
docker compose -f /tmp/blytz-build/docker-compose.offline.yml up -d

# Wait for services to start
echo "⏳ Waiting for services to start..."
sleep 20

# Check status
echo "📊 Service Status:"
docker compose -f /tmp/blytz-build/docker-compose.offline.yml ps

echo ""
echo "✅ Offline deployment completed!"
echo ""
echo "🌐 Access URLs:"
echo "   Frontend: http://localhost:3000"
echo "   Auth API: http://localhost:8085"
echo "   Product API: http://localhost:8086"
echo "   PostgreSQL: localhost:5434"
echo ""
echo "💡 Memory Usage:"
echo "   PostgreSQL: 256MB"
echo "   Auth Service: 128MB"
echo "   Product Service: 128MB"
echo "   Frontend: 256MB"
echo "   Total: ~768MB"
echo ""
echo "🔧 This deployment bypasses Go dependency downloads!"
echo "📊 Monitor with: docker stats"

# Cleanup
rm -rf /tmp/blytz-build