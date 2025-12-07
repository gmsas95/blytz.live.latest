package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

// Response struct for API responses
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
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

// JWT mock - simple token generation
func generateJWT(userID string) string {
	return fmt.Sprintf("mock-jwt-token-%s-%d", userID, time.Now().Unix())
}

func main() {
	// Create Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Auth service is healthy and working",
			Data: map[string]interface{}{
				"service": "auth-service",
				"version": "v2.0-working",
				"status":  "healthy",
				"time":    time.Now(),
			},
		})
	})

	// Register endpoint
	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid registration data",
				Error:   err.Error(),
			})
			return
		}

		// Check if user already exists
		for _, user := range users {
			if strings.ToLower(user.Email) == strings.ToLower(req.Email) {
				c.JSON(http.StatusConflict, Response{
					Success: false,
					Message: "Email already registered",
					Error:   "User with this email already exists",
				})
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

		c.JSON(http.StatusCreated, Response{
			Success: true,
			Message: "Registration successful",
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"id":     newUser.ID,
					"name":   newUser.Name,
					"email":  newUser.Email,
					"role":   newUser.Role,
					"status": newUser.Status,
				},
			},
		})
	})

	// Login endpoint
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid login request",
				Error:   err.Error(),
			})
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
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Invalid email or password",
				Error:   "Authentication failed",
			})
			return
		}

		// Generate JWT token (mock)
		token := generateJWT(authenticatedUser.ID)

		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Login successful",
			Data: map[string]interface{}{
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
			},
		})
	})

	// Logout endpoint
	r.POST("/api/v1/auth/logout", func(c *gin.Context) {
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Logout successful",
		})
	})

	// Profile endpoint (mock)
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

		// For demo, return first user
		if len(users) > 0 {
			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: "Profile retrieved successfully",
				Data: map[string]interface{}{
					"user": users[0],
				},
			})
		} else {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Message: "User not found",
				Error:   "No users in database",
			})
		}
	})

	// Users list endpoint (admin)
	r.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Users retrieved successfully",
			Data: map[string]interface{}{
				"users": users,
				"count": len(users),
			},
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