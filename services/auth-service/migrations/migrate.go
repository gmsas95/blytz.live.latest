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
		fmt.Println("✅ Auth Service migrations completed successfully")
	case "down":
		if err := migrateDown(db); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("✅ Auth Service migrations rolled back successfully")
	default:
		fmt.Println("Usage: go run migrate.go [up|down]")
		os.Exit(1)
	}
}

func migrateUp(db *sql.DB) error {
	ctx := context.Background()

	// Create users table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			phone_number VARCHAR(20),
			avatar_url TEXT,
			role VARCHAR(50) DEFAULT 'user' NOT NULL,
			is_active BOOLEAN DEFAULT true NOT NULL,
			email_verified BOOLEAN DEFAULT false NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			last_login_at TIMESTAMP WITH TIME ZONE
		);
	`); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create refresh_tokens table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token TEXT UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			is_revoked BOOLEAN DEFAULT false NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create refresh_tokens table: %w", err)
	}

	// Create user_sessions table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			session_token TEXT UNIQUE NOT NULL,
			ip_address INET,
			user_agent TEXT,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create user_sessions table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);",
		"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);",
		"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);",
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);",
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);",
	}

	for _, indexSQL := range indexes {
		if _, err := db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create updated_at trigger
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

	if _, err := db.ExecContext(ctx, `
		CREATE TRIGGER update_users_updated_at 
			BEFORE UPDATE ON users 
			FOR EACH ROW 
			EXECUTE FUNCTION update_updated_at_column();
	`); err != nil {
		return fmt.Errorf("failed to create users updated_at trigger: %w", err)
	}

	return nil
}

func migrateDown(db *sql.DB) error {
	ctx := context.Background()

	// Drop tables in reverse order
	tables := []string{
		"user_sessions",
		"refresh_tokens",
		"users",
	}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// Drop trigger function
	if _, err := db.ExecContext(ctx, "DROP FUNCTION IF EXISTS update_updated_at_column();"); err != nil {
		return fmt.Errorf("failed to drop update_updated_at_function: %w", err)
	}

	return nil
}
