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
		fmt.Println("✅ Product Service migrations completed successfully")
	case "down":
		if err := migrateDown(db); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("✅ Product Service migrations rolled back successfully")
	default:
		fmt.Println("Usage: go run migrate.go [up|down]")
		os.Exit(1)
	}
}

func migrateUp(db *sql.DB) error {
	ctx := context.Background()

	// Create products table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL,
			category_id UUID,
			seller_id UUID NOT NULL,
			sku VARCHAR(100) UNIQUE,
			stock_quantity INTEGER DEFAULT 0 NOT NULL,
			min_bid_amount DECIMAL(10,2) DEFAULT 1.00 NOT NULL,
			buy_it_now_price DECIMAL(10,2),
			images JSONB DEFAULT '[]'::jsonb,
			tags JSONB DEFAULT '[]'::jsonb,
			status VARCHAR(50) DEFAULT 'active' NOT NULL,
			is_featured BOOLEAN DEFAULT false NOT NULL,
			weight DECIMAL(8,2),
			dimensions JSONB,
			shipping_info JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			deleted_at TIMESTAMP WITH TIME ZONE
		);
	`); err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	// Create categories table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			description TEXT,
			parent_id UUID REFERENCES categories(id),
			image_url TEXT,
			sort_order INTEGER DEFAULT 0 NOT NULL,
			is_active BOOLEAN DEFAULT true NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create categories table: %w", err)
	}

	// Create product_reviews table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS product_reviews (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			user_id UUID NOT NULL,
			rating INTEGER CHECK (rating >= 1 AND rating <= 5) NOT NULL,
			title VARCHAR(255),
			comment TEXT,
			verified_purchase BOOLEAN DEFAULT false NOT NULL,
			helpful_count INTEGER DEFAULT 0 NOT NULL,
			status VARCHAR(50) DEFAULT 'approved' NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create product_reviews table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_products_seller_id ON products(seller_id);",
		"CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);",
		"CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);",
		"CREATE INDEX IF NOT EXISTS idx_products_featured ON products(is_featured);",
		"CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);",
		"CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock_quantity);",
		"CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);",
		"CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);",
		"CREATE INDEX IF NOT EXISTS idx_categories_active ON categories(is_active);",
		"CREATE INDEX IF NOT EXISTS idx_product_reviews_product_id ON product_reviews(product_id);",
		"CREATE INDEX IF NOT EXISTS idx_product_reviews_user_id ON product_reviews(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_product_reviews_rating ON product_reviews(rating);",
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
		"CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();",
		"CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();",
		"CREATE TRIGGER update_product_reviews_updated_at BEFORE UPDATE ON product_reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();",
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
		"product_reviews",
		"products",
		"categories",
	}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}
