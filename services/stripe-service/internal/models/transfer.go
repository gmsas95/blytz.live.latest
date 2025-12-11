package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Transfer represents a Stripe Transfer
type Transfer struct {
	ID                   string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	StripeTransferID     string         `gorm:"not null;uniqueIndex" json:"stripe_transfer_id"`
	ConnectedAccountID   string         `gorm:"not null;index" json:"connected_account_id"`
	Amount               int64          `gorm:"not null" json:"amount"` // Amount in cents
	Currency             string         `gorm:"not null" json:"currency"`
	Status               string         `gorm:"not null" json:"status"`
	DestinationPaymentID  string         `gorm:"index" json:"destination_payment_id,omitempty"`
	Metadata             string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	DeletedAt            *time.Time      `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate sets up default values for Transfer
func (t *Transfer) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}