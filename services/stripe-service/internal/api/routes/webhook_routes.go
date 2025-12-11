package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/handlers"
)

// WebhookRoutes defines the webhook routes
func WebhookRoutes(router *gin.RouterGroup, webhookHandler *handlers.WebhookHandler) {
	// Apply webhook middleware for signature verification
	router.Use(webhookHandler.WebhookMiddleware())

	// Webhook endpoints
	webhooks := router.Group("/webhooks")
	{
		// Stripe webhook endpoint
		webhooks.POST("/stripe", webhookHandler.HandleStripeWebhook)
	}
}

// WebhookAdminRoutes defines the admin routes for webhooks
func WebhookAdminRoutes(router *gin.RouterGroup, webhookHandler *handlers.WebhookHandler, adminAuthMiddleware gin.HandlerFunc) {
	// Admin webhook routes
	webhooks := router.Group("/webhooks")
	webhooks.Use(adminAuthMiddleware)
	{
		// Get webhook events with pagination
		webhooks.GET("", webhookHandler.GetWebhookEvents)
		
		// Get a specific webhook event
		webhooks.GET("/:id", webhookHandler.GetWebhookEvent)
		
		// Replay a webhook event
		webhooks.POST("/:id/replay", webhookHandler.ReplayWebhookEvent)
	}
}