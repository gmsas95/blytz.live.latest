package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	shared_metrics "github.com/gmsas95/blytz-mvp/shared/pkg/metrics"
	"github.com/gmsas95/blytz-mvp/services/search-service/internal/api/handlers"
	"github.com/gmsas95/blytz-mvp/services/search-service/internal/api/routes"
	"github.com/gmsas95/blytz-mvp/services/search-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/search-service/internal/repository"
	"github.com/gmsas95/blytz-mvp/services/search-service/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize logger
	logger, err := shared_utils.NewDevelopmentLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Database configuration
	dbConfig := &config.DatabaseConfig{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/blytz_db?sslmode=disable"),
		Environment:  getEnv("ENVIRONMENT", "development"),
	}

	// Initialize database
	db, err := config.InitDB(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run migrations
	if err := config.MigrateDatabase(db); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	logger.Info("Database connected and migrated successfully")

	// Initialize Redis
	redisClient := initRedis(logger)

	// Initialize repositories
	searchRepo := repository.NewSearchRepository(db)

	// Initialize services
	searchService := services.NewSearchService(searchRepo, redisClient, logger)

	// Initialize handlers
	searchHandler := handlers.NewSearchHandler(searchService)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware using shared package
	router.Use(shared_utils.RequestIDMiddleware())

	// Metrics middleware
	router.Use(shared_metrics.MetricsMiddleware("search-service"))

	// Setup routes
	routes.SetupRoutes(router, searchHandler)

	// Set build info and start time
	shared_metrics.SetBuildInfo("v1.0.0", "unknown", time.Now().Format(time.RFC3339))
	shared_metrics.SetStartTime(float64(time.Now().Unix()))

	// Start server
	port := getEnv("PORT", "8095")
	
	logger.Info("Search Service starting",
		zap.String("port", port),
		zap.String("database", "PostgreSQL"),
		zap.String("cache", "Redis"),
		zap.String("version", "v1.0.0"),
	)

	fmt.Printf("🚀 Search Service starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("📈 Metrics: http://localhost:%s/metrics\n", port)
	fmt.Printf("🔍 Search endpoint: http://localhost:%s/api/v1/search\n", port)
	fmt.Printf("📝 Index endpoint: http://localhost:%s/api/v1/index\n", port)
	fmt.Printf("🗄️  Database: PostgreSQL with full-text search\n")
	fmt.Printf("⚡ Cache: Redis\n")
	fmt.Printf("🔧 Max connections: 200\n")
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// initRedis initializes Redis connection
func initRedis(logger *zap.Logger) *redis.Client {
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Warn("Failed to parse Redis URL, using default config", zap.Error(err))
		opt = &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
	}

	// Create Redis client
	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		logger.Warn("Failed to connect to Redis, caching will be disabled", zap.Error(err))
		return nil // Return nil to disable caching
	}

	logger.Info("Redis connected successfully", zap.String("addr", opt.Addr))
	return client
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}