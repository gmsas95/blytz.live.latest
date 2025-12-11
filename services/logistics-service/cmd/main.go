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
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz.live.latest/services/logistics-service/internal/api/handlers"
	"github.com/gmsas95/blytz.live.latest/services/logistics-service/internal/config"
	"github.com/gmsas95/blytz.live.latest/services/logistics-service/internal/models"
	"github.com/gmsas95/blytz.live.latest/services/logistics-service/internal/services"
	"github.com/gmsas95/blytz.live.latest/shared/pkg/auth"
	"github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	logger, err := utils.NewProductionLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database connection
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Auto-migrate database schema
	if err := db.AutoMigrate(
		&models.Shipment{},
		&models.TrackingEvent{},
	); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	// Initialize services
	logisticsService := services.NewLogisticsService(db, logger, cfg)
	ninjaVanService := services.NewNinjaVanService(db, logger, cfg)

	// Initialize handlers
	logisticsHandler := handlers.NewLogisticsHandler(logisticsService, logger)
	ninjaVanHandler := handlers.NewNinjaVanHandler(ninjaVanService, logger)

	// Setup router
	router := setupRouter(cfg, logger, logisticsHandler, ninjaVanHandler)

	// Start server
	startServer(router, cfg, logger)
}

func initDatabase(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully")
	return db, nil
}

func setupRouter(cfg *config.Config, logger *zap.Logger, logisticsHandler *handlers.LogisticsHandler, ninjaVanHandler *handlers.NinjaVanHandler) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(utils.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		health := utils.NewHealthStatus("ok")
		health.AddService("database", "connected", "Database connection established")
		health.AddService("ninjavan_api", "configured", "NinjaVan API configured")
		
		utils.SendSuccessResponse(c, http.StatusOK, health)
	})

	// Public webhook endpoint (no auth required)
	router.POST("/api/v1/logistics/webhook/ninjavan", ninjaVanHandler.ProcessWebhook)

	// Protected routes
	protected := router.Group("/api/v1/logistics")
	protected.Use(auth.MockAuthMiddleware()) // Using mock auth for development
	{
		// Shipment endpoints
		protected.GET("/shipments", logisticsHandler.ListShipments)
		protected.GET("/shipments/:id", logisticsHandler.GetShipment)
		protected.POST("/shipments", logisticsHandler.CreateShipment)
		protected.PUT("/shipments/:id", logisticsHandler.UpdateShipmentStatus)
		protected.POST("/track", logisticsHandler.TrackShipment)

		// Carrier endpoints
		protected.GET("/carriers", logisticsHandler.ListCarriers)

		// Rate calculation endpoints
		protected.POST("/rates", logisticsHandler.CalculateShippingRates)

		// NinjaVan integration endpoints
		protected.POST("/ninjavan/shipments", ninjaVanHandler.CreateNinjaVanShipment)
		protected.POST("/ninjavan/shipments/:id/cancel", ninjaVanHandler.CancelNinjaVanShipment)
		protected.POST("/ninjavan/tariff", ninjaVanHandler.GetShippingCost)
		protected.GET("/ninjavan/pudo-points", ninjaVanHandler.GetPUDOPoints)
	}

	return router
}

func startServer(router *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	port := utils.GetEnv("PORT", "8087")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting logistics service", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
