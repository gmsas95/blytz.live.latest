package main

import (
	"log"
	"database/sql"

	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/api"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/repository"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/services"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}
	logger.Info("Starting auction service", zap.String("port", cfg.Port))

	// Initialize database connection
	// Initialize database connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Get underlying SQL DB for repository
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get underlying SQL DB", zap.Error(err))
	}

	auctionRepo := repository.NewPostgresRepo(sqlDB, logger)

	// Initialize services
	auctionService := services.NewAuctionService(auctionRepo, logger, db)

	// Setup router
	router := api.SetupRouter(auctionService, logger, cfg)

	// Start server
	log.Printf("Auction service starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}