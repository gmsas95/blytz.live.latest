package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/shared/pkg/jwt"
	"go.uber.org/zap"
)

// AuthClient provides authentication service integration
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
	jwtConfig  *jwt.Config
}

// NewAuthClient creates a new authentication client
func NewAuthClient(authServiceURL string) *AuthClient {
	logger, _ := zap.NewProduction()
	return &AuthClient{
		baseURL: authServiceURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:    logger,
		jwtConfig: jwt.NewConfig(),
	}
}

// ValidateToken validates a JWT token and returns user information
func (c *AuthClient) ValidateToken(ctx context.Context, tokenString string) (*UserInfo, error) {
	// First, try local JWT validation
	claims, err := c.jwtConfig.ValidateToken(tokenString)
	if err != nil {
		c.logger.Warn("JWT validation failed",
			zap.String("token", tokenString[:min(len(tokenString), 20)]+"..."),
			zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract user information from claims
	userID, err := jwt.ExtractUserID(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user ID: %w", err)
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("email not found in token claims")
	}

	role, err := jwt.ExtractRole(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to extract role: %w", err)
	}

	// Log successful authentication
	c.logger.Info("Token validated successfully",
		zap.String("user_id", userID),
		zap.String("email", email),
		zap.String("role", role))

	return &UserInfo{
		ID:    userID,
		Email: email,
		Role:  role,
	}, nil
}

// ValidateTokenWithService validates token via auth service (fallback)
func (c *AuthClient) ValidateTokenWithService(ctx context.Context, tokenString string) (*UserInfo, error) {
	// Create request to auth service
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/auth/validate", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Success bool       `json:"success"`
		User    UserInfo    `json:"user"`
		Message string      `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("auth service error: %s", result.Message)
	}

	return &result.User, nil
}

// UserInfo represents authenticated user information
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// GinAuthMiddleware creates a Gin middleware for authentication
func GinAuthMiddleware(authClient *AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token (assuming Bearer token format)
		token := strings.TrimSpace(authHeader)
		if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimSpace(token[7:])
		}

		// Validate token format
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token format",
			})
			c.Abort()
			return
		}

		// Validate token
		userInfo, err := authClient.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authentication token",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("userID", userInfo.ID)
		c.Set("userEmail", userInfo.Email)
		c.Set("userRole", userInfo.Role)

		c.Next()
	}
}

// MockAuthMiddleware provides a mock authentication middleware for development
func MockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set mock user information for development
		c.Set("userID", "mock-user-123")
		c.Set("userEmail", "mock@example.com")
		c.Set("userRole", "user")
		c.Next()
	}
}