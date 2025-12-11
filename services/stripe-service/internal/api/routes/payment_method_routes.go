package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/handlers"
)

// PaymentMethodRoutes defines all payment method related routes
func PaymentMethodRoutes(router *gin.RouterGroup, paymentMethodHandler *handlers.PaymentMethodHandler, authMiddleware gin.HandlerFunc) {
	// Payment method routes with authentication
	paymentMethods := router.Group("/payment-methods")
	paymentMethods.Use(authMiddleware)
	{
		// Get available payment method configurations
		paymentMethods.GET("/configs", paymentMethodHandler.GetPaymentMethodConfigs)

		// Create custom payment method
		paymentMethods.POST("/custom", paymentMethodHandler.CreateCustomPaymentMethod)

		// Create MbWay payment method
		paymentMethods.POST("/mbway", paymentMethodHandler.CreateMbWayPaymentMethod)

		// Create TWINT payment method
		paymentMethods.POST("/twint", paymentMethodHandler.CreateTWINTPaymentMethod)

		// Create crypto payment method
		paymentMethods.POST("/crypto", paymentMethodHandler.CreateCryptoPaymentMethod)

		// Get payment method by ID
		paymentMethods.GET("/:id", paymentMethodHandler.GetPaymentMethod)

		// List payment methods for user
		paymentMethods.GET("/", paymentMethodHandler.ListPaymentMethods)

		// Delete payment method
		paymentMethods.DELETE("/:id", paymentMethodHandler.DeletePaymentMethod)
	}
}

// PaymentMethodAdminRoutes defines admin-only payment method routes
func PaymentMethodAdminRoutes(router *gin.RouterGroup, paymentMethodHandler *handlers.PaymentMethodHandler, adminAuthMiddleware gin.HandlerFunc) {
	// Admin payment method routes
	paymentMethods := router.Group("/payment-methods")
	paymentMethods.Use(adminAuthMiddleware)
	{
		// List all payment methods (admin only)
		paymentMethods.GET("/admin/all", paymentMethodHandler.ListAllPaymentMethods)

		// Update payment method status (admin only)
		paymentMethods.PUT("/admin/:id/status", paymentMethodHandler.UpdatePaymentMethodStatus)
	}
}