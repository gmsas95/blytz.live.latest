package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StripeEvent represents a Stripe webhook event
type StripeEvent struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	StripeEventID string    `gorm:"not null;uniqueIndex" json:"stripe_event_id"`
	EventType   string    `gorm:"not null;index" json:"event_type"`
	Processed  bool      `gorm:"default:false;index" json:"processed"`
	EventData   string    `gorm:"type:jsonb;not null" json:"event_data"`
	ErrorMessage string   `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate sets up default values for StripeEvent
func (s *StripeEvent) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}