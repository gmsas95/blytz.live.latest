package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/handlers"
	"github.com/gmsas95/blytz-mvp/shared/pkg/auth"
)

// ConnectedAccountRoutes sets up routes for connected account operations
func ConnectedAccountRoutes(
	router *gin.RouterGroup,
	handler *handlers.ConnectedAccountHandler,
	authMiddleware gin.HandlerFunc,
) {
	accounts := router.Group("/accounts")
	accounts.Use(authMiddleware)
	{
		// User routes (accessible by account owners)
		accounts.POST("", handler.CreateConnectedAccount)
		accounts.GET("/me", handler.GetConnectedAccountByUserID)
		accounts.GET("/:id", handler.GetConnectedAccount)
		accounts.PUT("/:id", handler.UpdateConnectedAccount)
		accounts.DELETE("/:id", handler.DeleteConnectedAccount)
		accounts.POST("/:id/onboarding", handler.GetAccountOnboardingLink)
		accounts.GET("/:id/login", handler.GetAccountLoginLink)
		accounts.GET("/:id/balance", handler.GetAccountBalance)
		accounts.GET("/:id/transactions", handler.GetAccountTransactions)
	}
}

// ConnectedAccountAdminRoutes sets up admin-only routes for connected account operations
func ConnectedAccountAdminRoutes(
	router *gin.RouterGroup,
	handler *handlers.ConnectedAccountHandler,
	adminAuthMiddleware gin.HandlerFunc,
) {
	adminAccounts := router.Group("/admin/accounts")
	adminAccounts.Use(adminAuthMiddleware)
	{
		// Admin-only routes
		adminAccounts.GET("", handler.ListConnectedAccounts)
		adminAccounts.GET("/:id", handler.GetConnectedAccount)
		adminAccounts.PUT("/:id/status", func(c *gin.Context) {
			// This would be an admin-only status update endpoint
			// Implementation would go here
		})
		adminAccounts.DELETE("/:id", handler.DeleteConnectedAccount)
	}
}

// SetupAuthMiddleware creates the authentication middleware
func SetupAuthMiddleware(authClient *auth.AuthClient) gin.HandlerFunc {
	return auth.GinAuthMiddleware(authClient)
}

// SetupAdminAuthMiddleware creates the admin authentication middleware
func SetupAdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is admin
		userRole, exists := c.Get("userRole")
		if !exists || userRole != "admin" {
			c.JSON(403, gin.H{
				"success": false,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}