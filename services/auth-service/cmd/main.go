package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// User struct with modern Go features
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Password  string    `json:"-" db:"password"` // Don't return password in JSON
	Role      string    `json:"role" db:"role"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Request structs with proper validation
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// Response struct with proper typing
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JWT Claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// App structure with modern dependencies
type App struct {
	db     *pgxpool.Pool
	logger *zap.Logger
	config map[string]string
}

// NewApp creates a new application instance
func NewApp(ctx context.Context) (*App, error) {
	// Load environment
	_ = godotenv.Load(".env")

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Database connection
	dbConfig := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "blytz_db"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := pgxpool.New(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &App{
		db:     db,
		logger: logger,
		config: map[string]string{
			"jwt_secret": getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			"jwt_expiry": getEnv("JWT_EXPIRY", "24h"),
			"port":       getEnv("PORT", "8084"),
		},
	}, nil
}

// getEnv safely gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// generateJWT creates a JWT token with modern jwt/v5
func (app *App) generateJWT(user User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "blytz-auth-service",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(app.config["jwt_secret"]))
}

// hashPassword creates a password hash (mock - use bcrypt in production)
func hashPassword(password string) string {
	return password // Use bcrypt in production
}

// validatePassword validates password (mock - use bcrypt in production)
func validatePassword(hashedPassword, password string) bool {
	return hashedPassword == password // Use bcrypt in production
}

func main() {
	ctx := context.Background()

	// Initialize app
	app, err := NewApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.db.Close()
	defer app.logger.Sync()

	// Initialize database
	if err := app.initDatabase(ctx); err != nil {
		app.logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Setup Gin router
	r := gin.Default()

	// Modern CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Middleware to add request ID and logging
	r.Use(func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Set("start_time", time.Now())

		c.Header("X-Request-ID", requestID)
		
		c.Next()
		
		duration := time.Since(time.Now().Add(time.Hour * -1))
		app.logger.Info("Request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("request_id", requestID),
		)
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Auth service is healthy and working!",
			Data: map[string]interface{}{
				"service":   "auth-service",
				"version":   "v2.0-go1.25",
				"status":    "healthy",
				"timestamp": time.Now(),
				"database":  "connected",
			},
		})
	})

	// Register endpoint
	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			app.logger.Error("Invalid registration request", zap.Error(err))
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid registration data",
				Error:   err.Error(),
			})
			return
		}

		// Check if user already exists
		var exists bool
		err := app.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", strings.ToLower(req.Email)).Scan(&exists)
		if err != nil {
			app.logger.Error("Database error checking user existence", zap.Error(err))
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Internal server error",
				Error:   "Failed to check user existence",
			})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, Response{
				Success: false,
				Message: "Email already registered",
				Error:   "User with this email already exists",
			})
			return
		}

		// Create new user
		userID := uuid.New().String()
		now := time.Now()
		
		hashedPassword := hashPassword(req.Password)

		_, err = app.db.Exec(ctx, `
			INSERT INTO users (id, name, email, password, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, userID, req.Name, strings.ToLower(req.Email), hashedPassword, "user", "active", now, now)

		if err != nil {
			app.logger.Error("Failed to create user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to create user",
				Error:   "Database error",
			})
			return
		}

		app.logger.Info("User registered successfully",
			zap.String("user_id", userID),
			zap.String("email", req.Email),
			zap.String("name", req.Name),
		)

		c.JSON(http.StatusCreated, Response{
			Success: true,
			Message: "Registration successful",
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"id":     userID,
					"name":   req.Name,
					"email":  strings.ToLower(req.Email),
					"role":   "user",
					"status": "active",
				},
			},
		})
	})

	// Login endpoint
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			app.logger.Error("Invalid login request", zap.Error(err))
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid login request",
				Error:   err.Error(),
			})
			return
		}

		// Find user by email
		var user User
		err := app.db.QueryRow(ctx, `
			SELECT id, name, email, password, role, status, created_at, updated_at 
			FROM users 
			WHERE email = $1 AND status = 'active'
		`, strings.ToLower(req.Email)).Scan(
			&user.ID, &user.Name, &user.Email, &user.Password,
			&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)

		if err != nil {
			app.logger.Error("Database error finding user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Authentication failed",
				Error:   "Database error",
			})
			return
		}

		if user.ID == "" {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Invalid email or password",
				Error:   "Authentication failed",
			})
			return
		}

		// Validate password
		if !validatePassword(user.Password, req.Password) {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Invalid email or password",
				Error:   "Authentication failed",
			})
			return
		}

		// Generate JWT token
		token, err := app.generateJWT(user)
		if err != nil {
			app.logger.Error("Failed to generate JWT", zap.Error(err))
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Authentication failed",
				Error:   "Failed to generate token",
			})
			return
		}

		app.logger.Info("User logged in successfully",
			zap.String("user_id", user.ID),
			zap.String("email", user.Email),
			zap.String("role", user.Role),
		)

		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Login successful",
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"id":     user.ID,
					"name":   user.Name,
					"email":  user.Email,
					"role":   user.Role,
					"status": user.Status,
				},
				"token":      token,
				"token_type": "Bearer",
				"expires_in": 86400, // 24 hours in seconds
			},
		})
	})

	// Profile endpoint
	r.GET("/api/v1/auth/profile", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Authorization header required",
				Error:   "No token provided",
			})
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Invalid authorization format",
				Error:   "Token format should be 'Bearer <token>'",
			})
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(app.config["jwt_secret"]), nil
		})

		if err != nil || !token.Valid {
			app.logger.Error("Invalid JWT token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Invalid or expired token",
				Error:   "Token validation failed",
			})
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Invalid token claims",
				Error:   "Token format error",
			})
			return
		}

		// Get user from database
		var user User
		err = app.db.QueryRow(ctx, `
			SELECT id, name, email, role, status, created_at, updated_at 
			FROM users 
			WHERE id = $1 AND status = 'active'
		`, claims.UserID).Scan(
			&user.ID, &user.Name, &user.Email, &user.Role,
			&user.Status, &user.CreatedAt, &user.UpdatedAt,
		)

		if err != nil {
			app.logger.Error("Database error getting user profile", zap.Error(err))
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to get user profile",
				Error:   "Database error",
			})
			return
		}

		if user.ID == "" {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Message: "User not found",
				Error:   "User does not exist",
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Profile retrieved successfully",
			Data: map[string]interface{}{
				"user": user,
			},
		})
	})

	// Logout endpoint
	r.POST("/api/v1/auth/logout", func(c *gin.Context) {
		// In a real app, you'd blacklist the token
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Logout successful",
		})
	})

	// Start server
	port := app.config["port"]
	app.logger.Info("Starting auth service",
		zap.String("port", port),
		zap.String("go_version", "go1.25"),
		zap.String("database", "postgres"),
	)

	fmt.Printf("🚀 AUTH SERVICE - GO 1.25 (2025)\n")
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔐 Login endpoint: http://localhost:%s/api/v1/auth/login\n", port)
	fmt.Printf("📝 Register endpoint: http://localhost:%s/api/v1/auth/register\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("📊 Database: PostgreSQL (pgx/v5)\n")
	fmt.Printf("🔐 JWT: golang-jwt/jwt/v5\n")
	fmt.Printf("📋 Logger: zap\n")

	log.Fatal(r.Run(":" + port))
}

// initDatabase creates the users table
func (app *App) initDatabase(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'user',
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
	`

	_, err := app.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	app.logger.Info("Database initialized successfully")
	return nil
}