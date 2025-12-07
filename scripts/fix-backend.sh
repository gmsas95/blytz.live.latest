#!/bin/bash

# 🔥 BACKEND RECOVERY SCRIPT - EMERGENCY FIX
# This script will fix all Go module issues and restore backend functionality

echo "🚨 STARTING EMERGENCY BACKEND RECOVERY..."

# Backup current go.mod files
echo "📦 Backing up go.mod files..."
find services/ -name "go.mod" -exec cp {} {}.backup \;

# Fix the problematic go-internal dependency across all services
echo "🔧 Fixing go-internal dependency issues..."

for service_dir in services/*/; do
    if [ -f "$service_dir/go.mod" ]; then
        echo "🔧 Fixing $(basename $service_dir)..."
        cd "$service_dir"
        
        # Remove problematic dependency
        go mod edit -droprequire github.com/rogpeppe/go-internal
        
        # Clean go.mod and go.sum
        rm -f go.sum
        
        # Re-download dependencies
        go mod download
        
        # Tidy modules
        go mod tidy
        
        echo "✅ Fixed $(basename $service_dir)"
        cd ../../
    fi
done

echo "🎉 GO MODULE RECOVERY COMPLETE!"
echo "🏗️ Now services should compile..."