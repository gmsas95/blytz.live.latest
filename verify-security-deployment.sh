#!/bin/bash

echo "🔍 Verifying Security Implementation Deployment..."
echo "================================================"

# Check if authentication middleware is actually used in services
echo ""
echo "1. Checking Gateway Authentication Integration:"
grep -n "GinAuthMiddleware\|auth.*Middleware" /home/sas/blytzmvp-clean/services/gateway/cmd/main.go || echo "❌ No authentication middleware found in gateway"

# Check for remaining mock authentication
echo ""
echo "2. Checking for Remaining Mock Authentication:"
echo "Mock user IDs found:"
grep -rn "mock-user-id\|dummy-user-id" /home/sas/blytzmvp-clean/services/ | grep -v ".git" | head -5

echo ""
echo "MockAuthMiddleware usage:"
grep -rn "MockAuthMiddleware" /home/sas/blytzmvp-clean/services/ | grep -v ".git" | head -5

# Check JWT validation implementation
echo ""
echo "3. Checking JWT Validation Implementation:"
grep -rn "ValidateToken\|jwt.*Validate" /home/sas/blytzmvp-clean/services/ | grep -v ".git" | head -5

# Check WebSocket authentication
echo ""
echo "4. Checking WebSocket Authentication:"
grep -rn "user_id.*Query\|Query.*user_id" /home/sas/blytzmvp-clean/services/chat-service/ | head -3

echo ""
echo "✅ Verification complete!"
echo ""
echo "🔧 TO FIX: Run these commands to deploy security:"
echo "1. Add auth middleware to gateway"
echo "2. Replace mock-user-id with real authentication"
echo "3. Implement JWT validation for WebSocket"
echo "4. Remove MockAuthMiddleware usage"