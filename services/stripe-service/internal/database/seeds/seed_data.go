package seeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Seeder handles database seeding
type Seeder struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSeeder creates a new seeder instance
func NewSeeder(db *gorm.DB, logger *zap.Logger) *Seeder {
	return &Seeder{
		db:     db,
		logger: logger,
	}
}

// SeedAll seeds all test data
func (s *Seeder) SeedAll() error {
	s.logger.Info("Starting database seeding")
	
	if err := s.SeedConnectedAccounts(); err != nil {
		return fmt.Errorf("failed to seed connected accounts: %w", err)
	}
	
	if err := s.SeedPaymentIntents(); err != nil {
		return fmt.Errorf("failed to seed payment intents: %w", err)
	}
	
	if err := s.SeedTransfers(); err != nil {
		return fmt.Errorf("failed to seed transfers: %w", err)
	}
	
	if err := s.SeedPayouts(); err != nil {
		return fmt.Errorf("failed to seed payouts: %w", err)
	}
	
	if err := s.SeedApplicationFees(); err != nil {
		return fmt.Errorf("failed to seed application fees: %w", err)
	}
	
	if err := s.SeedStripeEvents(); err != nil {
		return fmt.Errorf("failed to seed stripe events: %w", err)
	}
	
	s.logger.Info("Database seeding completed successfully")
	return nil
}

// SeedConnectedAccounts creates sample connected accounts
func (s *Seeder) SeedConnectedAccounts() error {
	accounts := []models.ConnectedAccount{
		{
			ID:                uuid.New().String(),
			UserID:            uuid.New().String(),
			StripeAccountID:   "acct_test123456789",
			AccountType:       "express",
			Country:           "US",
			Currency:          "USD",
			VerificationStatus: "verified",
			ChargesEnabled:    true,
			PayoutsEnabled:    true,
			Metadata:          s.buildMetadata(map[string]string{"business_type": "individual", "platform": "blytz"}),
		},
		{
			ID:                uuid.New().String(),
			UserID:            uuid.New().String(),
			StripeAccountID:   "acct_test987654321",
			AccountType:       "custom",
			Country:           "US",
			Currency:          "USD",
			VerificationStatus: "pending",
			ChargesEnabled:    false,
			PayoutsEnabled:    false,
			Metadata:          s.buildMetadata(map[string]string{"business_type": "business", "platform": "blytz"}),
		},
		{
			ID:                uuid.New().String(),
			UserID:            uuid.New().String(),
			StripeAccountID:   "acct_test111111111",
			AccountType:       "express",
			Country:           "GB",
			Currency:          "GBP",
			VerificationStatus: "verified",
			ChargesEnabled:    true,
			PayoutsEnabled:    true,
			Metadata:          s.buildMetadata(map[string]string{"business_type": "individual", "platform": "blytz", "region": "EU"}),
		},
	}

	for _, account := range accounts {
		var existing models.ConnectedAccount
		result := s.db.Where("stripe_account_id = ?", account.StripeAccountID).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.Create(&account).Error; err != nil {
				return fmt.Errorf("failed to create connected account %s: %w", account.StripeAccountID, err)
			}
			s.logger.Info("Created connected account", zap.String("stripe_account_id", account.StripeAccountID))
		}
	}

	return nil
}

// SeedPaymentIntents creates sample payment intents
func (s *Seeder) SeedPaymentIntents() error {
	// Get connected accounts for foreign key reference
	var accounts []models.ConnectedAccount
	if err := s.db.Find(&accounts).Error; err != nil {
		return fmt.Errorf("failed to fetch connected accounts: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no connected accounts found, cannot seed payment intents")
	}

	paymentIntents := []models.PaymentIntent{
		{
			ID:                    uuid.New().String(),
			StripePaymentIntentID: "pi_test123456789",
			Amount:                10000, // $100.00 in cents
			Currency:              "USD",
			Status:                "succeeded",
			UserID:                uuid.New().String(),
			AuctionID:             stringPtr(uuid.New().String()),
			ConnectedAccountID:    stringPtr(accounts[0].ID),
			ApplicationFeeAmount:  500, // $5.00 fee
			Metadata:              s.buildMetadata(map[string]string{"auction_id": "auc_123", "buyer_id": "user_456"}),
		},
		{
			ID:                    uuid.New().String(),
			StripePaymentIntentID: "pi_test987654321",
			Amount:                25000, // $250.00 in cents
			Currency:              "USD",
			Status:                "processing",
			UserID:                uuid.New().String(),
			AuctionID:             stringPtr(uuid.New().String()),
			ConnectedAccountID:    stringPtr(accounts[0].ID),
			ApplicationFeeAmount:  1250, // $12.50 fee
			Metadata:              s.buildMetadata(map[string]string{"auction_id": "auc_789", "buyer_id": "user_101"}),
		},
		{
			ID:                    uuid.New().String(),
			StripePaymentIntentID: "pi_test111111111",
			Amount:                5000, // $50.00 in cents
			Currency:              "GBP",
			Status:                "requires_payment_method",
			UserID:                uuid.New().String(),
			AuctionID:             stringPtr(uuid.New().String()),
			ConnectedAccountID:    stringPtr(accounts[2].ID), // UK account
			ApplicationFeeAmount:  250, // £2.50 fee
			Metadata:              s.buildMetadata(map[string]string{"auction_id": "auc_222", "buyer_id": "user_333"}),
		},
	}

	for _, payment := range paymentIntents {
		var existing models.PaymentIntent
		result := s.db.Where("stripe_payment_intent_id = ?", payment.StripePaymentIntentID).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.Create(&payment).Error; err != nil {
				return fmt.Errorf("failed to create payment intent %s: %w", payment.StripePaymentIntentID, err)
			}
			s.logger.Info("Created payment intent", zap.String("stripe_payment_intent_id", payment.StripePaymentIntentID))
		}
	}

	return nil
}

// SeedTransfers creates sample transfers
func (s *Seeder) SeedTransfers() error {
	// Get connected accounts for foreign key reference
	var accounts []models.ConnectedAccount
	if err := s.db.Find(&accounts).Error; err != nil {
		return fmt.Errorf("failed to fetch connected accounts: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no connected accounts found, cannot seed transfers")
	}

	transfers := []models.Transfer{
		{
			ID:                   uuid.New().String(),
			StripeTransferID:     "tr_test123456789",
			ConnectedAccountID:   accounts[0].ID,
			Amount:               9500, // $95.00 in cents (after fees)
			Currency:             "USD",
			Status:               "completed",
			DestinationPaymentID: "py_test123456789",
			Metadata:             s.buildMetadata(map[string]string{"payment_intent": "pi_test123456789", "type": "auction_payout"}),
		},
		{
			ID:                   uuid.New().String(),
			StripeTransferID:     "tr_test987654321",
			ConnectedAccountID:   accounts[0].ID,
			Amount:               23750, // $237.50 in cents (after fees)
			Currency:             "USD",
			Status:               "in_transit",
			DestinationPaymentID: "py_test987654321",
			Metadata:             s.buildMetadata(map[string]string{"payment_intent": "pi_test987654321", "type": "auction_payout"}),
		},
		{
			ID:                   uuid.New().String(),
			StripeTransferID:     "tr_test111111111",
			ConnectedAccountID:   accounts[2].ID, // UK account
			Amount:               4750, // £47.50 in cents (after fees)
			Currency:             "GBP",
			Status:               "pending",
			DestinationPaymentID: "py_test111111111",
			Metadata:             s.buildMetadata(map[string]string{"payment_intent": "pi_test111111111", "type": "auction_payout"}),
		},
	}

	for _, transfer := range transfers {
		var existing models.Transfer
		result := s.db.Where("stripe_transfer_id = ?", transfer.StripeTransferID).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.Create(&transfer).Error; err != nil {
				return fmt.Errorf("failed to create transfer %s: %w", transfer.StripeTransferID, err)
			}
			s.logger.Info("Created transfer", zap.String("stripe_transfer_id", transfer.StripeTransferID))
		}
	}

	return nil
}

// SeedPayouts creates sample payouts
func (s *Seeder) SeedPayouts() error {
	// Get connected accounts for foreign key reference
	var accounts []models.ConnectedAccount
	if err := s.db.Find(&accounts).Error; err != nil {
		return fmt.Errorf("failed to fetch connected accounts: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no connected accounts found, cannot seed payouts")
	}

	payouts := []models.Payout{
		{
			ID:                 uuid.New().String(),
			StripePayoutID:     "po_test123456789",
			ConnectedAccountID: accounts[0].ID,
			Amount:             9500, // $95.00 in cents
			Currency:           "USD",
			Status:             "paid",
			ArrivalDate:        time.Now().AddDate(0, 0, -2), // 2 days ago
			Metadata:           s.buildMetadata(map[string]string{"batch_id": "batch_123", "type": "weekly_payout"}),
		},
		{
			ID:                 uuid.New().String(),
			StripePayoutID:     "po_test987654321",
			ConnectedAccountID: accounts[0].ID,
			Amount:             23750, // $237.50 in cents
			Currency:           "USD",
			Status:             "in_transit",
			ArrivalDate:        time.Now().AddDate(0, 0, 2), // 2 days from now
			Metadata:           s.buildMetadata(map[string]string{"batch_id": "batch_456", "type": "weekly_payout"}),
		},
		{
			ID:                 uuid.New().String(),
			StripePayoutID:     "po_test111111111",
			ConnectedAccountID: accounts[2].ID, // UK account
			Amount:             4750, // £47.50 in cents
			Currency:           "GBP",
			Status:             "pending",
			ArrivalDate:        time.Now().AddDate(0, 0, 5), // 5 days from now
			Metadata:           s.buildMetadata(map[string]string{"batch_id": "batch_789", "type": "weekly_payout"}),
		},
	}

	for _, payout := range payouts {
		var existing models.Payout
		result := s.db.Where("stripe_payout_id = ?", payout.StripePayoutID).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.Create(&payout).Error; err != nil {
				return fmt.Errorf("failed to create payout %s: %w", payout.StripePayoutID, err)
			}
			s.logger.Info("Created payout", zap.String("stripe_payout_id", payout.StripePayoutID))
		}
	}

	return nil
}

// SeedApplicationFees creates sample application fees
func (s *Seeder) SeedApplicationFees() error {
	// Get payment intents for foreign key reference
	var paymentIntents []models.PaymentIntent
	if err := s.db.Find(&paymentIntents).Error; err != nil {
		return fmt.Errorf("failed to fetch payment intents: %w", err)
	}

	if len(paymentIntents) == 0 {
		return fmt.Errorf("no payment intents found, cannot seed application fees")
	}

	applicationFees := []models.ApplicationFee{
		{
			ID:              uuid.New().String(),
			StripeFeeID:     "fee_test123456789",
			PaymentIntentID: paymentIntents[0].ID,
			Amount:          500, // $5.00 in cents
			Currency:        "USD",
		},
		{
			ID:              uuid.New().String(),
			StripeFeeID:     "fee_test987654321",
			PaymentIntentID: paymentIntents[1].ID,
			Amount:          1250, // $12.50 in cents
			Currency:        "USD",
		},
		{
			ID:              uuid.New().String(),
			StripeFeeID:     "fee_test111111111",
			PaymentIntentID: paymentIntents[2].ID,
			Amount:          250, // £2.50 in cents
			Currency:        "GBP",
		},
	}

	for _, fee := range applicationFees {
		var existing models.ApplicationFee
		result := s.db.Where("stripe_fee_id = ?", fee.StripeFeeID).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.Create(&fee).Error; err != nil {
				return fmt.Errorf("failed to create application fee %s: %w", fee.StripeFeeID, err)
			}
			s.logger.Info("Created application fee", zap.String("stripe_fee_id", fee.StripeFeeID))
		}
	}

	return nil
}

// SeedStripeEvents creates sample stripe events
func (s *Seeder) SeedStripeEvents() error {
	events := []models.StripeEvent{
		{
			ID:          uuid.New().String(),
			StripeEventID: "evt_test123456789",
			EventType:   "payment_intent.succeeded",
			Processed:   true,
			EventData:   s.buildEventData(map[string]interface{}{"payment_intent": "pi_test123456789", "amount": 10000}),
		},
		{
			ID:          uuid.New().String(),
			StripeEventID: "evt_test987654321",
			EventType:   "transfer.created",
			Processed:   false,
			EventData:   s.buildEventData(map[string]interface{}{"transfer": "tr_test987654321", "amount": 23750}),
		},
		{
			ID:          uuid.New().String(),
			StripeEventID: "evt_test111111111",
			EventType:   "payout.created",
			Processed:   false,
			EventData:   s.buildEventData(map[string]interface{}{"payout": "po_test111111111", "amount": 4750}),
		},
		{
			ID:          uuid.New().String(),
			StripeEventID: "evt_test222222222",
			EventType:   "account.updated",
			Processed:   true,
			EventData:   s.buildEventData(map[string]interface{}{"account": "acct_test123456789", "charges_enabled": true}),
		},
	}

	for _, event := range events {
		var existing models.StripeEvent
		result := s.db.Where("stripe_event_id = ?", event.StripeEventID).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.Create(&event).Error; err != nil {
				return fmt.Errorf("failed to create stripe event %s: %w", event.StripeEventID, err)
			}
			s.logger.Info("Created stripe event", zap.String("stripe_event_id", event.StripeEventID))
		}
	}

	return nil
}

// Cleanup removes all seeded data
func (s *Seeder) Cleanup() error {
	s.logger.Info("Cleaning up seeded data")
	
	tables := []string{
		"stripe_events",
		"application_fees",
		"payouts",
		"transfers",
		"payment_intents",
		"connected_accounts",
	}
	
	for _, table := range tables {
		if err := s.db.Exec("DELETE FROM " + table).Error; err != nil {
			return fmt.Errorf("failed to cleanup table %s: %w", table, err)
		}
	}
	
	s.logger.Info("Seeded data cleanup completed")
	return nil
}

// Helper function to build metadata JSON
func (s *Seeder) buildMetadata(data map[string]string) string {
	metadata, _ := json.Marshal(data)
	return string(metadata)
}

// Helper function to build event data JSON
func (s *Seeder) buildEventData(data map[string]interface{}) string {
	eventData, _ := json.Marshal(data)
	return string(eventData)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// SeedTestData is a convenience function to seed test data
func SeedTestData(db *gorm.DB, logger *zap.Logger) error {
	seeder := NewSeeder(db, logger)
	return seeder.SeedAll()
}

// CleanupTestData is a convenience function to clean up test data
func CleanupTestData(db *gorm.DB, logger *zap.Logger) error {
	seeder := NewSeeder(db, logger)
	return seeder.Cleanup()
}