-- Migration: Create Stripe Tables
-- Description: Creates all tables required for the Stripe service
-- Version: 001
-- Date: 2025-12-11

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create connected_accounts table
CREATE TABLE IF NOT EXISTS connected_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    stripe_account_id VARCHAR(255) NOT NULL UNIQUE,
    account_type VARCHAR(50) NOT NULL CHECK (account_type IN ('individual', 'business', 'express')),
    country VARCHAR(2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    verification_status VARCHAR(50) DEFAULT 'pending',
    charges_enabled BOOLEAN DEFAULT false,
    payouts_enabled BOOLEAN DEFAULT false,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create payment_intents table
CREATE TABLE IF NOT EXISTS payment_intents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_payment_intent_id VARCHAR(255) NOT NULL UNIQUE,
    amount BIGINT NOT NULL, -- Amount in cents
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    user_id UUID NOT NULL,
    auction_id UUID,
    connected_account_id UUID,
    application_fee_amount BIGINT DEFAULT 0, -- Fee amount in cents
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create transfers table
CREATE TABLE IF NOT EXISTS transfers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_transfer_id VARCHAR(255) NOT NULL UNIQUE,
    connected_account_id UUID NOT NULL,
    amount BIGINT NOT NULL, -- Amount in cents
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    destination_payment_id VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create payouts table
CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_payout_id VARCHAR(255) NOT NULL UNIQUE,
    connected_account_id UUID NOT NULL,
    amount BIGINT NOT NULL, -- Amount in cents
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    arrival_date TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create application_fees table
CREATE TABLE IF NOT EXISTS application_fees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_fee_id VARCHAR(255) NOT NULL UNIQUE,
    payment_intent_id UUID NOT NULL,
    amount BIGINT NOT NULL, -- Amount in cents
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create stripe_events table
CREATE TABLE IF NOT EXISTS stripe_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_event_id VARCHAR(255) NOT NULL UNIQUE,
    event_type VARCHAR(255) NOT NULL,
    processed BOOLEAN DEFAULT false,
    event_data JSONB NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for connected_accounts
CREATE INDEX IF NOT EXISTS idx_connected_accounts_user_id ON connected_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_connected_accounts_stripe_account_id ON connected_accounts(stripe_account_id);
CREATE INDEX IF NOT EXISTS idx_connected_accounts_deleted_at ON connected_accounts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_connected_accounts_verification_status ON connected_accounts(verification_status);

-- Create indexes for payment_intents
CREATE INDEX IF NOT EXISTS idx_payment_intents_stripe_payment_intent_id ON payment_intents(stripe_payment_intent_id);
CREATE INDEX IF NOT EXISTS idx_payment_intents_user_id ON payment_intents(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_intents_auction_id ON payment_intents(auction_id);
CREATE INDEX IF NOT EXISTS idx_payment_intents_connected_account_id ON payment_intents(connected_account_id);
CREATE INDEX IF NOT EXISTS idx_payment_intents_status ON payment_intents(status);
CREATE INDEX IF NOT EXISTS idx_payment_intents_deleted_at ON payment_intents(deleted_at);

-- Create indexes for transfers
CREATE INDEX IF NOT EXISTS idx_transfers_stripe_transfer_id ON transfers(stripe_transfer_id);
CREATE INDEX IF NOT EXISTS idx_transfers_connected_account_id ON transfers(connected_account_id);
CREATE INDEX IF NOT EXISTS idx_transfers_destination_payment_id ON transfers(destination_payment_id);
CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
CREATE INDEX IF NOT EXISTS idx_transfers_deleted_at ON transfers(deleted_at);

-- Create indexes for payouts
CREATE INDEX IF NOT EXISTS idx_payouts_stripe_payout_id ON payouts(stripe_payout_id);
CREATE INDEX IF NOT EXISTS idx_payouts_connected_account_id ON payouts(connected_account_id);
CREATE INDEX IF NOT EXISTS idx_payouts_status ON payouts(status);
CREATE INDEX IF NOT EXISTS idx_payouts_arrival_date ON payouts(arrival_date);
CREATE INDEX IF NOT EXISTS idx_payouts_deleted_at ON payouts(deleted_at);

-- Create indexes for application_fees
CREATE INDEX IF NOT EXISTS idx_application_fees_stripe_fee_id ON application_fees(stripe_fee_id);
CREATE INDEX IF NOT EXISTS idx_application_fees_payment_intent_id ON application_fees(payment_intent_id);
CREATE INDEX IF NOT EXISTS idx_application_fees_deleted_at ON application_fees(deleted_at);

-- Create indexes for stripe_events
CREATE INDEX IF NOT EXISTS idx_stripe_events_stripe_event_id ON stripe_events(stripe_event_id);
CREATE INDEX IF NOT EXISTS idx_stripe_events_event_type ON stripe_events(event_type);
CREATE INDEX IF NOT EXISTS idx_stripe_events_processed ON stripe_events(processed);
CREATE INDEX IF NOT EXISTS idx_stripe_events_deleted_at ON stripe_events(deleted_at);

-- Add foreign key constraints
ALTER TABLE payment_intents 
    ADD CONSTRAINT fk_payment_intents_connected_account 
    FOREIGN KEY (connected_account_id) 
    REFERENCES connected_accounts(id) 
    ON DELETE SET NULL;

ALTER TABLE transfers 
    ADD CONSTRAINT fk_transfers_connected_account 
    FOREIGN KEY (connected_account_id) 
    REFERENCES connected_accounts(id) 
    ON DELETE CASCADE;

ALTER TABLE payouts 
    ADD CONSTRAINT fk_payouts_connected_account 
    FOREIGN KEY (connected_account_id) 
    REFERENCES connected_accounts(id) 
    ON DELETE CASCADE;

ALTER TABLE application_fees 
    ADD CONSTRAINT fk_application_fees_payment_intent 
    FOREIGN KEY (payment_intent_id) 
    REFERENCES payment_intents(id) 
    ON DELETE CASCADE;

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_connected_accounts_updated_at 
    BEFORE UPDATE ON connected_accounts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_intents_updated_at 
    BEFORE UPDATE ON payment_intents 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transfers_updated_at 
    BEFORE UPDATE ON transfers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payouts_updated_at 
    BEFORE UPDATE ON payouts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_application_fees_updated_at 
    BEFORE UPDATE ON application_fees 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_stripe_events_updated_at 
    BEFORE UPDATE ON stripe_events 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create a migration tracking table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Record this migration as applied
INSERT INTO schema_migrations (version) VALUES ('001_create_stripe_tables')
ON CONFLICT (version) DO NOTHING;