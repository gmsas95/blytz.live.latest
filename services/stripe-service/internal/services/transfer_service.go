package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TransferService handles transfer operations
type TransferService struct {
	db           *gorm.DB
	stripeClient *StripeClient
	config       *config.StripeConfig
	logger       *zap.Logger
}

// NewTransferService creates a new transfer service instance
func NewTransferService(db *gorm.DB, stripeClient *StripeClient, config *config.StripeConfig, logger *zap.Logger) *TransferService {
	return &TransferService{
		db:           db,
		stripeClient: stripeClient,
		config:       config,
		logger:       logger,
	}
}

// CreateTransfer creates a new transfer to a connected account
func (s *TransferService) CreateTransfer(ctx context.Context, req *CreateTransferRequest) (*models.Transfer, error) {
	s.logger.Info("Creating transfer",
		zap.Float64("amount", req.Amount),
		zap.String("currency", req.Currency),
		zap.String("connected_account_id", req.ConnectedAccountID),
	)

	// Validate request
	if err := s.validateCreateTransferRequest(req); err != nil {
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

	// Calculate transfer amount after fees
	transferAmount := s.CalculateTransferAmount(req.Amount, req.FeePercent)

	// Create Stripe transfer parameters
	stripeParams := &TransferParams{
		Amount:            transferAmount,
		Currency:          req.Currency,
		Destination:       connectedAccount.StripeAccountID,
		TransferGroup:     req.TransferGroup,
		SourceTransaction: req.SourceTransaction,
	}

	// Create transfer with Stripe
	stripeTransfer, err := s.stripeClient.CreateTransfer(ctx, stripeParams)
	if err != nil {
		s.logger.Error("Failed to create Stripe transfer",
			zap.Error(err),
			zap.String("connected_account_id", req.ConnectedAccountID),
		)
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}

	// Create local transfer record
	transfer := &models.Transfer{
		StripeTransferID:     stripeTransfer.ID,
		ConnectedAccountID:   req.ConnectedAccountID,
		Amount:               int64(transferAmount * 100), // Convert to cents
		Currency:             req.Currency,
		Status:               string(stripeTransfer.DestinationPayment.Status),
		DestinationPaymentID: req.DestinationPaymentID,
	}

	// Add metadata as JSON
	metadata := map[string]interface{}{
		"service":         "blytz-stripe-service",
		"original_amount": req.Amount,
		"fee_percent":     req.FeePercent,
		"transfer_amount": transferAmount,
	}
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}
	metadataJSON, _ := json.Marshal(metadata)
	transfer.Metadata = string(metadataJSON)

	// Save to database
	if err := s.db.Create(transfer).Error; err != nil {
		s.logger.Error("Failed to save transfer to database",
			zap.Error(err),
			zap.String("stripe_transfer_id", stripeTransfer.ID),
		)
		return nil, fmt.Errorf("failed to save transfer: %w", err)
	}

	s.logger.Info("Transfer created successfully",
		zap.String("transfer_id", transfer.ID),
		zap.String("stripe_transfer_id", stripeTransfer.ID),
		zap.Float64("amount", transferAmount),
		zap.String("currency", req.Currency),
	)

	return transfer, nil
}

// CreateBatchTransfers creates multiple transfers in a batch
func (s *TransferService) CreateBatchTransfers(ctx context.Context, reqs []*CreateTransferRequest) ([]*models.Transfer, error) {
	s.logger.Info("Creating batch transfers",
		zap.Int("count", len(reqs)),
	)

	var transfers []*models.Transfer
	var errors []error

	for i, req := range reqs {
		transfer, err := s.CreateTransfer(ctx, req)
		if err != nil {
			s.logger.Error("Failed to create transfer in batch",
				zap.Error(err),
				zap.Int("index", i),
				zap.String("connected_account_id", req.ConnectedAccountID),
			)
			errors = append(errors, fmt.Errorf("transfer %d failed: %w", i+1, err))
			continue
		}
		transfers = append(transfers, transfer)
	}

	if len(errors) > 0 {
		return transfers, fmt.Errorf("batch transfer completed with %d errors: %v", len(errors), errors)
	}

	s.logger.Info("Batch transfers created successfully",
		zap.Int("success_count", len(transfers)),
		zap.Int("total_count", len(reqs)),
	)

	return transfers, nil
}

// GetTransfer retrieves a transfer by ID
func (s *TransferService) GetTransfer(ctx context.Context, transferID string) (*models.Transfer, error) {
	s.logger.Info("Retrieving transfer",
		zap.String("transfer_id", transferID),
	)

	var transfer models.Transfer
	if err := s.db.Where("id = ?", transferID).First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transfer not found")
		}
		return nil, fmt.Errorf("failed to retrieve transfer: %w", err)
	}

	// Get latest data from Stripe
	stripeTransfer, err := s.stripeClient.GetTransfer(ctx, transfer.StripeTransferID)
	if err != nil {
		s.logger.Error("Failed to get Stripe transfer", zap.Error(err), zap.String("stripe_transfer_id", transfer.StripeTransferID))
		// Continue with local data if Stripe is unavailable
	} else {
		// Update local record with latest Stripe data
		transfer.Status = string(stripeTransfer.DestinationPayment.Status)
		s.db.Save(&transfer)
	}

	return &transfer, nil
}

// GetTransferHistory retrieves transfer history for an account
func (s *TransferService) GetTransferHistory(ctx context.Context, connectedAccountID string, page, perPage int) ([]*models.Transfer, int64, error) {
	s.logger.Info("Getting transfer history",
		zap.String("connected_account_id", connectedAccountID),
		zap.Int("page", page),
		zap.Int("per_page", perPage),
	)

	var transfers []*models.Transfer
	var total int64

	// Calculate offset
	offset := (page - 1) * perPage

	// Get total count
	if err := s.db.Model(&models.Transfer{}).Where("connected_account_id = ?", connectedAccountID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transfers: %w", err)
	}

	// Get transfers with pagination
	if err := s.db.Where("connected_account_id = ?", connectedAccountID).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&transfers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve transfer history: %w", err)
	}

	s.logger.Info("Transfer history retrieved successfully",
		zap.String("connected_account_id", connectedAccountID),
		zap.Int("count", len(transfers)),
		zap.Int64("total", total),
	)

	return transfers, total, nil
}

// ReverseTransfer reverses a transfer if needed
func (s *TransferService) ReverseTransfer(ctx context.Context, transferID string, amount float64) (*models.Transfer, error) {
	s.logger.Info("Reversing transfer",
		zap.String("transfer_id", transferID),
		zap.Float64("amount", amount),
	)

	// Get transfer from database
	var transfer models.Transfer
	if err := s.db.Where("id = ?", transferID).First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transfer not found")
		}
		return nil, fmt.Errorf("failed to retrieve transfer: %w", err)
	}

	// Check if transfer can be reversed
	if transfer.Status != "paid" && transfer.Status != "complete" {
		return nil, fmt.Errorf("transfer cannot be reversed. Current status: %s", transfer.Status)
	}

	// Convert amount to cents for Stripe
	amountCents := int64(amount * 100)

	// Reverse transfer with Stripe
	reversal, err := s.stripeClient.ReverseTransfer(ctx, transfer.StripeTransferID, amountCents)
	if err != nil {
		s.logger.Error("Failed to reverse transfer with Stripe",
			zap.Error(err),
			zap.String("transfer_id", transferID),
		)
		return nil, fmt.Errorf("failed to reverse transfer: %w", err)
	}

	s.logger.Info("Stripe transfer reversed",
		zap.String("reversal_id", reversal.ID),
		zap.String("transfer_id", transfer.StripeTransferID),
	)

	// Update local transfer record
	transfer.Status = "reversed"
	if err := s.db.Save(&transfer).Error; err != nil {
		s.logger.Error("Failed to update transfer status in database",
			zap.Error(err),
			zap.String("transfer_id", transferID),
		)
		return nil, fmt.Errorf("failed to update transfer: %w", err)
	}

	s.logger.Info("Transfer reversed successfully",
		zap.String("transfer_id", transferID),
		zap.Float64("amount", amount),
	)

	return &transfer, nil
}

// CalculateTransferAmount calculates the amount after fees
func (s *TransferService) CalculateTransferAmount(amount, feePercent float64) float64 {
	feeAmount := amount * (feePercent / 100)
	return amount - feeAmount
}

// Helper methods

func (s *TransferService) validateCreateTransferRequest(req *CreateTransferRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("transfer amount must be greater than 0")
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

	if req.FeePercent < 0 || req.FeePercent > 100 {
		return fmt.Errorf("fee percent must be between 0 and 100")
	}

	return nil
}

// Request structs

// CreateTransferRequest represents a request to create a transfer
type CreateTransferRequest struct {
	Amount               float64           `json:"amount" binding:"required,min=0.01"`
	Currency             string            `json:"currency" binding:"required,len=3"`
	ConnectedAccountID   string            `json:"connected_account_id" binding:"required,uuid"`
	TransferGroup        string            `json:"transfer_group,omitempty"`
	SourceTransaction    string            `json:"source_transaction,omitempty"`
	DestinationPaymentID string            `json:"destination_payment_id,omitempty"`
	FeePercent           float64           `json:"fee_percent" binding:"min=0,max=100"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}
