package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	shared_errors "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
)

// User struct for authentication
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginRequest struct
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest struct
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}


// Mock user database
var users = []User{
	{
		ID:        "demo-user-123",
		Email:     "demo@blytz.app",
		Password:  "demo123",
		Name:      "Demo User",
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now(),
	},
	{
		ID:        "admin-user-456",
		Email:     "admin@blytz.app",
		Password:  "admin123",
		Name:      "Admin User",
		Role:      "admin",
		Status:    "active",
		CreatedAt: time.Now(),
	},
}

// JWT mock - simple token generation using shared package
func generateJWT(userID string) (string, error) {
	// For mock version, we'll create a simple JWT token
	// In production, this would use the shared JWT utilities
	return shared_utils.GenerateMockJWT(userID), nil
}

func main() {
	// Create Gin router
	r := gin.Default()

	// CORS middleware using shared package
	r.Use(shared_utils.CORSMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
			"service": "auth-service",
			"version": "v2.0-working",
			"status":  "healthy",
			"time":    time.Now(),
		})
	})

	// Register endpoint
	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			shared_utils.SendErrorResponse(c, shared_errors.ErrInvalidRequestBody)
			return
		}

		// Check if user already exists
		for _, user := range users {
			if strings.ToLower(user.Email) == strings.ToLower(req.Email) {
				shared_utils.SendErrorResponse(c, shared_errors.ConflictError("USER_EXISTS", "Email already registered"))
				return
			}
		}

		// Create new user
		newUser := User{
			ID:        fmt.Sprintf("user-%d", time.Now().UnixNano()),
			Name:      req.Name,
			Email:     strings.ToLower(req.Email),
			Password:  req.Password, // In production, hash this
			Role:      "user",
			Status:    "active",
			CreatedAt: time.Now(),
		}

		users = append(users, newUser)

		shared_utils.SendSuccessResponse(c, http.StatusCreated, map[string]interface{}{
			"user": map[string]interface{}{
				"id":     newUser.ID,
				"name":   newUser.Name,
				"email":  newUser.Email,
				"role":   newUser.Role,
				"status": newUser.Status,
			},
		})
	})

	// Login endpoint
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			shared_utils.SendErrorResponse(c, shared_errors.ErrInvalidRequestBody)
			return
		}

		// Find user by email
		var authenticatedUser *User
		for i := range users {
			if strings.ToLower(users[i].Email) == strings.ToLower(req.Email) {
				if users[i].Password == req.Password {
					authenticatedUser = &users[i]
					break
				}
			}
		}

		if authenticatedUser == nil {
			shared_utils.SendErrorResponse(c, shared_errors.AuthenticationError("INVALID_CREDENTIALS", "Invalid email or password"))
			return
		}

		// Generate JWT token (mock)
		token, err := generateJWT(authenticatedUser.ID)
		if err != nil {
			shared_utils.SendErrorResponse(c, shared_errors.InternalServerError("TOKEN_GENERATION_FAILED", "Failed to generate token"))
			return
		}

		shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
			"user": map[string]interface{}{
				"id":     authenticatedUser.ID,
				"name":   authenticatedUser.Name,
				"email":  authenticatedUser.Email,
				"role":   authenticatedUser.Role,
				"status": authenticatedUser.Status,
			},
			"token":      token,
			"token_type": "Bearer",
			"expires_in": 3600,
		})
	})

	// Logout endpoint
	r.POST("/api/v1/auth/logout", func(c *gin.Context) {
		shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
			"message": "Logout successful",
		})
	})

	// Profile endpoint (mock)
	r.GET("/api/v1/auth/profile", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			shared_utils.SendErrorResponse(c, shared_errors.AuthenticationError("NO_TOKEN", "Authorization header required"))
			return
		}

		// For demo, return first user
		if len(users) > 0 {
			shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
				"user": users[0],
			})
		} else {
			shared_utils.SendErrorResponse(c, shared_errors.NotFoundError("NO_USERS", "No users in database"))
		}
	})

	// Users list endpoint (admin)
	r.GET("/api/v1/users", func(c *gin.Context) {
		shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
			"users": users,
			"count": len(users),
		})
	})

	// Start server
	port := "8084"
	fmt.Printf("🚀 Auth Service starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔐 Login endpoint: http://localhost:%s/api/v1/auth/login\n", port)
	fmt.Printf("📝 Register endpoint: http://localhost:%s/api/v1/auth/register\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))

	log.Fatal(r.Run(":" + port))
}