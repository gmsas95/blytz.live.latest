package main

import (
	"log"

	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/api"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/repository"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/services"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.LoadConfig()
	logger.Info("Starting auction service", zap.String("port", cfg.Port))

	// Initialize database connection
	db, err := config.InitDB(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize repositories
	auctionRepo := repository.NewAuctionRepository(db, logger)
	bidRepo := repository.NewBidRepository(db, logger)

	// Initialize services
	auctionService := services.NewAuctionService(auctionRepo, bidRepo, logger, cfg)

	// Setup router
	router := api.SetupRouter(auctionService, logger, cfg)

	// Start server
	log.Printf("Auction service starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}