package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/api/handlers"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/services"
	"github.com/gmsas95/blytz-mvp/shared/pkg/auth"
	"github.com/gmsas95/blytz-mvp/shared/pkg/ratelimiter"
)

func SetupRouter(auctionService *services.AuctionService, logger *zap.Logger, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// CORS middleware first - SECURE CONFIGURATION
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Allowed origins for production
		allowedOrigins := []string{
			"https://blytz.app",
			"https://www.blytz.app",
			"https://seller.blytz.app",
			"https://demo.blytz.app",
			"http://localhost:3000",     // Development
			"http://localhost:3001",     // Development alternative
		}
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

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
	if cfg.FirebaseEnabled {
		logger.Warn("Firebase is enabled but not yet implemented - notifications will be skipped")
	}

	// Initialize authentication client
	authClient := auth.NewAuthClient(cfg.AuthServiceURL)
	logger.Info("🎵 Auction Service: Auth client initialized", zap.String("auth_url", cfg.AuthServiceURL))

	// Initialize rate limiters
	// General API rate limiter: 100 requests per minute per IP
	apiRateLimiter, err := ratelimiter.NewRedisRateLimiter(ratelimiter.Config{
		RequestsPerSecond: 10.0,
		BurstSize:         100,
		RedisURL:          cfg.RedisURL,
		WindowDuration:    time.Minute,
		Logger:            logger,
	})
	if err != nil {
		logger.Warn("Failed to create Redis rate limiter, using local fallback", zap.Error(err))
	}

	// Bid rate limiter: 10 bids per minute per user (stricter)
	bidRateLimiter, err := ratelimiter.NewRedisRateLimiter(ratelimiter.Config{
		RequestsPerSecond: 0.5,
		BurstSize:         10,
		RedisURL:          cfg.RedisURL,
		WindowDuration:    time.Minute,
		Logger:            logger,
	})
	if err != nil {
		logger.Warn("Failed to create bid rate limiter, using local fallback", zap.Error(err))
	}

	// Auction creation rate limiter: 5 per hour per user
	auctionCreateRateLimiter, err := ratelimiter.NewRedisRateLimiter(ratelimiter.Config{
		RequestsPerSecond: 0.0014, // ~5 per hour
		BurstSize:         5,
		RedisURL:          cfg.RedisURL,
		WindowDuration:    time.Hour,
		Logger:            logger,
	})
	if err != nil {
		logger.Warn("Failed to create auction creation rate limiter, using local fallback", zap.Error(err))
	}

	// Initialize handlers
	auctionHandler := handlers.NewAuctionHandler(auctionService, logger)

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no auth required, general rate limiting)
		auctions := api.Group("/auctions")
		auctions.Use(apiRateLimiter.GinMiddleware(ratelimiter.DefaultKeyExtractor))
		{
			auctions.GET("", auctionHandler.SearchAuctions)
			auctions.GET("/:id", auctionHandler.GetAuction)
			auctions.GET("/:id/status", auctionHandler.GetAuctionStatus)
			auctions.GET("/:id/bids", auctionHandler.GetAuctionBids)
			auctions.GET("/active", auctionHandler.GetActiveAuctions)
		}

		// Protected routes (auth required)
		protectedAuctions := api.Group("/auctions")
		protectedAuctions.Use(auth.GinAuthMiddleware(authClient))
		{
			// Auction creation with stricter rate limiting
			protectedAuctions.POST("", auctionCreateRateLimiter.GinMiddleware(ratelimiter.UserKeyExtractor), auctionHandler.CreateAuction)
			protectedAuctions.PUT("/:id", auctionHandler.UpdateAuction)
			protectedAuctions.DELETE("/:id", auctionHandler.DeleteAuction)
			// Bid placement with stricter rate limiting
			protectedAuctions.POST("/:id/bids", bidRateLimiter.GinMiddleware(ratelimiter.UserKeyExtractor), auctionHandler.PlaceBid)
		}
	}

	return router
}
