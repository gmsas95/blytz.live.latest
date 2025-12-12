package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LiveKitHandler struct {
	logger    *zap.Logger
	apiKey    string
	apiSecret string
	serverURL string
}

func NewLiveKitHandler(logger *zap.Logger) *LiveKitHandler {
	return &LiveKitHandler{
		logger:    logger,
		apiKey:    getEnv("LIVEKIT_API_KEY", "devkey"),
		apiSecret: getEnv("LIVEKIT_API_SECRET", "secret"),
		serverURL: getEnv("LIVEKIT_SERVER_URL", "ws://localhost:7880"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// LiveKitGrant represents permissions for a LiveKit token
type LiveKitGrant struct {
	RoomJoin     bool   `json:"roomJoin"`
	RoomCreate   bool   `json:"roomCreate"`
	Room         string `json:"room"`
	CanPublish   bool   `json:"canPublish"`
	CanSubscribe bool   `json:"canSubscribe"`
}

// LiveKitClaims represents JWT claims for LiveKit
type LiveKitClaims struct {
	Exp      int64         `json:"exp"`
	Iss      string        `json:"iss"`
	Nbf      int64         `json:"nbf"`
	Sub      string        `json:"sub"`
	Name     string        `json:"name,omitempty"`
	Video    *LiveKitGrant `json:"video,omitempty"`
	Metadata string        `json:"metadata,omitempty"`
}

// GenerateToken generates a LiveKit token for connecting to a room
func (h *LiveKitHandler) GenerateToken(c *gin.Context) {
	// Require authentication
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	room := c.Query("room")
	role := c.Query("role") // "viewer" or "broadcaster"

	if room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room parameter is required"})
		return
	}

	if role == "" {
		role = "viewer" // Default to viewer
	}

	// Validate role
	if role != "viewer" && role != "broadcaster" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be 'viewer' or 'broadcaster'"})
		return
	}

	// Check if we're in development mode (secret not configured)
	if h.apiSecret == "secret" {
		h.logger.Warn("LiveKit API secret not configured, using development mode")
	}

	// Generate real JWT token
	token, err := h.generateJWT(userID, room, role)
	if err != nil {
		h.logger.Error("Failed to generate LiveKit token",
			zap.String("room", room),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("Generated LiveKit token",
		zap.String("room", room),
		zap.String("role", role),
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"token":     token,
			"room":      room,
			"role":      role,
			"serverUrl": h.serverURL,
		},
	})
}

// generateJWT creates a signed JWT token for LiveKit
func (h *LiveKitHandler) generateJWT(userID, room, role string) (string, error) {
	now := time.Now()
	exp := now.Add(6 * time.Hour) // Token valid for 6 hours

	// Set permissions based on role
	grant := &LiveKitGrant{
		RoomJoin:     true,
		Room:         room,
		CanSubscribe: true,
	}

	if role == "broadcaster" {
		grant.CanPublish = true
		grant.RoomCreate = true
	}

	claims := LiveKitClaims{
		Exp:   exp.Unix(),
		Iss:   h.apiKey,
		Nbf:   now.Unix(),
		Sub:   userID,
		Video: grant,
	}

	// Create JWT header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Base64URL encode
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	signingInput := headerB64 + "." + claimsB64
	mac := hmac.New(sha256.New, []byte(h.apiSecret))
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}
