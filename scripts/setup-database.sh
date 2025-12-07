#!/bin/bash
# Database Setup Script for Blytz MVP Production
# This script sets up PostgreSQL database for all services

set -e

echo "🗄️  Setting up PostgreSQL Database for Blytz MVP"
echo "=================================================="

# Database configuration
DB_NAME="blytz_mvp"
DB_USER="blytz"
DB_PASSWORD="blytz_password_2025"
DB_HOST="localhost"
DB_PORT="5432"

echo "📋 Database Configuration:"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo

# Check if PostgreSQL is running
if ! pg_isready -h $DB_HOST -p $DB_PORT; then
    echo "❌ PostgreSQL is not running on $DB_HOST:$DB_PORT"
    echo "Please start PostgreSQL service:"
    echo "  sudo systemctl start postgresql"
    echo "  # OR"
    echo "  docker run -d --name postgres -e POSTGRES_PASSWORD=$DB_PASSWORD -p $DB_PORT:5432 postgres:15"
    exit 1
fi

echo "✅ PostgreSQL is running"

# Create database and user
echo "🔧 Creating database and user..."

# Connect to PostgreSQL as superuser and create database/user
sudo -u postgres psql << EOF
-- Create database user
CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';

-- Create database
CREATE DATABASE $DB_NAME OWNER $DB_USER;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

-- Connect to database and create extensions
\c $DB_NAME;

-- Create required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

EOF

echo "✅ Database and user created successfully"

# Create .env file for database configuration
echo "📝 Creating database configuration..."

cat > /home/sas/blytzmvp-clean/.env.database << EOF
# Database Configuration
DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD

# Service Database URLs
AUTH_DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable
PRODUCT_DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable
AUCTION_DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable
ORDER_DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable
PAYMENT_DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable
CHAT_DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable
LOGISTICS_DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable
EOF

echo "✅ Database configuration created at .env.database"

# Test database connection
echo "🔍 Testing database connection..."
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" > /dev/null 2>&1; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed"
    exit 1
fi

echo
echo "🎉 Database setup completed successfully!"
echo
echo "📋 Next Steps:"
echo "1. Source the database configuration:"
echo "   source .env.database"
echo
echo "2. Run database migrations for each service:"
echo "   cd services/auth-service && go run migrations/migrate.go up"
echo "   cd services/product-service && go run migrations/migrate.go up"
echo "   # ... repeat for all services"
echo
echo "3. Test database connectivity:"
echo "   cd services/auth-service && go run main.go"
echo
echo "🚀 Your PostgreSQL database is ready for production!"