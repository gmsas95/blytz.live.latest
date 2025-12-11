package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PayoutService handles payout operations
type PayoutService struct {
	db          *gorm.DB
	stripeClient *StripeClient
	config      *config.StripeConfig
	logger      *zap.Logger
}

// NewPayoutService creates a new payout service instance
func NewPayoutService(db *gorm.DB, stripeClient *StripeClient, config *config.StripeConfig, logger *zap.Logger) *PayoutService {
	return &PayoutService{
		db:           db,
		stripeClient: stripeClient,
		config:       config,
		logger:       logger,
	}
}

// CreatePayout creates a new manual payout from a connected account
func (s *PayoutService) CreatePayout(ctx context.Context, req *CreatePayoutRequest) (*models.Payout, error) {
	s.logger.Info("Creating payout",
		zap.Float64("amount", req.Amount),
		zap.String("currency", req.Currency),
		zap.String("connected_account_id", req.ConnectedAccountID),
	)

	// Validate request
	if err := s.validateCreatePayoutRequest(req); err != nil {
		return nil, err
	}

	// Check if connected account exists and is active
	var connectedAccount models.ConnectedAccount
	if err := s.db.Where("id = ? AND payouts_enabled = ?", req.ConnectedAccountID, true).First(&connectedAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("connected account not found or not enabled for payouts")
		}
		return nil, fmt.Errorf("failed to verify connected account: %w", err)
	}

	// Create Stripe payout parameters
	stripeParams := &PayoutParams{
		Amount:              req.Amount,
		Currency:            req.Currency,
		Destination:         connectedAccount.StripeAccountID,
		StatementDescriptor: req.StatementDescriptor,
	}

	// Create payout with Stripe
	stripePayout, err := s.stripeClient.CreatePayout(ctx, stripeParams)
	if err != nil {
		s.logger.Error("Failed to create Stripe payout",
			zap.Error(err),
			zap.String("connected_account_id", req.ConnectedAccountID),
		)
		return nil, fmt.Errorf("failed to create payout: %w", err)
	}

	// Create local payout record
	payout := &models.Payout{
		StripePayoutID:     stripePayout.ID,
		ConnectedAccountID:  req.ConnectedAccountID,
		Amount:              int64(req.Amount * 100), // Convert to cents
		Currency:            req.Currency,
		Status:              string(stripePayout.Status),
	}

	// Set arrival date if available
	if stripePayout.ArrivalDate > 0 {
		arrivalTime := time.Unix(stripePayout.ArrivalDate, 0)
		payout.ArrivalDate = &arrivalTime
	}

	// Add metadata as JSON
	metadata := map[string]interface{}{
		"service":            "blytz-stripe-service",
		"manual_payout":      true,
		"requested_by":       req.RequestedBy,
	}
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}
	metadataJSON, _ := json.Marshal(metadata)
	payout.Metadata = string(metadataJSON)

	// Save to database
	if err := s.db.Create(payout).Error; err != nil {
		s.logger.Error("Failed to save payout to database",
			zap.Error(err),
			zap.String("stripe_payout_id", stripePayout.ID),
		)
		return nil, fmt.Errorf("failed to save payout: %w", err)
	}

	s.logger.Info("Payout created successfully",
		zap.String("payout_id", payout.ID),
		zap.String("stripe_payout_id", stripePayout.ID),
		zap.Float64("amount", req.Amount),
		zap.String("currency", req.Currency),
	)

	return payout, nil
}

// CreateScheduledPayout creates a scheduled payout
func (s *PayoutService) CreateScheduledPayout(ctx context.Context, req *CreateScheduledPayoutRequest) (*models.Payout, error) {
	s.logger.Info("Creating scheduled payout",
		zap.String("connected_account_id", req.ConnectedAccountID),
		zap.String("schedule", req.Schedule),
	)

	// For scheduled payouts, we would typically use Stripe's payout schedules
	// For now, we'll create a regular payout with scheduling metadata
	createReq := &CreatePayoutRequest{
		Amount:               req.Amount,
		Currency:             req.Currency,
		ConnectedAccountID:    req.ConnectedAccountID,
		StatementDescriptor:   fmt.Sprintf("Scheduled payout - %s", req.Schedule),
		RequestedBy:          req.RequestedBy,
		Metadata: map[string]string{
			"scheduled_payout": "true",
			"schedule":         req.Schedule,
			"frequency":        req.Frequency,
		},
	}

	return s.CreatePayout(ctx, createReq)
}

// GetPayout retrieves a payout by ID
func (s *PayoutService) GetPayout(ctx context.Context, payoutID string) (*models.Payout, error) {
	s.logger.Info("Retrieving payout",
		zap.String("payout_id", payoutID),
	)

	var payout models.Payout
	if err := s.db.Where("id = ?", payoutID).First(&payout).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payout not found")
		}
		return nil, fmt.Errorf("failed to retrieve payout: %w", err)
	}

	// Get latest data from Stripe
	stripePayout, err := s.stripeClient.GetPayout(ctx, payout.StripePayoutID)
	if err != nil {
		s.logger.Error("Failed to get Stripe payout", zap.Error(err), zap.String("stripe_payout_id", payout.StripePayoutID))
		// Continue with local data if Stripe is unavailable
	} else {
		// Update local record with latest Stripe data
		payout.Status = string(stripePayout.Status)
		if stripePayout.ArrivalDate > 0 {
			arrivalTime := time.Unix(stripePayout.ArrivalDate, 0)
			payout.ArrivalDate = &arrivalTime
		}
		s.db.Save(&payout)
	}

	return &payout, nil
}

// GetPayoutHistory retrieves payout history for an account
func (s *PayoutService) GetPayoutHistory(ctx context.Context, connectedAccountID string, page, perPage int) ([]*models.Payout, int64, error) {
	s.logger.Info("Getting payout history",
		zap.String("connected_account_id", connectedAccountID),
		zap.Int("page", page),
		zap.Int("per_page", perPage),
	)

	var payouts []*models.Payout
	var total int64

	// Calculate offset
	offset := (page - 1) * perPage

	// Get total count
	if err := s.db.Model(&models.Payout{}).Where("connected_account_id = ?", connectedAccountID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payouts: %w", err)
	}

	// Get payouts with pagination
	if err := s.db.Where("connected_account_id = ?", connectedAccountID).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&payouts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve payout history: %w", err)
	}

	s.logger.Info("Payout history retrieved successfully",
		zap.String("connected_account_id", connectedAccountID),
		zap.Int("count", len(payouts)),
		zap.Int64("total", total),
	)

	return payouts, total, nil
}

// CancelPayout cancels a pending payout
func (s *PayoutService) CancelPayout(ctx context.Context, payoutID string) (*models.Payout, error) {
	s.logger.Info("Cancelling payout",
		zap.String("payout_id", payoutID),
	)

	// Get payout from database
	var payout models.Payout
	if err := s.db.Where("id = ?", payoutID).First(&payout).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payout not found")
		}
		return nil, fmt.Errorf("failed to retrieve payout: %w", err)
	}

	// Check if payout can be cancelled
	if payout.Status != "pending" {
		return nil, fmt.Errorf("payout cannot be cancelled. Current status: %s", payout.Status)
	}

	// Cancel payout with Stripe
	_, err := s.stripeClient.CancelPayout(ctx, payout.StripePayoutID)
	if err != nil {
		s.logger.Error("Failed to cancel payout with Stripe",
			zap.Error(err),
			zap.String("payout_id", payoutID),
		)
		return nil, fmt.Errorf("failed to cancel payout: %w", err)
	}

	// Update local payout record
	payout.Status = "canceled"
	if err := s.db.Save(&payout).Error; err != nil {
		s.logger.Error("Failed to update payout status in database",
			zap.Error(err),
			zap.String("payout_id", payoutID),
		)
		return nil, fmt.Errorf("failed to update payout: %w", err)
	}

	s.logger.Info("Payout cancelled successfully",
		zap.String("payout_id", payoutID),
	)

	return &payout, nil
}

// GetAvailableBalance checks available balance for payouts
func (s *PayoutService) GetAvailableBalance(ctx context.Context, connectedAccountID string) (*BalanceResponse, error) {
	s.logger.Info("Getting available balance",
		zap.String("connected_account_id", connectedAccountID),
	)

	// Get connected account
	var connectedAccount models.ConnectedAccount
	if err := s.db.Where("id = ?", connectedAccountID).First(&connectedAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("connected account not found")
		}
		return nil, fmt.Errorf("failed to retrieve connected account: %w", err)
	}

	// Get balance from Stripe
	stripeBalance, err := s.stripeClient.GetBalance(ctx, connectedAccount.StripeAccountID)
	if err != nil {
		s.logger.Error("Failed to get balance from Stripe",
			zap.Error(err),
			zap.String("connected_account_id", connectedAccountID),
		)
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert Stripe balance to our response format
	response := &BalanceResponse{
		ConnectedAccountID: connectedAccountID,
		Currency:          "usd", // Default to USD, would need to handle multiple currencies
	}

	// Parse available balance
	for _, available := range stripeBalance.Available {
		response.Available += available.Amount
		response.Currency = string(available.Currency)
	}

	// Parse pending balance
	for _, pending := range stripeBalance.Pending {
		response.Pending += pending.Amount
	}

	s.logger.Info("Balance retrieved successfully",
		zap.String("connected_account_id", connectedAccountID),
		zap.Int64("available", response.Available),
		zap.Int64("pending", response.Pending),
	)

	return response, nil
}

// Helper methods

func (s *PayoutService) validateCreatePayoutRequest(req *CreatePayoutRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("payout amount must be greater than 0")
	}

	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if len(req.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter ISO currency code")
	}

	if req.ConnectedAccountID == "" {
		return fmt.Errorf("connected account ID is required")
	}

	if req.RequestedBy == "" {
		return fmt.Errorf("requested by field is required")
	}

	return nil
}

// Request/Response structs

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

// BalanceResponse represents the balance response
type BalanceResponse struct {
	ConnectedAccountID string `json:"connected_account_id"`
	Available         int64  `json:"available"` // Amount in cents
	Pending          int64  `json:"pending"`  // Amount in cents
	Currency         string  `json:"currency"`
}