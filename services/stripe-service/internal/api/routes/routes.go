package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/handlers"
)

// SetupRouter configures all routes for the stripe service
func SetupRouter(
	connectedAccountHandler *handlers.ConnectedAccountHandler,
	paymentHandler *handlers.PaymentHandler,
	transferHandler *handlers.TransferHandler,
	payoutHandler *handlers.PayoutHandler,
	webhookHandler *handlers.WebhookHandler,
	paymentMethodHandler *handlers.PaymentMethodHandler,
) *gin.Engine {
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "healthy",
				"service": "stripe-service",
				"version": "1.0.0",
			})
		})

		// Connected account routes will be added by the main function
		// Payment routes will be added by the main function
		// Transfer and payout routes will be added by the main function
		// Webhook routes will be added by the main function
		// Payment method routes will be added by the main function
	}

	return router
}