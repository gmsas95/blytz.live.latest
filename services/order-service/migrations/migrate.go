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
		fmt.Println("✅ Order Service migrations completed successfully")
	case "down":
		if err := migrateDown(db); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("✅ Order Service migrations rolled back successfully")
	default:
		fmt.Println("Usage: go run migrate.go [up|down]")
		os.Exit(1)
	}
}

func migrateUp(db *sql.DB) error {
	ctx := context.Background()

	// Create orders table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			auction_id UUID REFERENCES auctions(id),
			seller_id UUID NOT NULL,
			total_amount DECIMAL(10,2) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending' NOT NULL,
			payment_status VARCHAR(50) DEFAULT 'pending' NOT NULL,
			shipping_address JSONB,
			billing_address JSONB,
			items JSONB NOT NULL DEFAULT '[]'::jsonb,
			subtotal DECIMAL(10,2) NOT NULL,
			tax_amount DECIMAL(10,2) DEFAULT 0.00 NOT NULL,
			shipping_amount DECIMAL(10,2) DEFAULT 0.00 NOT NULL,
			discount_amount DECIMAL(10,2) DEFAULT 0.00 NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD' NOT NULL,
			notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	// Create order_items table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS order_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id),
			auction_id UUID REFERENCES auctions(id),
			quantity INTEGER NOT NULL DEFAULT 1,
			unit_price DECIMAL(10,2) NOT NULL,
			total_price DECIMAL(10,2) NOT NULL,
			product_snapshot JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}

	// Create order_status_history table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS order_status_history (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			status VARCHAR(50) NOT NULL,
			previous_status VARCHAR(50),
			comment TEXT,
			created_by UUID,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create order_status_history table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_orders_seller_id ON orders(seller_id);",
		"CREATE INDEX IF NOT EXISTS idx_orders_auction_id ON orders(auction_id);",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);",
		"CREATE INDEX IF NOT EXISTS idx_orders_payment_status ON orders(payment_status);",
		"CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);",
		"CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);",
		"CREATE INDEX IF NOT EXISTS idx_order_items_auction_id ON order_items(auction_id);",
		"CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);",
		"CREATE INDEX IF NOT EXISTS idx_order_status_history_status ON order_status_history(status);",
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
		"CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();",
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
		"order_status_history",
		"order_items",
		"orders",
	}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}
