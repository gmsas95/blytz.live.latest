package main

import (
	"log"

	"github.com/gmsas95/blytz.live.latest/services/logistics-service/internal/api"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting logistics service")

	// Setup router
	router := api.SetupRouter(logger)

	// Start server
	port := "8091"
	log.Printf("Logistics service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}