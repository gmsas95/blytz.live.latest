package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/handlers"
)

// TransferRoutes defines transfer-related routes
func TransferRoutes(router *gin.RouterGroup, transferHandler *handlers.TransferHandler, authMiddleware gin.HandlerFunc) {
	transfers := router.Group("/transfers")
	{
		// Public routes (if any)
		
		// Protected routes
		transfers.Use(authMiddleware)
		{
			// Create transfer
			transfers.POST("", transferHandler.CreateTransfer)
			
			// Create batch transfers
			transfers.POST("/batch", transferHandler.CreateBatchTransfers)
			
			// Get transfer by ID
			transfers.GET("/:id", transferHandler.GetTransfer)
			
			// Get transfer history for account
			transfers.GET("/account/:account_id", transferHandler.GetTransferHistory)
			
			// Reverse transfer
			transfers.POST("/:id/reverse", transferHandler.ReverseTransfer)
		}
	}
}

// TransferAdminRoutes defines admin-only transfer routes
func TransferAdminRoutes(router *gin.RouterGroup, transferHandler *handlers.TransferHandler, adminAuthMiddleware gin.HandlerFunc) {
	transfers := router.Group("/transfers")
	{
		// Admin-only routes
		transfers.Use(adminAuthMiddleware)
		{
			// Admin can access all transfer operations
			transfers.GET("/admin/all", transferHandler.GetTransferHistory) // Would need to modify handler for admin access
		}
	}
}

// PayoutRoutes defines payout-related routes
func PayoutRoutes(router *gin.RouterGroup, payoutHandler *handlers.PayoutHandler, authMiddleware gin.HandlerFunc) {
	payouts := router.Group("/payouts")
	{
		// Public routes (if any)
		
		// Protected routes
		payouts.Use(authMiddleware)
		{
			// Create payout
			payouts.POST("", payoutHandler.CreatePayout)
			
			// Create scheduled payout
			payouts.POST("/scheduled", payoutHandler.CreateScheduledPayout)
			
			// Get payout by ID
			payouts.GET("/:id", payoutHandler.GetPayout)
			
			// Get payout history for account
			payouts.GET("/account/:account_id", payoutHandler.GetPayoutHistory)
			
			// Cancel payout
			payouts.POST("/:id/cancel", payoutHandler.CancelPayout)
			
			// Get available balance for account
			payouts.GET("/balance/:account_id", payoutHandler.GetAvailableBalance)
		}
	}
}

// PayoutAdminRoutes defines admin-only payout routes
func PayoutAdminRoutes(router *gin.RouterGroup, payoutHandler *handlers.PayoutHandler, adminAuthMiddleware gin.HandlerFunc) {
	payouts := router.Group("/payouts")
	{
		// Admin-only routes
		payouts.Use(adminAuthMiddleware)
		{
			// Admin can access all payout operations
			payouts.GET("/admin/all", payoutHandler.GetPayoutHistory) // Would need to modify handler for admin access
		}
	}
}

// TransferPayoutRoutes combines all transfer and payout routes
func TransferPayoutRoutes(
	router *gin.RouterGroup, 
	transferHandler *handlers.TransferHandler,
	payoutHandler *handlers.PayoutHandler,
	authMiddleware gin.HandlerFunc,
	adminAuthMiddleware gin.HandlerFunc,
) {
	// Transfer routes
	TransferRoutes(router, transferHandler, authMiddleware)
	TransferAdminRoutes(router, transferHandler, adminAuthMiddleware)
	
	// Payout routes
	PayoutRoutes(router, payoutHandler, authMiddleware)
	PayoutAdminRoutes(router, payoutHandler, adminAuthMiddleware)
}