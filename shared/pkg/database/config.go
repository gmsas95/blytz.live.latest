package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Config holds database configuration
type Config struct {
	DatabaseURL string
	RedisURL    string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	RedisPoolSize int
}

// NewConfig creates a new database configuration with optimized defaults
func NewConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/blytz_db"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		MaxOpenConns: 50,        // Optimized for 3000+ concurrent users
		MaxIdleConns: 25,        // Half of max connections for efficiency
		ConnMaxLifetime: 300 * time.Second, // 5 minutes max connection lifetime
		ConnMaxIdleTime: 60 * time.Second,  // 1 minute max idle time
		RedisPoolSize: 50,       // Optimized Redis connection pool
	}
}

// Database holds database connections
type Database struct {
	DB     *sql.DB
	Redis  *redis.Client
	Logger *zap.Logger
}

// NewDatabase creates new database connections with optimized settings
func NewDatabase(config *Config) (*Database, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Connect to PostgreSQL with optimized connection pool
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for high performance
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize Redis client with optimized settings
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: "",
		DB:       0, // Default DB
		PoolSize: config.RedisPoolSize,
		MinIdleConns: 10, // Minimum idle connections
		MaxRetries: 3,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout: 4 * time.Second,
	})

	// Test Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis connection failed, continuing without cache", zap.Error(err))
		// Don't fail completely if Redis is not available
	}

	database := &Database{
		DB:     db,
		Redis:  rdb,
		Logger: logger,
	}

	logger.Info("Database connections established",
		zap.Int("max_open_conns", config.MaxOpenConns),
		zap.Int("max_idle_conns", config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", config.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", config.ConnMaxIdleTime),
		zap.Int("redis_pool_size", config.RedisPoolSize),
	)

	return database, nil
}

// Close closes database connections
func (d *Database) Close() error {
	var errs []error

	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("database close error: %w", err))
		}
	}

	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("redis close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors occurred: %v", errs)
	}

	return nil
}

// Health checks both database connections
func (d *Database) Health(ctx context.Context) map[string]string {
	status := make(map[string]string)

	// Check PostgreSQL
	if err := d.DB.PingContext(ctx); err != nil {
		status["database"] = "unhealthy"
		d.Logger.Error("Database health check failed", zap.Error(err))
	} else {
		status["database"] = "healthy"
	}

	// Check Redis
	if err := d.Redis.Ping(ctx).Err(); err != nil {
		status["redis"] = "unhealthy"
		d.Logger.Error("Redis health check failed", zap.Error(err))
	} else {
		status["redis"] = "healthy"
	}

	return status
}

// GetStats returns connection pool statistics
func (d *Database) GetStats() sql.DBStats {
	return d.DB.Stats()
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}