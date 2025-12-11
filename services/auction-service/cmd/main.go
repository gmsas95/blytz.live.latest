package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/api/handlers"
	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/config"
	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/models"
	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/repository"
	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/services"
	"github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
	"github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize logger
	zapLogger, err := utils.NewProductionLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("🎵 Auction Service: Starting server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("🎵 Auction Service: Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := initDatabase(cfg, zapLogger)
	if err != nil {
		zapLogger.Fatal("🎵 Auction Service: Failed to initialize database", zap.Error(err))
	}

	// Run database migrations
	if err := runMigrations(db); err != nil {
		zapLogger.Fatal("🎵 Auction Service: Failed to run migrations", zap.Error(err))
	}

	// Initialize repository
	// Get underlying SQL DB for repository
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Fatal("🎵 Auction Service: Failed to get underlying SQL DB", zap.Error(err))
	}
	repo := repository.NewPostgresRepo(sqlDB, zapLogger)

	// Initialize services
	auctionService := services.NewAuctionService(repo, zapLogger, db)

	// Initialize handlers
	auctionHandler := handlers.NewAuctionHandler(auctionService, zapLogger)

	// Setup router
	router := setupRouter(auctionHandler, cfg, zapLogger)

	// Start server
	startServer(router, cfg, zapLogger)
}

// initDatabase initializes the database connection using GORM
func initDatabase(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	logger.Info("🎵 Auction Service: Connecting to database...", 
		zap.String("host", cfg.PostgresHost),
		zap.String("database", cfg.PostgresDB))

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("🎵 Auction Service: Database connected successfully")
	return db, nil
}

// runMigrations runs database migrations
func runMigrations(db *gorm.DB) error {
	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.Auction{},
		&models.Bid{},
		&models.AuctionImage{},
		&models.AuctionWatcher{},
		&models.AuctionActivity{},
		&models.User{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// setupRouter sets up the Gin router with all routes and middleware
func setupRouter(auctionHandler *handlers.AuctionHandler, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(utils.CORSMiddleware())

	// Request logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("Request processed",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		)
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		health := utils.NewHealthStatus("ok")
		health.AddService("database", "healthy", "Database connection")
		health.AddService("bidding", "operational", "Bidding system")
		health.AddService("auctions", "operational", "Auction management")
		
		utils.SendSuccessResponse(c, http.StatusOK, health)
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auction routes
		auctions := v1.Group("/auctions")
		{
			auctions.GET("", auctionHandler.SearchAuctions)
			auctions.GET("/active", func(c *gin.Context) {
				req := &models.SearchAuctionsRequest{
					Status: models.AuctionStatusActive,
					Page:   1,
					Limit:  20,
				}
				response, err := auctionHandler.GetAuctionService().SearchAuctions(c.Request.Context(), req)
				if err != nil {
					utils.SendErrorResponse(c, err)
					return
				}
				utils.SendSuccessResponse(c, http.StatusOK, response)
			})
			auctions.GET("/:id", auctionHandler.GetAuction)
			auctions.GET("/:id/bids", auctionHandler.GetAuctionBids)
			auctions.GET("/seller/:seller_id", auctionHandler.GetSellerAuctions)
		}

		// Protected auction routes (require authentication)
		protected := v1.Group("/auctions")
		protected.Use(func(c *gin.Context) {
			// TODO: Add proper authentication middleware
			// For now, we'll simulate authentication with a mock user ID
			c.Set("userID", "mock-user-id")
			c.Next()
		})
		{
			protected.POST("", auctionHandler.CreateAuction)
			protected.PUT("/:id", auctionHandler.UpdateAuction)
			protected.DELETE("/:id", auctionHandler.DeleteAuction)
			protected.POST("/:id/bid", func(c *gin.Context) {
				auctionID := c.Param("id")
				bidderID := c.GetString("userID")
				if bidderID == "" {
					utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
					return
				}

				var req models.PlaceBidRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
					return
				}

				response, err := auctionHandler.GetAuctionService().PlaceBid(c.Request.Context(), auctionID, bidderID, &req)
				if err != nil {
					utils.SendErrorResponse(c, err)
					return
				}

				utils.SendSuccessResponseWithMessage(c, http.StatusCreated, "Bid placed successfully", response)
			})
			protected.GET("/my", auctionHandler.GetMyAuctions)
			protected.POST("/:id/end", auctionHandler.EndAuction)
			protected.POST("/:id/start", auctionHandler.StartAuction)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(func(c *gin.Context) {
			// TODO: Add admin authentication middleware
			c.Set("userID", "admin-user-id")
			c.Next()
		})
		{
			admin.GET("/auctions", auctionHandler.SearchAuctions)
			admin.GET("/stats", func(c *gin.Context) {
				stats := gin.H{
					"total_auctions": 100,
					"active_auctions": 25,
					"total_bids": 1500,
					"total_revenue": 50000.0,
				}
				utils.SendSuccessResponse(c, http.StatusOK, stats)
			})
		}
	}

	return router
}

// startServer starts the HTTP server with graceful shutdown
func startServer(router *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("🎵 Auction Service: Starting server", 
			zap.String("port", cfg.Port),
			zap.String("environment", cfg.Environment))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("🎵 Auction Service: Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🎵 Auction Service: Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("🎵 Auction Service: Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("🎵 Auction Service: Server shutdown complete")
	}
}
