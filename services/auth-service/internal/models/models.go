package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// User represents a user in the system
type User struct {
	ID            string             `json:"id" gorm:"primaryKey"`
	Email         string             `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash  string             `json:"-" gorm:"not null"` // Hashed password, never returned in JSON
	DisplayName   string             `json:"display_name"`
	PhoneNumber   string             `json:"phone_number,omitempty"`
	AvatarURL     string             `json:"avatar_url,omitempty"`
	IsActive      bool               `json:"is_active" gorm:"default:true"`
	Role          string             `json:"role" gorm:"default:user"`
	EmailVerified bool               `json:"email_verified" gorm:"default:false"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	LastLoginAt   pgtype.Timestamptz `json:"last_login_at"`
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=50"`
	PhoneNumber string `json:"phone_number,omitempty" binding:"omitempty,e164"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
	Token   string `json:"token,omitempty"`
}

// ValidateTokenRequest represents token validation request
type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// ValidateTokenResponse represents token validation response
type ValidateTokenResponse struct {
	Valid   bool   `json:"valid"`
	UserID  string `json:"user_id,omitempty"`
	Email   string `json:"email,omitempty"`
	Message string `json:"message,omitempty"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name,omitempty" binding:"omitempty,min=2,max=50"`
	PhoneNumber string `json:"phone_number,omitempty" binding:"omitempty,e164"`
	AvatarURL   string `json:"avatar_url,omitempty" binding:"omitempty,url"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// RefreshToken represents a refresh token for JWT
type RefreshToken struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	IsRevoked bool      `json:"is_revoked" gorm:"default:false"`
}


// AuthRequest represents authentication request
type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// VerifyResponse represents verification response
type VerifyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}
