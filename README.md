# Blytz Live Auction Platform

A real-time livestream commerce platform built with Go microservices and modern web technologies.

## 🚀 Quick Start

```bash
# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8080/health
```

## 📁 Project Structure

```
blytz/
├── services/           # Go microservices
├── frontend/           # Next.js web application  
├── blytz_flutter_app/  # Flutter mobile app
├── shared/             # Shared Go packages
├── functions/          # Firebase cloud functions
├── docs/               # Documentation
├── scripts/            # Utility scripts
└── config/             # Configuration files
```

## 🔧 Services

- **Auth Service** (8084) - Authentication & authorization
- **Product Service** (8082) - Product catalog
- **Auction Service** (8083) - Real-time auctions
- **Order Service** (8085) - Order processing
- **Payment Service** (8086) - Payment processing
- **Chat Service** (8088) - Real-time chat
- **Logistics Service** (8087) - Shipping management
- **Gateway** (8080) - API gateway

## 🌐 Access Points

- **Web App**: http://localhost:3000
- **API Gateway**: http://localhost:8080
- **Documentation**: See `/docs` directory

## 📚 Documentation

- `/docs/development/SETUP.md` - Development setup
- `/docs/deployment/DEPLOYMENT.md` - Deployment guide
- `/docs/api/` - API specifications

## 🛠️ Development

See individual service directories for development instructions.

## 📄 License

MIT License - see LICENSE file for details.
