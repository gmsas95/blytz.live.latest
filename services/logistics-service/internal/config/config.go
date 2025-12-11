package config

import (
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DatabaseURL string
	Environment string
	LogLevel    string
	// NinjaVan Logistics Configuration
	NinjaVanClientID  string `env:"NINJAVAN_CLIENT_ID"`
	NinjaVanClientKey  string `env:"NINJAVAN_CLIENT_KEY"`
	NinjaVanEnvironment string `env:"NINJAVAN_ENVIRONMENT"`
	NinjaVanCountryCode string `env:"NINJAVAN_COUNTRY_CODE"`
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/logistics_db"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		// NinjaVan Configuration
		NinjaVanClientID:  getEnv("NINJAVAN_CLIENT_ID", ""),
		NinjaVanClientKey:  getEnv("NINJAVAN_CLIENT_KEY", ""),
		NinjaVanEnvironment: getEnv("NINJAVAN_ENVIRONMENT", "sandbox"),
		NinjaVanCountryCode: getEnv("NINJAVAN_COUNTRY_CODE", "my"),
	}
}

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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func InitDB(cfg *Config) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
}