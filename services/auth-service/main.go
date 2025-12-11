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

	shared_utils "github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
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
	authMiddleware := middleware.AuthMiddleware(authService)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware using shared package
	router.Use(shared_utils.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
			"service":  "auth-service",
			"version":  "v2.0-database",
			"status":   "healthy",
			"database": "connected",
			"time":     time.Now(),
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
		port = "8085"  // Updated to use consistent port 8085 as per docker-compose.yml
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
