package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents JWT claims structure
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IssuedAt int64  `json:"iat"`
	ExpiresAt int64 `json:"exp"`
}

// GetAudience implements jwt.Claims interface
func (c JWTClaims) GetAudience() (string, error) {
	return "", nil
}

// GetExpirationTime implements jwt.Claims interface
func (c JWTClaims) GetExpirationTime() time.Time {
	return time.Unix(c.ExpiresAt, 0)
}

// GetIssuedAt implements jwt.Claims interface
func (c JWTClaims) GetIssuedAt() time.Time {
	return time.Unix(c.IssuedAt, 0)
}

// GetIssuer implements jwt.Claims interface
func (c JWTClaims) GetIssuer() (string, error) {
	return "", nil
}

// GetNotBefore implements jwt.Claims interface
func (c JWTClaims) GetNotBefore() time.Time {
	return time.Unix(c.ExpiresAt, 0)
}

// GetSubject implements jwt.Claims interface
func (c JWTClaims) GetSubject() (string, error) {
	return c.UserID, nil
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	SecretKey      string
	ExpirationTime time.Duration
	Issuer         string
}

// NewJWTConfig creates a new JWT configuration
func NewJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:      GetEnv("JWT_SECRET", "your-secret-key"),
		ExpirationTime: 24 * time.Hour,
		Issuer:         GetEnv("JWT_ISSUER", "blytz.live.latest"),
	}
}

// GenerateToken generates a new JWT token
func GenerateToken(userID, email, role string, config *JWTConfig) (string, error) {
	if config == nil {
		config = NewJWTConfig()
	}

	now := time.Now()
	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		IssuedAt: now.Unix(),
		ExpiresAt: now.Add(config.ExpirationTime).Unix(),
	}

	// Create a standard claims map to avoid interface issues
	standardClaims := jwt.MapClaims{
		"user_id":   claims.UserID,
		"email":     claims.Email,
		"role":      claims.Role,
		"expires_at": claims.ExpiresAt,
		"iss":       "blytz.live.latest",
		"iat":       time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, standardClaims)
	return token.SignedString([]byte(config.SecretKey))
}

// ValidateToken validates a JWT token and returns claims
func ValidateToken(tokenString string, config *JWTConfig) (*JWTClaims, error) {
	if config == nil {
		config = NewJWTConfig()
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid user_id claim")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid email claim")
		}

		role, ok := claims["role"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid role claim")
		}

		iat, ok := claims["iat"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid iat claim")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid exp claim")
		}

		return &JWTClaims{
			UserID:   userID,
			Email:    email,
			Role:     role,
			IssuedAt: int64(iat),
			ExpiresAt: int64(exp),
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken generates a new token with extended expiration
func RefreshToken(tokenString string, config *JWTConfig) (string, error) {
	claims, err := ValidateToken(tokenString, config)
	if err != nil {
		return "", fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Check if token is still valid (not expired yet)
	if time.Now().Unix() > claims.ExpiresAt {
		return "", fmt.Errorf("token has expired")
	}

	// Generate new token with same claims but extended expiration
	return GenerateToken(claims.UserID, claims.Email, claims.Role, config)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	// Check if header starts with "Bearer "
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	return authHeader
}

// IsTokenExpired checks if a token is expired
func IsTokenExpired(claims *JWTClaims) bool {
	return time.Now().Unix() > claims.ExpiresAt
}

// GetTokenRemainingTime returns remaining time until token expires
func GetTokenRemainingTime(claims *JWTClaims) time.Duration {
	remaining := claims.ExpiresAt - time.Now().Unix()
	if remaining < 0 {
		return 0
	}
	return time.Duration(remaining) * time.Second
}

// GenerateRefreshToken generates a refresh token
func GenerateRefreshToken(userID string, config *JWTConfig) (string, error) {
	if config == nil {
		config = NewJWTConfig()
	}

	// Refresh tokens have longer expiration (7 days)
	refreshConfig := *config
	refreshConfig.ExpirationTime = 7 * 24 * time.Hour

	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"type":     "refresh",
		"iat":      now.Unix(),
		"exp":      now.Add(refreshConfig.ExpirationTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(refreshConfig.SecretKey))
}

// ValidateRefreshToken validates a refresh token
func ValidateRefreshToken(tokenString string, config *JWTConfig) (string, error) {
	if config == nil {
		config = NewJWTConfig()
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "refresh" {
			return "", fmt.Errorf("invalid refresh token type")
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("invalid user_id claim in refresh token")
		}

		return userID, nil
	}

	return "", fmt.Errorf("invalid refresh token")
}