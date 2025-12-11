package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentIntent represents a Stripe Payment Intent
type PaymentIntent struct {
	ID                    string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	StripePaymentIntentID string         `gorm:"not null;uniqueIndex" json:"stripe_payment_intent_id"`
	Amount                int64          `gorm:"not null" json:"amount"` // Amount in cents
	Currency              string         `gorm:"not null" json:"currency"`
	Status                string         `gorm:"not null" json:"status"`
	UserID                string         `gorm:"not null;index" json:"user_id"`
	AuctionID             *string        `gorm:"index" json:"auction_id,omitempty"`
	ConnectedAccountID    *string        `gorm:"index" json:"connected_account_id,omitempty"`
	ApplicationFeeAmount  int64          `gorm:"default:0" json:"application_fee_amount"` // Fee amount in cents
	PaymentMethodTypes    string         `gorm:"type:text" json:"payment_method_types"` // JSON array of payment method types
	Metadata              string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	DeletedAt             *time.Time      `gorm:"index" json:"deleted_at,omitempty"`
}

// PaymentMethodType represents supported payment method types
type PaymentMethodType string

const (
	// Traditional payment methods
	PaymentMethodTypeCard         PaymentMethodType = "card"
	PaymentMethodTypeAlipay       PaymentMethodType = "alipay"
	PaymentMethodTypeSEPA        PaymentMethodType = "sepa_debit"
	PaymentMethodTypeIdeal        PaymentMethodType = "ideal"
	PaymentMethodTypeSofort      PaymentMethodType = "sofort"

	// New payment methods from stripe-go v84.0.0
	PaymentMethodTypeCustom      PaymentMethodType = "custom"
	PaymentMethodTypeMbWay       PaymentMethodType = "mbway"
	PaymentMethodTypeTWINT       PaymentMethodType = "twint"
	PaymentMethodTypeCrypto      PaymentMethodType = "crypto"
)

// PaymentMethodConfig represents configuration for specific payment methods
type PaymentMethodConfig struct {
	Type              PaymentMethodType       `json:"type"`
	Enabled           bool                    `json:"enabled"`
	MinimumAmount     int64                   `json:"minimum_amount"`     // in cents
	MaximumAmount     int64                   `json:"maximum_amount"`     // in cents
	SupportedCurrencies []string                `json:"supported_currencies"`
	RequiresVerification bool                    `json:"requires_verification"`
	AdditionalData    map[string]interface{}   `json:"additional_data,omitempty"`
}

// CustomPaymentMethodData represents data for custom payment methods
type CustomPaymentMethodData struct {
	PaymentMethodType string                 `json:"payment_method_type"`
	CustomType       string                 `json:"custom_type"`
	Provider         string                 `json:"provider"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// MbWayPaymentMethodData represents data for MbWay payment method
type MbWayPaymentMethodData struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	CountryCode string `json:"country_code" binding:"required"`
}

// TWINTPaymentMethodData represents data for TWINT payment method
type TWINTPaymentMethodData struct {
	QRCode     string `json:"qr_code,omitempty"`
	DeviceID    string `json:"device_id,omitempty"`
	MerchantID  string `json:"merchant_id,omitempty"`
}

// CryptoPaymentMethodData represents data for Crypto payment method
type CryptoPaymentMethodData struct {
	CryptoType      string `json:"crypto_type" binding:"required"` // e.g., "bitcoin", "ethereum", "usdc"
	WalletAddress   string `json:"wallet_address" binding:"required"`
	Network         string `json:"network,omitempty"` // e.g., "mainnet", "polygon", "arbitrum"
	TransactionHash string `json:"transaction_hash,omitempty"`
}

// GetDefaultPaymentMethodConfigs returns default configurations for all payment methods
func GetDefaultPaymentMethodConfigs() map[PaymentMethodType]PaymentMethodConfig {
	return map[PaymentMethodType]PaymentMethodConfig{
		PaymentMethodTypeCard: {
			Type:              PaymentMethodTypeCard,
			Enabled:           true,
			MinimumAmount:     50,  // $0.50 in cents
			MaximumAmount:     999999, // $9,999.99 in cents
			SupportedCurrencies: []string{"usd", "eur", "gbp"},
			RequiresVerification: false,
		},
		PaymentMethodTypeAlipay: {
			Type:              PaymentMethodTypeAlipay,
			Enabled:           true,
			MinimumAmount:     50,
			MaximumAmount:     999999,
			SupportedCurrencies: []string{"usd", "eur", "cny", "hkd", "sgd"},
			RequiresVerification: false,
		},
		PaymentMethodTypeSEPA: {
			Type:              PaymentMethodTypeSEPA,
			Enabled:           true,
			MinimumAmount:     50,
			MaximumAmount:     999999,
			SupportedCurrencies: []string{"eur"},
			RequiresVerification: true,
		},
		PaymentMethodTypeIdeal: {
			Type:              PaymentMethodTypeIdeal,
			Enabled:           true,
			MinimumAmount:     50,
			MaximumAmount:     50000, // $500.00 in cents
			SupportedCurrencies: []string{"eur"},
			RequiresVerification: false,
		},
		PaymentMethodTypeSofort: {
			Type:              PaymentMethodTypeSofort,
			Enabled:           true,
			MinimumAmount:     50,
			MaximumAmount:     999999,
			SupportedCurrencies: []string{"eur"},
			RequiresVerification: true,
		},
		// New payment methods from stripe-go v84.0.0
		PaymentMethodTypeCustom: {
			Type:              PaymentMethodTypeCustom,
			Enabled:           true,
			MinimumAmount:     50,
			MaximumAmount:     999999,
			SupportedCurrencies: []string{"usd", "eur", "gbp"},
			RequiresVerification: true,
			AdditionalData: map[string]interface{}{
				"requires_provider_setup": true,
				"supported_providers":    []string{"paypal", "apple_pay", "google_pay"},
			},
		},
		PaymentMethodTypeMbWay: {
			Type:              PaymentMethodTypeMbWay,
			Enabled:           true,
			MinimumAmount:     100, // €1.00 in cents
			MaximumAmount:     100000, // €1,000.00 in cents
			SupportedCurrencies: []string{"eur"},
			RequiresVerification: false,
			AdditionalData: map[string]interface{}{
				"requires_phone_verification": true,
				"country_code":               "PT",
			},
		},
		PaymentMethodTypeTWINT: {
			Type:              PaymentMethodTypeTWINT,
			Enabled:           true,
			MinimumAmount:     50,
			MaximumAmount:     999999,
			SupportedCurrencies: []string{"chf"},
			RequiresVerification: false,
			AdditionalData: map[string]interface{}{
				"requires_qr_code": true,
				"country_code":     "CH",
			},
		},
		PaymentMethodTypeCrypto: {
			Type:              PaymentMethodTypeCrypto,
			Enabled:           true,
			MinimumAmount:     500, // $5.00 in cents
			MaximumAmount:     999999,
			SupportedCurrencies: []string{"usd", "eur"},
			RequiresVerification: true,
			AdditionalData: map[string]interface{}{
				"supported_crypto_types": []string{"bitcoin", "ethereum", "usdc"},
				"requires_wallet_verification": true,
			},
		},
	}
}

// BeforeCreate sets up default values for PaymentIntent
func (p *PaymentIntent) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	
	// Set default payment method types if not specified
	if p.PaymentMethodTypes == "" {
		p.PaymentMethodTypes = `["card"]`
	}
	
	return nil
}

// GetPaymentMethodTypes returns the payment method types as a slice
func (p *PaymentIntent) GetPaymentMethodTypes() []PaymentMethodType {
	// This would parse the JSON string and return a slice
	// For now, return default
	return []PaymentMethodType{PaymentMethodTypeCard}
}

// SetPaymentMethodTypes sets the payment method types from a slice
func (p *PaymentIntent) SetPaymentMethodTypes(types []PaymentMethodType) {
	// This would serialize the slice to JSON
	// For now, set a basic implementation
	if len(types) > 0 {
		p.PaymentMethodTypes = `["` + string(types[0]) + `"]`
	}
}

// SupportsPaymentMethod checks if a specific payment method is supported
func (p *PaymentIntent) SupportsPaymentMethod(methodType PaymentMethodType) bool {
	types := p.GetPaymentMethodTypes()
	for _, t := range types {
		if t == methodType {
			return true
		}
	}
	return false
}

// IsValidPaymentMethod checks if a payment method type is valid
func IsValidPaymentMethod(methodType string) bool {
	switch PaymentMethodType(methodType) {
	case PaymentMethodTypeCard, PaymentMethodTypeAlipay, PaymentMethodTypeSEPA,
		 PaymentMethodTypeIdeal, PaymentMethodTypeSofort, PaymentMethodTypeCustom,
		 PaymentMethodTypeMbWay, PaymentMethodTypeTWINT, PaymentMethodTypeCrypto:
		return true
	default:
		return false
	}
}

// GetPaymentMethodConfig returns configuration for a specific payment method
func GetPaymentMethodConfig(methodType PaymentMethodType) (PaymentMethodConfig, bool) {
	config, exists := GetDefaultPaymentMethodConfigs()[methodType]
	return config, exists
}