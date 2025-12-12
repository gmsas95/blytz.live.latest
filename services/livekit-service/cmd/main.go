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

	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	"github.com/gmsas95/blytz-mvp/services/livekit-service/internal/api/handlers"
	"github.com/gmsas95/blytz-mvp/services/livekit-service/internal/services"
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

	// Get configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8089" // Default port as per docker-compose.yml
	}

	livekitURL := os.Getenv("LIVEKIT_URL")
	if livekitURL == "" {
		livekitURL = "ws://localhost:7880" // Default LiveKit URL
	}

	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	if livekitAPIKey == "" {
		livekitAPIKey = "devkey" // Default for development
	}

	livekitAPISecret := os.Getenv("LIVEKIT_API_SECRET")
	if livekitAPISecret == "" {
		livekitAPISecret = "devsecret" // Default for development
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/blytz_db?sslmode=disable"
	}

	// Initialize database connection
	db, err := shared_utils.InitDatabaseFromURL(databaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize LiveKit service
	livekitService, err := services.NewLiveKitService(db, logger, livekitURL, livekitAPIKey, livekitAPISecret)
	if err != nil {
		logger.Fatal("Failed to initialize LiveKit service", zap.Error(err))
	}
	defer livekitService.Close()

	// Initialize handlers
	livekitHandler := handlers.NewLiveKitHandler(livekitService, logger)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware using shared package
	router.Use(shared_utils.CORSMiddleware())

	// Add request ID middleware
	router.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
		c.Set("request_id", requestID)
		c.Next()
	})

	// Health check endpoint
	router.GET("/api/v1/health", livekitHandler.HealthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// LiveKit routes
		livekit := v1.Group("/livekit")
		{
			// Room management
			livekit.GET("/rooms", livekitHandler.ListRooms)
			livekit.POST("/rooms", livekitHandler.CreateRoom)
			livekit.GET("/rooms/:id", livekitHandler.GetRoom)
			livekit.POST("/rooms/:id/start", livekitHandler.StartRoom)
			livekit.POST("/rooms/:id/stop", livekitHandler.StopRoom)
			livekit.GET("/rooms/:id/participants", livekitHandler.GetRoomParticipants)
			livekit.GET("/rooms/:id/stats", livekitHandler.GetRoomStats)
			livekit.GET("/rooms/active", livekitHandler.GetActiveRooms)

			// Room participation
			livekit.POST("/rooms/:id/join", livekitHandler.JoinRoom)
			livekit.POST("/rooms/:id/leave", livekitHandler.LeaveRoom)

			// Token management
			livekit.POST("/tokens", livekitHandler.GenerateToken)

			// Webhook
			livekit.POST("/webhook", livekitHandler.ProcessWebhook)
		}
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("LiveKit Service starting",
			zap.String("port", port),
			zap.String("livekit_url", livekitURL),
			zap.String("database", "PostgreSQL"),
			zap.String("version", "v1.0.0"),
		)

		fmt.Printf("🚀 LiveKit Service starting on port %s\n", port)
		fmt.Printf("📊 Health check: http://localhost:%s/api/v1/health\n", port)
		fmt.Printf("🎥 LiveKit URL: %s\n", livekitURL)
		fmt.Printf("🗄️  Database: PostgreSQL\n")
		fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
		fmt.Printf("\n📋 Available endpoints:\n")
		fmt.Printf("  GET    /api/v1/livekit/rooms                    - List all rooms\n")
		fmt.Printf("  POST   /api/v1/livekit/rooms                    - Create new room\n")
		fmt.Printf("  GET    /api/v1/livekit/rooms/:id                - Get room by ID\n")
		fmt.Printf("  POST   /api/v1/livekit/rooms/:id/start           - Start room\n")
		fmt.Printf("  POST   /api/v1/livekit/rooms/:id/stop            - Stop room\n")
		fmt.Printf("  GET    /api/v1/livekit/rooms/:id/participants   - Get room participants\n")
		fmt.Printf("  GET    /api/v1/livekit/rooms/:id/stats          - Get room stats\n")
		fmt.Printf("  GET    /api/v1/livekit/rooms/active            - Get active rooms\n")
		fmt.Printf("  POST   /api/v1/livekit/rooms/:id/join           - Join room\n")
		fmt.Printf("  POST   /api/v1/livekit/rooms/:id/leave          - Leave room\n")
		fmt.Printf("  POST   /api/v1/livekit/tokens                  - Generate access token\n")
		fmt.Printf("  POST   /api/v1/livekit/webhook                 - Process webhook\n")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down LiveKit service...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("Server shutdown completed")
	}
}
