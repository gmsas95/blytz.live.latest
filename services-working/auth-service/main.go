package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Simple user struct
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Don't return password
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Request structs
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Auth service with in-memory storage
type AuthService struct {
	userList  []User
	mu        sync.RWMutex
	jwtStore  map[string]string // token -> user_id
}

// New auth service
func NewAuthService() *AuthService {
	return &AuthService{
		userList: []User{
			{
				ID:        "demo-user-123",
				Name:      "Demo User",
				Email:     "demo@blytz.app",
				Password:  "demo123",
				Role:      "user",
				Status:    "active",
				CreatedAt: time.Now(),
			},
			{
				ID:        "admin-user-456",
				Name:      "Admin User",
				Email:     "admin@blytz.app",
				Password:  "admin123",
				Role:      "admin",
				Status:    "active",
				CreatedAt: time.Now(),
			},
		},
		jwtStore: make(map[string]string),
	}
}

// Generate simple JWT token
func (s *AuthService) generateToken(userID string) string {
	token := fmt.Sprintf("token-%s-%d", userID, time.Now().Unix())
	s.jwtStore[token] = userID
	return token
}

// Validate JWT token
func (s *AuthService) validateToken(token string) (string, bool) {
	userID, exists := s.jwtStore[token]
	return userID, exists
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// JSON response helper
func writeJSONResponse(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// Health check
func (s *AuthService) health(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Auth service is healthy and working!",
		Data: map[string]interface{}{
			"service":   "auth-service",
			"version":   "v2.0-working",
			"status":    "healthy",
			"timestamp": time.Now(),
			"users":     len(s.userList),
		},
	})
}

// Register endpoint
func (s *AuthService) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid registration data",
			Error:   err.Error(),
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	for _, user := range s.userList {
		if strings.ToLower(user.Email) == strings.ToLower(req.Email) {
			writeJSONResponse(w, http.StatusConflict, Response{
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

	s.userList = append(s.userList, newUser)

	log.Printf("User registered: %s (%s)", newUser.Name, newUser.Email)

	writeJSONResponse(w, http.StatusCreated, Response{
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
}

// Login endpoint
func (s *AuthService) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid login request",
			Error:   err.Error(),
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find user by email
	var authenticatedUser *User
	for i := range s.userList {
		if strings.ToLower(s.userList[i].Email) == strings.ToLower(req.Email) {
			if s.userList[i].Password == req.Password {
				authenticatedUser = &s.userList[i]
				break
			}
		}
	}

	if authenticatedUser == nil {
		writeJSONResponse(w, http.StatusUnauthorized, Response{
			Success: false,
			Message: "Invalid email or password",
			Error:   "Authentication failed",
		})
		return
	}

	// Generate token
	token := s.generateToken(authenticatedUser.ID)

	log.Printf("User logged in: %s (%s)", authenticatedUser.Name, authenticatedUser.Email)

	writeJSONResponse(w, http.StatusOK, Response{
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
}

// Profile endpoint
func (s *AuthService) profile(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSONResponse(w, http.StatusUnauthorized, Response{
			Success: false,
			Message: "Authorization header required",
			Error:   "No token provided",
		})
		return
	}

	// Extract token from "Bearer <token>"
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		writeJSONResponse(w, http.StatusUnauthorized, Response{
			Success: false,
			Message: "Invalid authorization format",
			Error:   "Token format should be 'Bearer <token>'",
		})
		return
	}

	// Validate token
	userID, valid := s.validateToken(token)
	if !valid {
		writeJSONResponse(w, http.StatusUnauthorized, Response{
			Success: false,
			Message: "Invalid or expired token",
			Error:   "Token validation failed",
		})
		return
	}

	// Find user by ID
	s.mu.RLock()
	defer s.mu.RUnlock()

	var user *User
	for i := range s.userList {
		if s.userList[i].ID == userID {
			user = &s.userList[i]
			break
		}
	}

	if user == nil {
		writeJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Message: "User not found",
			Error:   "User does not exist",
		})
		return
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Profile retrieved successfully",
		Data: map[string]interface{}{
			"user": *user,
		},
	})
}

// Users list endpoint
func (s *AuthService) usersList(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Users retrieved successfully",
		Data: map[string]interface{}{
			"users": s.userList,
			"count": len(s.userList),
		},
	})
}

// Start auth service
func main() {
	service := NewAuthService()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085" // Different port to avoid conflicts
	}

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/auth/register", corsMiddleware(service.register))
	http.Handle("/api/v1/auth/login", corsMiddleware(service.login))
	http.Handle("/api/v1/auth/profile", corsMiddleware(service.profile))
	http.Handle("/api/v1/users", corsMiddleware(service.usersList))

	fmt.Printf("🚀 AUTH SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔐 Login endpoint: http://localhost:%s/api/v1/auth/login\n", port)
	fmt.Printf("📝 Register endpoint: http://localhost:%s/api/v1/auth/register\n", port)
	fmt.Printf("👥 Users list: http://localhost:%s/api/v1/users\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("📊 Total users: %d\n", len(service.userList))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}