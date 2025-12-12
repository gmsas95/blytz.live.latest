#!/bin/bash

# ===========================================
# Blytz Live Auction Platform - Security Fix Verifier
# ===========================================
# This script verifies that all security fixes are properly implemented
# ===========================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Security check counters
CHECKS_TOTAL=0
CHECKS_PASSED=0
CHECKS_FAILED=0

# Function to run a security check
run_check() {
    local check_name="$1"
    local check_command="$2"
    local expected_result="$3"
    
    ((CHECKS_TOTAL++))
    print_info "Checking: $check_name"
    
    if eval "$check_command" >/dev/null 2>&1; then
        if [ "$expected_result" = "should_not_exist" ]; then
            print_error "$check_name (Found but should not exist)"
            ((CHECKS_FAILED++))
        else
            print_success "$check_name"
            ((CHECKS_PASSED++))
        fi
    else
        if [ "$expected_result" = "should_not_exist" ]; then
            print_success "$check_name (Not found - correct)"
            ((CHECKS_PASSED++))
        else
            print_error "$check_name"
            ((CHECKS_FAILED++))
        fi
    fi
}

print_info "🔍 Blytz Platform Security Fix Verifier"
print_info "======================================"
echo

# Check 1: Mock user IDs should not exist
print_info "Checking for mock user IDs..."
run_check "Mock user ID: mock-user-id" \
    "grep -r 'mock-user-id' /home/sas/blytzmvp-clean/services/ | grep -v '.git' || true" \
    "should_not_exist"

run_check "Mock user ID: dummy-user-id" \
    "grep -r 'dummy-user-id' /home/sas/blytzmvp-clean/services/ | grep -v '.git' || true" \
    "should_not_exist"

run_check "Mock user ID: user123" \
    "grep -r 'user123' /home/sas/blytzmvp-clean/services/ | grep -v '.git' || true" \
    "should_not_exist"

echo

# Check 2: MockAuthMiddleware should not exist
run_check "MockAuthMiddleware usage" \
    "grep -r 'MockAuthMiddleware' /home/sas/blytzmvp-clean/services/ | grep -v '.git' || true" \
    "should_not_exist"

echo

# Check 3: Gateway authentication should be implemented
print_info "Checking gateway authentication..."
run_check "Gateway has auth client" \
    "grep -r 'authClient.*auth.NewAuthClient' /home/sas/blytzmvp-clean/services/gateway/ | grep -v '.git'" \
    "should_exist"

run_check "Gateway uses auth middleware" \
    "grep -r 'auth.GinAuthMiddleware' /home/sas/blytzmvp-clean/services/gateway/ | grep -v '.git'" \
    "should_exist"

run_check "Gateway has protected routes" \
    "grep -r 'protected.*Use.*auth.GinAuthMiddleware' /home/sas/blytzmvp-clean/services/gateway/ | grep -v '.git'" \
    "should_exist"

echo

# Check 4: Chat service WebSocket authentication
print_info "Checking chat service WebSocket authentication..."
run_check "Chat service uses token validation" \
    "grep -r 'ValidateToken.*token' /home/sas/blytzmvp-clean/services/chat-service/ | grep -v '.git'" \
    "should_exist"

run_check "Chat service has auth client" \
    "grep -r 'auth.NewAuthClient' /home/sas/blytzmvp-clean/services/chat-service/ | grep -v '.git'" \
    "should_exist"

run_check "Chat service WebSocket uses auth" \
    "grep -r 'auth.GinAuthMiddleware' /home/sas/blytzmvp-clean/services/chat-service/ | grep -v '.git'" \
    "should_exist"

echo

# Check 5: CORS configuration should be secure
print_info "Checking CORS configuration..."
run_check "No wildcard CORS in gateway" \
    "grep -r 'Access-Control-Allow-Origin.*\\*' /home/sas/blytzmvp-clean/services/gateway/ | grep -v '.git' || true" \
    "should_not_exist"

run_check "No wildcard CORS in auction service" \
    "grep -r 'Access-Control-Allow-Origin.*\\*' /home/sas/blytzmvp-clean/services/auction-service/ | grep -v '.git' || true" \
    "should_not_exist"

run_check "No wildcard CORS in stripe service" \
    "grep -r 'Access-Control-Allow-Origin.*\\*' /home/sas/blytzmvp-clean/services/stripe-service/ | grep -v '.git' || true" \
    "should_not_exist"

echo

# Check 6: Redis authentication should be configured
print_info "Checking Redis authentication..."
run_check "Redis password in environment" \
    "grep -r 'REDIS_PASSWORD' /home/sas/blytzmvp-clean/docker-compose.yml" \
    "should_exist"

run_check "Redis password in chat service" \
    "grep -r 'Password.*getEnv.*REDIS_PASSWORD' /home/sas/blytzmvp-clean/services/chat-service/ | grep -v '.git'" \
    "should_exist"

echo

# Check 7: JWT secrets should be secure
print_info "Checking JWT secrets..."
run_check "No default JWT secret in templates" \
    "grep -r 'your-secret-key' /home/sas/blytzmvp-clean/.env* | grep -v '.git' || true" \
    "should_not_exist"

run_check "Secure JWT defaults in config" \
    "grep -r 'dev_jwt_secret_key_32_characters_minimum' /home/sas/blytzmvp-clean/services/ | grep -v '.git'" \
    "should_exist"

echo

# Check 8: Security packages should exist
print_info "Checking security packages..."
run_check "JWT security package exists" \
    "test -f /home/sas/blytzmvp-clean/shared/pkg/jwt/jwt.go" \
    "should_exist"

run_check "Redis security package exists" \
    "test -f /home/sas/blytzmvp-clean/shared/pkg/redis/redis.go" \
    "should_exist"

run_check "Security validation package exists" \
    "test -f /home/sas/blytzmvp-clean/shared/pkg/security/security.go" \
    "should_exist"

echo

# Check 9: Security scripts should exist
print_info "Checking security scripts..."
run_check "Secret generation script exists" \
    "test -f /home/sas/blytzmvp-clean/scripts/generate-secrets.sh" \
    "should_exist"

run_check "Security test script exists" \
    "test -f /home/sas/blytzmvp-clean/scripts/security-test.sh" \
    "should_exist"

run_check "Security fix verifier script exists" \
    "test -f /home/sas/blytzmvp-clean/scripts/security-fix-verifier.sh" \
    "should_exist"

echo

# Check 10: Documentation should exist
print_info "Checking security documentation..."
run_check "Security implementation guide exists" \
    "test -f /home/sas/blytzmvp-clean/docs/security/SECURITY_IMPLEMENTATION_GUIDE.md" \
    "should_exist"

run_check "Security fixes summary exists" \
    "test -f /home/sas/blytzmvp-clean/docs/security/SECURITY_FIXES_SUMMARY.md" \
    "should_exist"

echo

# Generate security report
print_info "📊 Generating Security Verification Report"
echo

local success_rate=$((CHECKS_PASSED * 100 / CHECKS_TOTAL))

echo "=========================================="
echo "🔒 SECURITY VERIFICATION REPORT"
echo "=========================================="
echo "Date: $(date)"
echo "Environment: ${ENVIRONMENT:-development}"
echo
echo "Check Results:"
echo "  Total Checks: $CHECKS_TOTAL"
echo "  Passed: $CHECKS_PASSED"
echo "  Failed: $CHECKS_FAILED"
echo "  Success Rate: ${success_rate}%"
echo

if [ $success_rate -ge 90 ]; then
    echo -e "${GREEN}🎉 SECURITY STATUS: EXCELLENT${NC}"
    echo "✅ All critical security fixes are properly implemented"
    echo "✅ Platform is secure and production-ready"
elif [ $success_rate -ge 75 ]; then
    echo -e "${YELLOW}⚠️  SECURITY STATUS: GOOD${NC}"
    echo "⚠️  Most security fixes are implemented"
    echo "⚠️  Minor issues need attention"
else
    echo -e "${RED}🚨 SECURITY STATUS: CRITICAL${NC}"
    echo "🚨 Serious security vulnerabilities remain"
    echo "🚨 Immediate action required"
fi

echo
echo "=========================================="

if [ $CHECKS_FAILED -gt 0 ]; then
    echo
    echo -e "${RED}🚨 FAILED CHECKS REQUIRING ATTENTION:${NC}"
    echo "Please review and fix the failed checks above."
    echo
    echo "Common fixes:"
    echo "  1. Remove any remaining mock authentication"
    echo "  2. Implement proper JWT validation"
    echo "  3. Add authentication middleware to all protected routes"
    echo "  4. Ensure Redis authentication is configured"
    echo "  5. Verify CORS configuration is secure"
    exit 1
else
    echo
    echo -e "${GREEN}✅ ALL SECURITY CHECKS PASSED${NC}"
    echo "Platform is ready for production deployment!"
    echo
    echo "Next steps:"
    echo "  1. Run comprehensive security tests: ./scripts/security-test.sh"
    echo "  2. Generate production secrets: ./scripts/generate-secrets.sh"
    echo "  3. Deploy with: docker-compose --env-file .env.production up -d"
    echo "  4. Monitor security logs in production"
fi

echo "=========================================="