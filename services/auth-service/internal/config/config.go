package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the auth service
type Config struct {
	DatabaseURL      string
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresHost     string `env:"POSTGRES_HOST"`
	PostgresPort     string `env:"POSTGRES_PORT"`
	PostgresDB       string `env:"POSTGRES_DB"`
	BetterAuthSecret string `env:"BETTER_AUTH_SECRET"`
	JWTSecret        string `env:"JWT_SECRET"`
	RedisURL         string `env:"REDIS_URL"`
	ServicePort      string `env:"PORT"`
	Environment      string `env:"ENVIRONMENT"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for development)
	_ = godotenv.Load()

	cfg := &Config{
		PostgresUser:     getEnvOrDefault("POSTGRES_USER", "blytz"),
		PostgresPassword: getEnvOrDefault("POSTGRES_PASSWORD", ""),
		PostgresHost:     getEnvOrDefault("POSTGRES_HOST", "postgres"),
		PostgresPort:     getEnvOrDefault("POSTGRES_PORT", "5432"),
		PostgresDB:       getEnvOrDefault("POSTGRES_DB", "blytz_prod"),
		BetterAuthSecret: os.Getenv("BETTER_AUTH_SECRET"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		RedisURL:         getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
		ServicePort:      getEnvOrDefault("PORT", "8084"),
		Environment:      getEnvOrDefault("NODE_ENV", "development"),
	}

	// Validate required secrets in production
	if cfg.IsProduction() {
		if cfg.JWTSecret == "" {
			return nil, fmt.Errorf("JWT_SECRET environment variable is required in production")
		}
		if cfg.BetterAuthSecret == "" {
			return nil, fmt.Errorf("BETTER_AUTH_SECRET environment variable is required in production")
		}
	} else {
		// Development-only fallbacks (clearly marked as unsafe)
		if cfg.JWTSecret == "" {
			cfg.JWTSecret = "UNSAFE-DEV-ONLY-jwt-secret-do-not-use-in-production"
		}
		if cfg.BetterAuthSecret == "" {
			cfg.BetterAuthSecret = "UNSAFE-DEV-ONLY-auth-secret-do-not-use-in-production"
		}
	}

	// Check if DATABASE_URL is provided (Dokploy style)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		cfg.DatabaseURL = databaseURL
		// Parse the DATABASE_URL to extract components for fallback usage
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
		// Construct the database URL from individual components (original behavior)
		encodedPassword := url.QueryEscape(cfg.PostgresPassword)
		cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.PostgresUser, encodedPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)
	}

	return cfg, nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
