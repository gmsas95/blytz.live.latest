package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	shared_errors "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/services"
)

type AuthMiddleware struct {
	authService *services.AuthService
	logger      *zap.Logger
}

func NewAuthMiddleware(authService *services.AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth middleware that requires valid JWT token
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			shared_utils.SendErrorResponse(c, shared_errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			shared_utils.SendErrorResponse(c, shared_errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// Validate token
		token := tokenParts[1]
		tokenResponse, err := m.authService.ValidateToken(token)
		if err != nil || !tokenResponse.Valid {
			m.logger.Warn("Invalid token provided",
				zap.String("token", token[:min(len(token), 20)]+"..."),
				zap.Error(err))
			shared_utils.SendErrorResponse(c, shared_errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", tokenResponse.UserID)
		c.Set("user_email", tokenResponse.Email)
		c.Set("user_role", tokenResponse.Role)

		m.logger.Debug("User authenticated successfully",
			zap.String("user_id", tokenResponse.UserID),
			zap.String("email", tokenResponse.Email),
			zap.String("role", tokenResponse.Role))

		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		userRole, exists := c.Get("user_role")
		if !exists {
			shared_utils.SendErrorResponse(c, shared_errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check if user has required role
		userRoleStr, ok := userRole.(string)
		if !ok || userRoleStr != requiredRole {
			m.logger.Warn("User lacks required role",
				zap.String("user_role", userRoleStr),
				zap.String("required_role", requiredRole))
			shared_utils.SendErrorResponse(c, shared_errors.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}

// OptionalAuth middleware that optionally validates token if provided
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		// Validate token
		token := tokenParts[1]
		tokenResponse, err := m.authService.ValidateToken(token)
		if err != nil || !tokenResponse.Valid {
			// Invalid token, continue without authentication
			m.logger.Debug("Invalid optional token provided",
				zap.String("token", token[:min(len(token), 20)]+"..."))
			c.Next()
			return
		}

		// Set user context if token is valid
		c.Set("user_id", tokenResponse.UserID)
		c.Set("user_email", tokenResponse.Email)
		c.Set("user_role", tokenResponse.Role)

		m.logger.Debug("User optionally authenticated",
			zap.String("user_id", tokenResponse.UserID),
			zap.String("email", tokenResponse.Email),
			zap.String("role", tokenResponse.Role))

		c.Next()
	}
}

// ExtractUserID extracts user ID from context
func ExtractUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// ExtractUserEmail extracts user email from context
func ExtractUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}
	emailStr, ok := email.(string)
	return emailStr, ok
}

// ExtractUserRole extracts user role from context
func ExtractUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	return roleStr, ok
}

// RequireOwnership middleware that requires user to own the resource
func (m *AuthMiddleware) RequireOwnership(resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user ID
		userID, exists := ExtractUserID(c)
		if !exists {
			shared_utils.SendErrorResponse(c, shared_errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// Get resource ID from URL parameters
		resourceID := c.Param(resourceIDParam)
		if resourceID == "" {
			shared_utils.SendErrorResponse(c, shared_errors.ErrInvalidRequest)
			c.Abort()
			return
		}

		// For now, simple ownership check (user ID == resource ID)
		// In production, this would check database for actual ownership
		if userID != resourceID {
			m.logger.Warn("User attempted to access resource they don't own",
				zap.String("user_id", userID),
				zap.String("resource_id", resourceID))
			shared_utils.SendErrorResponse(c, shared_errors.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
