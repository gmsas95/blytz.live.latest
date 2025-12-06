package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Service configuration
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Stripe      StripeConfig      `mapstructure:"stripe"`
	Auth        AuthConfig        `mapstructure:"auth"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Environment string            `mapstructure:"environment"`
}

// Server configuration
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// Database configuration
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  string `mapstructure:"max_lifetime"`
}

// Stripe configuration
type StripeConfig struct {
	SecretKey              string  `mapstructure:"secret_key"`
	PublishableKey         string  `mapstructure:"publishable_key"`
	WebhookSecret         string  `mapstructure:"webhook_secret"`
	PlatformAccountID     string  `mapstructure:"platform_account_id"`
	ConnectClientID        string  `mapstructure:"connect_client_id"`
	ApplicationFeePercent  float64 `mapstructure:"application_fee_percent"`
	MinApplicationFee     int64   `mapstructure:"min_application_fee"`
	DefaultCurrency        string  `mapstructure:"default_currency"`
	WebhookEndpoint       string  `mapstructure:"webhook_endpoint"`
	SuccessURL           string  `mapstructure:"success_url"`
	CancelURL            string  `mapstructure:"cancel_url"`
}

// Authentication configuration
type AuthConfig struct {
	JWTSecret         string        `mapstructure:"jwt_secret"`
	TokenExpiry       time.Duration `mapstructure:"token_expiry"`
	RefreshExpiry     time.Duration `mapstructure:"refresh_expiry"`
	AllowedOrigins    []string      `mapstructure:"allowed_origins"`
}

// Logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
}

// Load configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnv("PORT", "8090"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "blytz"),
			Password:     getEnv("DB_PASSWORD", ""),
			Database:     getEnv("DB_NAME", "blytz_prod"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnv("DB_MAX_LIFETIME", "1h"),
		},
		Stripe: StripeConfig{
			SecretKey:              getEnv("STRIPE_SECRET_KEY", ""),
			PublishableKey:         getEnv("STRIPE_PUBLISHABLE_KEY", ""),
			WebhookSecret:         getEnv("STRIPE_WEBHOOK_SECRET", ""),
			PlatformAccountID:     getEnv("STRIPE_PLATFORM_ACCOUNT_ID", ""),
			ConnectClientID:        getEnv("STRIPE_CONNECT_CLIENT_ID", ""),
			ApplicationFeePercent:  getEnvAsFloat("STRIPE_APPLICATION_FEE_PERCENT", 0.05), // 5%
			MinApplicationFee:     getEnvAsInt64("STRIPE_MIN_APPLICATION_FEE", 50), // $0.50 in cents
			DefaultCurrency:        getEnv("STRIPE_DEFAULT_CURRENCY", "usd"),
			WebhookEndpoint:       getEnv("STRIPE_WEBHOOK_ENDPOINT", "/api/v1/webhooks/stripe"),
			SuccessURL:           getEnv("STRIPE_SUCCESS_URL", "https://blytz.app/payment/success"),
			CancelURL:            getEnv("STRIPE_CANCEL_URL", "https://blytz.app/payment/cancel"),
		},
		Auth: AuthConfig{
			JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			TokenExpiry:    getEnvAsDuration("TOKEN_EXPIRY", "1h"),
			RefreshExpiry:  getEnvAsDuration("REFRESH_EXPIRY", "24h"),
			AllowedOrigins: []string{
				"https://blytz.app",
				"https://www.blytz.app",
				"https://demo.blytz.app",
				"https://seller.blytz.app",
				"http://localhost:3000",
				"http://localhost:8080",
			},
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate configuration
func (c *Config) Validate() error {
	// Validate Stripe configuration
	if c.Stripe.SecretKey == "" {
		return ErrMissingStripeSecretKey
	}

	if c.Stripe.WebhookSecret == "" {
		return ErrMissingStripeWebhookSecret
	}

	if c.Stripe.ConnectClientID == "" {
		return ErrMissingStripeConnectClientID
	}

	// Validate database configuration
	if c.Database.Host == "" {
		return ErrMissingDatabaseHost
	}

	if c.Database.User == "" {
		return ErrMissingDatabaseUser
	}

	if c.Database.Database == "" {
		return ErrMissingDatabaseName
	}

	// Validate auth configuration
	if c.Auth.JWTSecret == "" {
		return ErrMissingJWTSecret
	}

	return nil
}

// IsDevelopment checks if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction checks if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetDatabaseConnectionString returns the database connection string
func (c *Config) GetDatabaseConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

// Configuration errors
var (
	ErrMissingStripeSecretKey        = errors.New("STRIPE_SECRET_KEY is required")
	ErrMissingStripeWebhookSecret    = errors.New("STRIPE_WEBHOOK_SECRET is required")
	ErrMissingStripeConnectClientID  = errors.New("STRIPE_CONNECT_CLIENT_ID is required")
	ErrMissingDatabaseHost          = errors.New("DB_HOST is required")
	ErrMissingDatabaseUser          = errors.New("DB_USER is required")
	ErrMissingDatabaseName          = errors.New("DB_NAME is required")
	ErrMissingJWTSecret           = errors.New("JWT_SECRET is required")
)

// Default configuration for testing
func DefaultConfig() *Config {
	return &Config{
		Environment: "test",
		Server: ServerConfig{
			Port:         "8090",
			Host:         "localhost",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         "5432",
			User:         "blytz",
			Password:     "test",
			Database:     "blytz_test",
			SSLMode:      "disable",
			MaxOpenConns: 5,
			MaxIdleConns: 2,
			MaxLifetime:  "30m",
		},
		Stripe: StripeConfig{
			SecretKey:              "sk_test_1234567890",
			PublishableKey:         "pk_test_1234567890",
			WebhookSecret:         "whsec_1234567890",
			PlatformAccountID:     "acct_1234567890",
			ConnectClientID:        "ca_1234567890",
			ApplicationFeePercent:  0.05,
			MinApplicationFee:     50,
			DefaultCurrency:        "usd",
			WebhookEndpoint:       "/api/v1/webhooks/stripe",
			SuccessURL:           "http://localhost:3000/payment/success",
			CancelURL:            "http://localhost:3000/payment/cancel",
		},
		Auth: AuthConfig{
			JWTSecret:      "test-jwt-secret",
			TokenExpiry:    1 * time.Hour,
			RefreshExpiry:  24 * time.Hour,
			AllowedOrigins: []string{"http://localhost:3000"},
		},
		Logging: LoggingConfig{
			Level:  "debug",
			Format: "text",
			Output: "stdout",
		},
	}
}