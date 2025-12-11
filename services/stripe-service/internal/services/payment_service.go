package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/refund"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentService handles payment processing operations
type PaymentService struct {
	db          *gorm.DB
	stripeClient *StripeClient
	config      *config.StripeConfig
	logger      *zap.Logger
}

// NewPaymentService creates a new payment service instance
func NewPaymentService(db *gorm.DB, stripeClient *StripeClient, config *config.StripeConfig, logger *zap.Logger) *PaymentService {
	return &PaymentService{
		db:          db,
		stripeClient: stripeClient,
		config:      config,
		logger:      logger,
	}
}

// CreatePaymentIntent creates a new Stripe Payment Intent
func (s *PaymentService) CreatePaymentIntent(ctx context.Context, req *CreatePaymentIntentRequest) (*models.PaymentIntent, error) {
	s.logger.Info("Creating payment intent",
		zap.Float64("amount", req.Amount),
		zap.String("currency", req.Currency),
		zap.String("user_id", req.UserID),
		zap.String("connected_account_id", req.ConnectedAccountID),
	)

	// Calculate application fee if connected account is provided
	var applicationFeeAmount float64
	if req.ConnectedAccountID != "" {
		applicationFeeAmount = req.Amount * (s.config.PlatformFeePercent / 100)
	}

	// Create Stripe payment intent parameters
	stripeParams := &PaymentIntentParams{
		Amount:                req.Amount,
		Currency:              req.Currency,
		CustomerID:            req.CustomerID,
		UserID:                req.UserID,
		ConnectedAccountID:    req.ConnectedAccountID,
		ApplicationFeeAmount:  applicationFeeAmount,
		PaymentMethodTypes:    req.PaymentMethodTypes,
	}

	// Create payment intent with Stripe
	stripePI, err := s.stripeClient.CreatePaymentIntent(ctx, stripeParams)
	if err != nil {
		s.logger.Error("Failed to create Stripe payment intent",
			zap.Error(err),
			zap.String("user_id", req.UserID),
		)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Create local payment intent record
	paymentIntent := &models.PaymentIntent{
		StripePaymentIntentID: stripePI.ID,
		Amount:                stripePI.Amount,
		Currency:              string(stripePI.Currency),
		Status:                string(stripePI.Status),
		UserID:                req.UserID,
		ApplicationFeeAmount:  int64(applicationFeeAmount * 100), // Convert to cents
	}

	// Add optional fields
	if req.AuctionID != "" {
		paymentIntent.AuctionID = &req.AuctionID
	}
	if req.ConnectedAccountID != "" {
		paymentIntent.ConnectedAccountID = &req.ConnectedAccountID
	}

	// Add metadata as JSON
	metadata := map[string]interface{}{
		"service":            "blytz-stripe-service",
		"platform_fee_percent": s.config.PlatformFeePercent,
		"requires_action":    stripePI.NextAction != nil,
	}
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}
	metadataJSON, _ := json.Marshal(metadata)
	paymentIntent.Metadata = string(metadataJSON)

	// Save to database
	if err := s.db.Create(paymentIntent).Error; err != nil {
		s.logger.Error("Failed to save payment intent to database",
			zap.Error(err),
			zap.String("stripe_payment_intent_id", stripePI.ID),
		)
		return nil, fmt.Errorf("failed to save payment intent: %w", err)
	}

	s.logger.Info("Payment intent created successfully",
		zap.String("payment_intent_id", paymentIntent.ID),
		zap.String("stripe_payment_intent_id", stripePI.ID),
		zap.Int64("amount", stripePI.Amount),
		zap.String("currency", string(stripePI.Currency)),
	)

	return paymentIntent, nil
}

// ConfirmPayment confirms and captures a payment
func (s *PaymentService) ConfirmPayment(ctx context.Context, req *ConfirmPaymentRequest) (*models.PaymentIntent, error) {
	s.logger.Info("Confirming payment",
		zap.String("payment_intent_id", req.PaymentIntentID),
		zap.String("payment_method_id", req.PaymentMethodID),
	)

	// Get payment intent from database
	var paymentIntent models.PaymentIntent
	if err := s.db.Where("stripe_payment_intent_id = ?", req.PaymentIntentID).First(&paymentIntent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment intent not found")
		}
		return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
	}

	// Confirm payment with Stripe
	stripePI, err := s.stripeClient.ConfirmPaymentIntent(ctx, req.PaymentIntentID, req.PaymentMethodID)
	if err != nil {
		s.logger.Error("Failed to confirm payment with Stripe",
			zap.Error(err),
			zap.String("payment_intent_id", req.PaymentIntentID),
		)
		return nil, fmt.Errorf("failed to confirm payment: %w", err)
	}

	// Update local payment intent record
	updates := map[string]interface{}{
		"status": string(stripePI.Status),
		"updated_at": time.Now(),
	}

	// Update metadata if needed
	if stripePI.NextAction != nil {
		metadata := map[string]interface{}{
			"requires_action": true,
			"next_action_type": string(stripePI.NextAction.Type),
		}
		metadataJSON, _ := json.Marshal(metadata)
		updates["metadata"] = string(metadataJSON)
	}

	if err := s.db.Model(&paymentIntent).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update payment intent in database",
			zap.Error(err),
			zap.String("payment_intent_id", paymentIntent.ID),
		)
		return nil, fmt.Errorf("failed to update payment intent: %w", err)
	}

	// Refresh the payment intent from database
	if err := s.db.Where("id = ?", paymentIntent.ID).First(&paymentIntent).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve updated payment intent: %w", err)
	}

	s.logger.Info("Payment confirmed successfully",
		zap.String("payment_intent_id", paymentIntent.ID),
		zap.String("status", string(stripePI.Status)),
	)

	return &paymentIntent, nil
}

// GetPaymentStatus retrieves the status of a payment
func (s *PaymentService) GetPaymentStatus(ctx context.Context, paymentIntentID string) (*models.PaymentIntent, error) {
	s.logger.Info("Getting payment status",
		zap.String("payment_intent_id", paymentIntentID),
	)

	// Get payment intent from database
	var paymentIntent models.PaymentIntent
	if err := s.db.Where("stripe_payment_intent_id = ?", paymentIntentID).First(&paymentIntent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment intent not found")
		}
		return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
	}

	// Get latest status from Stripe
	stripePI, err := s.stripeClient.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		s.logger.Error("Failed to get payment intent from Stripe",
			zap.Error(err),
			zap.String("payment_intent_id", paymentIntentID),
		)
		return nil, fmt.Errorf("failed to get payment status: %w", err)
	}

	// Update local status if it differs
	if string(stripePI.Status) != paymentIntent.Status {
		if err := s.db.Model(&paymentIntent).Update("status", string(stripePI.Status)).Error; err != nil {
			s.logger.Error("Failed to update payment status in database",
				zap.Error(err),
				zap.String("payment_intent_id", paymentIntent.ID),
			)
		}
		paymentIntent.Status = string(stripePI.Status)
	}

	s.logger.Info("Payment status retrieved successfully",
		zap.String("payment_intent_id", paymentIntent.ID),
		zap.String("status", paymentIntent.Status),
	)

	return &paymentIntent, nil
}

// RefundPayment processes a refund for a payment
func (s *PaymentService) RefundPayment(ctx context.Context, req *RefundPaymentRequest) (*stripe.Refund, error) {
	s.logger.Info("Processing refund",
		zap.String("payment_intent_id", req.PaymentIntentID),
		zap.Float64("amount", req.Amount),
		zap.String("reason", req.Reason),
	)

	// Get payment intent from database
	var paymentIntent models.PaymentIntent
	if err := s.db.Where("stripe_payment_intent_id = ?", req.PaymentIntentID).First(&paymentIntent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment intent not found")
		}
		return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
	}

	// Check if payment is in a refundable state
	if paymentIntent.Status != "succeeded" {
		return nil, fmt.Errorf("payment cannot be refunded. Current status: %s", paymentIntent.Status)
	}

	// Convert amount to cents for Stripe
	amountCents := int64(req.Amount * 100)

	// Create refund parameters
	refundParams := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentIntentID),
		Amount:        stripe.Int64(amountCents),
		Metadata: map[string]string{
			"service":       "blytz-stripe-service",
			"refund_reason": req.Reason,
			"refunded_by":   req.RefundedBy,
		},
	}

	// Add reason if provided
	if req.Reason != "" {
		refundParams.Reason = stripe.String(req.Reason)
	}

	// Process refund with Stripe
	stripeRefund, err := refund.New(refundParams)
	if err != nil {
		s.logger.Error("Failed to process refund with Stripe",
			zap.Error(err),
			zap.String("payment_intent_id", req.PaymentIntentID),
		)
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	s.logger.Info("Refund processed successfully",
		zap.String("refund_id", stripeRefund.ID),
		zap.String("payment_intent_id", req.PaymentIntentID),
		zap.Int64("amount", stripeRefund.Amount),
		zap.String("status", string(stripeRefund.Status)),
	)

	return stripeRefund, nil
}

// GetPaymentHistory retrieves payment history for a user
func (s *PaymentService) GetPaymentHistory(ctx context.Context, userID string, page, perPage int) ([]*models.PaymentIntent, int64, error) {
	s.logger.Info("Getting payment history",
		zap.String("user_id", userID),
		zap.Int("page", page),
		zap.Int("per_page", perPage),
	)

	var paymentIntents []*models.PaymentIntent
	var total int64

	// Calculate offset
	offset := (page - 1) * perPage

	// Get total count
	if err := s.db.Model(&models.PaymentIntent{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payment intents: %w", err)
	}

	// Get payment intents with pagination
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&paymentIntents).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve payment history: %w", err)
	}

	s.logger.Info("Payment history retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(paymentIntents)),
		zap.Int64("total", total),
	)

	return paymentIntents, total, nil
}

// Request/Response structs for payment service

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

// RefundPaymentRequest represents a request to refund a payment
type RefundPaymentRequest struct {
	PaymentIntentID string  `json:"payment_intent_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,min=0.01"`
	Reason          string  `json:"reason" binding:"omitempty,oneof=duplicate fraudulent requested_by_customer"`
	RefundedBy      string  `json:"refunded_by" binding:"required,uuid"`
}