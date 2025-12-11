package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	shared_errors "github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/models"
	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/repository"
)

type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	jwtSecret        string
	logger           *zap.Logger
}

func NewAuthService(db interface{}, jwtSecret string, logger *zap.Logger) *AuthService {
	// Type assertion to handle both sql.DB and other database types
	var sqlDB *sql.DB
	switch v := db.(type) {
	case *sql.DB:
		sqlDB = v
	default:
		// For now, we'll handle this case - in production this should be properly typed
		panic("database connection must be *sql.DB")
	}

	return &AuthService{
		userRepo:         repository.NewUserRepository(sqlDB),
		refreshTokenRepo: repository.NewRefreshTokenRepository(sqlDB),
		jwtSecret:        jwtSecret,
		logger:           logger,
	}
}

// RegisterUser creates a new user account
func (s *AuthService) RegisterUser(user *models.User) error {
	ctx := context.Background()

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, user.Email)
	if err != nil {
		s.logger.Error("Failed to check email existence", zap.Error(err))
		return shared_errors.NewDatabaseError("EMAIL_CHECK_FAILED", "Failed to check email existence")
	}

	if exists {
		return shared_errors.NewConflictError("EMAIL_EXISTS", "Email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return shared_errors.NewInternalError("PASSWORD_HASH_FAILED", "Failed to hash password")
	}

	user.PasswordHash = string(hashedPassword)
	user.IsActive = true
	user.EmailVerified = false
	user.Role = "user" // Default role

	// Create user
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return shared_errors.NewDatabaseError("USER_CREATION_FAILED", "Failed to create user")
	}

	s.logger.Info("User registered successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email))

	return nil
}

// LoginUser authenticates a user and returns JWT tokens
func (s *AuthService) LoginUser(email, password string) (string, error) {
	ctx := context.Background()

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("Login attempt with non-existent email",
			zap.String("email", email))
		return "", shared_errors.NewAuthenticationError("INVALID_CREDENTIALS", "Invalid email or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.logger.Warn("Login attempt with invalid password",
			zap.String("email", email),
			zap.String("user_id", user.ID))
		return "", shared_errors.NewAuthenticationError("INVALID_CREDENTIALS", "Invalid email or password")
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		s.logger.Warn("Failed to update last login",
			zap.String("user_id", user.ID),
			zap.Error(err))
		// Don't fail login if we can't update last login
	}

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		s.logger.Error("Failed to generate JWT",
			zap.String("user_id", user.ID),
			zap.Error(err))
		return "", shared_errors.NewInternalError("TOKEN_GENERATION_FAILED", "Failed to generate authentication token")
	}

	s.logger.Info("User logged in successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email))

	return token, nil
}

// ValidateToken validates a JWT token and returns user information
func (s *AuthService) ValidateToken(tokenString string) (*models.ValidateTokenResponse, error) {
	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return &models.ValidateTokenResponse{Valid: false}, nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return &models.ValidateTokenResponse{Valid: false}, nil
		}

		email, _ := claims["email"].(string)

		return &models.ValidateTokenResponse{
			Valid:   true,
			UserID:  userID,
			Email:   email,
		}, nil
	}

	return &models.ValidateTokenResponse{Valid: false}, nil
}

// GenerateJWT creates a JWT token for a user
func (s *AuthService) GenerateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// GenerateRefreshToken creates a refresh token
func (s *AuthService) GenerateRefreshToken(userID string) (string, error) {
	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	// Store in database
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7), // 7 days
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
	ctx := context.Background()

	err := s.refreshTokenRepo.Create(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	return token, nil
}

// RefreshAccessToken generates a new access token from refresh token
func (s *AuthService) RefreshAccessToken(refreshTokenString string) (string, error) {
	ctx := context.Background()

	// Validate refresh token
	refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshTokenString)
	if err != nil {
		return "", shared_errors.NewAuthenticationError("INVALID_REFRESH_TOKEN", "Invalid or expired refresh token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		return "", shared_errors.NewAuthenticationError("USER_NOT_FOUND", "User not found")
	}

	// Generate new access token
	return s.generateJWT(user)
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	ctx := context.Background()

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, shared_errors.NewNotFoundError("USER_NOT_FOUND", "User not found")
	}

	// Clear password hash for security
	user.PasswordHash = ""

	return user, nil
}

// UpdateUserProfile updates user profile information
func (s *AuthService) UpdateUserProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) error {
	// Get current user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return shared_errors.NewNotFoundError("USER_NOT_FOUND", "User not found")
	}

	// Update fields if provided
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	// Save changes
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update user profile",
			zap.String("user_id", userID),
			zap.Error(err))
		return shared_errors.NewDatabaseError("PROFILE_UPDATE_FAILED", "Failed to update profile")
	}

	s.logger.Info("User profile updated",
		zap.String("user_id", userID))

	return nil
}

// Logout revokes a refresh token
func (s *AuthService) Logout(refreshTokenString string) error {
	ctx := context.Background()

	err := s.refreshTokenRepo.Revoke(ctx, refreshTokenString)
	if err != nil {
		s.logger.Warn("Failed to revoke refresh token during logout", zap.Error(err))
		// Don't fail logout if token doesn't exist
	}

	return nil
}

// LogoutAll revokes all refresh tokens for a user
func (s *AuthService) LogoutAll(userID string) error {
	ctx := context.Background()

	err := s.refreshTokenRepo.RevokeAllForUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to revoke all refresh tokens",
			zap.String("user_id", userID),
			zap.Error(err))
		return shared_errors.NewDatabaseError("LOGOUT_ALL_FAILED", "Failed to logout from all devices")
	}

	s.logger.Info("User logged out from all devices",
		zap.String("user_id", userID))

	return nil
}

// generateJWT creates a JWT token for a user (private method)
func (s *AuthService) generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
