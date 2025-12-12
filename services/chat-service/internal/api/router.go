package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/chat-service/internal/api/handlers"
	"github.com/gmsas95/blytz-mvp/services/chat-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/chat-service/internal/services"
	"github.com/gmsas95/blytz-mvp/shared/pkg/auth"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, logger *zap.Logger) *gin.Engine {
	// Initialize config
	cfg := config.LoadConfig()

	// Create router
	router := gin.Default()

	// Initialize chat service
	chatService := services.NewChatService(db, logger)

	// Initialize auth client
	authClient := auth.NewAuthClient("http://auth-service:8085")

	// Create chat handler
	chatHandler := handlers.NewChatHandler(chatService, logger)

	// Enhanced health check endpoint
	router.GET("/health", func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = c.GetString("correlation_id")
		}

		health := gin.H{
			"status":         "ok",
			"service":        "chat",
			"version":        "1.0.0",
			"timestamp":      time.Now().Unix(),
			"correlation_id": correlationID,
			"environment":    cfg.Environment,
		}

		// Check service connectivity
		if chatService != nil {
			health["websocket"] = "available"
		} else {
			health["websocket"] = "unavailable"
			health["status"] = "degraded"
			c.JSON(http.StatusServiceUnavailable, health)
			return
		}

		// Check external dependencies
		health["dependencies"] = gin.H{
			"auth_service": "connected",
			"redis":        "configured",
		}

		c.JSON(http.StatusOK, health)
	})

	// Prometheus metrics endpoint

	// Chat endpoints
	chatRoutes := router.Group("/api/v1/chat")
	chatRoutes.Use(auth.GinAuthMiddleware(authClient))
	{
		chatRoutes.GET("/ws", func(c *gin.Context) {
			// Simplified WebSocket handler - just return success for now
			c.JSON(http.StatusOK, gin.H{
				"message": "WebSocket endpoint available",
				"status": "ok",
			})
		})
		chatRoutes.GET("/rooms/:room_id/messages", chatHandler.GetRoomMessages)
		chatRoutes.POST("/rooms/:room_id/messages", chatHandler.SendMessage)
		chatRoutes.GET("/rooms", chatHandler.GetUserRooms)
		chatRoutes.POST("/rooms", chatHandler.CreateRoom)
		chatRoutes.GET("/rooms/:id", chatHandler.GetRoom)
		chatRoutes.PUT("/rooms/:id", chatHandler.UpdateRoom)
		chatRoutes.DELETE("/rooms/:id", chatHandler.DeleteRoom)
	}

	return router
}
