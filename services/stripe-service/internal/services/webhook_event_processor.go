package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"github.com/stripe/stripe-go/v84"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebhookEventProcessor handles specific Stripe webhook event types
type WebhookEventProcessor struct {
	db           *gorm.DB
	stripeClient *StripeClient
	logger       *zap.Logger
}

// NewWebhookEventProcessor creates a new webhook event processor
func NewWebhookEventProcessor(db *gorm.DB, stripeClient *StripeClient, logger *zap.Logger) *WebhookEventProcessor {
	return &WebhookEventProcessor{
		db:           db,
		stripeClient: stripeClient,
		logger:       logger,
	}
}

// ProcessPaymentIntentEvent processes payment intent events
func (p *WebhookEventProcessor) ProcessPaymentIntentEvent(ctx context.Context, event *stripe.Event) error {
	switch event.Type {
	case stripe.EventTypePaymentIntentSucceeded:
		return p.handlePaymentIntentSucceeded(ctx, event)
	case stripe.EventTypePaymentIntentPaymentFailed:
		return p.handlePaymentIntentPaymentFailed(ctx, event)
	case stripe.EventTypePaymentIntentCanceled:
		return p.handlePaymentIntentCanceled(ctx, event)
	default:
		p.logger.Warn("Unknown payment intent event type", zap.String("event_type", string(event.Type)))
		return nil
	}
}

// ProcessAccountEvent processes connected account events
func (p *WebhookEventProcessor) ProcessAccountEvent(ctx context.Context, event *stripe.Event) error {
	switch event.Type {
	case stripe.EventTypeAccountUpdated:
		return p.handleAccountUpdated(ctx, event)
	case stripe.EventTypeAccountApplicationAuthorized:
		return p.handleAccountApplicationAuthorized(ctx, event)
	case stripe.EventTypeAccountApplicationDeauthorized:
		return p.handleAccountApplicationDeauthorized(ctx, event)
	default:
		p.logger.Warn("Unknown account event type", zap.String("event_type", string(event.Type)))
		return nil
	}
}

// ProcessTransferEvent processes transfer events
func (p *WebhookEventProcessor) ProcessTransferEvent(ctx context.Context, event *stripe.Event) error {
	switch event.Type {
	case stripe.EventTypeTransferCreated:
		return p.handleTransferCreated(ctx, event)
	case "transfer.completed":
		return p.handleTransferCompleted(ctx, event)
	case "transfer.failed":
		return p.handleTransferFailed(ctx, event)
	default:
		p.logger.Warn("Unknown transfer event type", zap.String("event_type", string(event.Type)))
		return nil
	}
}

// ProcessPayoutEvent processes payout events
func (p *WebhookEventProcessor) ProcessPayoutEvent(ctx context.Context, event *stripe.Event) error {
	switch event.Type {
	case stripe.EventTypePayoutCreated:
		return p.handlePayoutCreated(ctx, event)
	case stripe.EventTypePayoutPaid:
		return p.handlePayoutPaid(ctx, event)
	case stripe.EventTypePayoutFailed:
		return p.handlePayoutFailed(ctx, event)
	default:
		p.logger.Warn("Unknown payout event type", zap.String("event_type", string(event.Type)))
		return nil
	}
}

// ProcessApplicationFeeEvent processes application fee events
func (p *WebhookEventProcessor) ProcessApplicationFeeEvent(ctx context.Context, event *stripe.Event) error {
	switch event.Type {
	case stripe.EventTypeApplicationFeeCreated:
		return p.handleApplicationFeeCreated(ctx, event)
	case stripe.EventTypeApplicationFeeRefunded:
		return p.handleApplicationFeeRefunded(ctx, event)
	default:
		p.logger.Warn("Unknown application fee event type", zap.String("event_type", string(event.Type)))
		return nil
	}
}

// Payment Intent Event Handlers

func (p *WebhookEventProcessor) handlePaymentIntentSucceeded(ctx context.Context, event *stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to unmarshal payment intent: %w", err)
	}

	p.logger.Info("Payment intent succeeded",
		zap.String("payment_intent_id", paymentIntent.ID),
		zap.Int64("amount", paymentIntent.Amount),
		zap.String("currency", string(paymentIntent.Currency)))

	// Update payment intent in database
	var dbPaymentIntent models.PaymentIntent
	result := p.db.Where("stripe_payment_intent_id = ?", paymentIntent.ID).First(&dbPaymentIntent)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new payment intent record if not exists
			dbPaymentIntent = models.PaymentIntent{
				StripePaymentIntentID: paymentIntent.ID,
				Amount:               paymentIntent.Amount,
				Currency:             string(paymentIntent.Currency),
				Status:               string(paymentIntent.Status),
				Metadata:             p.getMetadataJSON(paymentIntent.Metadata),
			}

			// Extract user ID from metadata
			if userID, ok := paymentIntent.Metadata["user_id"]; ok {
				dbPaymentIntent.UserID = userID
			}

			// Extract auction ID from metadata
			if auctionID, ok := paymentIntent.Metadata["auction_id"]; ok {
				dbPaymentIntent.AuctionID = &auctionID
			}

			// Extract connected account ID from metadata
			if connectedAccountID, ok := paymentIntent.Metadata["connected_account"]; ok {
				dbPaymentIntent.ConnectedAccountID = &connectedAccountID
			}

			// Extract application fee amount
			if paymentIntent.ApplicationFeeAmount > 0 {
				dbPaymentIntent.ApplicationFeeAmount = paymentIntent.ApplicationFeeAmount
			}

			if err := p.db.Create(&dbPaymentIntent).Error; err != nil {
				return fmt.Errorf("failed to create payment intent: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query payment intent: %w", result.Error)
		}
	} else {
		// Update existing payment intent
		updates := map[string]interface{}{
			"status":   string(paymentIntent.Status),
			"metadata": p.getMetadataJSON(paymentIntent.Metadata),
		}

		if paymentIntent.ApplicationFeeAmount > 0 {
			updates["application_fee_amount"] = paymentIntent.ApplicationFeeAmount
		}

		if err := p.db.Model(&dbPaymentIntent).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update payment intent: %w", err)
		}
	}

	return nil
}

func (p *WebhookEventProcessor) handlePaymentIntentPaymentFailed(ctx context.Context, event *stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to unmarshal payment intent: %w", err)
	}

	p.logger.Info("Payment intent failed",
		zap.String("payment_intent_id", paymentIntent.ID),
		zap.String("last_payment_error", paymentIntent.LastPaymentError.Error()))

	// Update payment intent status in database
	result := p.db.Model(&models.PaymentIntent{}).
		Where("stripe_payment_intent_id = ?", paymentIntent.ID).
		Updates(map[string]interface{}{
			"status": string(paymentIntent.Status),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update payment intent status: %w", result.Error)
	}

	return nil
}

func (p *WebhookEventProcessor) handlePaymentIntentCanceled(ctx context.Context, event *stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to unmarshal payment intent: %w", err)
	}

	p.logger.Info("Payment intent canceled",
		zap.String("payment_intent_id", paymentIntent.ID))

	// Update payment intent status in database
	result := p.db.Model(&models.PaymentIntent{}).
		Where("stripe_payment_intent_id = ?", paymentIntent.ID).
		Updates(map[string]interface{}{
			"status": string(paymentIntent.Status),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update payment intent status: %w", result.Error)
	}

	return nil
}

// Account Event Handlers

func (p *WebhookEventProcessor) handleAccountUpdated(ctx context.Context, event *stripe.Event) error {
	var account stripe.Account
	if err := json.Unmarshal(event.Data.Raw, &account); err != nil {
		return fmt.Errorf("failed to unmarshal account: %w", err)
	}

	p.logger.Info("Account updated",
		zap.String("account_id", account.ID),
		zap.Bool("charges_enabled", account.ChargesEnabled),
		zap.Bool("payouts_enabled", account.PayoutsEnabled))

	// Update connected account in database
	updates := map[string]interface{}{
		"charges_enabled": account.ChargesEnabled,
		"payouts_enabled": account.PayoutsEnabled,
		"metadata":        p.getMetadataJSON(account.Metadata),
	}

	// Update verification status if available
	if account.Requirements != nil {
		if account.Requirements.CurrentlyDue != nil && len(account.Requirements.CurrentlyDue) == 0 {
			updates["verification_status"] = "verified"
		} else {
			updates["verification_status"] = "pending"
		}
	}

	result := p.db.Model(&models.ConnectedAccount{}).
		Where("stripe_account_id = ?", account.ID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update connected account: %w", result.Error)
	}

	return nil
}

func (p *WebhookEventProcessor) handleAccountApplicationAuthorized(ctx context.Context, event *stripe.Event) error {
	var account stripe.Account
	if err := json.Unmarshal(event.Data.Raw, &account); err != nil {
		return fmt.Errorf("failed to unmarshal account: %w", err)
	}

	p.logger.Info("Account application authorized",
		zap.String("account_id", account.ID))

	// Update connected account status
	result := p.db.Model(&models.ConnectedAccount{}).
		Where("stripe_account_id = ?", account.ID).
		Updates(map[string]interface{}{
			"verification_status": "authorized",
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update connected account: %w", result.Error)
	}

	return nil
}

func (p *WebhookEventProcessor) handleAccountApplicationDeauthorized(ctx context.Context, event *stripe.Event) error {
	var account stripe.Account
	if err := json.Unmarshal(event.Data.Raw, &account); err != nil {
		return fmt.Errorf("failed to unmarshal account: %w", err)
	}

	p.logger.Info("Account application deauthorized",
		zap.String("account_id", account.ID))

	// Update connected account status
	result := p.db.Model(&models.ConnectedAccount{}).
		Where("stripe_account_id = ?", account.ID).
		Updates(map[string]interface{}{
			"verification_status": "deauthorized",
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update connected account: %w", result.Error)
	}

	return nil
}

// Transfer Event Handlers

func (p *WebhookEventProcessor) handleTransferCreated(ctx context.Context, event *stripe.Event) error {
	var transfer stripe.Transfer
	if err := json.Unmarshal(event.Data.Raw, &transfer); err != nil {
		return fmt.Errorf("failed to unmarshal transfer: %w", err)
	}

	p.logger.Info("Transfer created",
		zap.String("transfer_id", transfer.ID),
		zap.Int64("amount", transfer.Amount),
		zap.String("destination", transfer.Destination.ID))

	// Create or update transfer in database
	var dbTransfer models.Transfer
	result := p.db.Where("stripe_transfer_id = ?", transfer.ID).First(&dbTransfer)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new transfer record
			dbTransfer = models.Transfer{
				StripeTransferID:     transfer.ID,
				ConnectedAccountID:   transfer.Destination.ID,
				Amount:               transfer.Amount,
				Currency:             string(transfer.Currency),
				Status:               string(transfer.DestinationPayment.Status),
				DestinationPaymentID: transfer.SourceTransaction.ID,
				Metadata:             p.getMetadataJSON(transfer.Metadata),
			}

			if err := p.db.Create(&dbTransfer).Error; err != nil {
				return fmt.Errorf("failed to create transfer: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query transfer: %w", result.Error)
		}
	} else {
		// Update existing transfer
		updates := map[string]interface{}{
			"status":   string(transfer.DestinationPayment.Status),
			"metadata": p.getMetadataJSON(transfer.Metadata),
		}

		if transfer.SourceTransaction != nil {
			updates["destination_payment_id"] = transfer.SourceTransaction.ID
		}

		if err := p.db.Model(&dbTransfer).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update transfer: %w", err)
		}
	}

	return nil
}

func (p *WebhookEventProcessor) handleTransferCompleted(ctx context.Context, event *stripe.Event) error {
	var transfer stripe.Transfer
	if err := json.Unmarshal(event.Data.Raw, &transfer); err != nil {
		return fmt.Errorf("failed to unmarshal transfer: %w", err)
	}

	p.logger.Info("Transfer completed",
		zap.String("transfer_id", transfer.ID))

	// Update transfer status in database
	result := p.db.Model(&models.Transfer{}).
		Where("stripe_transfer_id = ?", transfer.ID).
		Updates(map[string]interface{}{
			"status": string(transfer.DestinationPayment.Status),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update transfer status: %w", result.Error)
	}

	return nil
}

func (p *WebhookEventProcessor) handleTransferFailed(ctx context.Context, event *stripe.Event) error {
	var transfer stripe.Transfer
	if err := json.Unmarshal(event.Data.Raw, &transfer); err != nil {
		return fmt.Errorf("failed to unmarshal transfer: %w", err)
	}

	p.logger.Info("Transfer failed",
		zap.String("transfer_id", transfer.ID))

	// Update transfer status in database
	result := p.db.Model(&models.Transfer{}).
		Where("stripe_transfer_id = ?", transfer.ID).
		Updates(map[string]interface{}{
			"status": string(transfer.DestinationPayment.Status),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update transfer status: %w", result.Error)
	}

	return nil
}

// Payout Event Handlers

func (p *WebhookEventProcessor) handlePayoutCreated(ctx context.Context, event *stripe.Event) error {
	var payout stripe.Payout
	if err := json.Unmarshal(event.Data.Raw, &payout); err != nil {
		return fmt.Errorf("failed to unmarshal payout: %w", err)
	}

	p.logger.Info("Payout created",
		zap.String("payout_id", payout.ID),
		zap.Int64("amount", payout.Amount),
		zap.String("destination", payout.Destination.ID))

	// Create or update payout in database
	var dbPayout models.Payout
	result := p.db.Where("stripe_payout_id = ?", payout.ID).First(&dbPayout)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new payout record
			dbPayout = models.Payout{
				StripePayoutID:     payout.ID,
				ConnectedAccountID: payout.Destination.ID,
				Amount:             payout.Amount,
				Currency:           string(payout.Currency),
				Status:             string(payout.Status),
				Metadata:           p.getMetadataJSON(payout.Metadata),
			}

			if payout.ArrivalDate != 0 {
				arrivalTime := time.Unix(payout.ArrivalDate, 0)
				dbPayout.ArrivalDate = &arrivalTime
			}

			if err := p.db.Create(&dbPayout).Error; err != nil {
				return fmt.Errorf("failed to create payout: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query payout: %w", result.Error)
		}
	} else {
		// Update existing payout
		updates := map[string]interface{}{
			"status":   string(payout.Status),
			"metadata": p.getMetadataJSON(payout.Metadata),
		}

		if payout.ArrivalDate != 0 {
			arrivalTime := time.Unix(payout.ArrivalDate, 0)
			updates["arrival_date"] = &arrivalTime
		}

		if err := p.db.Model(&dbPayout).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update payout: %w", err)
		}
	}

	return nil
}

func (p *WebhookEventProcessor) handlePayoutPaid(ctx context.Context, event *stripe.Event) error {
	var payout stripe.Payout
	if err := json.Unmarshal(event.Data.Raw, &payout); err != nil {
		return fmt.Errorf("failed to unmarshal payout: %w", err)
	}

	p.logger.Info("Payout paid",
		zap.String("payout_id", payout.ID))

	// Update payout status in database
	result := p.db.Model(&models.Payout{}).
		Where("stripe_payout_id = ?", payout.ID).
		Updates(map[string]interface{}{
			"status": string(payout.Status),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update payout status: %w", result.Error)
	}

	return nil
}

func (p *WebhookEventProcessor) handlePayoutFailed(ctx context.Context, event *stripe.Event) error {
	var payout stripe.Payout
	if err := json.Unmarshal(event.Data.Raw, &payout); err != nil {
		return fmt.Errorf("failed to unmarshal payout: %w", err)
	}

	p.logger.Info("Payout failed",
		zap.String("payout_id", payout.ID))

	// Update payout status in database
	result := p.db.Model(&models.Payout{}).
		Where("stripe_payout_id = ?", payout.ID).
		Updates(map[string]interface{}{
			"status": string(payout.Status),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update payout status: %w", result.Error)
	}

	return nil
}

// Application Fee Event Handlers

func (p *WebhookEventProcessor) handleApplicationFeeCreated(ctx context.Context, event *stripe.Event) error {
	var fee stripe.ApplicationFee
	if err := json.Unmarshal(event.Data.Raw, &fee); err != nil {
		return fmt.Errorf("failed to unmarshal application fee: %w", err)
	}

	p.logger.Info("Application fee created",
		zap.String("fee_id", fee.ID),
		zap.Int64("amount", fee.Amount),
		zap.String("payment_intent_id", fee.FeeSource.Charge))

	// Create or update application fee in database
	var dbFee models.ApplicationFee
	result := p.db.Where("stripe_fee_id = ?", fee.ID).First(&dbFee)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new application fee record
			dbFee = models.ApplicationFee{
				StripeFeeID:     fee.ID,
				PaymentIntentID: fee.FeeSource.Charge,
				Amount:          fee.Amount,
				Currency:        string(fee.Currency),
			}

			if err := p.db.Create(&dbFee).Error; err != nil {
				return fmt.Errorf("failed to create application fee: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query application fee: %w", result.Error)
		}
	}

	return nil
}

func (p *WebhookEventProcessor) handleApplicationFeeRefunded(ctx context.Context, event *stripe.Event) error {
	var fee stripe.ApplicationFee
	if err := json.Unmarshal(event.Data.Raw, &fee); err != nil {
		return fmt.Errorf("failed to unmarshal application fee: %w", err)
	}

	p.logger.Info("Application fee refunded",
		zap.String("fee_id", fee.ID))

	// Note: Application fee refunds are handled by Stripe automatically
	// We might want to track refund status in the future
	return nil
}

// Helper function to convert metadata to JSON string
func (p *WebhookEventProcessor) getMetadataJSON(metadata map[string]string) string {
	if metadata == nil {
		return "{}"
	}
	
	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		p.logger.Error("Failed to marshal metadata", zap.Error(err))
		return "{}"
	}
	
	return string(jsonBytes)
}