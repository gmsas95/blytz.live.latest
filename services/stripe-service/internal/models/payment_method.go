package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentMethod represents a payment method in the system
type PaymentMethod struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Type           string    `gorm:"not null;index" json:"type"` // card, custom, mbway, twint, crypto, etc.
	CustomerID     string    `gorm:"not null;index" json:"customer_id"`
	StripeID       string    `gorm:"uniqueIndex" json:"stripe_id"`
	Provider       string    `gorm:"index" json:"provider"` // stripe, paypal, etc.
	IsReusable     bool      `gorm:"default:false" json:"is_reusable"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	Metadata       JSON      `gorm:"type:jsonb" json:"metadata"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Customer Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

// CustomPaymentMethod represents a custom payment method (PayPal, Apple Pay, etc.)
type CustomPaymentMethod struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PaymentMethodID string `gorm:"not null;uniqueIndex" json:"payment_method_id"`
	Provider     string    `gorm:"not null" json:"provider"` // paypal, apple_pay, google_pay
	ProviderID   string    `gorm:"not null;index" json:"provider_id"`
	Email        string    `gorm:"index" json:"email"`
	Phone        string    `json:"phone"`
	DeviceData   JSON      `gorm:"type:jsonb" json:"device_data"`
	VerificationData JSON  `gorm:"type:jsonb" json:"verification_data"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	PaymentMethod PaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"payment_method,omitempty"`
}

// MbWayPaymentMethod represents an MbWay payment method
type MbWayPaymentMethod struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PaymentMethodID string `gorm:"not null;uniqueIndex" json:"payment_method_id"`
	PhoneNumber  string    `gorm:"not null" json:"phone_number"`
	CountryCode  string    `gorm:"not null" json:"country_code"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	VerificationCode string `json:"verification_code"`
	VerificationExpiresAt *time.Time `json:"verification_expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	PaymentMethod PaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"payment_method,omitempty"`
}

// TWINTPaymentMethod represents a TWINT payment method
type TWINTPaymentMethod struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PaymentMethodID string `gorm:"not null;uniqueIndex" json:"payment_method_id"`
	QRCode       string    `gorm:"uniqueIndex" json:"qr_code"`
	DeviceID     string    `gorm:"index" json:"device_id"`
	MerchantID   string    `gorm:"index" json:"merchant_id"`
	CertificateData JSON  `gorm:"type:jsonb" json:"certificate_data"`
	IsPaired     bool      `gorm:"default:false" json:"is_paired"`
	PairedAt     *time.Time `json:"paired_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	PaymentMethod PaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"payment_method,omitempty"`
}

// CryptoPaymentMethod represents a cryptocurrency payment method
type CryptoPaymentMethod struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PaymentMethodID string `gorm:"not null;uniqueIndex" json:"payment_method_id"`
	CryptoType   string    `gorm:"not null;index" json:"crypto_type"` // bitcoin, ethereum, usdc, etc.
	WalletAddress string   `gorm:"not null;index" json:"wallet_address"`
	Network      string    `gorm:"not null" json:"network"` // mainnet, polygon, arbitrum, etc.
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	VerificationSignature string `json:"verification_signature"`
	VerificationData JSON   `gorm:"type:jsonb" json:"verification_data"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	PaymentMethod PaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"payment_method,omitempty"`
}

// Customer represents a customer in the system
type Customer struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	StripeID  string    `gorm:"uniqueIndex" json:"stripe_id"`
	Metadata  JSON      `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// JSON is a custom type for handling JSON data in GORM
type JSON map[string]interface{}

// BeforeCreate hook for PaymentMethod
func (pm *PaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == "" {
		pm.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate hook for CustomPaymentMethod
func (cpm *CustomPaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if cpm.ID == "" {
		cpm.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate hook for MbWayPaymentMethod
func (mpm *MbWayPaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if mpm.ID == "" {
		mpm.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate hook for TWINTPaymentMethod
func (tpm *TWINTPaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if tpm.ID == "" {
		tpm.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate hook for CryptoPaymentMethod
func (cpm *CryptoPaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if cpm.ID == "" {
		cpm.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate hook for Customer
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// PaymentMethodStatus represents the status of a payment method
type PaymentMethodStatus string

const (
	PaymentMethodStatusActive   PaymentMethodStatus = "active"
	PaymentMethodStatusInactive PaymentMethodStatus = "inactive"
	PaymentMethodStatusExpired  PaymentMethodStatus = "expired"
	PaymentMethodStatusFailed   PaymentMethodStatus = "failed"
)

// CryptoNetwork represents the cryptocurrency network
type CryptoNetwork string

const (
	CryptoNetworkMainnet  CryptoNetwork = "mainnet"
	CryptoNetworkPolygon  CryptoNetwork = "polygon"
	CryptoNetworkArbitrum CryptoNetwork = "arbitrum"
	CryptoNetworkOptimism CryptoNetwork = "optimism"
	CryptoNetworkBSC      CryptoNetwork = "bsc"
	CryptoNetworkAvalanche CryptoNetwork = "avalanche"
)

// CryptoType represents the type of cryptocurrency
type CryptoType string

const (
	CryptoTypeBitcoin  CryptoType = "bitcoin"
	CryptoTypeEthereum CryptoType = "ethereum"
	CryptoTypeUSDC     CryptoType = "usdc"
	CryptoTypeUSDT     CryptoType = "usdt"
	CryptoTypeDAI      CryptoType = "dai"
	CryptoTypeWBTC     CryptoType = "wbtc"
)

// CustomPaymentProvider represents the custom payment provider
type CustomPaymentProvider string

const (
	CustomPaymentProviderPayPal    CustomPaymentProvider = "paypal"
	CustomPaymentProviderApplePay  CustomPaymentProvider = "apple_pay"
	CustomPaymentProviderGooglePay CustomPaymentProvider = "google_pay"
	CustomPaymentProviderKlarna    CustomPaymentProvider = "klarna"
	CustomPaymentProviderAfterpay  CustomPaymentProvider = "afterpay"
)
