package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Payment Status Constants
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusSuccess    = "success"
	PaymentStatusFailed     = "failed"
	PaymentStatusCancelled   = "cancelled"
	PaymentStatusRefunded   = "refunded"
	PaymentStatusExpired    = "expired"
)

// Payment Type Constants
const (
	PaymentTypeFPX         = "FPX"
	PaymentTypeCreditCard = "credit_card"
	PaymentTypeDebitCard  = "debit_card"
	PaymentTypeOnlineBank = "online_banking"
	PaymentTypeEWallet    = "ewallet"
	PaymentTypeBankTransfer = "bank_transfer"
	PaymentTypeQRPayment   = "qr_payment"
)

// Fiuu Payment Methods
const (
	FiuuMethodFPX      = "FPX"
	FiuuMethodMaybank  = "MAYBANK2U"
	FiuuMethodCIMB     = "CIMB_CLICKS"
	FiuuMethodPublicBank = "PUBLIBANK"
	FiuuMethodRHB      = "RHB_NOW"
	FiuuMethodHLB      = "HLB_CONNECT"
	FiuuMethodAmbank   = "AMBANK"
	FiuuMethodTNG      = "TNG"
	FiuuMethodGrabPay  = "GRABPAY"
	FiuuMethodBoost    = "BOOST"
	FiuuMethodShopeePay = "SHOPEEPAY"
	FiuuMethodDuitNow  = "DUITNOW"
	FiuuMethodVisa     = "VISA"
	FiuuMethodMastercard = "MASTERCARD"
	FiuuMethodAmex     = "AMEX"
)

// Enhanced Payment Model
type Payment struct {
	ID               string          `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID          string          `json:"order_id" gorm:"not null;index"`
	TransactionID    string          `json:"transaction_id" gorm:"index"`
	PaymentID        string          `json:"payment_id" gorm:"index"` // Fiuu payment ID
	Amount           float64         `json:"amount" gorm:"not null"`
	Currency         string          `json:"currency" gorm:"not null;default:MYR"`
	Description      string          `json:"description" gorm:"type:text"`
	Status           string          `json:"status" gorm:"not null;default:pending"`
	PaymentType      string          `json:"payment_type" gorm:"not null"`
	PaymentMethod    string          `json:"payment_method" gorm:"not null"`
	CustomerName     string          `json:"customer_name" gorm:"not null"`
	CustomerEmail    string          `json:"customer_email" gorm:"not null"`
	CustomerPhone    string          `json:"customer_phone" gorm:"not null"`
	PaymentURL       string          `json:"payment_url" gorm:""`
	QRCode          string          `json:"qr_code" gorm:"type:text"`
	ExpiresAt        time.Time       `json:"expires_at" gorm:"not null"`
	PaidAt           *time.Time      `json:"paid_at"`
	FailedAt         *time.Time      `json:"failed_at"`
	CancelledAt      *time.Time      `json:"cancelled_at"`
	RefundedAt       *time.Time      `json:"refunded_at"`
	FailureReason    string          `json:"failure_reason" gorm:""`
	RefundReason     string          `json:"refund_reason" gorm:""`
	RefundAmount     float64         `json:"refund_amount" gorm:"default:0"`
	Fees             float64         `json:"fees" gorm:"default:0"`
	NetAmount        float64         `json:"net_amount" gorm:"default:0"`
	WebhookEvents    string          `json:"webhook_events" gorm:"type:text"` // JSON array
	Metadata         string          `json:"metadata" gorm:"type:text"` // Additional data
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	
	// Relationships
	Order           *Order          `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Refunds         []PaymentRefund `json:"refunds,omitempty" gorm:"foreignKey:PaymentID"`
	WebhookLogs     []WebhookLog    `json:"webhook_logs,omitempty" gorm:"foreignKey:PaymentID"`
}

// Payment Refund Model
type PaymentRefund struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PaymentID     string    `json:"payment_id" gorm:"not null;index"`
	RefundID      string    `json:"refund_id" gorm:"not null;index"`
	RefundAmount  float64   `json:"refund_amount" gorm:"not null"`
	Reason        string    `json:"reason" gorm:"not null"`
	Status        string    `json:"status" gorm:"not null;default:processing"` // processing, success, failed
	ProcessedAt   *time.Time `json:"processed_at"`
	FailureReason string    `json:"failure_reason" gorm:""`
	Reference     string    `json:"reference" gorm:""`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	// Relationships
	Payment       *Payment `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
}

// Webhook Log Model
type WebhookLog struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PaymentID  string    `json:"payment_id" gorm:"not null;index"`
	Event      string    `json:"event" gorm:"not null"` // payment.success, payment.failed, etc.
	Payload    string    `json:"payload" gorm:"type:text"` // JSON payload
	Signature  string    `json:"signature" gorm:"not null"` // Webhook signature
	IsVerified bool      `json:"is_verified" gorm:"default:false"`
	Processed  bool      `json:"processed" gorm:"default:false"`
	Response   string    `json:"response" gorm:"type:text"` // Response data
	StatusCode int       `json:"status_code" gorm:"default:200"`
	IPAddress  string    `json:"ip_address" gorm:""`
	UserAgent   string    `json:"user_agent" gorm:""`
	CreatedAt  time.Time `json:"created_at"`
	
	// Relationships
	Payment    *Payment `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
}

// Payment Method Model (from Fiuu)
type PaymentMethod struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	IconURL      string  `json:"icon_url"`
	Type         string  `json:"type"` // banking, ewallet, card
	Fees         float64 `json:"fees"`
	MinAmount    float64 `json:"min_amount"`
	MaxAmount    float64 `json:"max_amount"`
	IsActive     bool    `json:"is_active"`
	Currency     string  `json:"currency"`
	Description  string  `json:"description"`
	CountryCode  string  `json:"country_code"`
	ProcessingTime string `json:"processing_time"`
}

// Customer Model (for payment)
type Customer struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required"`
	Address      string `json:"address"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
}

// Fiuu API Request/Response Models

// CreatePaymentRequest for Fiuu API
type FiuuCreatePaymentRequest struct {
	MerchantID    string    `json:"merchant_id"`
	OrderID       string    `json:"order_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Description   string    `json:"description"`
	Customer      Customer  `json:"customer"`
	PaymentType   string    `json:"payment_type"`
	ReturnURL     string    `json:"return_url"`
	CallbackURL   string    `json:"callback_url"`
	SuccessURL    string    `json:"success_url"`
	FailURL       string    `json:"fail_url"`
	ExpiryMinutes int       `json:"expiry_minutes"`
	Metadata      string    `json:"metadata"`
}

// FiuuCreatePaymentResponse from Fiuu API
type FiuuCreatePaymentResponse struct {
	Status      string `json:"status"`
	PaymentID   string `json:"payment_id"`
	TransactionID string `json:"transaction_id"`
	PaymentURL  string `json:"payment_url"`
	QRCode      string `json:"qr_code"`
	ExpiresAt   string `json:"expires_at"`
	Instructions string `json:"instructions,omitempty"`
}

// FiuuPaymentStatusResponse from Fiuu API
type FiuuPaymentStatusResponse struct {
	PaymentID      string    `json:"payment_id"`
	OrderID        string    `json:"order_id"`
	Status         string    `json:"status"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	PaidAt         string    `json:"paid_at"`
	FailedAt       string    `json:"failed_at"`
	PaymentMethod   string    `json:"payment_method"`
	TransactionID  string    `json:"transaction_id"`
	FailureReason  string    `json:"failure_reason"`
	BankName       string    `json:"bank_name"`
	Reference      string    `json:"reference"`
	Metadata       string    `json:"metadata"`
}

// FiuuRefundRequest for Fiuu API
type FiuuRefundRequest struct {
	PaymentID    string  `json:"payment_id"`
	RefundAmount float64 `json:"refund_amount"`
	Reason       string  `json:"reason"`
	Reference    string  `json:"reference"`
}

// FiuuRefundResponse from Fiuu API
type FiuuRefundResponse struct {
	Status                string    `json:"status"`
	RefundID              string    `json:"refund_id"`
	PaymentID             string    `json:"payment_id"`
	RefundAmount          float64   `json:"refund_amount"`
	StatusText            string    `json:"status_text"`
	EstimatedCompletion   string    `json:"estimated_completion"`
	ProcessingFee         float64   `json:"processing_fee"`
	NetRefundAmount      float64   `json:"net_refund_amount"`
	CreatedAt             string    `json:"created_at"`
}

// FiuuPaymentMethodsResponse from Fiuu API
type FiuuPaymentMethodsResponse struct {
	Methods []PaymentMethod `json:"methods"`
	Total   int            `json:"total"`
	Currency string         `json:"currency"`
}

// Request Models

// CreatePaymentRequest represents payment creation request
type CreatePaymentRequest struct {
	OrderID       string          `json:"order_id" binding:"required"`
	Amount        float64         `json:"amount" binding:"required,min=0.01"`
	Currency      string          `json:"currency" binding:"required,oneof=MYR"`
	Description   string          `json:"description" binding:"required,min=1,max=200"`
	PaymentType   string          `json:"payment_type" binding:"required"`
	PaymentMethod string          `json:"payment_method" binding:"required"`
	Customer      Customer        `json:"customer" binding:"required"`
	ReturnURL     string          `json:"return_url" binding:"required,url"`
	CallbackURL   string          `json:"callback_url" binding:"required,url"`
	SuccessURL    string          `json:"success_url" binding:"omitempty,url"`
	FailURL       string          `json:"fail_url" binding:"omitempty,url"`
	ExpiryMinutes int             `json:"expiry_minutes" binding:"omitempty,min=5,max=1440"`
	Metadata      string          `json:"metadata"`
}

// UpdatePaymentRequest represents payment update request
type UpdatePaymentRequest struct {
	Description   string `json:"description" binding:"omitempty,min=1,max=200"`
	ExpiryMinutes int    `json:"expiry_minutes" binding:"omitempty,min=5,max=1440"`
}

// RefundPaymentRequest represents refund request
type RefundPaymentRequest struct {
	RefundAmount float64 `json:"refund_amount" binding:"required,min=0.01"`
	Reason       string  `json:"reason" binding:"required,min=1,max=500"`
	Reference    string  `json:"reference" binding:"max=100"`
}

// SearchPaymentsRequest represents search payments request
type SearchPaymentsRequest struct {
	Query         string    `json:"query" form:"query"`
	OrderID       string    `json:"order_id" form:"order_id"`
	Status        string    `json:"status" form:"status"`
	PaymentType   string    `json:"payment_type" form:"payment_type"`
	PaymentMethod string    `json:"payment_method" form:"payment_method"`
	CustomerEmail string    `json:"customer_email" form:"customer_email"`
	CustomerPhone string    `json:"customer_phone" form:"customer_phone"`
	MinAmount     float64   `json:"min_amount" form:"min_amount"`
	MaxAmount     float64   `json:"max_amount" form:"max_amount"`
	StartDate     string    `json:"start_date" form:"start_date"`
	EndDate       string    `json:"end_date" form:"end_date"`
	SortBy        string    `json:"sort_by" form:"sort_by"` // created_at, amount, status
	SortOrder     string    `json:"sort_order" form:"sort_order"` // asc, desc
	Page          int       `json:"page" form:"page"`
	Limit         int       `json:"limit" form:"limit"`
}

// Response Models

// PaymentResponse represents payment response
type PaymentResponse struct {
	Payment         Payment           `json:"payment"`
	PaymentURL      string            `json:"payment_url"`
	QRCode          string            `json:"qr_code"`
	TimeRemaining   string            `json:"time_remaining"`
	IsExpired       bool              `json:"is_expired"`
	CanRetry        bool              `json:"can_retry"`
	AvailableMethods []PaymentMethod  `json:"available_methods"`
}

// PaymentsResponse represents payments list response
type PaymentsResponse struct {
	Payments    []Payment `json:"payments"`
	Total       int64     `json:"total"`
	Page        int       `json:"page"`
	Limit       int       `json:"limit"`
	TotalPages  int       `json:"total_pages"`
	HasNext     bool      `json:"has_next"`
	HasPrevious bool      `json:"has_previous"`
}

// PaymentMethodsResponse represents payment methods response
type PaymentMethodsResponse struct {
	Methods     []PaymentMethod `json:"methods"`
	Total       int            `json:"total"`
	Currency    string         `json:"currency"`
}

// RefundResponse represents refund response
type RefundResponse struct {
	Refund               PaymentRefund `json:"refund"`
	EstimatedCompletion   string        `json:"estimated_completion"`
	NetRefundAmount      float64       `json:"net_refund_amount"`
	ProcessingFee        float64       `json:"processing_fee"`
}

// Webhook Messages (from Fiuu)

// FiuuWebhookPaymentSuccess from Fiuu
type FiuuWebhookPaymentSuccess struct {
	Event         string    `json:"event"`
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	PaidAt        string    `json:"paid_at"`
	PaymentMethod string    `json:"payment_method"`
	TransactionID string    `json:"transaction_id"`
	BankName      string    `json:"bank_name"`
	Reference     string    `json:"reference"`
	Timestamp     string    `json:"timestamp"`
	Signature     string    `json:"signature"`
	Metadata      string    `json:"metadata"`
}

// FiuuWebhookPaymentFailed from Fiuu
type FiuuWebhookPaymentFailed struct {
	Event         string    `json:"event"`
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	Status        string    `json:"status"`
	FailureReason string    `json:"failure_reason"`
	FailedAt      string    `json:"failed_at"`
	Timestamp     string    `json:"timestamp"`
	Signature     string    `json:"signature"`
	Metadata      string    `json:"metadata"`
}

// FiuuWebhookPaymentRefunded from Fiuu
type FiuuWebhookPaymentRefunded struct {
	Event          string    `json:"event"`
	PaymentID      string    `json:"payment_id"`
	RefundID       string    `json:"refund_id"`
	OrderID        string    `json:"order_id"`
	Status         string    `json:"status"`
	RefundAmount   float64   `json:"refund_amount"`
	RefundedAt     string    `json:"refunded_at"`
	Reason         string    `json:"reason"`
	Timestamp      string    `json:"timestamp"`
	Signature      string    `json:"signature"`
	Metadata       string    `json:"metadata"`
}

// Order Model (simplified for payment service)
type Order struct {
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	TotalAmount   float64 `json:"total_amount"`
	Status        string  `json:"status"`
	PaymentStatus string  `json:"payment_status"`
	CreatedAt     string  `json:"created_at"`
}

// Helper Methods

// BeforeCreate hook for Payment
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = generateUUID()
	}
	if p.Currency == "" {
		p.Currency = "MYR"
	}
	if p.Status == "" {
		p.Status = PaymentStatusPending
	}
	if p.ExpiresAt.IsZero() {
		p.ExpiresAt = time.Now().Add(24 * time.Hour) // Default 24 hours
	}
	return nil
}

// BeforeCreate hook for PaymentRefund
func (r *PaymentRefund) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = generateUUID()
	}
	if r.Status == "" {
		r.Status = PaymentStatusProcessing
	}
	return nil
}

// BeforeCreate hook for WebhookLog
func (w *WebhookLog) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = generateUUID()
	}
	return nil
}

// IsExpired checks if payment is expired
func (p *Payment) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// GetTimeRemaining returns time remaining as string
func (p *Payment) GetTimeRemaining() string {
	if p.IsExpired() {
		return "expired"
	}
	
	duration := time.Until(p.ExpiresAt)
	if duration <= 0 {
		return "expired"
	}
	
	if duration < time.Minute {
		return "less than 1 minute"
	} else if duration < time.Hour {
		return fmt.Sprintf("%.0f minutes", duration.Minutes())
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%.0f hours", duration.Hours())
	} else {
		return fmt.Sprintf("%.0f days", duration.Hours()/24)
	}
}

// CanRetry checks if payment can be retried
func (p *Payment) CanRetry() bool {
	return p.Status == PaymentStatusFailed || p.Status == PaymentStatusCancelled || p.IsExpired()
}

// SetWebhookEvents sets webhook events from string array
func (p *Payment) SetWebhookEvents(events []string) {
	if len(events) == 0 {
		p.WebhookEvents = ""
		return
	}
	data, _ := json.Marshal(events)
	p.WebhookEvents = string(data)
}

// GetWebhookEvents returns webhook events as string array
func (p *Payment) GetWebhookEvents() []string {
	if p.WebhookEvents == "" {
		return []string{}
	}
	var events []string
	json.Unmarshal([]byte(p.WebhookEvents), &events)
	return events
}

// GetStatusText returns human readable status text
func (p *Payment) GetStatusText() string {
	switch p.Status {
	case PaymentStatusPending:
		return "Pending Payment"
	case PaymentStatusProcessing:
		return "Processing"
	case PaymentStatusSuccess:
		return "Payment Successful"
	case PaymentStatusFailed:
		return "Payment Failed"
	case PaymentStatusCancelled:
		return "Payment Cancelled"
	case PaymentStatusRefunded:
		return "Payment Refunded"
	case PaymentStatusExpired:
		return "Payment Expired"
	default:
		return "Unknown Status"
	}
}

// GetPaymentMethodText returns human readable payment method text
func (p *Payment) GetPaymentMethodText() string {
	switch p.PaymentMethod {
	case FiuuMethodFPX:
		return "FPX Online Banking"
	case FiuuMethodMaybank:
		return "Maybank2u"
	case FiuuMethodCIMB:
		return "CIMB Clicks"
	case FiuuMethodPublicBank:
		return "Public Bank Online"
	case FiuuMethodRHB:
		return "RHB Now"
	case FiuuMethodHLB:
		return "Hong Leong Online"
	case FiuuMethodAmbank:
		return "AmBank Online"
	case FiuuMethodTNG:
		return "Touch 'n Go eWallet"
	case FiuuMethodGrabPay:
		return "GrabPay"
	case FiuuMethodBoost:
		return "Boost"
	case FiuuMethodShopeePay:
		return "ShopeePay"
	case FiuuMethodDuitNow:
		return "DuitNow"
	case FiuuMethodVisa:
		return "Visa Card"
	case FiuuMethodMastercard:
		return "Mastercard"
	case FiuuMethodAmex:
		return "American Express"
	default:
		return p.PaymentMethod
	}
}

// Helper function to generate UUID (simplified)
func generateUUID() string {
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
}