package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// StripeConfig holds all configuration for Stripe service
type StripeConfig struct {
	// Database configuration
	DatabaseURL      string
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresHost     string `env:"POSTGRES_HOST"`
	PostgresPort     string `env:"POSTGRES_PORT"`
	PostgresDB       string `env:"POSTGRES_DB"`

	// Redis configuration
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// Service configuration
	ServicePort string
	Environment string
	LogLevel    string

	// Stripe configuration
	SecretKey          string
	PublishableKey     string
	WebhookSecret      string
	ConnectClientID    string
	ConnectRedirectURI string
	PlatformFeePercent float64
	PayoutSchedule     string

	// JWT configuration (for auth with other services)
	JWTSecret string

	// External service URLs
	AuthServiceURL string
	GatewayURL     string
}

// Load loads configuration from environment variables
func Load() (*StripeConfig, error) {
	// Load .env file if it exists (for development)
	_ = godotenv.Load()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := &StripeConfig{
		// Database defaults
		PostgresUser:     getEnvOrDefault("POSTGRES_USER", "blytz"),
		PostgresPassword: getEnvOrDefault("POSTGRES_PASSWORD", ""),
		PostgresHost:     getEnvOrDefault("POSTGRES_HOST", "postgres"),
		PostgresPort:     getEnvOrDefault("POSTGRES_PORT", "5432"),
		PostgresDB:       getEnvOrDefault("POSTGRES_DB", "blytz_prod"),

		// Redis defaults
		RedisURL:      getEnvOrDefault("REDIS_URL", "redis:6379"),
		RedisPassword: getEnvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 8), // Use DB 8 for stripe service

		// Service defaults
		ServicePort: getEnvOrDefault("PORT", "8095"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),

		// Stripe defaults
		SecretKey:          getEnvOrDefault("STRIPE_SECRET_KEY", ""),
		PublishableKey:     getEnvOrDefault("STRIPE_PUBLISHABLE_KEY", ""),
		WebhookSecret:      getEnvOrDefault("STRIPE_WEBHOOK_SECRET", ""),
		ConnectClientID:    getEnvOrDefault("STRIPE_CONNECT_CLIENT_ID", ""),
		ConnectRedirectURI: getEnvOrDefault("STRIPE_CONNECT_REDIRECT_URI", ""),
		PlatformFeePercent: getEnvAsFloat("STRIPE_PLATFORM_FEE_PERCENT", 5.0),
		PayoutSchedule:     getEnvOrDefault("STRIPE_PAYOUT_SCHEDULE", "daily"),

		// JWT defaults
		JWTSecret: getEnvOrDefault("JWT_SECRET", "jwt-secret-key-change-in-production"),

		// External service URLs
		AuthServiceURL: getEnvOrDefault("AUTH_SERVICE_URL", "http://auth-service:8084"),
		GatewayURL:     getEnvOrDefault("GATEWAY_URL", "http://gateway-service:8092"),
	}

	// Check if DATABASE_URL is provided (Dokploy style)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		cfg.DatabaseURL = databaseURL
		// Parse DATABASE_URL to extract components for fallback usage
		if parsedURL, err := url.Parse(databaseURL); err == nil {
			if parsedURL.User != nil {
				cfg.PostgresUser = parsedURL.User.Username()
				if password, ok := parsedURL.User.Password(); ok {
					cfg.PostgresPassword = password
				}
			}
			if parsedURL.Hostname() != "" {
				cfg.PostgresHost = parsedURL.Hostname()
			}
			if parsedURL.Port() != "" {
				cfg.PostgresPort = parsedURL.Port()
			}
			// Extract database name from path
			if len(parsedURL.Path) > 1 {
				cfg.PostgresDB = parsedURL.Path[1:] // Remove leading slash
			}
		}
	} else {
		// Construct database URL from individual components (original behavior)
		encodedPassword := url.QueryEscape(cfg.PostgresPassword)
		cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.PostgresUser, encodedPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)
	}

	// Validate required configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// validate checks that all required configuration fields are set
func (c *StripeConfig) validate() error {
	if c.SecretKey == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is required")
	}

	if c.PublishableKey == "" {
		return fmt.Errorf("STRIPE_PUBLISHABLE_KEY is required")
	}

	if c.WebhookSecret == "" {
		return fmt.Errorf("STRIPE_WEBHOOK_SECRET is required")
	}

	if c.ConnectClientID == "" {
		return fmt.Errorf("STRIPE_CONNECT_CLIENT_ID is required")
	}

	if c.ConnectRedirectURI == "" {
		return fmt.Errorf("STRIPE_CONNECT_REDIRECT_URI is required")
	}

	if c.PlatformFeePercent < 0 || c.PlatformFeePercent > 100 {
		return fmt.Errorf("STRIPE_PLATFORM_FEE_PERCENT must be between 0 and 100")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *StripeConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *StripeConfig) IsProduction() bool {
	return c.Environment == "production"
}

// IsTest returns true if running in test mode
func (c *StripeConfig) IsTest() bool {
	return c.Environment == "test"
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer or returns default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsFloat gets environment variable as float64 or returns default value
func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets environment variable as boolean or returns default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}