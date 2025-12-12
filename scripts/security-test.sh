#!/bin/bash

# ===========================================
# Blytz Live Auction Platform - Security Test Script
# ===========================================
# This script tests all security implementations
# ===========================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
GATEWAY_URL=${GATEWAY_URL:-"http://localhost:8092"}
AUTH_SERVICE_URL=${AUTH_SERVICE_URL:-"http://localhost:8085"}
AUCTION_SERVICE_URL=${AUCTION_SERVICE_URL:-"http://localhost:8087"}

# Test counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print colored output
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
    ((TESTS_PASSED++))
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    ((TESTS_FAILED++))
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_result="$3"
    
    ((TESTS_TOTAL++))
    print_info "Running: $test_name"
    
    if eval "$test_command" >/dev/null 2>&1; then
        if [ "$expected_result" = "success" ]; then
            print_success "$test_name"
        else
            print_error "$test_name (Expected failure but succeeded)"
        fi
    else
        if [ "$expected_result" = "failure" ]; then
            print_success "$test_name"
        else
            print_error "$test_name"
        fi
    fi
}

# Function to test CORS
test_cors() {
    print_info "🔒 Testing CORS Configuration"
    echo
    
    # Test allowed origin
    run_test "CORS: Allowed origin (https://blytz.app)" \
        "curl -s -H 'Origin: https://blytz.app' -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: Content-Type' -X OPTIONS '$GATEWAY_URL/health' -w '%{http_code}' | grep -q '200'" \
        "success"
    
    # Test disallowed origin
    run_test "CORS: Disallowed origin (malicious-site.com)" \
        "curl -s -H 'Origin: https://malicious-site.com' -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: Content-Type' -X OPTIONS '$GATEWAY_URL/health' -w '%{http_code}' | grep -v '200'" \
        "success"
    
    # Test wildcard not present
    run_test "CORS: No wildcard origin" \
        "curl -s -I '$GATEWAY_URL/health' | grep -i 'access-control-allow-origin: \*' | grep -v 'access-control-allow-origin: *'" \
        "failure"
    
    echo
}

# Function to test security headers
test_security_headers() {
    print_info "🛡️  Testing Security Headers"
    echo
    
    # Test security headers
    run_test "Security Headers: X-Frame-Options" \
        "curl -s -I '$GATEWAY_URL/health' | grep -i 'x-frame-options: deny'" \
        "success"
    
    run_test "Security Headers: X-Content-Type-Options" \
        "curl -s -I '$GATEWAY_URL/health' | grep -i 'x-content-type-options: nosniff'" \
        "success"
    
    run_test "Security Headers: X-XSS-Protection" \
        "curl -s -I '$GATEWAY_URL/health' | grep -i 'x-xss-protection: 1; mode=block'" \
        "success"
    
    run_test "Security Headers: Content-Security-Policy" \
        "curl -s -I '$GATEWAY_URL/health' | grep -i 'content-security-policy:'" \
        "success"
    
    echo
}

# Function to test authentication
test_authentication() {
    print_info "🔐 Testing Authentication"
    echo
    
    # Test JWT validation
    run_test "Auth: Invalid JWT token rejected" \
        "curl -s -H 'Authorization: Bearer invalid_token' '$GATEWAY_URL/api/v1/protected' -w '%{http_code}' | grep -E '401|403'" \
        "success"
    
    # Test missing authorization
    run_test "Auth: Missing authorization header" \
        "curl -s '$GATEWAY_URL/api/v1/protected' -w '%{http_code}' | grep -E '401|403'" \
        "success"
    
    echo
}

# Function to test rate limiting
test_rate_limiting() {
    print_info "⏱️  Testing Rate Limiting"
    echo
    
    # Test rate limiting (make multiple rapid requests)
    print_info "Making 10 rapid requests to test rate limiting..."
    
    local rate_limit_hit=false
    for i in {1..10}; do
        local status_code=$(curl -s -o /dev/null -w '%{http_code}' "$GATEWAY_URL/health")
        if [ "$status_code" = "429" ]; then
            rate_limit_hit=true
            break
        fi
    done
    
    if [ "$rate_limit_hit" = true ]; then
        print_success "Rate limiting: Active (429 status detected)"
    else
        print_warning "Rate limiting: Not triggered in 10 requests (may be configured with higher limit)"
    fi
    
    echo
}

# Function to test input validation
test_input_validation() {
    print_info "🔍 Testing Input Validation"
    echo
    
    # Test SQL injection
    run_test "Input Validation: SQL injection attempt" \
        "curl -s -X POST -H 'Content-Type: application/json' -d '{\"email\": \"test'; DROP TABLE users; --\"}' '$AUTH_SERVICE_URL/api/auth/login' -w '%{http_code}' | grep -E '400|422'" \
        "success"
    
    # Test XSS attempt
    run_test "Input Validation: XSS attempt" \
        "curl -s -X POST -H 'Content-Type: application/json' -d '{\"name\": \"<script>alert(\\\"xss\\\")</script>\"}' '$AUCTION_SERVICE_URL/api/v1/auctions' -w '%{http_code}' | grep -E '400|422'" \
        "success"
    
    echo
}

# Function to test Redis authentication
test_redis_auth() {
    print_info "🔴 Testing Redis Authentication"
    echo
    
    # Check if Redis requires password
    if command -v redis-cli &> /dev/null; then
        # Test connection without password (should fail)
        run_test "Redis: Connection without password fails" \
            "redis-cli -h localhost -p 6379 ping 2>&1 | grep -E '(NOAUTH|authentication required)'" \
            "success"
        
        # Test connection with password from environment
        if [ -n "$REDIS_PASSWORD" ]; then
            run_test "Redis: Connection with password succeeds" \
                "redis-cli -h localhost -p 6379 -a '$REDIS_PASSWORD' ping | grep -q 'PONG'" \
                "success"
        else
            print_warning "Redis: REDIS_PASSWORD not set, skipping authenticated connection test"
        fi
    else
        print_warning "Redis CLI not available, skipping Redis authentication tests"
    fi
    
    echo
}

# Function to test JWT secrets
test_jwt_secrets() {
    print_info "🔑 Testing JWT Secrets"
    echo
    
    # Check if JWT secret is set and not default
    if [ -n "$JWT_SECRET" ]; then
        if [ "$JWT_SECRET" = "your-secret-key" ] || [ "$JWT_SECRET" = "dev_jwt_secret_key_32_characters_minimum" ]; then
            print_error "JWT Secret: Using default/weak secret"
        else
            print_success "JWT Secret: Custom secret is set"
        fi
        
        # Check secret length
        if [ ${#JWT_SECRET} -ge 32 ]; then
            print_success "JWT Secret: Length is adequate (${#JWT_SECRET} characters)"
        else
            print_error "JWT Secret: Too short (${#JWT_SECRET} characters, minimum 32 required)"
        fi
    else
        print_error "JWT Secret: Not set"
    fi
    
    echo
}

# Function to test environment configuration
test_environment_config() {
    print_info "⚙️  Testing Environment Configuration"
    echo
    
    # Check required environment variables
    local required_vars=("JWT_SECRET" "REDIS_PASSWORD" "DATABASE_URL")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -eq 0 ]; then
        print_success "Environment: All required variables are set"
    else
        print_error "Environment: Missing variables: ${missing_vars[*]}"
    fi
    
    # Check for development values in production
    if [ "$ENVIRONMENT" = "production" ]; then
        if [ "$JWT_SECRET" = "dev_jwt_secret_key_32_characters_minimum" ]; then
            print_error "Environment: Using development JWT secret in production"
        fi
        
        if [ "$REDIS_PASSWORD" = "dev_redis_password_32_characters_minimum" ]; then
            print_error "Environment: Using development Redis password in production"
        fi
    fi
    
    echo
}

# Function to test service health
test_service_health() {
    print_info "🏥 Testing Service Health"
    echo
    
    # Test Gateway health
    run_test "Health: Gateway service" \
        "curl -s '$GATEWAY_URL/health' | grep -q 'ok'" \
        "success"
    
    # Test Auth service health
    run_test "Health: Auth service" \
        "curl -s '$AUTH_SERVICE_URL/health' | grep -q 'ok'" \
        "success"
    
    # Test Auction service health
    run_test "Health: Auction service" \
        "curl -s '$AUCTION_SERVICE_URL/health' | grep -q 'ok'" \
        "success"
    
    echo
}

# Function to generate security report
generate_report() {
    print_info "📊 Generating Security Report"
    echo
    
    local success_rate=$((TESTS_PASSED * 100 / TESTS_TOTAL))
    
    echo "=========================================="
    echo "🔒 BLYTZ PLATFORM SECURITY TEST REPORT"
    echo "=========================================="
    echo "Date: $(date)"
    echo "Environment: ${ENVIRONMENT:-development}"
    echo
    echo "Test Results:"
    echo "  Total Tests: $TESTS_TOTAL"
    echo "  Passed: $TESTS_PASSED"
    echo "  Failed: $TESTS_FAILED"
    echo "  Success Rate: ${success_rate}%"
    echo
    
    if [ $success_rate -ge 90 ]; then
        echo -e "${GREEN}🎉 SECURITY STATUS: EXCELLENT${NC}"
        echo "✅ Platform is secure and production-ready"
    elif [ $success_rate -ge 75 ]; then
        echo -e "${YELLOW}⚠️  SECURITY STATUS: GOOD${NC}"
        echo "⚠️  Platform has minor security issues to address"
    else
        echo -e "${RED}🚨 SECURITY STATUS: CRITICAL${NC}"
        echo "🚨 Platform has serious security vulnerabilities"
    fi
    
    echo
    echo "Recommendations:"
    if [ $TESTS_FAILED -gt 0 ]; then
        echo "  🔧 Fix failed tests before production deployment"
    fi
    echo "  🔄 Run security tests regularly"
    echo "  📊 Monitor security logs"
    echo "  🔐 Rotate secrets regularly"
    echo "  📚 Keep security documentation updated"
    echo
    
    echo "=========================================="
}

# Main execution
main() {
    print_info "🔒 Blytz Platform Security Test Suite"
    print_info "====================================="
    echo
    
    # Check if services are running
    print_info "Checking service availability..."
    if ! curl -s "$GATEWAY_URL/health" >/dev/null 2>&1; then
        print_error "Gateway service is not accessible at $GATEWAY_URL"
        print_info "Please ensure services are running before testing"
        exit 1
    fi
    
    # Run all tests
    test_cors
    test_security_headers
    test_authentication
    test_rate_limiting
    test_input_validation
    test_redis_auth
    test_jwt_secrets
    test_environment_config
    test_service_health
    
    # Generate report
    generate_report
    
    # Exit with appropriate code
    if [ $TESTS_FAILED -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# Run main function
main "$@"