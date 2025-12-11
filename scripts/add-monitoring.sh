#!/bin/bash

# Script to add monitoring capabilities to all Blytz services

set -e

SERVICES_DIR="services"
SERVICES=("auth-service" "product-service" "auction-service" "order-service" "payment-service" "chat-service" "logistics-service" "notification-service" "gateway" "livekit-service")

echo "🚀 Adding monitoring capabilities to all Blytz services..."

for service in "${SERVICES[@]}"; do
    echo "📊 Processing $service..."
    
    SERVICE_DIR="$SERVICES_DIR/$service"
    
    if [ ! -d "$SERVICE_DIR" ]; then
        echo "⚠️  Service directory $SERVICE_DIR not found, skipping..."
        continue
    fi
    
    # Find main.go file (could be in different locations)
    MAIN_FILE=$(find "$SERVICE_DIR" -name "main.go" -type f | head -1)
    
    if [ -z "$MAIN_FILE" ]; then
        echo "⚠️  No main.go found for $service, skipping..."
        continue
    fi
    
    echo "📝 Found main.go at: $MAIN_FILE"
    
    # Check if metrics are already added
    if grep -q "shared_metrics" "$MAIN_FILE"; then
        echo "✅ Metrics already added to $service"
        continue
    fi
    
    # Add metrics import
    if ! grep -q "shared_metrics" "$MAIN_FILE"; then
        sed -i '/shared_utils/a\	shared_metrics "github.com/gmsas95/blytz.live.latest/shared/pkg/metrics"' "$MAIN_FILE"
        echo "✅ Added metrics import to $service"
    fi
    
    # Add metrics middleware
    if grep -q "router.Use(shared_utils.CORSMiddleware())" "$MAIN_FILE"; then
        sed -i '/router.Use(shared_utils.CORSMiddleware())/a\\n\t// Metrics middleware\n\trouter.Use(shared_metrics.MetricsMiddleware("'$service'"))' "$MAIN_FILE"
        echo "✅ Added metrics middleware to $service"
    fi
    
    # Add metrics endpoint
    if grep -q "/health" "$MAIN_FILE"; then
        sed -i '/router.GET("\/health"/a\\n\t// Metrics endpoint\n\trouter.GET("\/metrics", shared_metrics.PrometheusHandler())' "$MAIN_FILE"
        echo "✅ Added metrics endpoint to $service"
    fi
    
    # Add build info and start time
    if grep -q "logger.Info" "$MAIN_FILE"; then
        sed -i '/logger.Info/i\\n\t// Set build info and start time\n\tshared_metrics.SetBuildInfo("v1.0.0", "unknown", time.Now().Format(time.RFC3339))\n\tshared_metrics.SetStartTime(float64(time.Now().Unix()))\n' "$MAIN_FILE"
        echo "✅ Added build info to $service"
    fi
    
    # Update go.mod
    GO_MOD_FILE="$SERVICE_DIR/go.mod"
    if [ -f "$GO_MOD_FILE" ]; then
        if ! grep -q "github.com/prometheus/client_golang" "$GO_MOD_FILE"; then
            sed -i '/go.uber.org\/zap/a\	github.com/prometheus/client_golang v1.19.0' "$GO_MOD_FILE"
            echo "✅ Added Prometheus dependency to $service"
        fi
    fi
    
    echo "✅ Completed monitoring setup for $service"
    echo ""
done

echo "🎉 Monitoring capabilities added to all services!"
echo ""
echo "📋 Next steps:"
echo "1. Run 'go mod tidy' in each service directory"
echo "2. Update Docker Compose to expose metrics ports"
echo "3. Start the monitoring stack"
echo "4. Verify metrics collection"
echo ""
echo "🔗 Monitoring endpoints will be available at:"
for service in "${SERVICES[@]}"; do
    echo "   - $service: http://localhost:908X/metrics"
done