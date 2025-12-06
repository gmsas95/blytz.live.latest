package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimiterConfig holds configuration for rate limiter
type RateLimiterConfig struct {
	RequestsPerSecond int
	BurstSize        int
	KeyExtractor     func(*gin.Context) string
}

// InMemoryRateLimiter provides in-memory rate limiting
type InMemoryRateLimiter struct {
	clients map[string]*ClientLimiter
	config  *RateLimiterConfig
	mutex   sync.RWMutex
}

// ClientLimiter tracks rate limits per client
type ClientLimiter struct {
	tokens    float64
	lastSeen  time.Time
	mutex     sync.Mutex
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(config *RateLimiterConfig) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		clients: make(map[string]*ClientLimiter),
		config:  config,
	}
}

// Middleware returns a Gin middleware for rate limiting
func (r *InMemoryRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := r.config.KeyExtractor(c)
		
		r.mutex.RLock()
		client, exists := r.clients[key]
		r.mutex.RUnlock()

		if !exists {
			r.mutex.Lock()
			client = &ClientLimiter{
				tokens:   float64(r.config.BurstSize),
				lastSeen: time.Now(),
			}
			r.clients[key] = client
			r.mutex.Unlock()
		}

		client.mutex.Lock()
		defer client.mutex.Unlock()

		now := time.Now()
		elapsed := now.Sub(client.lastSeen).Seconds()
		client.tokens += elapsed * float64(r.config.RequestsPerSecond)
		client.lastSeen = now

		if client.tokens > float64(r.config.BurstSize) {
			client.tokens = float64(r.config.BurstSize)
		}

		if client.tokens >= 1 {
			client.tokens--
			c.Next()
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": int((1 - client.tokens) * float64(time.Second) / float64(r.config.RequestsPerSecond)),
			})
			c.Abort()
		}
	}
}

// DefaultRateLimiter creates a default rate limiter middleware
func DefaultRateLimiter(requestsPerSecond, burstSize int) gin.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:        burstSize,
		KeyExtractor: func(c *gin.Context) string {
			// Use IP address as key
			return c.ClientIP()
		},
	}

	limiter := NewInMemoryRateLimiter(config)
	return limiter.Middleware()
}

// UserBasedRateLimiter creates a rate limiter based on user ID
func UserBasedRateLimiter(requestsPerSecond, burstSize int) gin.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:        burstSize,
		KeyExtractor: func(c *gin.Context) string {
			// Try to get user ID from context, fallback to IP
			if userID, exists := c.Get("user_id"); exists {
				return fmt.Sprintf("user:%v", userID)
			}
			return fmt.Sprintf("ip:%s", c.ClientIP())
		},
	}

	limiter := NewInMemoryRateLimiter(config)
	return limiter.Middleware()
}

// APIKeyRateLimiter creates a rate limiter based on API key
func APIKeyRateLimiter(requestsPerSecond, burstSize int) gin.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:        burstSize,
		KeyExtractor: func(c *gin.Context) string {
			// Try to get API key from header
			apiKey := c.GetHeader("X-API-Key")
			if apiKey != "" {
				return fmt.Sprintf("api_key:%s", apiKey)
			}
			return fmt.Sprintf("ip:%s", c.ClientIP())
		},
	}

	limiter := NewInMemoryRateLimiter(config)
	return limiter.Middleware()
}

// CleanupOldClients removes inactive clients to prevent memory leaks
func (r *InMemoryRateLimiter) CleanupOldClients(maxAge time.Duration) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	for key, client := range r.clients {
		client.mutex.Lock()
		if now.Sub(client.lastSeen) > maxAge {
			delete(r.clients, key)
		}
		client.mutex.Unlock()
	}
}

// GetStats returns rate limiter statistics
func (r *InMemoryRateLimiter) GetStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_clients"] = len(r.clients)
	stats["requests_per_second"] = r.config.RequestsPerSecond
	stats["burst_size"] = r.config.BurstSize

	return stats
}