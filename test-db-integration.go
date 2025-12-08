package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Get database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/blytz_mvp?sslmode=disable"
	}

	fmt.Println("🧪 Testing Database Integration...")
	fmt.Printf("Database URL: %s\n", databaseURL)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Test query
	var result int
	err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		log.Fatalf("❌ Failed to execute test query: %v", err)
	}

	fmt.Printf("✅ Database connection successful! Result: %d\n", result)

	// Test users table exists
	var tableExists bool
	err = conn.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'users'
		)
	`).Scan(&tableExists)
	if err != nil {
		log.Fatalf("❌ Failed to check users table: %v", err)
	}

	if tableExists {
		fmt.Println("✅ Users table exists!")
		
		// Count users
		var userCount int
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
		if err != nil {
			log.Fatalf("❌ Failed to count users: %v", err)
		}
		fmt.Printf("✅ User count: %d\n", userCount)
	} else {
		fmt.Println("⚠️  Users table does not exist - migrations needed")
	}

	fmt.Println("🎉 Database integration test completed successfully!")
}