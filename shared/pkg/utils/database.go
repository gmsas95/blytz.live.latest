package utils

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDatabaseConfig creates a new database configuration
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:            GetEnv("DB_HOST", "localhost"),
		Port:            GetEnv("DB_PORT", "5432"),
		User:            GetEnv("DB_USER", "postgres"),
		Password:        GetEnv("DB_PASSWORD", "postgres"),
		DBName:          GetEnv("DB_NAME", "blytz_db"),
		SSLMode:         GetEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// GetDatabaseURL builds database connection URL
func (cfg *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)
}

// InitDatabase initializes database connection
func InitDatabase(cfg *DatabaseConfig, zapLogger *zap.Logger) (*gorm.DB, error) {
	// Configure GORM logger
	var gormLogger logger.Interface
	if zapLogger != nil {
		gormLogger = logger.New(
			NewGormLoggerAdapter(zapLogger),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.GetDatabaseURL()), &gorm.Config{
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

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if zapLogger != nil {
		zapLogger.Info("Database connected successfully",
			zap.String("host", cfg.Host),
			zap.String("port", cfg.Port),
			zap.String("database", cfg.DBName),
		)
	}

	return db, nil
}

// InitDatabaseFromURL initializes database from URL string
func InitDatabaseFromURL(databaseURL string, zapLogger *zap.Logger) (*gorm.DB, error) {
	// Configure GORM logger
	var gormLogger logger.Interface
	if zapLogger != nil {
		gormLogger = logger.New(
			NewGormLoggerAdapter(zapLogger),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
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

	// Configure connection pool with sensible defaults
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if zapLogger != nil {
		zapLogger.Info("Database connected successfully from URL")
	}

	return db, nil
}

// GormLoggerAdapter adapts zap logger to GORM logger interface
type GormLoggerAdapter struct {
	logger *zap.Logger
}

// NewGormLoggerAdapter creates a new GORM logger adapter
func NewGormLoggerAdapter(logger *zap.Logger) *GormLoggerAdapter {
	return &GormLoggerAdapter{
		logger: logger,
	}
}

// Log implements GORM logger interface
func (a *GormLoggerAdapter) Log(level logger.LogLevel, msg string, data ...interface{}) {
	switch level {
	case logger.Info:
		a.logger.Info(fmt.Sprintf(msg, data...))
	case logger.Warn:
		a.logger.Warn(fmt.Sprintf(msg, data...))
	case logger.Error:
		a.logger.Error(fmt.Sprintf(msg, data...))
	default:
		a.logger.Debug(fmt.Sprintf(msg, data...))
	}
}

// Printf implements logger.Writer interface
func (a *GormLoggerAdapter) Printf(format string, v ...interface{}) {
	a.logger.Info(fmt.Sprintf(format, v...))
}

// HealthCheck performs database health check
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// CloseDatabase closes database connection
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}

// GetStats returns database connection stats
func GetStats(db *gorm.DB) (*sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return &stats, nil
}