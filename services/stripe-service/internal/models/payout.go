package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Payout represents a Stripe Payout
type Payout struct {
	ID                string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	StripePayoutID   string         `gorm:"not null;uniqueIndex" json:"stripe_payout_id"`
	ConnectedAccountID string         `gorm:"not null;index" json:"connected_account_id"`
	Amount            int64          `gorm:"not null" json:"amount"` // Amount in cents
	Currency          string         `gorm:"not null" json:"currency"`
	Status            string         `gorm:"not null" json:"status"`
	ArrivalDate       *time.Time     `json:"arrival_date,omitempty"`
	Metadata          string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         *time.Time      `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate sets up default values for Payout
func (p *Payout) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}