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
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz.live.latest/services/chat-service/internal/api/handlers"
	"github.com/gmsas95/blytz.live.latest/services/chat-service/internal/models"
	"github.com/gmsas95/blytz.live.latest/services/chat-service/internal/services"
	sharedUtils "github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	connections map[string]*websocket.Conn // userID -> connection
	rooms       map[string][]string        // roomID -> []userID
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *WebSocketMessage
	logger      *zap.Logger
}

// Client represents a WebSocket client
type Client struct {
	ID     string
	RoomID string
	Conn   *websocket.Conn
	Send   chan []byte
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"room_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	MessageID string      `json:"message_id,omitempty"`
}

func main() {
	// Load environment variables
	if err := loadEnvironment(); err != nil {
		log.Fatalf("Failed to load environment: %v", err)
	}

	// Initialize logger
	logger, err := sharedUtils.NewDevelopmentLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("💬 Chat Service: Starting server...")

	// Initialize database
	db, err := initDatabase(logger)
	if err != nil {
		logger.Fatal("💬 Chat Service: Failed to initialize database", zap.Error(err))
	}

	// Auto-migrate database models
	if err := migrateDatabase(db); err != nil {
		logger.Fatal("💬 Chat Service: Failed to migrate database", zap.Error(err))
	}

	// Initialize services
	chatService := services.NewChatService(db, logger)
	chatHandler := handlers.NewChatHandler(chatService, logger)

	// Initialize WebSocket manager
	wsManager := NewWebSocketManager(logger)
	go wsManager.run()

	// Setup Gin router
	router := setupRouter(db, chatHandler, wsManager, logger)

	// Start server
	port := sharedUtils.GetEnv("PORT", "8088")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		logger.Info("💬 Chat Service: Server starting on port " + port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("💬 Chat Service: Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("💬 Chat Service: Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("💬 Chat Service: Server forced to shutdown", zap.Error(err))
	}

	logger.Info("💬 Chat Service: Server exited")
}

// loadEnvironment loads environment variables
func loadEnvironment() error {
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	return nil
}

// loadEnvFile loads .env file
func loadEnvFile() error {
	// Check if .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return nil // .env file is optional
	}
	
	// Load .env file
	return nil // Using godotenv would be implemented here
}

// initDatabase initializes database connection
func initDatabase(logger *zap.Logger) (*gorm.DB, error) {
	databaseURL := sharedUtils.GetEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/chat_db")
	
	logger.Info("💬 Chat Service: Connecting to database", zap.String("url", databaseURL))

	// Initialize GORM with PostgreSQL
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("💬 Chat Service: Database connection established successfully")
	return db, nil
}

// migrateDatabase migrates database schema
func migrateDatabase(db *gorm.DB) error {
	// Auto-migrate all models
	return db.AutoMigrate(
		&models.ChatRoom{},
		&models.RoomMember{},
		&models.Message{},
		&models.MessageReaction{},
		&models.TypingIndicator{},
		&models.UserPresence{},
		&models.ChatSettings{},
		&models.User{},
	)
}

// setupRouter sets up Gin router with all routes
func setupRouter(db *gorm.DB, chatHandler *handlers.ChatHandler, wsManager *WebSocketManager, logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(sharedUtils.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", chatHandler.Health)

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c, wsManager, logger)
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Chat routes
		chat := v1.Group("/chat")
		{
			// Room management
			rooms := chat.Group("/rooms")
			{
				rooms.GET("", chatHandler.GetUserRooms)           // List all chat rooms for user
				rooms.POST("", chatHandler.CreateRoom)            // Create chat room
				rooms.GET("/:id", chatHandler.GetRoom)            // Get chat room by ID
				rooms.PUT("/:id", chatHandler.UpdateRoom)         // Update chat room
				rooms.DELETE("/:id", chatHandler.DeleteRoom)      // Delete chat room
				rooms.POST("/:id/members", chatHandler.AddRoomMembers) // Add members to room
				rooms.DELETE("/:id/members/:member_id", chatHandler.RemoveRoomMember) // Remove member from room
			}

			// Message management
			messages := chat.Group("/messages")
			{
				messages.GET("", chatHandler.GetRoomMessages)      // Get messages for room
				messages.POST("", chatHandler.SendMessage)         // Send message to room
				messages.PUT("/:message_id", chatHandler.UpdateMessage) // Update message
				messages.DELETE("/:message_id", chatHandler.DeleteMessage) // Delete message
				messages.GET("/search", chatHandler.SearchRoomMessages) // Search messages
			}

			// User chat history
			chat.GET("/history", chatHandler.GetUserRooms) // Get chat history for user (using GetUserRooms)

			// Typing indicators
			chat.POST("/rooms/:room_id/typing", chatHandler.SetTypingStatus) // Set typing status
			chat.GET("/rooms/:room_id/typing", chatHandler.GetTypingIndicators) // Get typing indicators
		}
	}

	return router
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(logger *zap.Logger) *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[string]*websocket.Conn),
		rooms:       make(map[string][]string),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *WebSocketMessage),
		logger:      logger,
	}
}

// run runs the WebSocket manager
func (m *WebSocketManager) run() {
	for {
		select {
		case client := <-m.register:
			m.registerClient(client)
		case client := <-m.unregister:
			m.unregisterClient(client)
		case message := <-m.broadcast:
			m.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (m *WebSocketManager) registerClient(client *Client) {
	m.connections[client.ID] = client.Conn
	m.rooms[client.RoomID] = append(m.rooms[client.RoomID], client.ID)
	m.logger.Info("💬 Chat Service: Client connected", 
		zap.String("user_id", client.ID),
		zap.String("room_id", client.RoomID))
}

// unregisterClient unregisters a client
func (m *WebSocketManager) unregisterClient(client *Client) {
	if _, ok := m.connections[client.ID]; ok {
		delete(m.connections, client.ID)
		
		// Remove from room
		if roomUsers, exists := m.rooms[client.RoomID]; exists {
			for i, userID := range roomUsers {
				if userID == client.ID {
					m.rooms[client.RoomID] = append(roomUsers[:i], roomUsers[i+1:]...)
					break
				}
			}
		}
		
		close(client.Send)
		m.logger.Info("💬 Chat Service: Client disconnected", 
			zap.String("user_id", client.ID),
			zap.String("room_id", client.RoomID))
	}
}

// broadcastMessage broadcasts a message to all clients in a room
func (m *WebSocketManager) broadcastMessage(message *WebSocketMessage) {
	if roomUsers, exists := m.rooms[message.RoomID]; exists {
		for _, userID := range roomUsers {
			if conn, ok := m.connections[userID]; ok {
				if err := conn.WriteJSON(message); err != nil {
					m.logger.Error("💬 Chat Service: Failed to send message", 
						zap.String("user_id", userID),
						zap.Error(err))
					conn.Close()
					delete(m.connections, userID)
				}
			}
		}
	}
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(c *gin.Context, wsManager *WebSocketManager, logger *zap.Logger) {
	// Get user ID from query parameter (in production, this would come from JWT)
	userID := c.Query("user_id")
	roomID := c.Query("room_id")
	
	if userID == "" || roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and room_id are required"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("💬 Chat Service: Failed to upgrade connection", zap.Error(err))
		return
	}

	// Create client
	client := &Client{
		ID:     userID,
		RoomID: roomID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	// Register client
	wsManager.register <- client

	// Start goroutines for reading and writing
	go client.readPump(wsManager, logger)
	go client.writePump(logger)
}

// readPump reads messages from WebSocket connection
func (c *Client) readPump(wsManager *WebSocketManager, logger *zap.Logger) {
	defer func() {
		wsManager.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WebSocketMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("💬 Chat Service: WebSocket error", zap.Error(err))
			}
			break
		}

		// Set message metadata
		message.UserID = c.ID
		message.RoomID = c.RoomID
		message.Timestamp = time.Now()

		// Broadcast message
		wsManager.broadcast <- &message
	}
}

// writePump writes messages to WebSocket connection
func (c *Client) writePump(logger *zap.Logger) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error("💬 Chat Service: Failed to write message", zap.Error(err))
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
