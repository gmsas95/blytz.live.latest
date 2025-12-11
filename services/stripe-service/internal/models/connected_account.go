package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConnectedAccount represents a Stripe Connect account
type ConnectedAccount struct {
	ID                string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID            string         `gorm:"not null;index" json:"user_id"`
	StripeAccountID    string         `gorm:"not null;uniqueIndex" json:"stripe_account_id"`
	AccountType        string         `gorm:"not null" json:"account_type"`
	Country            string         `gorm:"not null" json:"country"`
	Currency           string         `gorm:"not null" json:"currency"`
	VerificationStatus string         `gorm:"default:pending" json:"verification_status"`
	ChargesEnabled    bool           `gorm:"default:false" json:"charges_enabled"`
	PayoutsEnabled    bool           `gorm:"default:false" json:"payouts_enabled"`
	Metadata           string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          *time.Time      `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate sets the ID as a UUID if not set
func (c *ConnectedAccount) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}