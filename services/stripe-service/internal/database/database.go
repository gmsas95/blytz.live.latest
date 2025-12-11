package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "blytz_db"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    50,
		MaxIdleConns:    25,
		ConnMaxLifetime: 300 * time.Second,
		ConnMaxIdleTime: 60 * time.Second,
	}
}

// Database wraps GORM DB and SQL DB
type Database struct {
	DB     *gorm.DB
	SQLDB  *sql.DB
	Logger *zap.Logger
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *Config, zapLogger *zap.Logger) (*Database, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open GORM connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	database := &Database{
		DB:     db,
		SQLDB:  sqlDB,
		Logger: zapLogger,
	}

	// Test connection
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	zapLogger.Info("Database connection established successfully",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("dbname", cfg.DBName),
	)

	return database, nil
}

// Ping checks if database is reachable
func (d *Database) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.SQLDB.PingContext(ctx)
}

// AutoMigrate runs auto-migration for all models
func (d *Database) AutoMigrate() error {
	d.Logger.Info("Running database auto-migration")

	// Get migration files and run them first
	if err := d.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Run GORM auto-migration for models
	err := d.DB.AutoMigrate(
		&models.ConnectedAccount{},
		&models.PaymentIntent{},
		&models.Transfer{},
		&models.Payout{},
		&models.ApplicationFee{},
		&models.StripeEvent{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	d.Logger.Info("Database auto-migration completed successfully")
	return nil
}

// runMigrations executes SQL migration files
func (d *Database) runMigrations() error {
	migrationsPath := filepath.Join("internal", "database", "migrations")
	
	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		d.Logger.Info("Migrations directory not found, skipping SQL migrations")
		return nil
	}

	// Create schema_migrations table if it doesn't exist
	if err := d.createMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Read migration files
	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort files by name (version order)
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		version := file.Name()
		
		// Check if migration already applied
		applied, err := d.isMigrationApplied(version)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}
		if applied {
			d.Logger.Info("Migration already applied", zap.String("version", version))
			continue
		}

		// Read and execute migration
		migrationPath := filepath.Join(migrationsPath, file.Name())
		if err := d.executeMigration(migrationPath, version); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		d.Logger.Info("Migration applied successfully", zap.String("version", version))
	}

	return nil
}

// createMigrationTable creates schema_migrations table
func (d *Database) createMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := d.SQLDB.Exec(query)
	return err
}

// isMigrationApplied checks if a migration has been applied
func (d *Database) isMigrationApplied(version string) (bool, error) {
	var count int64
	err := d.DB.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// executeMigration executes a migration file
func (d *Database) executeMigration(filePath, version string) error {
	// Read migration file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration in a transaction
	return d.DB.Transaction(func(tx *gorm.DB) error {
		// Execute migration SQL
		if err := tx.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}

		// Record migration as applied
		return tx.Exec("INSERT INTO schema_migrations (version) VALUES (?) ON CONFLICT (version) DO NOTHING", version).Error
	})
}

// HealthCheck performs a comprehensive health check
func (d *Database) HealthCheck() map[string]interface{} {
	health := make(map[string]interface{})
	
	// Check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := d.SQLDB.PingContext(ctx); err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
		return health
	}
	
	// Get connection stats
	stats := d.SQLDB.Stats()
	health["status"] = "healthy"
	health["open_connections"] = stats.OpenConnections
	health["in_use"] = stats.InUse
	health["idle"] = stats.Idle
	health["wait_count"] = stats.WaitCount
	health["wait_duration"] = stats.WaitDuration.String()
	health["max_idle_closed"] = stats.MaxIdleClosed
	health["max_lifetime_closed"] = stats.MaxLifetimeClosed
	
	return health
}

// Close closes database connection
func (d *Database) Close() error {
	d.Logger.Info("Closing database connection")
	return d.SQLDB.Close()
}

// GetDB returns GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// GetSQLDB returns SQL database instance
func (d *Database) GetSQLDB() *sql.DB {
	return d.SQLDB
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}