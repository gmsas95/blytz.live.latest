package dtos

import (
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
)

// CreateCustomPaymentMethodRequest represents a request to create a custom payment method
type CreateCustomPaymentMethodRequest struct {
	Type       string                 `json:"type" binding:"required"`       // e.g., "paypal", "apple_pay", "google_pay"
	Provider   string                 `json:"provider" binding:"required"`
	CustomerID string                 `json:"customer_id" binding:"required"`
	Currency   string                 `json:"currency" binding:"required"`
	Data       map[string]interface{}   `json:"data"`
	Metadata   map[string]string        `json:"metadata,omitempty"`
}

// CreateMbWayPaymentMethodRequest represents a request to create an MbWay payment method
type CreateMbWayPaymentMethodRequest struct {
	PhoneNumber string            `json:"phone_number" binding:"required"`
	CountryCode string            `json:"country_code" binding:"required,len=2"`
	CustomerID  string            `json:"customer_id" binding:"required"`
	Currency    string            `json:"currency" binding:"required"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateTWINTPaymentMethodRequest represents a request to create a TWINT payment method
type CreateTWINTPaymentMethodRequest struct {
	QRCode     string `json:"qr_code,omitempty"`
	DeviceID    string `json:"device_id,omitempty"`
	MerchantID  string `json:"merchant_id,omitempty"`
	CustomerID  string `json:"customer_id" binding:"required"`
	Currency    string `json:"currency" binding:"required"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateCryptoPaymentMethodRequest represents a request to create a crypto payment method
type CreateCryptoPaymentMethodRequest struct {
	CryptoType    string `json:"crypto_type" binding:"required"`    // e.g., "bitcoin", "ethereum", "usdc"
	WalletAddress string `json:"wallet_address" binding:"required"`
	Network       string `json:"network,omitempty"`             // e.g., "mainnet", "polygon", "arbitrum"
	CustomerID    string `json:"customer_id" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// PaymentMethodResponse represents a payment method response
type PaymentMethodResponse struct {
	ID          string                    `json:"id"`
	Type        string                    `json:"type"`
	CustomerID  string                    `json:"customer_id"`
	Status      string                    `json:"status"`
	CreatedAt   time.Time                 `json:"created_at"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// PaymentMethodConfigResponse represents payment method configuration response
type PaymentMethodConfigResponse struct {
	Type              string   `json:"type"`
	Enabled           bool     `json:"enabled"`
	MinimumAmount     int64    `json:"minimum_amount"`     // in cents
	MaximumAmount     int64    `json:"maximum_amount"`     // in cents
	SupportedCurrencies []string `json:"supported_currencies"`
	RequiresVerification bool     `json:"requires_verification"`
	AdditionalData    map[string]interface{} `json:"additional_data,omitempty"`
}

// PaymentMethodListResponse represents a list of payment methods response
type PaymentMethodListResponse struct {
	PaymentMethods []PaymentMethodResponse `json:"payment_methods"`
	Pagination     PaginationResponse       `json:"pagination"`
}

// PaymentMethodConfigsResponse represents payment method configurations response
type PaymentMethodConfigsResponse struct {
	Configs   map[string]PaymentMethodConfigResponse `json:"configs"`
	Currency  string                              `json:"currency,omitempty"`
	Count      int                                 `json:"count"`
}

// PaginationResponse represents pagination information
type PaginationResponse struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// Helper functions to convert between models and DTOs

// ToPaymentMethodResponse converts a Stripe payment method to PaymentMethodResponse DTO
func ToPaymentMethodResponse(pm interface{}, paymentType string) *PaymentMethodResponse {
	response := &PaymentMethodResponse{
		Type:      paymentType,
		CreatedAt: time.Now(),
		Metadata: make(map[string]interface{}),
	}

	// Extract common fields based on payment method type
	switch paymentType {
	case "custom":
		if pm, ok := pm.(interface{ GetID() string; GetType() string }); ok {
			response.ID = pm.GetID()
			response.Status = "active"
		}
	case "mbway":
		if pm, ok := pm.(interface{ GetID() string; GetCustomer() string }); ok {
			response.ID = pm.GetID()
			response.CustomerID = pm.GetCustomer()
			response.Status = "active"
		}
	case "twint":
		if pm, ok := pm.(interface{ GetID() string; GetCustomer() string }); ok {
			response.ID = pm.GetID()
			response.CustomerID = pm.GetCustomer()
			response.Status = "active"
		}
	case "crypto":
		if pm, ok := pm.(interface{ GetID() string; GetCustomer() string }); ok {
			response.ID = pm.GetID()
			response.CustomerID = pm.GetCustomer()
			response.Status = "active"
		}
	default:
		// Handle generic payment method
		if pm, ok := pm.(interface{ GetID() string; GetCustomer() string }); ok {
			response.ID = pm.GetID()
			response.CustomerID = pm.GetCustomer()
			response.Status = "active"
		}
	}

	return response
}

// ToPaymentMethodConfigResponse converts a PaymentMethodConfig model to PaymentMethodConfigResponse DTO
func ToPaymentMethodConfigResponse(config models.PaymentMethodConfig) *PaymentMethodConfigResponse {
	return &PaymentMethodConfigResponse{
		Type:              string(config.Type),
		Enabled:           config.Enabled,
		MinimumAmount:     config.MinimumAmount,
		MaximumAmount:     config.MaximumAmount,
		SupportedCurrencies: config.SupportedCurrencies,
		RequiresVerification: config.RequiresVerification,
		AdditionalData:    config.AdditionalData,
	}
}

// ToPaymentMethodConfigsResponse converts a map of PaymentMethodConfig models to PaymentMethodConfigsResponse DTO
func ToPaymentMethodConfigsResponse(configs map[models.PaymentMethodType]models.PaymentMethodConfig, currency string) *PaymentMethodConfigsResponse {
	response := &PaymentMethodConfigsResponse{
		Configs:  make(map[string]PaymentMethodConfigResponse),
		Currency: currency,
		Count:     len(configs),
	}

	for methodType, config := range configs {
		response.Configs[string(methodType)] = *ToPaymentMethodConfigResponse(config)
	}

	return response
}

// Validation helper functions

// ValidateCreateCustomPaymentMethodRequest validates the create custom payment method request
func ValidateCreateCustomPaymentMethodRequest(req *CreateCustomPaymentMethodRequest) map[string]string {
	errors := make(map[string]string)

	if req.Type == "" {
		errors["type"] = "Payment method type is required"
	}

	if req.Provider == "" {
		errors["provider"] = "Provider is required"
	}

	if req.CustomerID == "" {
		errors["customer_id"] = "Customer ID is required"
	}

	// Validate custom payment method types
	validTypes := []string{"paypal", "apple_pay", "google_pay"}
	isValidType := false
	for _, validType := range validTypes {
		if req.Type == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		errors["type"] = "Invalid payment method type. Supported types: paypal, apple_pay, google_pay"
	}

	return errors
}

// ValidateCreateMbWayPaymentMethodRequest validates the create MbWay payment method request
func ValidateCreateMbWayPaymentMethodRequest(req *CreateMbWayPaymentMethodRequest) map[string]string {
	errors := make(map[string]string)

	if req.PhoneNumber == "" {
		errors["phone_number"] = "Phone number is required"
	}

	if req.CountryCode == "" {
		errors["country_code"] = "Country code is required"
	}

	if len(req.CountryCode) != 2 {
		errors["country_code"] = "Country code must be 2 characters"
	}

	if req.CustomerID == "" {
		errors["customer_id"] = "Customer ID is required"
	}

	// Validate Portuguese phone number format for MbWay
	if req.CountryCode == "PT" && len(req.PhoneNumber) < 9 {
		errors["phone_number"] = "Invalid Portuguese phone number format"
	}

	return errors
}

// ValidateCreateTWINTPaymentMethodRequest validates the create TWINT payment method request
func ValidateCreateTWINTPaymentMethodRequest(req *CreateTWINTPaymentMethodRequest) map[string]string {
	errors := make(map[string]string)

	if req.CustomerID == "" {
		errors["customer_id"] = "Customer ID is required"
	}

	// At least one identifier should be provided
	if req.QRCode == "" && req.DeviceID == "" && req.MerchantID == "" {
		errors["identifier"] = "At least one of qr_code, device_id, or merchant_id is required"
	}

	return errors
}

// ValidateCreateCryptoPaymentMethodRequest validates the create crypto payment method request
func ValidateCreateCryptoPaymentMethodRequest(req *CreateCryptoPaymentMethodRequest) map[string]string {
	errors := make(map[string]string)

	if req.CryptoType == "" {
		errors["crypto_type"] = "Crypto type is required"
	}

	if req.WalletAddress == "" {
		errors["wallet_address"] = "Wallet address is required"
	}

	if req.CustomerID == "" {
		errors["customer_id"] = "Customer ID is required"
	}

	// Validate crypto types
	validTypes := []string{"bitcoin", "ethereum", "usdc"}
	isValidType := false
	for _, validType := range validTypes {
		if req.CryptoType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		errors["crypto_type"] = "Invalid crypto type. Supported types: bitcoin, ethereum, usdc"
	}

	// Basic wallet address validation
	if len(req.WalletAddress) < 10 {
		errors["wallet_address"] = "Invalid wallet address format"
	}

	return errors
}