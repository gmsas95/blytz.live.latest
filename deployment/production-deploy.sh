#!/bin/bash

# 🚀 Blytz Platform - Production Deployment Script
# ================================================

set -e

echo "🎉 Blytz Platform Production Deployment"
echo "========================================="

# Configuration
DOMAIN=${DOMAIN:-"blytz.app"}
EMAIL=${EMAIL:-"admin@blytz.app"}
SERVER_IP=${SERVER_IP:-"your-server-ip"}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Pre-deployment checks
print_info "Running pre-deployment checks..."

# Check if production secrets exist
if [ ! -f ".env.production" ]; then
    print_error "Production secrets not found. Run ./scripts/generate-secrets.sh first"
    exit 1
fi

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    exit 1
fi

print_success "Pre-deployment checks passed"

# Generate production Docker Compose
cat > docker-compose.production.yml << 'EOF'
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: blytz-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: blytz_prod
      POSTGRES_USER: blytz
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./postgres-init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U blytz -d blytz_prod"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - blytz-network

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: blytz-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD} --maxmemory 512mb --maxmemory-policy allkeys-lru --maxclients 10000
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - blytz-network

  # Nginx Reverse Proxy
  nginx:
    image: nginx:alpine
    container_name: blytz-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - nginx_logs:/var/log/nginx
    depends_on:
      - gateway
    networks:
      - blytz-network

  # Gateway Service
  gateway:
    build:
      context: .
      dockerfile: services/gateway/Dockerfile
    container_name: blytz-gateway
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8080
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
      - CORS_ORIGINS=${CORS_ORIGINS}
    ports:
      - "8080:8080"
    depends_on:
      - redis
      - postgres
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

  # Auth Service
  auth-service:
    build:
      context: .
      dockerfile: services/auth-service/Dockerfile
    container_name: blytz-auth
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8085
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
      - BETTER_AUTH_SECRET=${BETTER_AUTH_SECRET}
    ports:
      - "8085:8085"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8085/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

  # Auction Service
  auction-service:
    build:
      context: .
      dockerfile: services/auction-service/Dockerfile
    container_name: blytz-auction
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8087
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8087:8087"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8087/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

  # Product Service
  product-service:
    build:
      context: .
      dockerfile: services/product-service/Dockerfile
    container_name: blytz-product
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8086
      - DATABASE_URL=${DATABASE_URL}
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8086:8086"
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8086/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

  # Chat Service
  chat-service:
    build:
      context: .
      dockerfile: services/chat-service/Dockerfile
    container_name: blytz-chat
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8090
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8090:8090"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

  # Order Service
  order-service:
    build:
      context: .
      dockerfile: services/order-service/Dockerfile
    container_name: blytz-order
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8088
      - DATABASE_URL=${DATABASE_URL}
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8088:8088"
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8088/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

  # Payment Service
  payment-service:
    build:
      context: .
      dockerfile: services/stripe-service/Dockerfile
    container_name: blytz-payment
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8089
      - DATABASE_URL=${DATABASE_URL}
      - JWT_SECRET=${JWT_SECRET}
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}
    ports:
      - "8089:8089"
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8089/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

  # Notification Service
  notification-service:
    build:
      context: .
      dockerfile: services/notification-service/Dockerfile
    container_name: blytz-notification
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8094
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
      - FIREBASE_PROJECT_ID=${FIREBASE_PROJECT_ID}
    ports:
      - "8094:8094"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8094/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

  # LiveKit Service
  livekit-service:
    build:
      context: .
      dockerfile: services/livekit-service/Dockerfile
    container_name: blytz-livekit
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - PORT=8093
      - JWT_SECRET=${JWT_SECRET}
      - LIVEKIT_API_KEY=${LIVEKIT_API_KEY}
      - LIVEKIT_API_SECRET=${LIVEKIT_API_SECRET}
    ports:
      - "8093:8093"
    depends_on:
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8093/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - blytz-network

volumes:
  postgres_data:
  redis_data:
  nginx_logs:

networks:
  blytz-network:
    driver: bridge
EOF

print_success "Production Docker Compose file created"

# Create nginx configuration
cat > nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream gateway {
        server gateway:8080;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=auth:10m rate=5r/s;

    # Security headers
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # Main server block
    server {
        listen 80;
        server_name blytz.app www.blytz.app;
        
        # Redirect HTTP to HTTPS
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name blytz.app www.blytz.app;

        # SSL configuration
        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
        ssl_prefer_server_ciphers off;

        # Security configuration
        client_max_body_size 10M;
        client_body_timeout 60s;
        client_header_timeout 60s;

        # Rate limiting
        limit_req zone=api burst=20 nodelay;
        limit_req zone=auth burst=10 nodelay;

        # Health check endpoint
        location /health {
            proxy_pass http://gateway;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            access_log off;
        }

        # API proxy with security headers
        location /api/ {
            proxy_pass http://gateway;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Forwarded-Host $host;
            proxy_set_header X-Forwarded-Port $server_port;
            
            # WebSocket support
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            
            # Timeouts
            proxy_connect_timeout 30s;
            proxy_send_timeout 30s;
            proxy_read_timeout 30s;
        }

        # Static files (if needed)
        location /static/ {
            alias /var/www/static/;
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # Security.txt
        location /.well-known/security.txt {
            return 200 "Contact: security@blytz.app\nExpires: 2025-12-31T23:59:59.000Z\n";
            add_header Content-Type text/plain;
        }

        # Block sensitive files
        location ~ /\. {
            deny all;
            access_log off;
            log_not_found off;
        }

        location ~ ~$ {
            deny all;
            access_log off;
            log_not_found off;
        }
    }
}
EOF

print_success "Nginx configuration created"

# Create deployment script
cat > deploy.sh << 'EOF'
#!/bin/bash

# Blytz Platform Production Deployment Script
echo "🚀 Deploying Blytz Platform to Production..."

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
EOF

chmod +x deploy.sh

print_success "Deployment scripts created!"
print_info ""
print_info "🎯 Next Steps:"
print_info "1. Copy these files to your VPS"
print_info "2. Update server IP in deployment script"
print_info "3. Run: ./deploy.sh"
print_info "4. Monitor deployment with: docker-compose -f docker-compose.production.yml logs -f"
print_info ""
print_success "🚀 Ready for production deployment!"