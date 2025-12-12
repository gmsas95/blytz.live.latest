package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/api/handlers"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/config"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/middleware"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/services"
	shared_metrics "github.com/gmsas95/blytz.live.latest/shared/pkg/metrics"
	"github.com/gmsas95/blytz.live.latest/shared/pkg/ratelimiter"
	shared_utils "github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
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

	// Load configuration (validates secrets in production)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Database connection
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Database connected successfully",
		zap.String("host", cfg.PostgresHost),
		zap.String("database", cfg.PostgresDB))

	// Initialize services
	authService := services.NewAuthService(db, cfg.JWTSecret, logger)

	// Initialize handlers and middleware
	authHandler := handlers.NewAuthHandler(authService)
	authMiddleware := middleware.AuthMiddleware(authService)

	// Initialize account lockout (5 failed attempts = 15 min lockout)
	accountLockout := middleware.DefaultAccountLockout(logger)
	authHandler.SetAccountLockout(accountLockout)

	// Initialize rate limiters
	// General API rate limiter: 100 requests per minute
	apiRateLimiter, err := ratelimiter.NewRedisRateLimiter(ratelimiter.Config{
		RequestsPerSecond: 100.0 / 60.0, // ~1.67 rps
		BurstSize:         100,
		RedisURL:          cfg.RedisURL,
		WindowDuration:    time.Minute,
		Logger:            logger,
	})
	if err != nil {
		logger.Warn("Failed to create API rate limiter, using local fallback", zap.Error(err))
	}

	// Auth endpoint rate limiter: 10 requests per minute (stricter)
	authRateLimiter, err := ratelimiter.NewRedisRateLimiter(ratelimiter.Config{
		RequestsPerSecond: 10.0 / 60.0, // ~0.16 rps
		BurstSize:         10,
		RedisURL:          cfg.RedisURL,
		WindowDuration:    time.Minute,
		Logger:            logger,
	})
	if err != nil {
		logger.Warn("Failed to create Auth rate limiter, using local fallback", zap.Error(err))
	}

	// Setup Gin router
	router := gin.Default()

	// CORS middleware using shared package
	router.Use(shared_utils.CORSMiddleware())

	// Metrics middleware
	router.Use(shared_metrics.MetricsMiddleware("auth-service"))

	// General API rate limiting (100 req/min)
	router.Use(apiRateLimiter.GinMiddleware(ratelimiter.DefaultKeyExtractor))

	// Metrics endpoint
	router.GET("/metrics", shared_metrics.PrometheusHandler())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
			"service":  "auth-service",
			"version":  "v2.1-security",
			"status":   "healthy",
			"database": "connected",
			"time":     time.Now(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public) - stricter rate limiting
		auth := v1.Group("/auth")
		auth.Use(authRateLimiter.GinMiddleware(ratelimiter.DefaultKeyExtractor))
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/validate", authHandler.Verify)
		}

		// Protected user routes
		users := v1.Group("/users")
		users.Use(authMiddleware)
		{
			users.GET("/profile", authHandler.GetProfile)
			users.PUT("/profile", authHandler.UpdateProfile)
			users.POST("/logout", authHandler.Logout)
			users.POST("/refresh", authHandler.RefreshToken)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(authMiddleware)
		{
			admin.GET("/users", func(c *gin.Context) {
				// TODO: Implement user listing for admin
				shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
					"message": "Admin user listing endpoint",
				})
			})
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085" // Updated to use consistent port 8085 as per docker-compose.yml
	}

	// Set build info and start time
	shared_metrics.SetBuildInfo("v2.0-database", "unknown", time.Now().Format(time.RFC3339))
	shared_metrics.SetStartTime(float64(time.Now().Unix()))

	logger.Info("Auth Service starting",
		zap.String("port", port),
		zap.String("database", "PostgreSQL"),
		zap.String("version", "v2.0-database"),
	)

	fmt.Printf("🚀 Auth Service starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("📈 Metrics: http://localhost:%s/metrics\n", port)
	fmt.Printf("🔐 Login endpoint: http://localhost:%s/api/v1/auth/login\n", port)
	fmt.Printf("📝 Register endpoint: http://localhost:%s/api/v1/auth/register\n", port)
	fmt.Printf("🗄️  Database: PostgreSQL\n")
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
