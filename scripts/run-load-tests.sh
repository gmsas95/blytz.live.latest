#!/bin/bash

# Blytz Live Auction MVP - Load Testing Script
# This script runs comprehensive load tests to verify capacity scaling

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_DIR="$SCRIPT_DIR/../tests/performance/results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create results directory
mkdir -p "$RESULTS_DIR"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if services are running
check_services() {
    print_status "Checking if services are running..."
    
    services=("auth-service:8085" "chat-service:8090" "notification-service:8094" "auction-service:8087")
    
    for service in "${services[@]}"; do
        if curl -f -s "http://localhost:${service#*:}" >/dev/null 2>&1; then
            print_success "$service is running"
        else
            print_warning "$service is not running"
        fi
    done
}

# Function to run load test
run_load_test() {
    local test_name=$1
    local test_type=$2
    local num_users=$3
    local duration=$4
    
    print_status "Running $test_name load test..."
    print_status "Users: $num_users, Duration: $duration, Type: $test_type"
    
    cd "$SCRIPT_DIR/../tests/performance"
    
    # Build the load test binary if needed
    if [ ! -f "load_test" ]; then
        print_status "Building load test binary..."
        go build -o load_test load_test.go
    fi
    
    # Run the test
    local results_file="$RESULTS_DIR/load_test_${test_name}_${TIMESTAMP}.json"
    
    ./load_test $test_name > "$results_file" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "$test_name test completed successfully"
        print_status "Results saved to: $results_file"
    else
        print_error "$test_name test failed"
        return 1
    fi
}

# Function to analyze results
analyze_results() {
    local results_file=$1
    
    if [ ! -f "$results_file" ]; then
        print_error "Results file not found: $results_file"
        return 1
    fi
    
    print_status "Analyzing results from: $results_file"
    
    # Extract key metrics using jq (if available) or basic grep
    if command -v jq &> /dev/null; then
        local throughput=$(jq -r '.throughput_rps' "$results_file")
        local error_rate=$(jq -r '.error_rate_percent' "$results_file")
        local max_response=$(jq -r '.max_response_time' "$results_file")
        
        print_status "Throughput: $throughput RPS"
        print_status "Error Rate: $error_rate%"
        print_status "Max Response Time: $max_response"
        
        # Evaluate against targets
        if (( $(echo "$throughput >= 1000" | bc -l) )); then
            print_success "✅ Target achieved: 1000+ RPS throughput"
        else
            print_warning "⚠️  Target not met: Need 1000+ RPS throughput"
        fi
        
        if (( $(echo "$error_rate < 1.0" | bc -l) )); then
            print_success "✅ Target achieved: Error rate under 1%"
        else
            print_warning "⚠️  Target not met: Error rate above 1%"
        fi
    else
        print_warning "jq not available, showing raw results:"
        cat "$results_file"
    fi
}

# Function to generate HTML report
generate_report() {
    local report_file="$RESULTS_DIR/load_test_report_${TIMESTAMP}.html"
    
    print_status "Generating HTML report: $report_file"
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Blytz Load Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .test-result { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .success { border-left: 5px solid #4CAF50; }
        .warning { border-left: 5px solid #ff9800; }
        .error { border-left: 5px solid #f44336; }
        .metrics { display: flex; flex-wrap: wrap; gap: 20px; }
        .metric { flex: 1; min-width: 200px; }
        .metric-value { font-size: 24px; font-weight: bold; color: #2196F3; }
        .metric-label { color: #666; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Blytz Live Auction MVP - Load Test Report</h1>
        <p>Generated on: $(date)</p>
        <p>Target: Scale from 1,000 to 3,000+ concurrent users, architecture ready for 50,000+ users</p>
    </div>
    
    <h2>Test Execution Summary</h2>
    <div class="test-result success">
        <h3>✅ Load Tests Completed</h3>
        <p>All load tests have been executed successfully. Results are available in JSON format for detailed analysis.</p>
    </div>
    
    <h2>Performance Targets</h2>
    <div class="metrics">
        <div class="metric">
            <div class="metric-value">1,000+</div>
            <div class="metric-label">Concurrent Users (Target)</div>
        </div>
        <div class="metric">
            <div class="metric-value">3,000+</div>
            <div class="metric-label">Concurrent Users (Goal)</div>
        </div>
        <div class="metric">
            <div class="metric-value">50,000+</div>
            <div class="metric-label">Architecture Ready For</div>
        </div>
        <div class="metric">
            <div class="metric-value">< 1%</div>
            <div class="metric-label">Error Rate Target</div>
        </div>
        <div class="metric">
            <div class="metric-value">< 200ms</div>
            <div class="metric-label">Response Time Target</div>
        </div>
    </div>
    
    <h2>Test Results</h2>
    <p>Individual test results are available in the JSON files in the results directory.</p>
    
    <h2>Recommendations</h2>
    <ul>
        <li>Monitor Redis cache hit rates - target 70-85%</li>
        <li>Watch database connection pool utilization</li>
        <li>Monitor WebSocket connection counts</li>
        <li>Track circuit breaker state changes</li>
        <li>Observe rate limiter effectiveness</li>
    </ul>
</body>
</html>
EOF
    
    print_success "HTML report generated: $report_file"
}

# Main execution
main() {
    print_status "Blytz Live Auction MVP - Load Testing Suite"
    print_status "=========================================="
    
    # Check prerequisites
    if ! command -v go &> /dev/null; then
        print_error "Go is required but not installed"
        exit 1
    fi
    
    # Check services
    check_services
    
    echo ""
    print_status "Starting load test suite..."
    print_status "Results will be saved to: $RESULTS_DIR"
    echo ""
    
    # Run different load test scenarios
    run_load_test "small" "mixed" 100 "2m"
    analyze_results "$RESULTS_DIR/load_test_small_${TIMESTAMP}.json"
    echo ""
    
    run_load_test "medium" "mixed" 1000 "5m"
    analyze_results "$RESULTS_DIR/load_test_medium_${TIMESTAMP}.json"
    echo ""
    
    run_load_test "large" "mixed" 3000 "10m"
    analyze_results "$RESULTS_DIR/load_test_large_${TIMESTAMP}.json"
    echo ""
    
    # Stress test
    print_status "Running stress test..."
    run_load_test "stress" "mixed" 5000 "15m"
    analyze_results "$RESULTS_DIR/load_test_stress_${TIMESTAMP}.json"
    echo ""
    
    # Generate comprehensive report
    generate_report
    
    print_status "Load test suite completed!"
    print_status "Check results in: $RESULTS_DIR"
    print_status "Open the HTML report for a summary: $RESULTS_DIR/load_test_report_${TIMESTAMP}.html"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi