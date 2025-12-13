# Modular Deployment for 2C8GB VPS

## Overview
This repository contains modular Docker Compose configurations for easier management of frontend and backend services on your 2C8GB VPS.

## Files

### Frontend
- `docker-compose.modular-frontend.yml` - Frontend only (192MB RAM)

### Backend
- `docker-compose.modular-backend.yml` - All backend services (640MB RAM)

## Usage

### Deploy Frontend Only
```bash
docker compose -f docker-compose.modular-frontend.yml up -d
```

### Deploy Backend Only
```bash
docker compose -f docker-compose.modular-backend.yml up -d
```

### Deploy Both (Step by Step)
```bash
# Step 1: Deploy backend first
docker compose -f docker-compose.modular-backend.yml up -d

# Step 2: Deploy frontend
docker compose -f docker-compose.modular-frontend.yml up -d
```

### Deploy Individual Services
```bash
# Deploy only PostgreSQL
docker compose -f docker-compose.modular-backend.yml up -d postgres

# Deploy only Auth Service
docker compose -f docker-compose.modular-backend.yml up -d auth-service

# Deploy only Product Service
docker compose -f docker-compose.modular-backend.yml up -d product-service
```

## Memory Usage

| Configuration | Services | Total RAM |
|-------------|-----------|------------|
| Frontend Only | Frontend | 192MB |
| Backend Only | PostgreSQL + Auth + Product + Auction + Gateway | 640MB |
| Both | All Services | 832MB |

## Service URLs

| Service | Port | URL |
|---------|-------|-----|
| Frontend | 3000 | http://localhost:3000 |
| Auth Service | 8085 | http://localhost:8085 |
| Product Service | 8086 | http://localhost:8086 |
| Auction Service | 8087 | http://localhost:8087 |
| Gateway | 8092 | http://localhost:8092 |

## Stop Services

```bash
# Stop frontend
docker compose -f docker-compose.modular-frontend.yml down

# Stop backend
docker compose -f docker-compose.modular-backend.yml down

# Stop all
docker compose -f docker-compose.modular-frontend.yml down
docker compose -f docker-compose.modular-backend.yml down
```

## Monitor Services

```bash
# Check running containers
docker ps

# Check resource usage
docker stats

# Check logs
docker compose -f docker-compose.modular-frontend.yml logs
docker compose -f docker-compose.modular-backend.yml logs
```

## Troubleshooting

### Low Memory Issues
1. Deploy services one by one
2. Use frontend-only configuration
3. Check system resources: `free -h`
4. Clean up unused containers: `docker system prune -f`

### Service Connection Issues
1. Ensure backend services are running before frontend
2. Check service health: `docker compose -f docker-compose.modular-backend.yml ps`
3. Check logs for connection errors

### Build Issues
1. Use pre-built images from registry
2. Build services individually
3. Check system memory during build: `free -h`