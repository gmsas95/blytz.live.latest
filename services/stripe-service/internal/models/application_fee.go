package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicationFee represents a Stripe Application Fee
type ApplicationFee struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	StripeFeeID    string    `gorm:"not null;uniqueIndex" json:"stripe_fee_id"`
	PaymentIntentID string    `gorm:"not null;index" json:"payment_intent_id"`
	Amount         int64     `gorm:"not null" json:"amount"` // Amount in cents
	Currency       string    `gorm:"not null" json:"currency"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate sets up default values for ApplicationFee
func (a *ApplicationFee) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}