package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Room represents a live streaming room
type Room struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	HostID      string    `gorm:"not null;index" json:"host_id"`
	IsActive    bool      `gorm:"default:false" json:"is_active"`
	IsPublic    bool      `gorm:"default:true" json:"is_public"`
	MaxParticipants int   `gorm:"default:100" json:"max_participants"`
	CurrentParticipants int `gorm:"default:0" json:"current_participants"`
	LiveKitRoomID string  `gorm:"uniqueIndex" json:"livekit_room_id"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// RoomParticipant represents a participant in a live room
type RoomParticipant struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	RoomID       string    `gorm:"not null;index" json:"room_id"`
	UserID       string    `gorm:"not null;index" json:"user_id"`
	Identity     string    `gorm:"not null;uniqueIndex" json:"identity"`
	IsHost       bool      `gorm:"default:false" json:"is_host"`
	HasAudio     bool      `gorm:"default:false" json:"has_audio"`
	HasVideo     bool      `gorm:"default:false" json:"has_video"`
	HasScreen    bool      `gorm:"default:false" json:"has_screen"`
	JoinedAt     time.Time `json:"joined_at"`
	LeftAt       *time.Time `json:"left_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Relationships
	Room Room `gorm:"foreignKey:RoomID" json:"room,omitempty"`
}

// RoomConfig represents configuration for creating a room
type RoomConfig struct {
	Name            string `json:"name" binding:"required,min=3,max=100"`
	Title           string `json:"title" binding:"required,min=1,max=200"`
	Description     string `json:"description" binding:"max=1000"`
	IsPublic        bool   `json:"is_public"`
	MaxParticipants int    `json:"max_participants" binding:"min=1,max=1000"`
}

// JoinRoomRequest represents request to join a room
type JoinRoomRequest struct {
	RoomID    string `json:"room_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
	Username  string `json:"username" binding:"required,min=2,max=50"`
	HasAudio  bool   `json:"has_audio"`
	HasVideo  bool   `json:"has_video"`
}

// LeaveRoomRequest represents request to leave a room
type LeaveRoomRequest struct {
	RoomID   string `json:"room_id" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
	Identity string `json:"identity" binding:"required"`
}

// TokenRequest represents request for LiveKit token
type TokenRequest struct {
	RoomID    string `json:"room_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
	Username  string `json:"username" binding:"required,min=2,max=50"`
	HasAudio  bool   `json:"has_audio"`
	HasVideo  bool   `json:"has_video"`
	IsHost    bool   `json:"is_host"`
}

// WebhookEvent represents LiveKit webhook event
type WebhookEvent struct {
	Event   string      `json:"event"`
	Room    RoomInfo    `json:"room"`
	Participant ParticipantInfo `json:"participant"`
}

// RoomInfo represents room information from webhook
type RoomInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Empty  bool   `json:"empty"`
}

// ParticipantInfo represents participant information from webhook
type ParticipantInfo struct {
	Identity     string `json:"identity"`
	State        string `json:"state"`
	Tracks       []TrackInfo `json:"tracks"`
	JoinedAt     time.Time `json:"joined_at"`
}

// TrackInfo represents track information from webhook
type TrackInfo struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Source   string `json:"source"`
	Muted    bool   `json:"muted"`
}

// TableName returns the table name for Room model
func (Room) TableName() string {
	return "livekit_rooms"
}

// TableName returns the table name for RoomParticipant model
func (RoomParticipant) TableName() string {
	return "livekit_room_participants"
}

// BeforeCreate generates UUID for Room
func (r *Room) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate generates UUID for RoomParticipant
func (rp *RoomParticipant) BeforeCreate(tx *gorm.DB) error {
	if rp.ID == "" {
		rp.ID = uuid.New().String()
	}
	return nil
}

// IsActiveRoom checks if room is currently active
func (r *Room) IsActiveRoom() bool {
	if !r.IsActive {
		return false
	}
	
	now := time.Now()
	
	// Check if room has started
	if r.StartTime != nil && now.Before(*r.StartTime) {
		return false
	}
	
	// Check if room has ended
	if r.EndTime != nil && now.After(*r.EndTime) {
		return false
	}
	
	return true
}

// CanJoin checks if a user can join the room
func (r *Room) CanJoin() bool {
	if !r.IsActiveRoom() {
		return false
	}
	
	if r.CurrentParticipants >= r.MaxParticipants {
		return false
	}
	
	return true
}