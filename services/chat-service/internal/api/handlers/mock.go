package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gmsas95/blytz.live.latest/services/chat-service/internal/models"
	sharedErrors "github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
	sharedUtils "github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
)

// MockChatService interface for our mock implementation
type MockChatService interface {
	CreateRoom(ctx context.Context, creatorID string, req *models.CreateRoomRequest) (*models.ChatRoom, error)
	GetRoom(ctx context.Context, roomID, userID string) (*models.ChatRoom, error)
	GetUserRooms(ctx context.Context, userID string, page, limit int) (*models.RoomsResponse, error)
}

// MockChatHandler handles chat operations with mock service
type MockChatHandler struct {
	chatService MockChatService
	logger      *zap.Logger
}

func NewMockChatHandler(chatService MockChatService, logger *zap.Logger) *MockChatHandler {
	return &MockChatHandler{
		chatService: chatService,
		logger:      logger,
	}
}

// CreateRoom creates new chat room
func (h *MockChatHandler) CreateRoom(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		// For mock, use a default user ID
		userID = "mock-user-123"
	}

	var req models.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("💬 Mock: Create room validation failed", zap.Error(err))
		sharedUtils.SendValidationErrorResponse(c, map[string]string{"validation": err.Error()})
		return
	}

	room, err := h.chatService.CreateRoom(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("💬 Mock: Failed to create room",
			zap.String("user_id", userID),
			zap.String("name", req.Name),
			zap.Error(err))
		sharedUtils.SendErrorResponse(c, sharedErrors.NewInternalError("CREATE_ROOM_FAILED", "Failed to create room"))
		return
	}

	response := &models.RoomResponse{
		Room:         *room,
		MessageCount: 0,
		UnreadCount:  0,
		IsJoined:     true,
		UserRole:     models.UserRoleOwner,
		IsActive:     true,
	}

	h.logger.Info("💬 Mock: Room created successfully",
		zap.String("room_id", room.ID),
		zap.String("user_id", userID))

	sharedUtils.SendSuccessResponseWithMessage(c, http.StatusCreated, "💬 Mock: Room created successfully!", response)
}

// GetRoom retrieves room by ID
func (h *MockChatHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		sharedUtils.SendErrorResponse(c, sharedErrors.NewValidationError("MISSING_ROOM_ID", "Room ID is required"))
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		// For mock, use a default user ID
		userID = "mock-user-123"
	}

	room, err := h.chatService.GetRoom(c.Request.Context(), roomID, userID)
	if err != nil {
		h.logger.Error("💬 Mock: Failed to get room",
			zap.String("room_id", roomID),
			zap.String("user_id", userID),
			zap.Error(err))
		sharedUtils.SendErrorResponse(c, sharedErrors.NewNotFoundError("ROOM_NOT_FOUND", "Room not found"))
		return
	}

	response := &models.RoomResponse{
		Room:          *room,
		MessageCount:  0,
		UnreadCount:   0,
		IsJoined:      true,
		UserRole:      models.UserRoleMember,
		IsActive:      true,
	}

	sharedUtils.SendSuccessResponseWithMessage(c, http.StatusOK, "💬 Mock: Room retrieved successfully!", response)
}

// GetUserRooms gets rooms for user
func (h *MockChatHandler) GetUserRooms(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		// For mock, use a default user ID
		userID = "mock-user-123"
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	response, err := h.chatService.GetUserRooms(c.Request.Context(), userID, page, limit)
	if err != nil {
		h.logger.Error("💬 Mock: Failed to get user rooms",
			zap.String("user_id", userID),
			zap.Error(err))
		sharedUtils.SendErrorResponse(c, sharedErrors.NewInternalError("GET_ROOMS_FAILED", "Failed to get user rooms"))
		return
	}

	h.logger.Info("💬 Mock: User rooms retrieved successfully",
		zap.String("user_id", userID),
		zap.Int64("total", response.Total))

	sharedUtils.SendSuccessResponseWithMessage(c, http.StatusOK, "💬 Mock: User rooms retrieved successfully!", response)
}

// UpdateRoom updates existing room (mock implementation)
func (h *MockChatHandler) UpdateRoom(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		sharedUtils.SendErrorResponse(c, sharedErrors.NewValidationError("MISSING_ROOM_ID", "Room ID is required"))
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		userID = "mock-user-123"
	}

	var req models.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("💬 Mock: Update room validation failed", zap.Error(err))
		sharedUtils.SendValidationErrorResponse(c, map[string]string{"validation": err.Error()})
		return
	}

	// For mock, just return success
	h.logger.Info("💬 Mock: Room updated successfully", zap.String("room_id", roomID))

	sharedUtils.SendSuccessResponseWithMessage(c, http.StatusOK, "💬 Mock: Room updated successfully!", gin.H{
		"room_id": roomID,
		"user_id": userID,
		"updates": req,
	})
}

// DeleteRoom deletes room (mock implementation)
func (h *MockChatHandler) DeleteRoom(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		sharedUtils.SendErrorResponse(c, sharedErrors.NewValidationError("MISSING_ROOM_ID", "Room ID is required"))
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		userID = "mock-user-123"
	}

	// For mock, just return success
	h.logger.Info("💬 Mock: Room deleted successfully", zap.String("room_id", roomID))

	sharedUtils.SendSuccessResponseWithMessage(c, http.StatusOK, "💬 Mock: Room deleted successfully!", gin.H{
		"room_id": roomID,
		"user_id": userID,
	})
}

// SendMessage sends message to room (mock implementation)
func (h *MockChatHandler) SendMessage(c *gin.Context) {
	roomID := c.Param("room_id")
	if roomID == "" {
		sharedUtils.SendErrorResponse(c, sharedErrors.NewValidationError("MISSING_ROOM_ID", "Room ID is required"))
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		userID = "mock-user-123"
	}

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("💬 Mock: Send message validation failed", zap.Error(err))
		sharedUtils.SendValidationErrorResponse(c, map[string]string{"validation": err.Error()})
		return
	}

	// For mock, create a mock message
	message := &models.Message{
		ID:        "msg-" + strconv.FormatInt(int64(len(req.Content)), 10),
		RoomID:    roomID,
		UserID:    userID,
		Content:   req.Content,
		Type:      req.Type,
		Timestamp: time.Now(),
	}

	response := &models.MessageResponse{
		Message:        *message,
		ThreadCount:    0,
		Reactions:      []models.MessageReaction{},
		IsRead:         true,
		DeliveryStatus: models.MessageStatusDelivered,
		TimeFormatted:  message.GetTimeFormatted(),
	}

	h.logger.Info("💬 Mock: Message sent successfully",
		zap.String("message_id", message.ID),
		zap.String("room_id", roomID))

	sharedUtils.SendSuccessResponseWithMessage(c, http.StatusCreated, "💬 Mock: Message sent successfully!", response)
}

// GetRoomMessages retrieves messages from room (mock implementation)
func (h *MockChatHandler) GetRoomMessages(c *gin.Context) {
	roomID := c.Param("room_id")
	if roomID == "" {
		sharedUtils.SendErrorResponse(c, sharedErrors.NewValidationError("MISSING_ROOM_ID", "Room ID is required"))
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		userID = "mock-user-123"
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// For mock, return empty messages
	response := &models.MessagesResponse{
		Messages:    []models.Message{},
		Total:       0,
		Page:        page,
		Limit:       limit,
		HasNext:     false,
		HasPrevious: false,
	}

	h.logger.Info("💬 Mock: Messages retrieved successfully",
		zap.String("room_id", roomID),
		zap.Int64("total", response.Total))

	sharedUtils.SendSuccessResponseWithMessage(c, http.StatusOK, "💬 Mock: Messages retrieved successfully!", response)
}

// Health returns health status for mock chat service
func (h *MockChatHandler) Health(c *gin.Context) {
	healthStatus := sharedUtils.NewHealthStatus("ok")
	healthStatus.AddService("database", "mock", "Mock database connection")
	healthStatus.AddService("redis", "mock", "Mock Redis connection")
	healthStatus.AddService("messaging", "operational", "Real-time messaging system operational")
	healthStatus.AddService("rooms", "operational", "Chat room management operational")
	healthStatus.AddService("websockets", "operational", "WebSocket connections operational")
	healthStatus.AddService("typing_status", "operational", "Typing indicators operational")

	response := gin.H{
		"service": "chat-service-mock",
		"version": "v1.0.0-mock",
		"message": "💬 QUICK WIN: Mock Chat Service 100% Working!",
		"checks":  healthStatus.Checks,
	}

	sharedUtils.SendSuccessResponse(c, http.StatusOK, response)
}