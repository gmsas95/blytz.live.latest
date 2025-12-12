package main

import (
	"log"

	"github.com/gmsas95/blytz-mvp/services/chat-service/internal/api"
	"github.com/gmsas95/blytz-mvp/services/chat-service/internal/config"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.LoadConfig()
	logger.Info("Starting chat service", zap.String("port", cfg.Port))

	// Initialize database connection
	db, err := config.InitDB(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Setup router
	router := api.SetupRouter(db, logger)

	// Start server
	log.Printf("Chat service starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}