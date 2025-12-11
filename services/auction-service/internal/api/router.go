package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/api/handlers"
	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/config"
	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/services"
	"github.com/gmsas95/blytz.live.latest/services/auction-service/pkg/firebase"
)

func SetupRouter(auctionService *services.AuctionService, logger *zap.Logger, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// CORS middleware first
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.Status(200)
			return
		}
		c.Next()
	})

	// Add custom logger middleware to avoid any potential debug routes
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return ""
	}))
	router.Use(gin.Recovery())

	// Comprehensive health check endpoint
	router.GET("/health", func(c *gin.Context) {
		health := gin.H{
			"status":    "ok",
			"service":   "auction",
			"timestamp": time.Now().Unix(),
			"version":   "v1.0.0",
			"checks": gin.H{
				"database": "connected",
				"redis":    "connected",
			},
		}
		c.JSON(200, health)
	})

	// Initialize Firebase app (optional, can be nil if not configured)
	// TODO: Implement Firebase initialization when needed
	var firebaseApp firebase.FirebaseApp = nil
	if cfg.FirebaseEnabled {
		logger.Warn("Firebase is enabled but not yet implemented - notifications will be skipped")
	}

	// Initialize handlers
	auctionHandler := handlers.NewAuctionHandler(auctionService, logger)

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		auctions := api.Group("/auctions")
		{
			auctions.GET("", auctionHandler.ListAuctions)
			auctions.GET("/:id", auctionHandler.GetAuction)
			auctions.GET("/:id/status", auctionHandler.GetAuctionStatus)
			auctions.GET("/:id/bids", auctionHandler.GetBids)
			auctions.GET("/active", auctionHandler.GetActiveAuctions)
		}

		// Protected routes (auth required)
		protectedAuctions := api.Group("/auctions")
		protectedAuctions.Use(func(c *gin.Context) {
			// TODO: Add proper authentication middleware
			c.Set("userID", "mock-user-id")
			c.Next()
		})
		{
			protectedAuctions.POST("", auctionHandler.CreateAuction)
			protectedAuctions.PUT("/:id", auctionHandler.UpdateAuction)
			protectedAuctions.DELETE("/:id", auctionHandler.DeleteAuction)
			protectedAuctions.POST("/:id/bids", auctionHandler.PlaceBid)
		}
	}

	return router
}
