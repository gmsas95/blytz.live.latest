package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Config holds Redis configuration
type Config struct {
	URL         string
	Password    string
	DB          int
	PoolSize    int
	MaxRetries  int
	DialTimeout time.Duration
	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

// NewConfig creates a new Redis configuration from environment variables
func NewConfig() *Config {
	return &Config{
		URL:         getEnv("REDIS_URL", "localhost:6379"),
		Password:    getEnv("REDIS_PASSWORD", ""),
		DB:          getEnvAsInt("REDIS_DB", 0),
		PoolSize:    getEnvAsInt("REDIS_POOL_SIZE", 50),
		MaxRetries:  getEnvAsInt("REDIS_MAX_RETRIES", 3),
		DialTimeout: getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout: getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
	}
}

// NewClient creates a new Redis client with secure configuration
func NewClient(config *Config, logger *zap.Logger) (*redis.Client, error) {
	if config == nil {
		config = NewConfig()
	}

	options := &redis.Options{
		Addr:         config.URL,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	client := redis.NewClient(options)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		if logger != nil {
			logger.Error("Failed to connect to Redis", 
				zap.String("url", config.URL),
				zap.Error(err))
		}
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully connected to Redis",
			zap.String("url", config.URL),
			zap.Int("db", config.DB),
			zap.Bool("password_set", config.Password != ""))
	}

	return client, nil
}

// NewClientWithEnv creates a new Redis client using environment variables
func NewClientWithEnv(logger *zap.Logger) (*redis.Client, error) {
	config := NewConfig()
	return NewClient(config, logger)
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := parseInt(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func parseInt(s string) (int, error) {
	var result int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid integer: %s", s)
		}
		result = result*10 + int(r-'0')
	}
	return result, nil
}