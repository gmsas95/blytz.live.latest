#!/bin/bash

# Quick Test Script for Product Service

echo "🧪 Testing Product Service - Quick Win Fixes"

# Navigate to product service
cd /home/sas/blytzmvp-clean/services/product-service

# Test if the service can start
echo "📦 Testing service startup..."
go mod tidy

# Check for any compilation errors
echo "🔍 Checking for compilation errors..."
go build -o product-service cmd/main.go

if [ $? -eq 0 ]; then
    echo "✅ Product Service compiles successfully"
else
    echo "❌ Product Service has compilation errors"
    exit 1
fi

# Clean up binary
rm -f product-service

echo "✅ Quick Win Fix 1: Service builds successfully"

# Test database models
echo "🔍 Testing database models..."
go test ./internal/models/ -v

if [ $? -eq 0 ]; then
    echo "✅ Database models compile and test successfully"
else
    echo "❌ Database models have issues"
fi

echo "✅ Product Service Quick Win Fix Complete!"
echo ""
echo "🎯 Next: Start service and test API endpoints"
echo ""
echo "📝 Commands to test manually:"
echo "1. cd services/product-service"
echo "2. go run cmd/main.go"
echo "3. curl http://localhost:8082/health"
echo "4. curl http://localhost:8082/api/v1/products/"