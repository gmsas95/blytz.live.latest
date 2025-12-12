package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Config holds JWT configuration
type Config struct {
	Secret         string
	ExpirationTime time.Duration
	Issuer         string
	Audience       string
}

// NewConfig creates a new JWT configuration from environment variables
func NewConfig() *Config {
	return &Config{
		Secret:         getEnv("JWT_SECRET", ""),
		ExpirationTime: getEnvAsDuration("JWT_EXPIRATION_TIME", 24*time.Hour),
		Issuer:         getEnv("JWT_ISSUER", "blytz-platform"),
		Audience:       getEnv("JWT_AUDIENCE", "blytz-users"),
	}
}

// ValidateSecret checks if the JWT secret is secure
func (c *Config) ValidateSecret() error {
	if c.Secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	// Check if it's the default weak secret
	if c.Secret == "your-secret-key" || 
	   c.Secret == "dev_jwt_secret_key_32_characters_minimum" ||
	   len(c.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET is too weak or using default value. Use a secure 32+ character secret")
	}

	return nil
}

// GenerateSecureSecret generates a secure random JWT secret
func GenerateSecureSecret() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure secret: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// NewToken creates a new JWT token with claims
func (c *Config) NewToken(claims jwt.MapClaims) (string, error) {
	if err := c.ValidateSecret(); err != nil {
		return "", err
	}

	// Add standard claims
	now := time.Now()
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(c.ExpirationTime).Unix()
	claims["iss"] = c.Issuer
	claims["aud"] = c.Audience

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.Secret))
}

// ValidateToken validates a JWT token and returns claims
func (c *Config) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	if err := c.ValidateSecret(); err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken creates a new token with extended expiration
func (c *Config) RefreshToken(oldTokenString string) (string, error) {
	claims, err := c.ValidateToken(oldTokenString)
	if err != nil {
		return "", fmt.Errorf("failed to validate old token: %w", err)
	}

	// Remove old expiration and issue new token
	delete(claims, "exp")
	delete(claims, "iat")
	
	return c.NewToken(claims)
}

// ExtractUserID extracts user ID from token claims
func ExtractUserID(claims jwt.MapClaims) (string, error) {
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in token claims")
	}
	return userID, nil
}

// ExtractRole extracts user role from token claims
func ExtractRole(claims jwt.MapClaims) (string, error) {
	role, ok := claims["role"].(string)
	if !ok {
		return "", fmt.Errorf("role not found in token claims")
	}
	return role, nil
}

// LogSecurityEvent logs JWT-related security events
func LogSecurityEvent(logger *zap.Logger, event string, userID string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.Time("timestamp", time.Now()),
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	logger.Warn("JWT Security Event", fields...)
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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