package dtos

import (
	"encoding/json"
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
)

// CreatePaymentIntentRequest represents a request to create a payment intent
type CreatePaymentIntentRequest struct {
	Amount             float64            `json:"amount" binding:"required,min=0.50"`
	Currency           string             `json:"currency" binding:"required,len=3"`
	CustomerID         string             `json:"customer_id" binding:"required"`
	UserID             string             `json:"user_id" binding:"required,uuid"`
	AuctionID          string             `json:"auction_id,omitempty" binding:"omitempty,uuid"`
	ConnectedAccountID string             `json:"connected_account_id,omitempty" binding:"omitempty,uuid"`
	PaymentMethodTypes []string           `json:"payment_method_types"`
	Metadata           map[string]string  `json:"metadata,omitempty"`
}

// ConfirmPaymentRequest represents a request to confirm a payment
type ConfirmPaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id" binding:"required"`
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
}

// RefundRequest represents a request to refund a payment
type RefundRequest struct {
	PaymentIntentID string  `json:"payment_intent_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,min=0.01"`
	Reason          string  `json:"reason" binding:"omitempty,oneof=duplicate fraudulent requested_by_customer"`
}

// PaymentStatusResponse represents a payment status response
type PaymentStatusResponse struct {
	ID                    string    `json:"id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id"`
	Amount                int64     `json:"amount"` // Amount in cents
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	UserID                string    `json:"user_id"`
	AuctionID             *string   `json:"auction_id,omitempty"`
	ConnectedAccountID    *string   `json:"connected_account_id,omitempty"`
	ApplicationFeeAmount  int64     `json:"application_fee_amount"` // Fee amount in cents
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	ClientSecret          string    `json:"client_secret,omitempty"`
	NextAction            *NextAction `json:"next_action,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// NextAction represents the next action required for a payment
type NextAction struct {
	Type         string `json:"type"`
	RedirectToURL *RedirectToURL `json:"redirect_to_url,omitempty"`
}

// RedirectToURL represents a redirect URL for payment authentication
type RedirectToURL struct {
	URL string `json:"url"`
	ReturnURL string `json:"return_url"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	ID            string    `json:"id"`
	Amount        int64     `json:"amount"` // Amount in cents
	Currency      string    `json:"currency"`
	PaymentIntent string    `json:"payment_intent"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// PaymentHistoryResponse represents a payment history response
type PaymentHistoryResponse struct {
	Payments    []PaymentStatusResponse `json:"payments"`
	Page        int                     `json:"page"`
	PerPage     int                     `json:"per_page"`
	Total       int64                   `json:"total"`
	TotalPages  int                     `json:"total_pages"`
	HasNext     bool                    `json:"has_next"`
	HasPrev     bool                    `json:"has_prev"`
}

// PaymentIntentResponse represents a payment intent creation response
type PaymentIntentResponse struct {
	ID                    string    `json:"id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id"`
	Amount                int64     `json:"amount"` // Amount in cents
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	UserID                string    `json:"user_id"`
	AuctionID             *string   `json:"auction_id,omitempty"`
	ConnectedAccountID    *string   `json:"connected_account_id,omitempty"`
	ApplicationFeeAmount  int64     `json:"application_fee_amount"` // Fee amount in cents
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	ClientSecret          string    `json:"client_secret"`
	NextAction            *NextAction `json:"next_action,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

// Helper functions to convert between models and DTOs

// ToPaymentIntentResponse converts a PaymentIntent model to PaymentIntentResponse DTO
func ToPaymentIntentResponse(pi *models.PaymentIntent, clientSecret string, nextAction *NextAction) *PaymentIntentResponse {
	response := &PaymentIntentResponse{
		ID:                    pi.ID,
		StripePaymentIntentID: pi.StripePaymentIntentID,
		Amount:                pi.Amount,
		Currency:              pi.Currency,
		Status:                pi.Status,
		UserID:                pi.UserID,
		AuctionID:             pi.AuctionID,
		ConnectedAccountID:    pi.ConnectedAccountID,
		ApplicationFeeAmount:  pi.ApplicationFeeAmount,
		CreatedAt:             pi.CreatedAt,
	}

	if clientSecret != "" {
		response.ClientSecret = clientSecret
	}

	if nextAction != nil {
		response.NextAction = nextAction
	}

	// Parse metadata JSON if present
	if pi.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(pi.Metadata), &metadata); err == nil {
			response.Metadata = metadata
		}
	}

	return response
}

// ToPaymentStatusResponse converts a PaymentIntent model to PaymentStatusResponse DTO
func ToPaymentStatusResponse(pi *models.PaymentIntent, clientSecret string, nextAction *NextAction) *PaymentStatusResponse {
	response := &PaymentStatusResponse{
		ID:                    pi.ID,
		StripePaymentIntentID: pi.StripePaymentIntentID,
		Amount:                pi.Amount,
		Currency:              pi.Currency,
		Status:                pi.Status,
		UserID:                pi.UserID,
		AuctionID:             pi.AuctionID,
		ConnectedAccountID:    pi.ConnectedAccountID,
		ApplicationFeeAmount:  pi.ApplicationFeeAmount,
		CreatedAt:             pi.CreatedAt,
		UpdatedAt:             pi.UpdatedAt,
	}

	if clientSecret != "" {
		response.ClientSecret = clientSecret
	}

	if nextAction != nil {
		response.NextAction = nextAction
	}

	// Parse metadata JSON if present
	if pi.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(pi.Metadata), &metadata); err == nil {
			response.Metadata = metadata
		}
	}

	return response
}

// ToPaymentHistoryResponse converts a slice of PaymentIntent models to PaymentHistoryResponse DTO
func ToPaymentHistoryResponse(payments []*models.PaymentIntent, page, perPage int, total int64) *PaymentHistoryResponse {
	paymentResponses := make([]PaymentStatusResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = *ToPaymentStatusResponse(payment, "", nil)
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return &PaymentHistoryResponse{
		Payments:   paymentResponses,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ToRefundResponse converts a Stripe refund to RefundResponse DTO
func ToRefundResponse(refund interface{}) *RefundResponse {
	// This would typically convert from a Stripe refund object
	// For now, we'll create a basic implementation
	response := &RefundResponse{
		CreatedAt: time.Now(),
	}

	// In a real implementation, you would extract fields from the Stripe refund object
	// based on the actual Stripe Go library structure

	return response
}

// Validation helper functions

// ValidateCreatePaymentIntentRequest validates the create payment intent request
func ValidateCreatePaymentIntentRequest(req *CreatePaymentIntentRequest) map[string]string {
	errors := make(map[string]string)

	if req.Amount < 0.50 {
		errors["amount"] = "Amount must be at least $0.50"
	}

	if len(req.Currency) != 3 {
		errors["currency"] = "Currency must be a 3-letter ISO currency code"
	}

	if req.CustomerID == "" {
		errors["customer_id"] = "Customer ID is required"
	}

	if req.UserID == "" {
		errors["user_id"] = "User ID is required"
	}

	// Set default payment method types if not provided
	if len(req.PaymentMethodTypes) == 0 {
		req.PaymentMethodTypes = []string{"card"}
	}

	return errors
}

// ValidateConfirmPaymentRequest validates the confirm payment request
func ValidateConfirmPaymentRequest(req *ConfirmPaymentRequest) map[string]string {
	errors := make(map[string]string)

	if req.PaymentIntentID == "" {
		errors["payment_intent_id"] = "Payment Intent ID is required"
	}

	if req.PaymentMethodID == "" {
		errors["payment_method_id"] = "Payment Method ID is required"
	}

	return errors
}

// ValidateRefundRequest validates the refund request
func ValidateRefundRequest(req *RefundRequest) map[string]string {
	errors := make(map[string]string)

	if req.PaymentIntentID == "" {
		errors["payment_intent_id"] = "Payment Intent ID is required"
	}

	if req.Amount < 0.01 {
		errors["amount"] = "Refund amount must be at least $0.01"
	}

	return errors
}