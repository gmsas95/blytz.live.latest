package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/handlers"
)

// PaymentRoutes defines all payment-related routes
func PaymentRoutes(router *gin.RouterGroup, paymentHandler *handlers.PaymentHandler, authMiddleware gin.HandlerFunc) {
	// Payment routes with authentication
	payments := router.Group("/payments")
	payments.Use(authMiddleware)
	{
		// Create payment intent
		payments.POST("/create-intent", paymentHandler.CreatePaymentIntent)

		// Confirm payment
		payments.POST("/confirm", paymentHandler.ConfirmPayment)

		// Get payment status
		payments.GET("/:id/status", paymentHandler.GetPaymentStatus)

		// Refund payment
		payments.POST("/:id/refund", paymentHandler.RefundPayment)

		// Get payment history for authenticated user
		payments.GET("/history", paymentHandler.GetPaymentHistory)
	}

	// Webhook route (no authentication required)
	router.POST("/webhooks/stripe", paymentHandler.WebhookHandler)
}

// PaymentAdminRoutes defines admin-only payment routes
func PaymentAdminRoutes(router *gin.RouterGroup, paymentHandler *handlers.PaymentHandler, adminAuthMiddleware gin.HandlerFunc) {
	// Admin payment routes
	payments := router.Group("/payments")
	payments.Use(adminAuthMiddleware)
	{
		// Get payment history for all users (admin only)
		payments.GET("/admin/history", paymentHandler.AdminGetPaymentHistory)
	}
}