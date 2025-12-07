package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// User represents a user in the system
type User struct {
	ID            string             `json:"id" db:"id"`
	Email         string             `json:"email" db:"email"`
	PasswordHash  string             `json:"-" db:"password_hash"`
	DisplayName   string             `json:"display_name" db:"display_name"`
	PhoneNumber   string             `json:"phone_number" db:"phone_number"`
	AvatarURL     string             `json:"avatar_url" db:"avatar_url"`
	Role          string             `json:"role" db:"role"`
	IsActive      bool               `json:"is_active" db:"is_active"`
	EmailVerified bool               `json:"email_verified" db:"email_verified"`
	CreatedAt     time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" db:"updated_at"`
	LastLoginAt   pgtype.Timestamptz `json:"last_login_at" db:"last_login_at"`
}

// RefreshToken represents a refresh token for JWT
type RefreshToken struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsRevoked bool      `json:"is_revoked" db:"is_revoked"`
}

// UserSession represents an active user session
type UserSession struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	SessionToken string    `json:"session_token" db:"session_token"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
}

// Request/Response DTOs

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=100"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,max=20"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type ValidateTokenResponse struct {
	Valid   bool   `json:"valid"`
	UserID  string `json:"user_id,omitempty"`
	Email   string `json:"email,omitempty"`
	Role    string `json:"role,omitempty"`
	Expires int64  `json:"expires,omitempty"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// NewUser creates a new user with default values
func NewUser(email, password, displayName, phoneNumber string) *User {
	return &User{
		ID:            uuid.New().String(),
		Email:         email,
		PasswordHash:  "", // Should be set by service layer
		DisplayName:   displayName,
		PhoneNumber:   phoneNumber,
		Role:          "user",
		IsActive:      true,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(userID, token string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
}

// NewUserSession creates a new user session
func NewUserSession(userID, sessionToken, ipAddress, userAgent string, expiresAt time.Time) *UserSession {
	return &UserSession{
		ID:           uuid.New().String(),
		UserID:       userID,
		SessionToken: sessionToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
}
