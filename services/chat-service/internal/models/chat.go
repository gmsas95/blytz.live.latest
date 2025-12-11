package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Message Status Constants
const (
	MessageStatusSending   = "sending"
	MessageStatusSent      = "sent"
	MessageStatusDelivered = "delivered"
	MessageStatusRead      = "read"
	MessageStatusFailed    = "failed"
)

// Message Type Constants
const (
	MessageTypeText      = "text"
	MessageTypeImage     = "image"
	MessageTypeVideo     = "video"
	MessageTypeAudio     = "audio"
	MessageTypeFile      = "file"
	MessageTypeSystem    = "system"
	MessageTypingStart   = "typing_start"
	MessageTypingStop    = "typing_stop"
	MessageUserJoin     = "user_join"
	MessageUserLeave    = "user_leave"
	MessageRoomCreated  = "room_created"
	MessageRoomDeleted  = "room_deleted"
)

// Room Type Constants
const (
	RoomTypeDirect     = "direct"     // 1-to-1 chat
	RoomTypeGroup      = "group"      // Group chat
	RoomTypeChannel    = "channel"    // Public channel
	RoomTypeSupport    = "support"    // Customer support
	RoomTypeAuction    = "auction"    // Auction room
	RoomTypeProduct    = "product"    // Product discussion
)

// Room Status Constants
const (
	RoomStatusActive   = "active"
	RoomStatusArchived = "archived"
	RoomStatusDeleted  = "deleted"
)

// User Role Constants
const (
	UserRoleOwner     = "owner"
	UserRoleAdmin     = "admin"
	UserRoleModerator = "moderator"
	UserRoleMember    = "member"
	UserRoleGuest     = "guest"
)

// Enhanced Message Model
type Message struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	RoomID       string         `json:"room_id" gorm:"not null;index"`
	UserID       string         `json:"user_id" gorm:"not null;index"`
	Content      string         `json:"content" gorm:"type:text"`
	Type         string         `json:"type" gorm:"not null;default:text"`
	Status       string         `json:"status" gorm:"not null;default:sent"`
	ReplyToID   *string        `json:"reply_to_id" gorm:"index"` // Thread support
	EditCount   int            `json:"edit_count" gorm:"default:0"`
	IsEdited     bool           `json:"is_edited" gorm:"default:false"`
	IsPinned     bool           `json:"is_pinned" gorm:"default:false"`
	IsDeleted    bool           `json:"is_deleted" gorm:"default:false"`
	Timestamp    time.Time      `json:"timestamp" gorm:"not null"`
	EditedAt     *time.Time     `json:"edited_at"`
	DeletedAt    *time.Time     `json:"deleted_at"`
	IPAddress    string         `json:"ip_address" gorm:""`
	UserAgent    string         `json:"user_agent" gorm:""`
	Metadata     string         `json:"metadata" gorm:"type:text"` // JSON data
	Reactions    string         `json:"reactions" gorm:"type:text"` // JSON array
	Attachments  string         `json:"attachments" gorm:"type:text"` // JSON array
	Mentions     string         `json:"mentions" gorm:"type:text"` // JSON array
	Hashtags     string         `json:"hashtags" gorm:"type:text"` // JSON array
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	
	// Relationships
	ReplyTo     *Message       `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID"`
	Threads     []Message      `json:"threads,omitempty" gorm:"foreignKey:ReplyToID"`
	Sender      *User          `json:"sender,omitempty" gorm:"foreignKey:UserID"`
	ReactionLog []MessageReaction `json:"reactions_log,omitempty" gorm:"foreignKey:MessageID"`
}

// Enhanced Chat Room Model
type ChatRoom struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	Type        string         `json:"type" gorm:"not null;default:direct"`
	Status      string         `json:"status" gorm:"not null;default:active"`
	IsPrivate   bool           `json:"is_private" gorm:"default:false"`
	IsMuted     bool           `json:"is_muted" gorm:"default:false"`
	IsPinned    bool           `json:"is_pinned" gorm:"default:false"`
	MaxMembers  int            `json:"max_members" gorm:"default:100"`
	CurrentMembers int          `json:"current_members" gorm:"default:0"`
	LastMessage *string        `json:"last_message_id" gorm:"index"`
	LastActivity time.Time      `json:"last_activity"`
	Avatar      string         `json:"avatar"`
	CoverImage  string         `json:"cover_image"`
	Category    string         `json:"category"`
	Tags        string         `json:"tags" gorm:"type:text"` // JSON array
	Settings    string         `json:"settings" gorm:"type:text"` // JSON config
	Metadata    string         `json:"metadata" gorm:"type:text"` // Additional data
	CreatedBy   string         `json:"created_by" gorm:"not null"`
	ArchivedAt  *time.Time     `json:"archived_at"`
	DeletedAt   *time.Time     `json:"deleted_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	
	// Relationships
	Members     []RoomMember   `json:"members,omitempty" gorm:"foreignKey:RoomID"`
	Messages    []Message      `json:"messages,omitempty" gorm:"foreignKey:RoomID"`
	PinnedMessages []Message   `json:"pinned_messages,omitempty" gorm:"foreignKey:RoomID"`
	Creator     *User          `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// Room Member Model
type RoomMember struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	RoomID     string    `json:"room_id" gorm:"not null;index"`
	UserID     string    `json:"user_id" gorm:"not null;index"`
	Role       string    `json:"role" gorm:"not null;default:member"`
	JoinedAt   time.Time `json:"joined_at" gorm:"not null"`
	LeftAt     *time.Time `json:"left_at"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	IsMuted    bool      `json:"is_muted" gorm:"default:false"`
	IsBanned   bool      `json:"is_banned" gorm:"default:false"`
	LastRead   *string   `json:"last_read_id" gorm:"index"` // Last read message ID
	NotificationSound string `json:"notification_sound" gorm:"default:true"`
	NotificationDesktop bool `json:"notification_desktop" gorm:"default:true"`
	NotificationMobile  bool `json:"notification_mobile" gorm:"default:true"`
	CustomNickname string `json:"custom_nickna"`
	CustomAvatar   string `json:"custom_avatar"`
	Permissions    string `json:"permissions" gorm:"type:text"` // JSON permissions
	Settings       string `json:"settings" gorm:"type:text"` // JSON user settings
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	
	// Relationships
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Room       *ChatRoom `json:"room,omitempty" gorm:"foreignKey:RoomID"`
}

// Message Reaction Model
type MessageReaction struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MessageID string    `json:"message_id" gorm:"not null;index"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	Emoji     string    `json:"emoji" gorm:"not null"`
	Reaction  string    `json:"reaction" gorm:"not null"` // like, love, laugh, wow, sad, angry
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relationships
	Message    *Message `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	User       *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Typing Indicator Model
type TypingIndicator struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	RoomID    string    `json:"room_id" gorm:"not null;index"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	StartedAt time.Time `json:"started_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	
	// Relationships
	User      *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Room      *ChatRoom `json:"room,omitempty" gorm:"foreignKey:RoomID"`
}

// User Presence Model
type UserPresence struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"not null;index"`
	Status      string    `json:"status" gorm:"not null;default:offline"` // online, idle, busy, offline
	IsOnline    bool      `json:"is_online" gorm:"default:false"`
	LastActive  time.Time `json:"last_active" gorm:"not null"`
	CurrentRoom *string   `json:"current_room_id" gorm:"index"` // Currently active room
	Device      string    `json:"device" gorm:""` // web, mobile, desktop
	IPAddress   string    `json:"ip_address" gorm:""`
	UserAgent   string    `json:"user_agent" gorm:""`
	CustomStatus string   `json:"custom_status" gorm:""`
	Emoji       string    `json:"emoji" gorm:""`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Relationships
	User       *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ActiveRoom *ChatRoom `json:"active_room,omitempty" gorm:"foreignKey:CurrentRoom"`
}

// Chat Settings Model
type ChatSettings struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       string    `json:"user_id" gorm:"not null;index"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	AllowDirect bool      `json:"allow_direct" gorm:"default:true"`
	AllowGroup   bool      `json:"allow_group" gorm:"default:true"`
	SoundEnabled bool      `json:"sound_enabled" gorm:"default:true"`
	DesktopNotifications bool `json:"desktop_notifications" gorm:"default:true"`
	MobileNotifications  bool `json:"mobile_notifications" gorm:"default:true"`
	EmailNotifications   bool `json:"email_notifications" gorm:"default:false"`
	MessagePreview      bool `json:"message_preview" gorm:"default:true"`
	TypingIndicators   bool `json:"typing_indicators" gorm:"default:true"`
	ReadReceipts       bool `json:"read_receipts" gorm:"default:true"`
	OnlineStatus       bool `json:"online_status" gorm:"default:true"`
	LastSeenPrivacy    bool `json:"last_seen_privacy" gorm:"default:false"`
	MaxUploadSize      int64 `json:"max_upload_size" gorm:"default:10485760"` // 10MB
	SupportedFormats   string `json:"supported_formats" gorm:"type:text"` // JSON array
	Theme              string `json:"theme" gorm:"default:light"`
	Language           string `json:"language" gorm:"default:en"`
	Timezone           string `json:"timezone" gorm:"default:UTC"`
	PrivacySettings    string `json:"privacy_settings" gorm:"type:text"` // JSON config
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	
	// Relationships
	User               *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// User Model for chat service
type User struct {
	UserID      string    `json:"user_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username    string    `json:"username" gorm:"not null;uniqueIndex"`
	Email       string    `json:"email" gorm:"not null;uniqueIndex"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Avatar      string    `json:"avatar"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio" gorm:"type:text"`
	IsOnline    bool      `json:"is_online" gorm:"default:false"`
	IsVerified  bool      `json:"is_verified" gorm:"default:false"`
	IsPremium   bool      `json:"is_premium" gorm:"default:false"`
	IsBanned    bool      `json:"is_banned" gorm:"default:false"`
	LastActive  time.Time `json:"last_active"`
	Preferences  string    `json:"preferences" gorm:"type:text"` // JSON settings
	Metadata    string    `json:"metadata" gorm:"type:text"` // Additional data
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Request Models

// SendMessageRequest represents send message request
type SendMessageRequest struct {
	Content      string   `json:"content" binding:"required,max=4000"`
	Type         string   `json:"type" binding:"omitempty,oneof=text image video audio file"`
	ReplyToID    *string  `json:"reply_to_id" binding:"omitempty,uuid"`
	Attachments  []string `json:"attachments" binding:"omitempty,dive,url"`
	Mentions     []string `json:"mentions" binding:"omitempty,dive,uuid"`
	IsTemporary  bool     `json:"is_temporary" gorm:"default:false"`
	TTL          int      `json:"ttl" binding:"omitempty,min=1,max=86400"` // Time to live in seconds
}

// CreateRoomRequest represents create room request
type CreateRoomRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=100"`
	Description string   `json:"description" binding:"max=500"`
	Type        string   `json:"type" binding:"required,oneof=direct group channel support auction product"`
	IsPrivate   bool     `json:"is_private"`
	MaxMembers  int      `json:"max_members" binding:"omitempty,min=2,max=1000"`
	Avatar      string   `json:"avatar" binding:"omitempty,url"`
	CoverImage  string   `json:"cover_image" binding:"omitempty,url"`
	Category    string   `json:"category" binding:"max=50"`
	Tags        []string `json:"tags" binding:"omitempty,dive,alphanum"`
	Members     []string `json:"members" binding:"required,min=1,dive,uuid"`
	Settings    string   `json:"settings"`
	Metadata    string   `json:"metadata"`
}

// UpdateRoomRequest represents update room request
type UpdateRoomRequest struct {
	Name        string   `json:"name" binding:"omitempty,min=2,max=100"`
	Description string   `json:"description" binding:"omitempty,max=500"`
	Avatar      string   `json:"avatar" binding:"omitempty,url"`
	CoverImage  string   `json:"cover_image" binding:"omitempty,url"`
	Category    string   `json:"category" binding:"omitempty,max=50"`
	Tags        []string `json:"tags" binding:"omitempty,dive,alphanum"`
	MaxMembers  int      `json:"max_members" binding:"omitempty,min=2,max=1000"`
	IsPrivate   bool     `json:"is_private"`
	Settings    string   `json:"settings"`
	Metadata    string   `json:"metadata"`
}

// AddMemberRequest represents add member request
type AddMemberRequest struct {
	UserIDs    []string `json:"user_ids" binding:"required,min=1,dive,uuid"`
	Role       string   `json:"role" binding:"omitempty,oneof=owner admin moderator member guest"`
	IsActive   bool     `json:"is_active" gorm:"default:true"`
	Permissions string   `json:"permissions"`
}

// UpdateMemberRequest represents update member request
type UpdateMemberRequest struct {
	Role                 string `json:"role" binding:"omitempty,oneof=owner admin moderator member guest"`
	IsMuted              bool   `json:"is_muted"`
	IsBanned             bool   `json:"is_banned"`
	NotificationSound    bool   `json:"notification_sound"`
	NotificationDesktop   bool   `json:"notification_desktop"`
	NotificationMobile    bool   `json:"notification_mobile"`
	CustomNickname       string `json:"custom_nickna" binding:"max=50"`
	CustomAvatar         string `json:"custom_avatar" binding:"omitempty,url"`
	Permissions          string `json:"permissions"`
}

// SearchMessagesRequest represents search messages request
type SearchMessagesRequest struct {
	Query       string `json:"query" form:"query"`
	RoomID      string `json:"room_id" form:"room_id"`
	UserID      string `json:"user_id" form:"user_id"`
	MessageType string `json:"message_type" form:"message_type"`
	HasMedia    bool   `json:"has_media" form:"has_media"`
	StartDate   string `json:"start_date" form:"start_date"`
	EndDate     string `json:"end_date" form:"end_date"`
	SortBy      string `json:"sort_by" form:"sort_by"` // timestamp, popularity
	SortOrder   string `json:"sort_order" form:"sort_order"` // asc, desc
	Page        int    `json:"page" form:"page"`
	Limit       int    `json:"limit" form:"limit"`
}

// SearchRoomsRequest represents search rooms request
type SearchRoomsRequest struct {
	Query     string `json:"query" form:"query"`
	Type      string `json:"type" form:"type"`
	Category  string `json:"category" form:"category"`
	IsPublic  bool   `json:"is_public" form:"is_public"`
	IsJoined  bool   `json:"is_joined" form:"is_joined"`
	SortBy    string `json:"sort_by" form:"sort_by"` // name, members, activity
	SortOrder string `json:"sort_order" form:"sort_order"` // asc, desc
	Page      int    `json:"page" form:"page"`
	Limit     int    `json:"limit" form:"limit"`
}

// Response Models

// MessageResponse represents message response
type MessageResponse struct {
	Message       Message          `json:"message"`
	Sender        *User           `json:"sender,omitempty"`
	ReplyTo       *Message        `json:"reply_to,omitempty"`
	ThreadCount   int             `json:"thread_count"`
	Reactions     []MessageReaction `json:"reactions,omitempty"`
	IsRead        bool            `json:"is_read"`
	DeliveryStatus string         `json:"delivery_status"`
	TimeFormatted string          `json:"time_formatted"`
}

// RoomResponse represents room response
type RoomResponse struct {
	Room          ChatRoom       `json:"room"`
	Creator       *User          `json:"creator,omitempty"`
	Members       []RoomMember   `json:"members,omitempty"`
	MessageCount  int            `json:"message_count"`
	UnreadCount   int            `json:"unread_count"`
	LastMessage   *Message       `json:"last_message,omitempty"`
	IsJoined      bool           `json:"is_joined"`
	UserRole      string         `json:"user_role"`
	IsActive      bool           `json:"is_active"`
}

// MessagesResponse represents messages list response
type MessagesResponse struct {
	Messages    []Message `json:"messages"`
	Total       int64     `json:"total"`
	Page        int       `json:"page"`
	Limit       int       `json:"limit"`
	HasNext     bool      `json:"has_next"`
	HasPrevious bool      `json:"has_previous"`
}

// RoomsResponse represents rooms list response
type RoomsResponse struct {
	Rooms      []ChatRoom `json:"rooms"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
	HasNext    bool       `json:"has_next"`
}

// WebSocket Messages

// WebSocketMessage represents WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"` // message, room_update, user_status, typing, etc.
	RoomID    string      `json:"room_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	MessageID string      `json:"message_id,omitempty"`
}

// ChatUpdateMessage represents chat update for WebSocket
type ChatUpdateMessage struct {
	Type      string      `json:"type"` // new_message, message_update, message_delete, typing, user_join, user_leave
	RoomID    string      `json:"room_id"`
	UserID    string      `json:"user_id,omitempty"`
	Message   *Message    `json:"message,omitempty"`
	User      *User       `json:"user,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// TypingMessage represents typing indicator for WebSocket
type TypingMessage struct {
	Type      string    `json:"type"` // typing_start, typing_stop
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	User      *User     `json:"user,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UserStatusMessage represents user status for WebSocket
type UserStatusMessage struct {
	Type       string    `json:"type"` // user_online, user_offline, user_status
	UserID     string    `json:"user_id"`
	User       *User     `json:"user,omitempty"`
	Status     string    `json:"status"`
	IsOnline   bool      `json:"is_online"`
	LastActive time.Time `json:"last_active"`
	Timestamp  time.Time `json:"timestamp"`
}

// Helper Methods

// Note: GORM hooks removed for simplified implementation

// SetAttachmentsArray sets attachments from string array
func (m *Message) SetAttachmentsArray(attachments []string) {
	if len(attachments) == 0 {
		m.Attachments = ""
		return
	}
	data, _ := json.Marshal(attachments)
	m.Attachments = string(data)
}

// GetAttachmentsArray returns attachments as string array
func (m *Message) GetAttachmentsArray() []string {
	if m.Attachments == "" {
		return []string{}
	}
	var attachments []string
	json.Unmarshal([]byte(m.Attachments), &attachments)
	return attachments
}

// SetMentionsArray sets mentions from string array
func (m *Message) SetMentionsArray(mentions []string) {
	if len(mentions) == 0 {
		m.Mentions = ""
		return
	}
	data, _ := json.Marshal(mentions)
	m.Mentions = string(data)
}

// GetMentionsArray returns mentions as string array
func (m *Message) GetMentionsArray() []string {
	if m.Mentions == "" {
		return []string{}
	}
	var mentions []string
	json.Unmarshal([]byte(m.Mentions), &mentions)
	return mentions
}

// IsEdited checks if message has been edited
func (m *Message) IsEditedMessage() bool {
	return m.IsEdited && m.EditCount > 0
}

// GetTimeFormatted returns formatted time string
func (m *Message) GetTimeFormatted() string {
	return formatTime(m.Timestamp)
}

// GetEditedTimeFormatted returns formatted edited time string
func (m *Message) GetEditedTimeFormatted() string {
	if m.EditedAt != nil {
		return formatTime(*m.EditedAt)
	}
	return ""
}

// SetTagsArray sets tags from string array
func (r *ChatRoom) SetTagsArray(tags []string) {
	if len(tags) == 0 {
		r.Tags = ""
		return
	}
	data, _ := json.Marshal(tags)
	r.Tags = string(data)
}

// GetTagsArray returns tags as string array
func (r *ChatRoom) GetTagsArray() []string {
	if r.Tags == "" {
		return []string{}
	}
	var tags []string
	json.Unmarshal([]byte(r.Tags), &tags)
	return tags
}

// GetTimeFormatted returns formatted time string
func (r *ChatRoom) GetTimeFormatted() string {
	return formatTime(r.CreatedAt)
}

// GetLastActivityFormatted returns formatted last activity time
func (r *ChatRoom) GetLastActivityFormatted() string {
	return formatTime(r.LastActivity)
}

// Helper function to generate UUID (simplified)
func generateUUID() string {
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
}

// Helper function to format time
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// Helper function to format relative time
func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	
	now := time.Now()
	diff := now.Sub(t)
	
	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	} else {
		return t.Format("Jan 2, 2006")
	}
}