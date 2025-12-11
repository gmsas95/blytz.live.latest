package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebhookService handles Stripe webhook events
type WebhookService struct {
	db                *gorm.DB
	stripeClient      *StripeClient
	config            *config.StripeConfig
	logger            *zap.Logger
	eventProcessor    *WebhookEventProcessor
}

// NewWebhookService creates a new webhook service
func NewWebhookService(db *gorm.DB, stripeClient *StripeClient, config *config.StripeConfig, logger *zap.Logger) *WebhookService {
	eventProcessor := NewWebhookEventProcessor(db, stripeClient, logger)
	
	return &WebhookService{
		db:             db,
		stripeClient:   stripeClient,
		config:         config,
		logger:         logger,
		eventProcessor: eventProcessor,
	}
}

// ProcessWebhookEvent processes an incoming webhook event
func (s *WebhookService) ProcessWebhookEvent(ctx context.Context, payload []byte, signatureHeader string) error {
	// Verify webhook signature
	event, err := s.VerifyWebhookSignature(payload, signatureHeader)
	if err != nil {
		s.logger.Error("Webhook signature verification failed", zap.Error(err))
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	// Check if event has already been processed (idempotency)
	var existingEvent models.StripeEvent
	result := s.db.Where("stripe_event_id = ?", event.ID).First(&existingEvent)
	if result.Error == nil {
		s.logger.Info("Event already processed", zap.String("event_id", event.ID))
		return nil
	} else if result.Error != gorm.ErrRecordNotFound {
		s.logger.Error("Error checking for existing event", zap.Error(result.Error))
		return fmt.Errorf("error checking for existing event: %w", result.Error)
	}

	// Store the event in the database
	stripeEvent := &models.StripeEvent{
		StripeEventID: event.ID,
		EventType:     string(event.Type),
		Processed:     false,
		EventData:     string(payload),
	}

	if err := s.db.Create(stripeEvent).Error; err != nil {
		s.logger.Error("Failed to store webhook event", zap.Error(err))
		return fmt.Errorf("failed to store webhook event: %w", err)
	}

	// Process the event based on its type
	var processErr error
	switch event.Type {
	case stripe.EventTypePaymentIntentSucceeded,
		 stripe.EventTypePaymentIntentPaymentFailed,
		 stripe.EventTypePaymentIntentCanceled:
		processErr = s.HandlePaymentIntentEvent(ctx, event)
	
	case stripe.EventTypeAccountUpdated,
		 stripe.EventTypeAccountApplicationAuthorized,
		 stripe.EventTypeAccountApplicationDeauthorized:
		processErr = s.HandleAccountEvent(ctx, event)
	
	case stripe.EventTypeTransferCreated:
		processErr = s.HandleTransferEvent(ctx, event)
	
	case stripe.EventTypePayoutCreated,
		 stripe.EventTypePayoutPaid,
		 stripe.EventTypePayoutFailed:
		processErr = s.HandlePayoutEvent(ctx, event)
	
	case stripe.EventTypeApplicationFeeCreated,
		 stripe.EventTypeApplicationFeeRefunded:
		processErr = s.HandleApplicationFeeEvent(ctx, event)
	
	default:
		s.logger.Info("Unhandled event type", zap.String("event_type", string(event.Type)))
	}

	// Update the event status
	updates := map[string]interface{}{
		"processed": true,
	}
	
	if processErr != nil {
		updates["error_message"] = processErr.Error()
		s.logger.Error("Error processing webhook event", 
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.Type)),
			zap.Error(processErr))
	}

	if err := s.db.Model(stripeEvent).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update webhook event status", zap.Error(err))
		return fmt.Errorf("failed to update webhook event status: %w", err)
	}

	return processErr
}

// VerifyWebhookSignature verifies the Stripe webhook signature
func (s *WebhookService) VerifyWebhookSignature(payload []byte, signatureHeader string) (*stripe.Event, error) {
	// Use Stripe's webhook utility for verification
	event, err := webhook.ConstructEvent(payload, signatureHeader, s.config.WebhookSecret)
	if err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", err)
	}

	return &event, nil
}

// HandlePaymentIntentEvent handles payment intent events
func (s *WebhookService) HandlePaymentIntentEvent(ctx context.Context, event *stripe.Event) error {
	s.logger.Info("Processing payment intent event", 
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)))

	return s.eventProcessor.ProcessPaymentIntentEvent(ctx, event)
}

// HandleAccountEvent handles connected account events
func (s *WebhookService) HandleAccountEvent(ctx context.Context, event *stripe.Event) error {
	s.logger.Info("Processing account event", 
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)))

	return s.eventProcessor.ProcessAccountEvent(ctx, event)
}

// HandleTransferEvent handles transfer events
func (s *WebhookService) HandleTransferEvent(ctx context.Context, event *stripe.Event) error {
	s.logger.Info("Processing transfer event", 
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)))

	return s.eventProcessor.ProcessTransferEvent(ctx, event)
}

// HandlePayoutEvent handles payout events
func (s *WebhookService) HandlePayoutEvent(ctx context.Context, event *stripe.Event) error {
	s.logger.Info("Processing payout event", 
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)))

	return s.eventProcessor.ProcessPayoutEvent(ctx, event)
}

// HandleApplicationFeeEvent handles application fee events
func (s *WebhookService) HandleApplicationFeeEvent(ctx context.Context, event *stripe.Event) error {
	s.logger.Info("Processing application fee event", 
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)))

	return s.eventProcessor.ProcessApplicationFeeEvent(ctx, event)
}

// GetEventByID retrieves a webhook event by ID
func (s *WebhookService) GetEventByID(eventID string) (*models.StripeEvent, error) {
	var stripeEvent models.StripeEvent
	result := s.db.Where("stripe_event_id = ?", eventID).First(&stripeEvent)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve webhook event: %w", result.Error)
	}
	return &stripeEvent, nil
}

// ListEvents retrieves webhook events with pagination
func (s *WebhookService) ListEvents(limit, offset int, eventType string) ([]models.StripeEvent, error) {
	var events []models.StripeEvent
	query := s.db.Order("created_at DESC")
	
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	result := query.Find(&events)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list webhook events: %w", result.Error)
	}
	
	return events, nil
}

// ReplayEvent replays a previously processed event
func (s *WebhookService) ReplayEvent(ctx context.Context, eventID string) error {
	// Get the event from the database
	stripeEvent, err := s.GetEventByID(eventID)
	if err != nil {
		return fmt.Errorf("failed to retrieve event for replay: %w", err)
	}

	// Parse the event data
	var event stripe.Event
	if err := json.Unmarshal([]byte(stripeEvent.EventData), &event); err != nil {
		return fmt.Errorf("failed to parse event data: %w", err)
	}

	// Process the event again
	var processErr error
	switch event.Type {
	case stripe.EventTypePaymentIntentSucceeded,
		 stripe.EventTypePaymentIntentPaymentFailed,
		 stripe.EventTypePaymentIntentCanceled:
		processErr = s.HandlePaymentIntentEvent(ctx, &event)
	
	case stripe.EventTypeAccountUpdated,
		 stripe.EventTypeAccountApplicationAuthorized,
		 stripe.EventTypeAccountApplicationDeauthorized:
		processErr = s.HandleAccountEvent(ctx, &event)
	
	case stripe.EventTypeTransferCreated:
		processErr = s.HandleTransferEvent(ctx, &event)
	
	case stripe.EventTypePayoutCreated,
		 stripe.EventTypePayoutPaid,
		 stripe.EventTypePayoutFailed:
		processErr = s.HandlePayoutEvent(ctx, &event)
	
	case stripe.EventTypeApplicationFeeCreated,
		 stripe.EventTypeApplicationFeeRefunded:
		processErr = s.HandleApplicationFeeEvent(ctx, &event)
	
	default:
		s.logger.Info("Unhandled event type during replay", zap.String("event_type", string(event.Type)))
	}

	return processErr
}