package dtos

import (
	"encoding/json"
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
)

// Transfer DTOs

// CreateTransferRequest represents a request to create a transfer
type CreateTransferRequest struct {
	Amount               float64            `json:"amount" binding:"required,min=0.01"`
	Currency             string             `json:"currency" binding:"required,len=3"`
	ConnectedAccountID    string             `json:"connected_account_id" binding:"required,uuid"`
	TransferGroup        string             `json:"transfer_group,omitempty"`
	SourceTransaction     string             `json:"source_transaction,omitempty"`
	DestinationPaymentID  string             `json:"destination_payment_id,omitempty"`
	FeePercent           float64            `json:"fee_percent" binding:"min=0,max=100"`
	Metadata             map[string]string  `json:"metadata,omitempty"`
}

// CreateBatchTransferRequest represents a request to create multiple transfers
type CreateBatchTransferRequest struct {
	Transfers []CreateTransferRequest `json:"transfers" binding:"required,min=1"`
}

// ReverseTransferRequest represents a request to reverse a transfer
type ReverseTransferRequest struct {
	Amount float64 `json:"amount" binding:"required,min=0.01"`
	Reason string `json:"reason" binding:"omitempty,oneof=duplicate fraudulent requested_by_customer"`
}

// TransferResponse represents a transfer response
type TransferResponse struct {
	ID                   string                 `json:"id"`
	StripeTransferID     string                 `json:"stripe_transfer_id"`
	ConnectedAccountID    string                 `json:"connected_account_id"`
	Amount               int64                  `json:"amount"` // Amount in cents
	Currency             string                 `json:"currency"`
	Status               string                 `json:"status"`
	DestinationPaymentID  string                 `json:"destination_payment_id,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// TransferHistoryResponse represents a transfer history response
type TransferHistoryResponse struct {
	Transfers  []TransferResponse `json:"transfers"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"total_pages"`
	HasNext    bool              `json:"has_next"`
	HasPrev    bool              `json:"has_prev"`
}

// Payout DTOs

// CreatePayoutRequest represents a request to create a payout
type CreatePayoutRequest struct {
	Amount               float64            `json:"amount" binding:"required,min=0.01"`
	Currency             string             `json:"currency" binding:"required,len=3"`
	ConnectedAccountID    string             `json:"connected_account_id" binding:"required,uuid"`
	StatementDescriptor   string             `json:"statement_descriptor,omitempty"`
	RequestedBy          string             `json:"requested_by" binding:"required,uuid"`
	Metadata             map[string]string  `json:"metadata,omitempty"`
}

// CreateScheduledPayoutRequest represents a request to create a scheduled payout
type CreateScheduledPayoutRequest struct {
	Amount               float64            `json:"amount" binding:"required,min=0.01"`
	Currency             string             `json:"currency" binding:"required,len=3"`
	ConnectedAccountID    string             `json:"connected_account_id" binding:"required,uuid"`
	Schedule             string             `json:"schedule" binding:"required"`
	Frequency            string             `json:"frequency" binding:"required,oneof=daily weekly monthly"`
	RequestedBy          string             `json:"requested_by" binding:"required,uuid"`
	Metadata             map[string]string  `json:"metadata,omitempty"`
}

// PayoutResponse represents a payout response
type PayoutResponse struct {
	ID                 string                 `json:"id"`
	StripePayoutID     string                 `json:"stripe_payout_id"`
	ConnectedAccountID  string                 `json:"connected_account_id"`
	Amount             int64                  `json:"amount"` // Amount in cents
	Currency           string                 `json:"currency"`
	Status             string                 `json:"status"`
	ArrivalDate        *time.Time             `json:"arrival_date,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// PayoutHistoryResponse represents a payout history response
type PayoutHistoryResponse struct {
	Payouts    []PayoutResponse `json:"payouts"`
	Page        int              `json:"page"`
	PerPage     int              `json:"per_page"`
	Total       int64            `json:"total"`
	TotalPages  int              `json:"total_pages"`
	HasNext     bool             `json:"has_next"`
	HasPrev     bool             `json:"has_prev"`
}

// BalanceResponse represents a balance response
type BalanceResponse struct {
	ConnectedAccountID string `json:"connected_account_id"`
	Available         int64  `json:"available"` // Amount in cents
	Pending          int64  `json:"pending"`  // Amount in cents
	Currency         string  `json:"currency"`
}

// Helper functions to convert between models and DTOs

// ToTransferResponse converts a Transfer model to TransferResponse DTO
func ToTransferResponse(transfer *models.Transfer) *TransferResponse {
	response := &TransferResponse{
		ID:                  transfer.ID,
		StripeTransferID:     transfer.StripeTransferID,
		ConnectedAccountID:    transfer.ConnectedAccountID,
		Amount:              transfer.Amount,
		Currency:            transfer.Currency,
		Status:              transfer.Status,
		DestinationPaymentID: transfer.DestinationPaymentID,
		CreatedAt:           transfer.CreatedAt,
		UpdatedAt:           transfer.UpdatedAt,
	}

	// Parse metadata JSON if present
	if transfer.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(transfer.Metadata), &metadata); err == nil {
			response.Metadata = metadata
		}
	}

	return response
}

// ToTransferHistoryResponse converts a slice of Transfer models to TransferHistoryResponse DTO
func ToTransferHistoryResponse(transfers []*models.Transfer, page, perPage int, total int64) *TransferHistoryResponse {
	transferResponses := make([]TransferResponse, len(transfers))
	for i, transfer := range transfers {
		transferResponses[i] = *ToTransferResponse(transfer)
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return &TransferHistoryResponse{
		Transfers:  transferResponses,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ToPayoutResponse converts a Payout model to PayoutResponse DTO
func ToPayoutResponse(payout *models.Payout) *PayoutResponse {
	response := &PayoutResponse{
		ID:                payout.ID,
		StripePayoutID:    payout.StripePayoutID,
		ConnectedAccountID: payout.ConnectedAccountID,
		Amount:            payout.Amount,
		Currency:          payout.Currency,
		Status:            payout.Status,
		CreatedAt:         payout.CreatedAt,
		UpdatedAt:         payout.UpdatedAt,
	}

	// Add arrival date if available
	if payout.ArrivalDate != nil && !payout.ArrivalDate.IsZero() {
		response.ArrivalDate = payout.ArrivalDate
	}

	// Parse metadata JSON if present
	if payout.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(payout.Metadata), &metadata); err == nil {
			response.Metadata = metadata
		}
	}

	return response
}

// ToPayoutHistoryResponse converts a slice of Payout models to PayoutHistoryResponse DTO
func ToPayoutHistoryResponse(payouts []*models.Payout, page, perPage int, total int64) *PayoutHistoryResponse {
	payoutResponses := make([]PayoutResponse, len(payouts))
	for i, payout := range payouts {
		payoutResponses[i] = *ToPayoutResponse(payout)
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return &PayoutHistoryResponse{
		Payouts:    payoutResponses,
		Page:        page,
		PerPage:     perPage,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}
}

// Validation helper functions

// ValidateCreateTransferRequest validates the create transfer request
func ValidateCreateTransferRequest(req *CreateTransferRequest) map[string]string {
	errors := make(map[string]string)

	if req.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	if req.Currency == "" {
		errors["currency"] = "Currency is required"
	}

	if len(req.Currency) != 3 {
		errors["currency"] = "Currency must be a 3-letter ISO currency code"
	}

	if req.ConnectedAccountID == "" {
		errors["connected_account_id"] = "Connected account ID is required"
	}

	if req.FeePercent < 0 || req.FeePercent > 100 {
		errors["fee_percent"] = "Fee percent must be between 0 and 100"
	}

	return errors
}

// ValidateCreatePayoutRequest validates the create payout request
func ValidateCreatePayoutRequest(req *CreatePayoutRequest) map[string]string {
	errors := make(map[string]string)

	if req.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	if req.Currency == "" {
		errors["currency"] = "Currency is required"
	}

	if len(req.Currency) != 3 {
		errors["currency"] = "Currency must be a 3-letter ISO currency code"
	}

	if req.ConnectedAccountID == "" {
		errors["connected_account_id"] = "Connected account ID is required"
	}

	if req.RequestedBy == "" {
		errors["requested_by"] = "Requested by field is required"
	}

	return errors
}

// ValidateCreateScheduledPayoutRequest validates the create scheduled payout request
func ValidateCreateScheduledPayoutRequest(req *CreateScheduledPayoutRequest) map[string]string {
	errors := make(map[string]string)

	if req.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	if req.Currency == "" {
		errors["currency"] = "Currency is required"
	}

	if len(req.Currency) != 3 {
		errors["currency"] = "Currency must be a 3-letter ISO currency code"
	}

	if req.ConnectedAccountID == "" {
		errors["connected_account_id"] = "Connected account ID is required"
	}

	if req.Schedule == "" {
		errors["schedule"] = "Schedule is required"
	}

	if req.Frequency == "" {
		errors["frequency"] = "Frequency is required"
	}

	if req.RequestedBy == "" {
		errors["requested_by"] = "Requested by field is required"
	}

	return errors
}

// ValidateReverseTransferRequest validates the reverse transfer request
func ValidateReverseTransferRequest(req *ReverseTransferRequest) map[string]string {
	errors := make(map[string]string)

	if req.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	return errors
}