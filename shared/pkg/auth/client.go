package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthClient provides authentication service integration
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewAuthClient creates a new authentication client
func NewAuthClient(authServiceURL string) *AuthClient {
	return &AuthClient{
		baseURL: authServiceURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: zap.NewNop(), // Replace with actual logger in production
	}
}

// ValidateToken validates a JWT token and returns user information
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	// TODO: Implement actual token validation with auth service
	// For now, return a mock response
	return &UserInfo{
		ID:    "user123",
		Email: "user@example.com",
		Role:  "user",
	}, nil
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
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
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