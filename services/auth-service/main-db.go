package main

import (
	"context"
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

	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/api/handlers"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/middleware"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/services"
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

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://blytz:blytz_password_2025@localhost:5432/blytz_mvp?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Database connected successfully")

	// Initialize services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
	}

	authService := services.NewAuthService(db, jwtSecret, logger)

	// Initialize handlers and middleware
	authHandler := handlers.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, shared_utils.Response{
			Success: true,
			Message: "Auth service is healthy and working",
			Data: map[string]interface{}{
				"service":  "auth-service",
				"version":  "v2.0-database",
				"status":   "healthy",
				"database": "connected",
				"time":     time.Now(),
			},
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/validate", authHandler.Validate)
		}

		// Protected user routes
		users := v1.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			users.GET("/profile", authHandler.GetProfile)
			users.PUT("/profile", authHandler.UpdateProfile)
			users.POST("/logout", authHandler.Logout)
			users.POST("/refresh", authHandler.RefreshToken)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(authMiddleware.RequireAuth())
		admin.Use(authMiddleware.RequireAdmin())
		{
			admin.GET("/users", func(c *gin.Context) {
				// TODO: Implement user listing for admin
				c.JSON(http.StatusOK, shared_utils.Response{
					Success: true,
					Message: "Admin user listing endpoint",
				})
			})
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	logger.Info("Auth Service starting",
		zap.String("port", port),
		zap.String("database", "PostgreSQL"),
		zap.String("version", "v2.0-database"),
	)

	fmt.Printf("🚀 Auth Service starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔐 Login endpoint: http://localhost:%s/api/v1/auth/login\n", port)
	fmt.Printf("📝 Register endpoint: http://localhost:%s/api/v1/auth/register\n", port)
	fmt.Printf("🗄️  Database: PostgreSQL\n")
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
