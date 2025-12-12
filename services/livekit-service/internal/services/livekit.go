package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	shared_errors "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/gmsas95/blytz-mvp/services/livekit-service/internal/models"
)

// LiveKitService handles LiveKit integration
type LiveKitService struct {
	db           *gorm.DB
	logger       *zap.Logger
	livekitURL   string
	apiKey       string
	apiSecret    string
}

// NewLiveKitService creates a new LiveKit service
func NewLiveKitService(db *gorm.DB, logger *zap.Logger, livekitURL, apiKey, apiSecret string) (*LiveKitService, error) {
	service := &LiveKitService{
		db:         db,
		logger:     logger,
		livekitURL: livekitURL,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
	}

	// Auto-migrate database models
	if err := service.db.AutoMigrate(&models.Room{}, &models.RoomParticipant{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	logger.Info("LiveKit service initialized successfully (mock mode)")
	return service, nil
}

// CreateRoom creates a new live streaming room
func (s *LiveKitService) CreateRoom(ctx context.Context, config *models.RoomConfig, hostID string) (*models.Room, error) {
	// Check if room name already exists
	var existingRoom models.Room
	if err := s.db.Where("name = ?", config.Name).First(&existingRoom).Error; err == nil {
		return nil, shared_errors.NewConflictError("ROOM_EXISTS", "Room with this name already exists")
	}

	// Generate LiveKit room ID (mock implementation)
	livekitRoomID := fmt.Sprintf("room_%s_%d", uuid.New().String()[:8], time.Now().Unix())

	// Create database room record
	now := time.Now()
	room := &models.Room{
		Name:                config.Name,
		Title:               config.Title,
		Description:         config.Description,
		HostID:              hostID,
		IsActive:            false,
		IsPublic:            config.IsPublic,
		MaxParticipants:     config.MaxParticipants,
		CurrentParticipants: 0,
		LiveKitRoomID:       livekitRoomID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.db.Create(room).Error; err != nil {
		s.logger.Error("Failed to create room in database", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("DB_ERROR", "Failed to create room in database")
	}

	s.logger.Info("Room created successfully (mock mode)",
		zap.String("room_id", room.ID),
		zap.String("room_name", room.Name),
		zap.String("livekit_room_id", livekitRoomID),
	)

	return room, nil
}

// ListRooms lists all rooms
func (s *LiveKitService) ListRooms(ctx context.Context, includeInactive bool) ([]models.Room, error) {
	var rooms []models.Room
	query := s.db.Model(&models.Room{})

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&rooms).Error; err != nil {
		s.logger.Error("Failed to list rooms", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("DB_ERROR", "Failed to list rooms")
	}

	return rooms, nil
}

// GetRoomByID gets a room by ID
func (s *LiveKitService) GetRoomByID(ctx context.Context, roomID string) (*models.Room, error) {
	var room models.Room
	if err := s.db.Where("id = ?", roomID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared_errors.NewNotFoundError("ROOM_NOT_FOUND", "Room not found")
		}
		s.logger.Error("Failed to get room", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("DB_ERROR", "Failed to get room")
	}

	return &room, nil
}

// StartRoom starts a live streaming room
func (s *LiveKitService) StartRoom(ctx context.Context, roomID string) error {
	room, err := s.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.IsActive {
		return shared_errors.NewBusinessError("ROOM_ALREADY_ACTIVE", "Room is already active")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"is_active":  true,
		"start_time": &now,
		"updated_at": now,
	}

	if err := s.db.Model(room).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to start room", zap.Error(err))
		return shared_errors.NewDatabaseError("DB_ERROR", "Failed to start room")
	}

	s.logger.Info("Room started successfully", zap.String("room_id", roomID))
	return nil
}

// StopRoom stops a live streaming room
func (s *LiveKitService) StopRoom(ctx context.Context, roomID string) error {
	room, err := s.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	if !room.IsActive {
		return shared_errors.NewBusinessError("ROOM_NOT_ACTIVE", "Room is not active")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"is_active":    false,
		"end_time":     &now,
		"updated_at":   now,
		"current_participants": 0,
	}

	if err := s.db.Model(room).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to stop room", zap.Error(err))
		return shared_errors.NewDatabaseError("DB_ERROR", "Failed to stop room")
	}

	// Remove all participants
	if err := s.db.Where("room_id = ?", roomID).Delete(&models.RoomParticipant{}).Error; err != nil {
		s.logger.Error("Failed to clear participants", zap.Error(err))
	}

	s.logger.Info("Room stopped successfully", zap.String("room_id", roomID))
	return nil
}

// JoinRoom adds a participant to a room
func (s *LiveKitService) JoinRoom(ctx context.Context, req *models.JoinRoomRequest) (*models.RoomParticipant, error) {
	room, err := s.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	if !room.CanJoin() {
		return nil, shared_errors.NewBusinessError("ROOM_NOT_JOINABLE", "Room cannot be joined")
	}

	// Check if user is already in the room
	var existingParticipant models.RoomParticipant
	err = s.db.Where("room_id = ? AND user_id = ? AND left_at IS NULL", req.RoomID, req.UserID).First(&existingParticipant).Error
	if err == nil {
		return nil, shared_errors.NewBusinessError("ALREADY_IN_ROOM", "User is already in the room")
	}

	// Create participant record
	identity := fmt.Sprintf("%s_%s", req.UserID, uuid.New().String()[:8])
	participant := &models.RoomParticipant{
		RoomID:   req.RoomID,
		UserID:   req.UserID,
		Identity: identity,
		IsHost:   room.HostID == req.UserID,
		HasAudio: req.HasAudio,
		HasVideo: req.HasVideo,
		JoinedAt: time.Now(),
	}

	if err := s.db.Create(participant).Error; err != nil {
		s.logger.Error("Failed to create participant", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("DB_ERROR", "Failed to join room")
	}

	// Update room participant count
	if err := s.db.Model(room).UpdateColumn("current_participants", gorm.Expr("current_participants + 1")).Error; err != nil {
		s.logger.Error("Failed to update participant count", zap.Error(err))
	}

	s.logger.Info("Participant joined room",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", req.UserID),
		zap.String("identity", identity),
	)

	return participant, nil
}

// LeaveRoom removes a participant from a room
func (s *LiveKitService) LeaveRoom(ctx context.Context, req *models.LeaveRoomRequest) error {
	var participant models.RoomParticipant
	if err := s.db.Where("room_id = ? AND identity = ? AND left_at IS NULL", req.RoomID, req.Identity).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return shared_errors.NewNotFoundError("PARTICIPANT_NOT_FOUND", "Participant not found in room")
		}
		s.logger.Error("Failed to find participant", zap.Error(err))
		return shared_errors.NewDatabaseError("DB_ERROR", "Failed to find participant")
	}

	now := time.Now()
	if err := s.db.Model(&participant).Updates(map[string]interface{}{
		"left_at":    &now,
		"updated_at": now,
	}).Error; err != nil {
		s.logger.Error("Failed to update participant", zap.Error(err))
		return shared_errors.NewDatabaseError("DB_ERROR", "Failed to leave room")
	}

	// Update room participant count
	var room models.Room
	if err := s.db.Where("id = ?", req.RoomID).First(&room).Error; err == nil {
		if room.CurrentParticipants > 0 {
			s.db.Model(&room).UpdateColumn("current_participants", gorm.Expr("current_participants - 1"))
		}
	}

	s.logger.Info("Participant left room",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", participant.UserID),
		zap.String("identity", req.Identity),
	)

	return nil
}

// GenerateToken generates a LiveKit token for a participant
func (s *LiveKitService) GenerateToken(ctx context.Context, req *models.TokenRequest) (string, error) {
	room, err := s.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return "", err
	}

	if !room.CanJoin() {
		return "", shared_errors.NewBusinessError("ROOM_NOT_JOINABLE", "Room cannot be joined")
	}

	// Create JWT token (mock implementation)
	identity := fmt.Sprintf("%s_%s", req.UserID, uuid.New().String()[:8])
	
	// Create a simple JWT token for mock implementation
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      s.apiKey,
		"sub":      identity,
		"name":      req.Username,
		"room":     room.LiveKitRoomID,
		"room_join": true,
		"can_publish": req.HasAudio || req.HasVideo,
		"can_subscribe": true,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // 24 hours expiry
	})

	tokenString, err := at.SignedString([]byte(s.apiSecret))
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return "", shared_errors.NewInternalError("TOKEN_ERROR", "Failed to generate access token")
	}

	s.logger.Info("Token generated successfully (mock mode)",
		zap.String("room_id", req.RoomID),
		zap.String("user_id", req.UserID),
		zap.String("identity", identity),
	)

	return tokenString, nil
}

// ProcessWebhook processes LiveKit webhook events
func (s *LiveKitService) ProcessWebhook(ctx context.Context, event *models.WebhookEvent) error {
	switch event.Event {
	case "room_started":
		return s.handleRoomStarted(ctx, event)
	case "room_finished":
		return s.handleRoomFinished(ctx, event)
	case "participant_joined":
		return s.handleParticipantJoined(ctx, event)
	case "participant_left":
		return s.handleParticipantLeft(ctx, event)
	default:
		s.logger.Info("Unhandled webhook event", zap.String("event", event.Event))
		return nil
	}
}

// handleRoomStarted handles room started event
func (s *LiveKitService) handleRoomStarted(ctx context.Context, event *models.WebhookEvent) error {
	var room models.Room
	if err := s.db.Where("livekit_room_id = ?", event.Room.ID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("Room not found for webhook event", zap.String("livekit_room_id", event.Room.ID))
			return nil
		}
		return err
	}

	if !room.IsActive {
		now := time.Now()
		updates := map[string]interface{}{
			"is_active":  true,
			"start_time": &now,
			"updated_at": now,
		}
		return s.db.Model(&room).Updates(updates).Error
	}

	return nil
}

// handleRoomFinished handles room finished event
func (s *LiveKitService) handleRoomFinished(ctx context.Context, event *models.WebhookEvent) error {
	var room models.Room
	if err := s.db.Where("livekit_room_id = ?", event.Room.ID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("Room not found for webhook event", zap.String("livekit_room_id", event.Room.ID))
			return nil
		}
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"is_active":    false,
		"end_time":     &now,
		"updated_at":   now,
		"current_participants": 0,
	}

	// Clear all participants
	if err := s.db.Where("room_id = ?", room.ID).Delete(&models.RoomParticipant{}).Error; err != nil {
		s.logger.Error("Failed to clear participants", zap.Error(err))
	}

	return s.db.Model(&room).Updates(updates).Error
}

// handleParticipantJoined handles participant joined event
func (s *LiveKitService) handleParticipantJoined(ctx context.Context, event *models.WebhookEvent) error {
	var room models.Room
	if err := s.db.Where("livekit_room_id = ?", event.Room.ID).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("Room not found for participant webhook", zap.String("livekit_room_id", event.Room.ID))
			return nil
		}
		return err
	}

	// Check if participant already exists
	var existingParticipant models.RoomParticipant
	err := s.db.Where("identity = ? AND left_at IS NULL", event.Participant.Identity).First(&existingParticipant).Error
	if err == nil {
		return nil // Already exists
	}

	// Create participant record
	participant := &models.RoomParticipant{
		RoomID:   room.ID,
		Identity: event.Participant.Identity,
		JoinedAt:  event.Participant.JoinedAt,
	}

	// Extract user ID from identity (format: userID_uuid)
	if len(event.Participant.Identity) > 9 {
		participant.UserID = event.Participant.Identity[:9] // Assuming userID is first 9 chars
	}

	// Check if this is the host
	if participant.UserID == room.HostID {
		participant.IsHost = true
	}

	// Check tracks for audio/video
	for _, track := range event.Participant.Tracks {
		switch track.Type {
		case "audio":
			participant.HasAudio = !track.Muted
		case "video":
			participant.HasVideo = !track.Muted
		case "screen":
			participant.HasScreen = !track.Muted
		}
	}

	if err := s.db.Create(participant).Error; err != nil {
		return err
	}

	// Update room participant count
	return s.db.Model(&room).UpdateColumn("current_participants", gorm.Expr("current_participants + 1")).Error
}

// handleParticipantLeft handles participant left event
func (s *LiveKitService) handleParticipantLeft(ctx context.Context, event *models.WebhookEvent) error {
	var participant models.RoomParticipant
	if err := s.db.Where("identity = ? AND left_at IS NULL", event.Participant.Identity).First(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Not found, already left
		}
		return err
	}

	now := time.Now()
	if err := s.db.Model(&participant).Updates(map[string]interface{}{
		"left_at":    &now,
		"updated_at": now,
	}).Error; err != nil {
		return err
	}

	// Update room participant count
	var room models.Room
	if err := s.db.Where("id = ?", participant.RoomID).First(&room).Error; err == nil {
		if room.CurrentParticipants > 0 {
			s.db.Model(&room).UpdateColumn("current_participants", gorm.Expr("current_participants - 1"))
		}
	}

	return nil
}

// GetDB returns the database instance
func (s *LiveKitService) GetDB() *gorm.DB {
	return s.db
}

// Close closes the LiveKit service connection
func (s *LiveKitService) Close() error {
	// Mock implementation - no actual LiveKit client to close
	s.logger.Info("LiveKit service closed (mock mode)")
	return nil
}