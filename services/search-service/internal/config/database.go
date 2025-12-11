package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	DatabaseURL string
	Environment string
}

// InitDB initializes database connection with optimized connection pooling
func InitDB(cfg *DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DatabaseURL

	// Remove duplicate SSL mode parameters if they exist
	if strings.Contains(dsn, "?sslmode=") && strings.Count(dsn, "?sslmode=") > 1 {
		// Find all SSL mode positions
		sslIndices := []int{}
		start := 0
		for {
			idx := strings.Index(dsn[start:], "?sslmode=")
			if idx == -1 {
				break
			}
			sslIndices = append(sslIndices, start+idx)
			start += idx + 1
		}

		// If we found more than one, remove the second one onwards
		if len(sslIndices) > 1 {
			// Keep everything up to the second SSL mode
			dsn = dsn[:sslIndices[1]]
		}
	}

	gormConfig := &gorm.Config{}

	if cfg.Environment == "development" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool with high-performance settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool parameters for high-performance search service
	sqlDB.SetMaxIdleConns(50)           // Increased from 10 for better performance
	sqlDB.SetMaxOpenConns(200)          // As specified: 200 max connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Connection lifetime
	sqlDB.SetConnMaxIdleTime(time.Minute * 30) // Idle connection timeout

	log.Println("Database connection established with optimized connection pooling (200 max connections)")
	return db, nil
}

// CreateSearchIndex creates the search index table with full-text search capabilities
func CreateSearchIndex(db *gorm.DB) error {
	// Create search_indices table with full-text search
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS search_indices (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			entity_type VARCHAR(50) NOT NULL,
			entity_id VARCHAR(255) NOT NULL,
			title VARCHAR(500) NOT NULL,
			description TEXT,
			tags TEXT,
			category VARCHAR(100),
			sub_category VARCHAR(100),
			price DECIMAL(10,2),
			status VARCHAR(50),
			location VARCHAR(255),
			seller_id VARCHAR(255),
			seller_name VARCHAR(255),
			rating DECIMAL(3,2),
			review_count INTEGER DEFAULT 0,
			view_count BIGINT DEFAULT 0,
			likes_count BIGINT DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			search_vector TSVECTOR,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			indexed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`).Error

	if err != nil {
		return fmt.Errorf("failed to create search_indices table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_search_indices_entity_type ON search_indices(entity_type)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_entity_id ON search_indices(entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_category ON search_indices(category)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_sub_category ON search_indices(sub_category)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_price ON search_indices(price)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_status ON search_indices(status)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_location ON search_indices(location)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_seller_id ON search_indices(seller_id)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_seller_name ON search_indices(seller_name)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_rating ON search_indices(rating)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_review_count ON search_indices(review_count)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_view_count ON search_indices(view_count)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_likes_count ON search_indices(likes_count)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_is_active ON search_indices(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_created_at ON search_indices(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_updated_at ON search_indices(updated_at)",
		// Full-text search index
		"CREATE INDEX IF NOT EXISTS idx_search_indices_search_vector ON search_indices USING GIN(search_vector)",
		// Composite indexes for common queries
		"CREATE INDEX IF NOT EXISTS idx_search_indices_type_status_active ON search_indices(entity_type, status, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_category_price ON search_indices(category, price)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_location_type ON search_indices(location, entity_type)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create search_analytics table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS search_analytics (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			query VARCHAR(500) NOT NULL,
			user_id VARCHAR(255),
			results INTEGER,
			clicks INTEGER DEFAULT 0,
			ip_address VARCHAR(45),
			user_agent TEXT,
			session_id VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`).Error

	if err != nil {
		return fmt.Errorf("failed to create search_analytics table: %w", err)
	}

	// Create analytics indexes
	analyticsIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_search_analytics_query ON search_analytics(query)",
		"CREATE INDEX IF NOT EXISTS idx_search_analytics_user_id ON search_analytics(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_search_analytics_results ON search_analytics(results)",
		"CREATE INDEX IF NOT EXISTS idx_search_analytics_clicks ON search_analytics(clicks)",
		"CREATE INDEX IF NOT EXISTS idx_search_analytics_session_id ON search_analytics(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_search_analytics_created_at ON search_analytics(created_at)",
	}

	for _, idx := range analyticsIndexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create analytics index: %w", err)
		}
	}

	// Create trigger to automatically update search_vector
	triggerSQL := `
		CREATE OR REPLACE FUNCTION update_search_vector()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.search_vector := 
				setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
				setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
				setweight(to_tsvector('english', COALESCE(NEW.category, '')), 'C') ||
				setweight(to_tsvector('english', COALESCE(NEW.sub_category, '')), 'C') ||
				setweight(to_tsvector('english', COALESCE(NEW.tags, '')), 'D') ||
				setweight(to_tsvector('english', COALESCE(NEW.seller_name, '')), 'D');
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		DROP TRIGGER IF EXISTS trigger_update_search_vector ON search_indices;
		CREATE TRIGGER trigger_update_search_vector
			BEFORE INSERT OR UPDATE ON search_indices
			FOR EACH ROW EXECUTE FUNCTION update_search_vector();
	`

	if err := db.Exec(triggerSQL).Error; err != nil {
		return fmt.Errorf("failed to create search_vector trigger: %w", err)
	}

	log.Println("Search database schema created successfully with full-text search capabilities")
	return nil
}

// MigrateDatabase runs database migrations
func MigrateDatabase(db *gorm.DB) error {
	// Auto-migrate the models
	if err := db.AutoMigrate(&SearchIndex{}, &SearchAnalytics{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create additional indexes and triggers
	if err := CreateSearchIndex(db); err != nil {
		return fmt.Errorf("failed to create search index: %w", err)
	}

	return nil
}

// SearchIndex is imported for migration
type SearchIndex struct {
	// This is defined here only for migration purposes
	// The actual model is in models/search.go
}

// SearchAnalytics is imported for migration purposes
type SearchAnalytics struct {
	// This is defined here only for migration purposes
	// The actual model is in models/search.go
}