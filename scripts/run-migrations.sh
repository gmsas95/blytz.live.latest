#!/bin/bash
# Database Migration Runner for Blytz MVP
# This script runs migrations for all services

set -e

echo "🗄️  Running Database Migrations for Blytz MVP"
echo "=============================================="

# Database configuration
DB_URL=${DATABASE_URL:-"postgres://blytz:blytz_password_2025@localhost:5432/blytz_mvp?sslmode=disable"}

echo "📋 Database Configuration:"
echo "  URL: ${DB_URL//*@*/***:***@}" # Hide password
echo

# Services with migrations
SERVICES=(
	"auth-service"
	"product-service" 
	"auction-service"
	"order-service"
)

# Function to run migration for a service
run_migration() {
	local service=$1
	local action=${2:-"up"}
	
	echo "🔄 Running migrations for $service..."
	
	cd "/home/sas/blytzmvp-clean/services/$service"
	
	if [ ! -f "migrations/migrate.go" ]; then
		echo "⚠️  No migrations found for $service, skipping..."
		return 0
	fi
	
	# Set database URL for migration
	export DATABASE_URL="$DB_URL"
	
	# Run migration
	if go run migrations/migrate.go "$action"; then
		echo "✅ $service migrations completed successfully"
	else
		echo "❌ $service migrations failed"
		return 1
	fi
	
	echo
}

# Parse command line arguments
ACTION=${1:-"up"}

case "$ACTION" in
	"up")
		echo "🚀 Running UP migrations for all services..."
		echo
		for service in "${SERVICES[@]}"; do
			run_migration "$service" "up"
		done
		;;
	"down")
		echo "🔽 Running DOWN migrations for all services..."
		echo
		for service in "${SERVICES[@]}"; do
			run_migration "$service" "down"
		done
		;;
	"service")
		if [ -z "$2" ]; then
			echo "Usage: $0 service <service-name> [up|down]"
			exit 1
		fi
		SERVICE_NAME=$2
		SERVICE_ACTION=${3:-"up"}
		echo "🔄 Running migrations for specific service: $SERVICE_NAME"
		echo
		run_migration "$SERVICE_NAME" "$SERVICE_ACTION"
		;;
	*)
		echo "Usage: $0 [up|down|service <service-name>]"
		echo
		echo "Examples:"
		echo "  $0 up                    # Run all UP migrations"
		echo "  $0 down                  # Run all DOWN migrations"
		echo "  $0 service auth-service    # Run UP migrations for auth-service only"
		echo "  $0 service auth-service down # Run DOWN migrations for auth-service only"
		exit 1
		;;
esac

echo "🎉 Migration process completed!"
echo
echo "📋 Next Steps:"
echo "1. Test database connectivity:"
echo "   cd services/auth-service && go run main-db.go"
echo
echo "2. Verify database schema:"
echo "   psql \$DATABASE_URL -c \"\\dt\""
echo
echo "3. Create seed data:"
echo "   ./scripts/create-seed-data.sh"
echo
echo "🚀 Your database is ready for production!"