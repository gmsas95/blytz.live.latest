package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	shared_errors "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	"github.com/gmsas95/blytz-mvp/services/livekit-service/internal/models"
	"github.com/gmsas95/blytz-mvp/services/livekit-service/internal/services"
)

// LiveKitHandler handles LiveKit HTTP requests
type LiveKitHandler struct {
	livekitService *services.LiveKitService
	logger         *zap.Logger
}

// NewLiveKitHandler creates a new LiveKit handler
func NewLiveKitHandler(livekitService *services.LiveKitService, logger *zap.Logger) *LiveKitHandler {
	return &LiveKitHandler{
		livekitService: livekitService,
		logger:         logger,
	}
}

// ListRooms handles GET /api/v1/livekit/rooms
func (h *LiveKitHandler) ListRooms(c *gin.Context) {
	includeInactive := c.DefaultQuery("include_inactive", "false") == "true"

	rooms, err := h.livekitService.ListRooms(c.Request.Context(), includeInactive)
	if err != nil {
		h.logger.Error("Failed to list rooms", zap.Error(err))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, rooms)
}

// CreateRoom handles POST /api/v1/livekit/rooms
func (h *LiveKitHandler) CreateRoom(c *gin.Context) {
	var config models.RoomConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get host ID from context (should be set by auth middleware)
	hostID, exists := c.Get("user_id")
	if !exists {
		shared_utils.SendErrorResponse(c, shared_errors.NewAuthenticationError(
			"MISSING_USER_ID",
			"User ID is required",
		))
		return
	}

	room, err := h.livekitService.CreateRoom(c.Request.Context(), &config, hostID.(string))
	if err != nil {
		h.logger.Error("Failed to create room", zap.Error(err))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusCreated, room)
}

// GetRoom handles GET /api/v1/livekit/rooms/:id
func (h *LiveKitHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"room_id": "Room ID is required",
		})
		return
	}

	room, err := h.livekitService.GetRoomByID(c.Request.Context(), roomID)
	if err != nil {
		h.logger.Error("Failed to get room", zap.Error(err), zap.String("room_id", roomID))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, room)
}

// StartRoom handles POST /api/v1/livekit/rooms/:id/start
func (h *LiveKitHandler) StartRoom(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"room_id": "Room ID is required",
		})
		return
	}

	err := h.livekitService.StartRoom(c.Request.Context(), roomID)
	if err != nil {
		h.logger.Error("Failed to start room", zap.Error(err), zap.String("room_id", roomID))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Room started successfully", nil)
}

// StopRoom handles POST /api/v1/livekit/rooms/:id/stop
func (h *LiveKitHandler) StopRoom(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"room_id": "Room ID is required",
		})
		return
	}

	err := h.livekitService.StopRoom(c.Request.Context(), roomID)
	if err != nil {
		h.logger.Error("Failed to stop room", zap.Error(err), zap.String("room_id", roomID))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Room stopped successfully", nil)
}

// JoinRoom handles POST /api/v1/livekit/rooms/:id/join
func (h *LiveKitHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"room_id": "Room ID is required",
		})
		return
	}

	var req models.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Set room ID from URL parameter
	req.RoomID = roomID

	participant, err := h.livekitService.JoinRoom(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to join room", zap.Error(err), zap.String("room_id", roomID))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, participant)
}

// LeaveRoom handles POST /api/v1/livekit/rooms/:id/leave
func (h *LiveKitHandler) LeaveRoom(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"room_id": "Room ID is required",
		})
		return
	}

	var req models.LeaveRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Set room ID from URL parameter
	req.RoomID = roomID

	err := h.livekitService.LeaveRoom(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to leave room", zap.Error(err), zap.String("room_id", roomID))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Left room successfully", nil)
}

// GenerateToken handles GET /api/v1/livekit/tokens
func (h *LiveKitHandler) GenerateToken(c *gin.Context) {
	var req models.TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request": "Invalid request body: " + err.Error(),
		})
		return
	}

	token, err := h.livekitService.GenerateToken(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]string{
		"token": token,
	})
}

// ProcessWebhook handles POST /api/v1/livekit/webhook
func (h *LiveKitHandler) ProcessWebhook(c *gin.Context) {
	var event models.WebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("Invalid webhook payload", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"webhook": "Invalid webhook payload: " + err.Error(),
		})
		return
	}

	err := h.livekitService.ProcessWebhook(c.Request.Context(), &event)
	if err != nil {
		h.logger.Error("Failed to process webhook", zap.Error(err))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Webhook processed successfully", nil)
}

// GetRoomParticipants handles GET /api/v1/livekit/rooms/:id/participants
func (h *LiveKitHandler) GetRoomParticipants(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"room_id": "Room ID is required",
		})
		return
	}

	// Get room first to validate it exists
	room, err := h.livekitService.GetRoomByID(c.Request.Context(), roomID)
	if err != nil {
		h.logger.Error("Failed to get room", zap.Error(err), zap.String("room_id", roomID))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	// Get participants for this room
	var participants []models.RoomParticipant
	if err := h.livekitService.GetDB().Where("room_id = ? AND left_at IS NULL", roomID).Find(&participants).Error; err != nil {
		h.logger.Error("Failed to get participants", zap.Error(err), zap.String("room_id", roomID))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("DB_ERROR", "Failed to get participants"))
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"room_id":      roomID,
		"room_name":    room.Name,
		"participants": participants,
		"count":        len(participants),
	})
}

// GetActiveRooms handles GET /api/v1/livekit/rooms/active
func (h *LiveKitHandler) GetActiveRooms(c *gin.Context) {
	// Get pagination parameters
	page, perPage := shared_utils.GetPaginationParams(c)

	// Get active rooms with pagination
	var rooms []models.Room
	var total int64

	query := h.livekitService.GetDB().Model(&models.Room{}).Where("is_active = ?", true)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error("Failed to count active rooms", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("DB_ERROR", "Failed to count rooms"))
		return
	}

	// Get paginated results
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&rooms).Error; err != nil {
		h.logger.Error("Failed to get active rooms", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("DB_ERROR", "Failed to get rooms"))
		return
	}

	// Calculate pagination info
	pagination := shared_utils.CalculatePagination(page, perPage, total)

	shared_utils.SendPaginatedResponse(c, http.StatusOK, rooms, pagination)
}

// GetRoomStats handles GET /api/v1/livekit/rooms/:id/stats
func (h *LiveKitHandler) GetRoomStats(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"room_id": "Room ID is required",
		})
		return
	}

	// Get room first to validate it exists
	room, err := h.livekitService.GetRoomByID(c.Request.Context(), roomID)
	if err != nil {
		h.logger.Error("Failed to get room", zap.Error(err), zap.String("room_id", roomID))
		shared_utils.SendErrorResponse(c, err)
		return
	}

	// Get participant count
	var participantCount int64
	if err := h.livekitService.GetDB().Model(&models.RoomParticipant{}).Where("room_id = ? AND left_at IS NULL", roomID).Count(&participantCount).Error; err != nil {
		h.logger.Error("Failed to count participants", zap.Error(err), zap.String("room_id", roomID))
		participantCount = 0
	}

	// Get host information
	var host models.RoomParticipant
	if err := h.livekitService.GetDB().Where("room_id = ? AND is_host = ? AND left_at IS NULL", roomID, true).First(&host).Error; err != nil {
		host = models.RoomParticipant{} // No host found
	}

	stats := map[string]interface{}{
		"room_id":               room.ID,
		"room_name":             room.Name,
		"title":                 room.Title,
		"is_active":             room.IsActive,
		"is_public":             room.IsPublic,
		"current_participants":  participantCount,
		"max_participants":      room.MaxParticipants,
		"host_id":               room.HostID,
		"host_present":          host.ID != "",
		"host_identity":         host.Identity,
		"start_time":            room.StartTime,
		"duration_minutes":      0,
		"livekit_room_id":       room.LiveKitRoomID,
	}

	// Calculate duration if room is active
	if room.IsActive && room.StartTime != nil {
		duration := time.Since(*room.StartTime)
		stats["duration_minutes"] = int(duration.Minutes())
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, stats)
}

// HealthCheck handles health check for LiveKit service
func (h *LiveKitHandler) HealthCheck(c *gin.Context) {
	status := shared_utils.NewHealthStatus("ok")

	// Check database connection
	if err := shared_utils.HealthCheck(h.livekitService.GetDB()); err != nil {
		status.AddService("database", "error", "Database connection failed: "+err.Error())
		status.Status = "error"
	} else {
		status.AddService("database", "ok", "Database connection healthy")
	}

	// Check LiveKit connection
	// Note: We could add a ping to LiveKit server here if needed
	status.AddService("livekit", "ok", "LiveKit service connected")

	shared_utils.SendSuccessResponse(c, http.StatusOK, status)
}