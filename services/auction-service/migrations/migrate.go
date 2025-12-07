package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run migrate.go [up|down]")
		os.Exit(1)
	}

	action := os.Args[1]

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://blytz:blytz_password_2025@localhost:5432/blytz_mvp?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	switch action {
	case "up":
		if err := migrateUp(db); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("✅ Auction Service migrations completed successfully")
	case "down":
		if err := migrateDown(db); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("✅ Auction Service migrations rolled back successfully")
	default:
		fmt.Println("Usage: go run migrate.go [up|down]")
		os.Exit(1)
	}
}

func migrateUp(db *sql.DB) error {
	ctx := context.Background()

	// Create auctions table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS auctions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			seller_id UUID NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			starting_price DECIMAL(10,2) NOT NULL,
			reserve_price DECIMAL(10,2),
			buy_it_now_price DECIMAL(10,2),
			current_bid DECIMAL(10,2) DEFAULT 0.00 NOT NULL,
			bid_count INTEGER DEFAULT 0 NOT NULL,
			start_time TIMESTAMP WITH TIME ZONE NOT NULL,
			end_time TIMESTAMP WITH TIME ZONE NOT NULL,
			status VARCHAR(50) DEFAULT 'draft' NOT NULL,
			auto_extend BOOLEAN DEFAULT true NOT NULL,
			anti_snipe_time INTEGER DEFAULT 300 NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create auctions table: %w", err)
	}

	// Create bids table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bids (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auction_id UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
			bidder_id UUID NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			bid_time TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			is_auto_bid BOOLEAN DEFAULT false NOT NULL,
			is_winning BOOLEAN DEFAULT false NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create bids table: %w", err)
	}

	// Create auction_watchers table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS auction_watchers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auction_id UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			notification_sent BOOLEAN DEFAULT false NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			UNIQUE(auction_id, user_id)
		);
	`); err != nil {
		return fmt.Errorf("failed to create auction_watchers table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_auctions_product_id ON auctions(product_id);",
		"CREATE INDEX IF NOT EXISTS idx_auctions_seller_id ON auctions(seller_id);",
		"CREATE INDEX IF NOT EXISTS idx_auctions_status ON auctions(status);",
		"CREATE INDEX IF NOT EXISTS idx_auctions_start_time ON auctions(start_time);",
		"CREATE INDEX IF NOT EXISTS idx_auctions_end_time ON auctions(end_time);",
		"CREATE INDEX IF NOT EXISTS idx_auctions_current_bid ON auctions(current_bid);",
		"CREATE INDEX IF NOT EXISTS idx_bids_auction_id ON bids(auction_id);",
		"CREATE INDEX IF NOT EXISTS idx_bids_bidder_id ON bids(bidder_id);",
		"CREATE INDEX IF NOT EXISTS idx_bids_amount ON bids(amount);",
		"CREATE INDEX IF NOT EXISTS idx_bids_bid_time ON bids(bid_time);",
		"CREATE INDEX IF NOT EXISTS idx_auction_watchers_auction_id ON auction_watchers(auction_id);",
		"CREATE INDEX IF NOT EXISTS idx_auction_watchers_user_id ON auction_watchers(user_id);",
	}

	for _, indexSQL := range indexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create updated_at trigger function (reuse from auth)
	if _, err := db.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`); err != nil {
		return fmt.Errorf("failed to create update_updated_at_function: %w", err)
	}

	// Create triggers
	triggers := []string{
		"CREATE TRIGGER update_auctions_updated_at BEFORE UPDATE ON auctions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();",
	}

	for _, triggerSQL := range triggers {
		if _, err := db.ExecContext(ctx, triggerSQL); err != nil {
			return fmt.Errorf("failed to create trigger: %w", err)
		}
	}

	return nil
}

func migrateDown(db *sql.DB) error {
	ctx := context.Background()

	// Drop tables in reverse order
	tables := []string{
		"auction_watchers",
		"bids",
		"auctions",
	}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}
