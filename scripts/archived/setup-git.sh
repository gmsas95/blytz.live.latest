#!/bin/bash

# Blytz Platform Git Setup Script
# This script will help you set up and push to a new Git repository

set -e

echo "🚀 Blytz Platform Git Setup Script"
echo "================================="

# Navigate to project directory
PROJECT_DIR="/home/sas/blytzmvp-clean"
cd "$PROJECT_DIR"

echo "📁 Working directory: $PROJECT_DIR"

# Check if we're in the right place
if [[ ! -f "docker-compose.yml" || ! -d "services" ]]; then
    echo "❌ Error: Not in the correct project directory!"
    echo "Please run this script from the blytzmvp-clean directory"
    exit 1
fi

echo ""
echo "🧹 Cleaning up any existing Git history..."

# Remove any existing Git history for fresh start
rm -rf .git 2>/dev/null || true
echo "✅ Removed existing Git history"

echo ""
echo "🔧 Setting up clean .gitignore..."

# Create clean, focused .gitignore
cat > .gitignore << 'EOF'
# Environment variables
.env
.env.local
.env.*.local
.env.production
.env.staging
.env.*.production
.env.*.staging

# Build artifacts
*.exe
*.dll
*.so
*.dylib
*.test
*.out
services/*/main
services/*/*-service

# Dependencies
node_modules/
vendor/
go.sum

# IDE files
.vscode/
.idea/
*.swp
*~

# OS files
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Docker volumes
data/
postgres-data/
redis-data/

# Next.js build artifacts
frontend/.next/
frontend/out/
frontend/build/

# Flutter build artifacts
blytz_flutter_app/build/
blytz_flutter_app/.dart_tool/

# SSL certificates
*.pem
*.key
*.crt

# Coverage and test files
coverage.out
coverage/
test-results/
playwright-report/

# Temporary files
tmp/
*.tmp
.cache/

# Database files
*.db
*.sqlite

# Package distributions
*.tar.gz
*.zip

# Firebase
firebase-debug.log
.firebaserc

# Security files
secrets/
*.secret
*.token
EOF

echo "✅ Created clean .gitignore"

echo ""
echo "🔧 Initializing Git repository..."

# Initialize new Git repository
git init
git branch -M main

echo "✅ Git repository initialized"

echo ""
echo "⚙️  Configuring Git user..."

# Get Git user info
read -p "Enter your Git name: " GIT_NAME
read -p "Enter your Git email: " GIT_EMAIL

git config user.name "$GIT_NAME"
git config user.email "$GIT_EMAIL"

echo "✅ Git user configured"
echo "   Name: $GIT_NAME"
echo "   Email: $GIT_EMAIL"

echo ""
echo "📋 Adding files to Git..."

# Add all files
git add .

# Show what will be committed
echo ""
echo "📊 Files to be committed:"
git status --porcelain | head -10
if [[ $(git status --porcelain | wc -l) -gt 10 ]]; then
    echo "... and $(($(git status --porcelain | wc -l) - 10)) more files"
fi

echo ""
echo "🎉 Creating initial commit..."

# Create initial commit
git commit -m "🎉 Initial commit: Complete 9-service Blytz Live Auction Platform

🏗️ Architecture:
- 9 Microservices: Auth, Product, Auction, Order, Payment, Chat, Logistics, Gateway, LiveKit
- PostgreSQL + Redis databases
- Docker containerization
- Next.js + Flutter frontends

🔧 Features:
- JWT authentication with Better Auth
- Real-time bidding with Redis
- Video streaming with LiveKit
- Payment processing (Fiuu)
- Real-time chat
- Shipping management (Ninjavan)
- API Gateway with rate limiting
- Health checks on all services

📦 Production-ready with:
- Comprehensive documentation
- Clean project structure
- Proper Docker orchestration
- Complete API routing
- Security best practices"

echo "✅ Initial commit created"

echo ""
echo "🌐 Setting up remote repository..."

# Get repository URL
echo "Choose your Git provider:"
echo "1) GitHub"
echo "2) GitLab" 
echo "3) Bitbucket"
echo "4) Custom URL"
read -p "Enter your choice (1-4): " CHOICE

case $CHOICE in
    1)
        read -p "Enter your GitHub username: " USERNAME
        REPO_URL="https://github.com/$USERNAME/blytz-live-auction.git"
        ;;
    2)
        read -p "Enter your GitLab username: " USERNAME
        REPO_URL="https://gitlab.com/$USERNAME/blytz-live-auction.git"
        ;;
    3)
        read -p "Enter your Bitbucket username: " USERNAME
        REPO_URL="https://bitbucket.org/$USERNAME/blytz-live-auction.git"
        ;;
    4)
        read -p "Enter your repository URL: " REPO_URL
        ;;
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo "📦 Repository URL: $REPO_URL"

# Add remote
git remote add origin "$REPO_URL"

echo "✅ Remote repository configured"

echo ""
echo "🚀 Pushing to remote repository..."

# Push to remote
echo "🔄 Pushing to main branch..."
git push -u origin main

echo ""
echo "🎉 SUCCESS! Your Blytz Live Auction Platform is now on Git!"
echo ""
echo "📊 Repository Statistics:"
echo "   - 9 microservices"
echo "   - 2 frontend applications" 
echo "   - Complete Docker setup"
echo "   - Comprehensive documentation"
echo "   - Production-ready configuration"
echo ""
echo "🌐 Repository URL: $REPO_URL"
echo ""
echo "📋 Next steps:"
echo "1. Visit your repository online"
echo "2. Clone and test: git clone $REPO_URL"
echo "3. Run: cd blytz-live-auction && docker-compose up -d"
echo "4. Test: curl http://localhost:8080/health"
echo ""
echo "🎯 Happy coding with your complete live auction platform!"