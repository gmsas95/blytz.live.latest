package services

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/gmsas95/blytz.live.latest/services/auth-service/internal/models"
)

func setupMockDB(t *testing.T) (*AuthService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}

	logger := zap.NewNop()
	authService := NewAuthService(db, "test-jwt-secret", logger)

	return authService, mock
}

func TestAuthService_RegisterUser(t *testing.T) {
	authService, mock := setupMockDB(t)

	t.Run("Successful registration", func(t *testing.T) {
		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: "password123",
			DisplayName:  "Test User",
		}

		// Expect email existence check
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1 AND is_active = true\)`).
			WithArgs(user.Email).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Expect user creation
		mock.ExpectExec(`INSERT INTO users`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := authService.RegisterUser(user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Duplicate email", func(t *testing.T) {
		authService, mock := setupMockDB(t)

		user := &models.User{
			Email:        "duplicate@example.com",
			PasswordHash: "password123",
			DisplayName:  "Test User",
		}

		// Expect email existence check - returns true (email exists)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1 AND is_active = true\)`).
			WithArgs(user.Email).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		err := authService.RegisterUser(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EMAIL_EXISTS")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthService_LoginUser(t *testing.T) {
	t.Run("Valid credentials", func(t *testing.T) {
		authService, mock := setupMockDB(t)

		// Generate a real bcrypt hash for "password123"
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		now := time.Now()

		// Use .* to match the multiline SQL query with flexible whitespace
		mock.ExpectQuery(`SELECT .* FROM users WHERE email = \$1`).
			WithArgs("test@example.com").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "email", "password_hash", "display_name", "phone_number",
				"avatar_url", "role", "is_active", "email_verified",
				"created_at", "updated_at", "last_login_at",
			}).AddRow(
				"user-123", "test@example.com", string(passwordHash), "Test User", "",
				"", "user", true, false,
				now, now, now,
			))

		mock.ExpectExec(`UPDATE users SET last_login_at`).
			WithArgs("user-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		token, err := authService.LoginUser("test@example.com", "password123")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid password", func(t *testing.T) {
		authService, mock := setupMockDB(t)

		// bcrypt hash for "password123"
		passwordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMye.IjVQhqXgXd5HwJHj8aXVRfKOjLk5Ky"

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1 AND is_active = true`).
			WithArgs("test@example.com").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "email", "password_hash", "display_name", "phone_number",
				"avatar_url", "role", "is_active", "email_verified",
				"created_at", "updated_at", "last_login_at",
			}).AddRow(
				"user-123", "test@example.com", passwordHash, "Test User", "",
				"", "user", true, false,
				nil, nil, nil,
			))

		_, err := authService.LoginUser("test@example.com", "wrongpassword")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "INVALID_CREDENTIALS")
	})

	t.Run("User not found", func(t *testing.T) {
		authService, mock := setupMockDB(t)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1 AND is_active = true`).
			WithArgs("nonexistent@example.com").
			WillReturnRows(sqlmock.NewRows([]string{}))

		_, err := authService.LoginUser("nonexistent@example.com", "password123")
		assert.Error(t, err)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	authService, _ := setupMockDB(t)

	t.Run("Valid token", func(t *testing.T) {
		// Generate a real token to validate
		user := &models.User{
			ID:    "user-123",
			Email: "test@example.com",
			Role:  "user",
		}

		token, err := authService.GenerateJWT(user)
		assert.NoError(t, err)

		response, err := authService.ValidateToken(token)
		assert.NoError(t, err)
		assert.True(t, response.Valid)
		assert.Equal(t, "user-123", response.UserID)
		assert.Equal(t, "test@example.com", response.Email)
	})

	t.Run("Invalid token", func(t *testing.T) {
		response, err := authService.ValidateToken("invalid.token.here")
		assert.NoError(t, err)
		assert.False(t, response.Valid)
	})

	t.Run("Empty token", func(t *testing.T) {
		response, err := authService.ValidateToken("")
		assert.NoError(t, err)
		assert.False(t, response.Valid)
	})
}

func TestAuthService_GenerateJWT(t *testing.T) {
	authService, _ := setupMockDB(t)

	user := &models.User{
		ID:    "user-123",
		Email: "test@example.com",
		Role:  "admin",
	}

	token, err := authService.GenerateJWT(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the generated token
	response, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.True(t, response.Valid)
	assert.Equal(t, user.ID, response.UserID)
	assert.Equal(t, user.Email, response.Email)
}
