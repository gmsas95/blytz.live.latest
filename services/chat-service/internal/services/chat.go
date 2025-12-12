package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz-mvp/services/chat-service/internal/models"
	sharedErrors "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
)

type ChatService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewChatService(db *gorm.DB, logger *zap.Logger) *ChatService {
	return &ChatService{
		db:     db,
		logger: logger,
	}
}

// === ROOM MANAGEMENT ===

// CreateRoom creates new chat room
func (s *ChatService) CreateRoom(ctx context.Context, creatorID string, req *models.CreateRoomRequest) (*models.ChatRoom, error) {
	s.logger.Info("💬 Chat Service: Creating room", 
		zap.String("creator_id", creatorID),
		zap.String("name", req.Name))

	// Validate request
	if err := s.validateCreateRoomRequest(req); err != nil {
		return nil, err
	}

	// Create room
	room := &models.ChatRoom{
		Name:           req.Name,
		Description:     req.Description,
		Type:           req.Type,
		IsPrivate:      req.IsPrivate,
		MaxMembers:     req.MaxMembers,
		Avatar:         req.Avatar,
		CoverImage:     req.CoverImage,
		Category:       req.Category,
		Settings:       req.Settings,
		Metadata:       req.Metadata,
		Status:         models.RoomStatusActive,
		CurrentMembers: len(req.Members) + 1, // Include creator
		LastActivity:   time.Now(),
		CreatedBy:      creatorID,
	}

	// Set tags array
	room.SetTagsArray(req.Tags)

	// Create room with members in transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create room
		if err := tx.Create(room).Error; err != nil {
			return fmt.Errorf("failed to create room: %w", err)
		}

		// Add creator as owner
		creatorMember := &models.RoomMember{
			RoomID:    room.ID,
			UserID:    creatorID,
			Role:      models.UserRoleOwner,
			JoinedAt:  time.Now(),
			IsActive:  true,
		}
		if err := tx.Create(creatorMember).Error; err != nil {
			return fmt.Errorf("failed to add creator to room: %w", err)
		}

		// Add other members
		for _, memberID := range req.Members {
			if memberID == creatorID {
				continue // Skip creator
			}

			member := &models.RoomMember{
				RoomID:    room.ID,
				UserID:    memberID,
				Role:      models.UserRoleMember,
				JoinedAt:  time.Now(),
				IsActive:  true,
			}
			if err := tx.Create(member).Error; err != nil {
				return fmt.Errorf("failed to add member to room: %w", err)
			}
		}

		// Create system message
		systemMessage := &models.Message{
			RoomID:    room.ID,
			UserID:    "system",
			Content:   fmt.Sprintf(`{"action": "room_created", "room_name": "%s", "creator": "%s"}`, room.Name, creatorID),
			Type:      models.MessageTypeSystem,
			Timestamp: time.Now(),
		}
		if err := tx.Create(systemMessage).Error; err != nil {
			s.logger.Error("Failed to create system message", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("💬 Chat Service: Failed to create room", zap.Error(err))
		return nil, err
	}

	s.logger.Info("💬 Chat Service: Room created successfully", 
		zap.String("room_id", room.ID),
		zap.String("name", room.Name))

	return room, nil
}

// GetRoom retrieves room by ID
func (s *ChatService) GetRoom(ctx context.Context, roomID, userID string) (*models.ChatRoom, error) {
	s.logger.Info("💬 Chat Service: Getting room", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID))

	var room models.ChatRoom
	err := s.db.Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ?", true).Preload("User")
	}).Preload("Creator").Where("id = ? AND status = ?", roomID, models.RoomStatusActive).First(&room).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("room not found")
		}
		return nil, err
	}

	// Check if user is member (for private rooms)
	if room.IsPrivate {
		isMember := false
		for _, member := range room.Members {
			if member.UserID == userID && member.IsActive {
				isMember = true
				break
			}
		}
		if !isMember {
			return nil, fmt.Errorf("access denied - not a member")
		}
	}

	s.logger.Info("💬 Chat Service: Room retrieved successfully", 
		zap.String("room_id", roomID))
	return &room, nil
}

// UpdateRoom updates existing room
func (s *ChatService) UpdateRoom(ctx context.Context, roomID, userID string, req *models.UpdateRoomRequest) (*models.ChatRoom, error) {
	s.logger.Info("💬 Chat Service: Updating room", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID))

	// Get existing room and check permissions
	var room models.ChatRoom
	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("id = ? AND status = ?", roomID, models.RoomStatusActive).First(&room).Error
		if err != nil {
			return err
		}

		// Check if user is owner or admin
		var member models.RoomMember
		err = tx.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&member).Error
		if err != nil {
			return fmt.Errorf("user not found in room")
		}

		if member.Role != models.UserRoleOwner && member.Role != models.UserRoleAdmin {
			return fmt.Errorf("insufficient permissions")
		}

		// Update fields
		if req.Name != "" {
			room.Name = req.Name
		}
		if req.Description != "" {
			room.Description = req.Description
		}
		if req.Avatar != "" {
			room.Avatar = req.Avatar
		}
		if req.CoverImage != "" {
			room.CoverImage = req.CoverImage
		}
		if req.Category != "" {
			room.Category = req.Category
		}
		if req.MaxMembers > 0 {
			room.MaxMembers = req.MaxMembers
		}
		if req.Tags != nil {
			room.SetTagsArray(req.Tags)
		}
		if req.Settings != "" {
			room.Settings = req.Settings
		}
		if req.Metadata != "" {
			room.Metadata = req.Metadata
		}

		// Update room
		if err := tx.Save(&room).Error; err != nil {
			return fmt.Errorf("failed to update room: %w", err)
		}

		// Create system message
		systemMessage := &models.Message{
			RoomID:    roomID,
			UserID:    "system",
			Content:   fmt.Sprintf(`{"action": "room_updated", "room_name": "%s", "updated_by": "%s"}`, room.Name, userID),
			Type:      models.MessageTypeSystem,
			Timestamp: time.Now(),
		}
		if err := tx.Create(systemMessage).Error; err != nil {
			s.logger.Error("Failed to create system message", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("💬 Chat Service: Failed to update room", zap.Error(err))
		return nil, err
	}

	s.logger.Info("💬 Chat Service: Room updated successfully", 
		zap.String("room_id", roomID))
	return &room, nil
}

// DeleteRoom deletes room
func (s *ChatService) DeleteRoom(ctx context.Context, roomID, userID string) error {
	s.logger.Info("💬 Chat Service: Deleting room", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID))

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Get room and check permissions
		var room models.ChatRoom
		err := tx.Where("id = ? AND status = ?", roomID, models.RoomStatusActive).First(&room).Error
		if err != nil {
			return fmt.Errorf("room not found")
		}

		// Check if user is owner
		var member models.RoomMember
		err = tx.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&member).Error
		if err != nil {
			return fmt.Errorf("user not found in room")
		}

		if member.Role != models.UserRoleOwner {
			return fmt.Errorf("only room owner can delete room")
		}

		// Delete room (soft delete)
		now := time.Now()
		room.Status = models.RoomStatusDeleted
		room.DeletedAt = &now
		if err := tx.Save(&room).Error; err != nil {
			return fmt.Errorf("failed to delete room: %w", err)
		}

		// Create system message
		systemMessage := &models.Message{
			RoomID:    roomID,
			UserID:    "system",
			Content:   fmt.Sprintf(`{"action": "room_deleted", "room_name": "%s", "deleted_by": "%s"}`, room.Name, userID),
			Type:      models.MessageTypeSystem,
			Timestamp: now,
		}
		if err := tx.Create(systemMessage).Error; err != nil {
			s.logger.Error("Failed to create system message", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("💬 Chat Service: Failed to delete room", zap.Error(err))
		return err
	}

	s.logger.Info("💬 Chat Service: Room deleted successfully", 
		zap.String("room_id", roomID))
	return nil
}

// === MEMBER MANAGEMENT ===

// AddRoomMember adds member to room
func (s *ChatService) AddRoomMember(ctx context.Context, roomID, userID string, req *models.AddMemberRequest) error {
	s.logger.Info("💬 Chat Service: Adding members to room", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
		zap.Int("member_count", len(req.UserIDs)))

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Get room and check permissions
		var room models.ChatRoom
		err := tx.Where("id = ? AND status = ?", roomID, models.RoomStatusActive).First(&room).Error
		if err != nil {
			return fmt.Errorf("room not found")
		}

		// Check if user has permission
		var member models.RoomMember
		err = tx.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&member).Error
		if err != nil {
			return fmt.Errorf("user not found in room")
		}

		if member.Role != models.UserRoleOwner && member.Role != models.UserRoleAdmin {
			return fmt.Errorf("insufficient permissions")
		}

		// Check room capacity
		if room.CurrentMembers+len(req.UserIDs) > room.MaxMembers {
			return fmt.Errorf("room is at maximum capacity")
		}

		// Add members
		for _, memberID := range req.UserIDs {
			// Check if already member
			var existingMember models.RoomMember
			err = tx.Where("room_id = ? AND user_id = ?", roomID, memberID).First(&existingMember).Error
			if err == nil {
				// Reactivate if inactive
				if !existingMember.IsActive {
					existingMember.IsActive = true
					existingMember.LeftAt = nil
					if err := tx.Save(&existingMember).Error; err != nil {
						return fmt.Errorf("failed to reactivate member: %w", err)
					}
				}
				continue
			}

			// Add new member
			role := req.Role
			if role == "" {
				role = models.UserRoleMember
			}

			newMember := &models.RoomMember{
				RoomID:    roomID,
				UserID:    memberID,
				Role:      role,
				JoinedAt:  time.Now(),
				IsActive:  true,
			}
			if err := tx.Create(newMember).Error; err != nil {
				return fmt.Errorf("failed to add member: %w", err)
			}
		}

		// Update room member count
		if err := tx.Model(&room).Update("current_members", gorm.Expr("current_members + ?", len(req.UserIDs))).Error; err != nil {
			return fmt.Errorf("failed to update member count: %w", err)
		}

		// Create system message
		systemMessage := &models.Message{
			RoomID:    roomID,
			UserID:    "system",
			Content:   fmt.Sprintf(`{"action": "members_added", "count": %d, "added_by": "%s"}`, len(req.UserIDs), userID),
			Type:      "user_join",
			Timestamp: time.Now(),
		}
		if err := tx.Create(systemMessage).Error; err != nil {
			s.logger.Error("Failed to create system message", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("💬 Chat Service: Failed to add members", zap.Error(err))
		return err
	}

	s.logger.Info("💬 Chat Service: Members added successfully", 
		zap.String("room_id", roomID),
		zap.Int("count", len(req.UserIDs)))
	return nil
}

// RemoveRoomMember removes member from room
func (s *ChatService) RemoveRoomMember(ctx context.Context, roomID, userID, memberID string) error {
	s.logger.Info("💬 Chat Service: Removing member from room", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
		zap.String("member_id", memberID))

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Get room and check permissions
		var room models.ChatRoom
		err := tx.Where("id = ? AND status = ?", roomID, models.RoomStatusActive).First(&room).Error
		if err != nil {
			return fmt.Errorf("room not found")
		}

		// Get member to remove
		var memberToRemove models.RoomMember
		err = tx.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, memberID, true).First(&memberToRemove).Error
		if err != nil {
			return fmt.Errorf("member not found")
		}

		// Check permissions (owner, admin, or removing self)
		var requestor models.RoomMember
		err = tx.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&requestor).Error
		if err != nil {
			return fmt.Errorf("requestor not found in room")
		}

		// Allow self-removal or owner/admin removal
		isSelf := userID == memberID
		isAdmin := requestor.Role == models.UserRoleOwner || requestor.Role == models.UserRoleAdmin

		if !isSelf && !isAdmin {
			return fmt.Errorf("insufficient permissions")
		}

		// Cannot remove owner
		if memberToRemove.Role == models.UserRoleOwner && !isSelf {
			return fmt.Errorf("cannot remove room owner")
		}

		// Deactivate member
		now := time.Now()
		memberToRemove.IsActive = false
		memberToRemove.LeftAt = &now
		if err := tx.Save(&memberToRemove).Error; err != nil {
			return fmt.Errorf("failed to remove member: %w", err)
		}

		// Update room member count
		if err := tx.Model(&room).Update("current_members", gorm.Expr("current_members - 1")).Error; err != nil {
			return fmt.Errorf("failed to update member count: %w", err)
		}

		// Create system message
		actionType := "member_removed"
		if isSelf {
			actionType = "member_left"
		}

		systemMessage := &models.Message{
			RoomID:    roomID,
			UserID:    "system",
			Content:   fmt.Sprintf(`{"action": "%s", "user_id": "%s", "removed_by": "%s"}`, actionType, memberID, userID),
			Type:      models.MessageUserLeave,
			Timestamp: now,
		}
		if err := tx.Create(systemMessage).Error; err != nil {
			s.logger.Error("Failed to create system message", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("💬 Chat Service: Failed to remove member", zap.Error(err))
		return err
	}

	s.logger.Info("💬 Chat Service: Member removed successfully", 
		zap.String("room_id", roomID),
		zap.String("member_id", memberID))
	return nil
}

// === MESSAGE MANAGEMENT ===

// SendMessage sends message to room
func (s *ChatService) SendMessage(ctx context.Context, roomID, userID string, req *models.SendMessageRequest) (*models.Message, error) {
	s.logger.Info("💬 Chat Service: Sending message", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID))

	// Validate user is member
	var member models.RoomMember
	err := s.db.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&member).Error
	if err != nil {
		return nil, fmt.Errorf("user not found in room")
	}

	// Validate room is active
	var room models.ChatRoom
	err = s.db.Where("id = ? AND status = ?", roomID, models.RoomStatusActive).First(&room).Error
	if err != nil {
		return nil, fmt.Errorf("room not found")
	}

	// Create message
	message := &models.Message{
		RoomID:    roomID,
		UserID:    userID,
		Content:   req.Content,
		Type:      req.Type,
		ReplyToID: req.ReplyToID,
		Timestamp: time.Now(),
		IPAddress: "127.0.0.1", // Would get from request context
		UserAgent:  "Chat Service",
	}

	if message.Type == "" {
		message.Type = models.MessageTypeText
	}

	// Set attachments array
	if req.Attachments != nil {
		message.SetAttachmentsArray(req.Attachments)
	}

	// Set mentions array
	if req.Mentions != nil {
		message.SetMentionsArray(req.Mentions)
	}

	// Create message and update room
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create message
		if err := tx.Create(message).Error; err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		// Update room last activity and last message
		updates := map[string]interface{}{
			"last_activity": time.Now(),
			"last_message": message.ID,
		}
		if err := tx.Model(&room).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update room: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("💬 Chat Service: Failed to send message", zap.Error(err))
		return nil, err
	}

	s.logger.Info("💬 Chat Service: Message sent successfully", 
		zap.String("message_id", message.ID),
		zap.String("room_id", roomID))

	return message, nil
}

// GetMessages retrieves messages from room
func (s *ChatService) GetMessages(ctx context.Context, roomID, userID string, page, limit int) (*models.MessagesResponse, error) {
	s.logger.Info("💬 Chat Service: Getting messages", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID))

	// Validate user is member
	var member models.RoomMember
	err := s.db.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&member).Error
	if err != nil {
		return nil, fmt.Errorf("user not found in room")
	}

	// Build query
	query := s.db.Model(&models.Message{}).
		Preload("Sender").
		Preload("ReplyTo").
		Where("room_id = ? AND is_deleted = ?", roomID, false).
		Order("timestamp DESC")

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	var messages []models.Message
	err = query.Limit(limit).Offset(offset).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse order for proper chronological display
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	hasNext := int64(page*limit) < total
	hasPrevious := page > 1

	response := &models.MessagesResponse{
		Messages:    messages,
		Total:       total,
		Page:        page,
		Limit:       limit,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}

	s.logger.Info("💬 Chat Service: Messages retrieved successfully", 
		zap.String("room_id", roomID),
		zap.Int64("total", total))

	return response, nil
}

// UpdateMessage updates existing message
func (s *ChatService) UpdateMessage(ctx context.Context, messageID, userID string, content string) (*models.Message, error) {
	s.logger.Info("💬 Chat Service: Updating message", 
		zap.String("message_id", messageID),
		zap.String("user_id", userID))

	var message models.Message
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Get message
		err := tx.Where("id = ? AND user_id = ? AND is_deleted = ?", messageID, userID, false).First(&message).Error
		if err != nil {
			return fmt.Errorf("message not found or unauthorized")
		}

		// Check if message is too old to edit (24 hours)
		if time.Since(message.Timestamp) > 24*time.Hour {
			return fmt.Errorf("message too old to edit")
		}

		// Update message
		message.Content = content
		message.IsEdited = true
		message.EditCount += 1
		now := time.Now()
		message.EditedAt = &now

		if err := tx.Save(&message).Error; err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("💬 Chat Service: Failed to update message", zap.Error(err))
		return nil, err
	}

	s.logger.Info("💬 Chat Service: Message updated successfully", 
		zap.String("message_id", messageID))
	return &message, nil
}

// DeleteMessage deletes message
func (s *ChatService) DeleteMessage(ctx context.Context, messageID, userID string) error {
	s.logger.Info("💬 Chat Service: Deleting message", 
		zap.String("message_id", messageID),
		zap.String("user_id", userID))

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Get message
		var message models.Message
		err := tx.Where("id = ? AND is_deleted = ?", messageID, false).First(&message).Error
		if err != nil {
			return fmt.Errorf("message not found")
		}

		// Check permissions (owner, admin, or message author)
		var member models.RoomMember
		err = tx.Where("room_id = ? AND user_id = ? AND is_active = ?", message.RoomID, userID, true).First(&member).Error
		if err != nil {
			return fmt.Errorf("user not found in room")
		}

		isAuthor := message.UserID == userID
		isAdmin := member.Role == models.UserRoleOwner || member.Role == models.UserRoleAdmin

		if !isAuthor && !isAdmin {
			return fmt.Errorf("insufficient permissions")
		}

		// Soft delete message
		now := time.Now()
		message.IsDeleted = true
		message.DeletedAt = &now

		if err := tx.Save(&message).Error; err != nil {
			return fmt.Errorf("failed to delete message: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("💬 Chat Service: Failed to delete message", zap.Error(err))
		return err
	}

	s.logger.Info("💬 Chat Service: Message deleted successfully", 
		zap.String("message_id", messageID))
	return nil
}

// === USER ROOMS & SEARCH ===

// GetUserRooms gets rooms for user
func (s *ChatService) GetUserRooms(ctx context.Context, userID string, page, limit int) (*models.RoomsResponse, error) {
	s.logger.Info("💬 Chat Service: Getting user rooms", zap.String("user_id", userID))

	// Build query
	query := s.db.Model(&models.ChatRoom{}).
		Joins("INNER JOIN room_members ON room_members.room_id = chat_rooms.id").
		Where("room_members.user_id = ? AND room_members.is_active = ? AND chat_rooms.status = ?", 
			userID, true, models.RoomStatusActive).
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Limit(5) // Limit members preview
		}).
		Preload("Creator").
		Order("chat_rooms.last_activity DESC")

	// Count total
	var total int64
	if err := s.db.Model(&models.ChatRoom{}).
		Joins("INNER JOIN room_members ON room_members.room_id = chat_rooms.id").
		Where("room_members.user_id = ? AND room_members.is_active = ? AND chat_rooms.status = ?", 
			userID, true, models.RoomStatusActive).
		Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	offset := (page - 1) * limit

	var rooms []models.ChatRoom
	err := query.Limit(limit).Offset(offset).Find(&rooms).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages

	response := &models.RoomsResponse{
		Rooms:      rooms,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    hasNext,
	}

	s.logger.Info("💬 Chat Service: User rooms retrieved successfully", 
		zap.String("user_id", userID),
		zap.Int64("total", total))

	return response, nil
}

// SearchMessages searches messages in room
func (s *ChatService) SearchMessages(ctx context.Context, roomID, userID string, req *models.SearchMessagesRequest) (*models.MessagesResponse, error) {
	s.logger.Info("💬 Chat Service: Searching messages", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
		zap.String("query", req.Query))

	// Validate user is member
	var member models.RoomMember
	err := s.db.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&member).Error
	if err != nil {
		return nil, fmt.Errorf("user not found in room")
	}

	// Build search query
	query := s.db.Model(&models.Message{}).
		Preload("Sender").
		Where("room_id = ? AND is_deleted = ?", roomID, false)

	// Apply filters
	if req.Query != "" {
		query = query.Where("content ILIKE ?", "%"+req.Query+"%")
	}
	if req.MessageType != "" {
		query = query.Where("type = ?", req.MessageType)
	}
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.HasMedia {
		query = query.Where("type IN ?", []string{models.MessageTypeImage, models.MessageTypeVideo, models.MessageTypeAudio, models.MessageTypeFile})
	}
	if req.StartDate != "" {
		// Parse start date (simplified)
		query = query.Where("timestamp >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		// Parse end date (simplified)
		query = query.Where("timestamp <= ?", req.EndDate)
	}

	// Apply sorting
	sortBy := "timestamp"
	if req.SortBy == "popularity" {
		sortBy = "created_at" // Would use reactions count
	}

	sortOrder := "DESC"
	if req.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(sortBy + " " + sortOrder)

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	offset := (req.Page - 1) * req.Limit

	var messages []models.Message
	err = query.Limit(req.Limit).Offset(offset).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	hasNext := int64(req.Page*req.Limit) < total

	response := &models.MessagesResponse{
		Messages: messages,
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
		HasNext:  hasNext,
	}

	s.logger.Info("💬 Chat Service: Messages searched successfully", 
		zap.String("room_id", roomID),
		zap.String("query", req.Query),
		zap.Int64("total", total))

	return response, nil
}

// === TYPING INDICATORS ===

// SetTypingIndicator sets user typing status
func (s *ChatService) SetTypingIndicator(ctx context.Context, roomID, userID string, isTyping bool) error {
	s.logger.Info("💬 Chat Service: Setting typing indicator", 
		zap.String("room_id", roomID),
		zap.String("user_id", userID),
		zap.Bool("is_typing", isTyping))

	if isTyping {
		// Create or update typing indicator
		typing := &models.TypingIndicator{
			RoomID:    roomID,
			UserID:    userID,
			IsActive:  true,
			StartedAt: time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Second), // Expire after 5 seconds
		}

		// Use upsert
		s.db.Where("room_id = ? AND user_id = ?", roomID, userID).
			Assign(typing).
			FirstOrCreate(typing)
	} else {
		// Remove typing indicator
		s.db.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&models.TypingIndicator{})
	}

	return nil
}

// GetTypingIndicators gets active typing indicators for room
func (s *ChatService) GetTypingIndicators(ctx context.Context, roomID string) ([]models.TypingIndicator, error) {
	s.logger.Info("💬 Chat Service: Getting typing indicators", zap.String("room_id", roomID))

	var typing []models.TypingIndicator
	err := s.db.Where("room_id = ? AND is_active = ? AND expires_at > ?", 
		roomID, true, time.Now()).
		Preload("User").
		Find(&typing).Error

	if err != nil {
		return nil, err
	}

	s.logger.Info("💬 Chat Service: Typing indicators retrieved", 
		zap.String("room_id", roomID),
		zap.Int("count", len(typing)))

	return typing, nil
}

// === HELPER METHODS ===

// validateCreateRoomRequest validates room creation request
func (s *ChatService) validateCreateRoomRequest(req *models.CreateRoomRequest) error {
	if req.Name == "" {
		return sharedErrors.NewValidationError("MISSING_ROOM_NAME", "Room name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return sharedErrors.NewValidationError("INVALID_ROOM_NAME", "Room name must be between 2 and 100 characters")
	}
	if req.MaxMembers <= 0 || req.MaxMembers > 1000 {
		return sharedErrors.NewValidationError("INVALID_MAX_MEMBERS", "Max members must be between 1 and 1000")
	}
	if len(req.Members) == 0 {
		return sharedErrors.NewValidationError("MISSING_MEMBERS", "At least one member is required")
	}
	if len(req.Members) >= req.MaxMembers {
		return sharedErrors.NewValidationError("TOO_MANY_MEMBERS", "Members cannot exceed max members")
	}
	return nil
}