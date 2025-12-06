package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"go.uber.org/zap"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
	RedisURL          string
	Logger            *zap.Logger
}

// RateLimiter implements Redis-based rate limiting
type RateLimiter struct {
	redisClient *redis.Client
	config      RateLimiterConfig
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	// Parse Redis URL
	redisAddr := config.RedisURL
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // No password by default
		DB:       0,
	})

	return &RateLimiter{
		redisClient: redisClient,
		config:      config,
	}
}

// Middleware returns the rate limiting middleware
func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		key := r.getRateLimitKey(c)
		
		// Check current request count
		val, err := r.redisClient.Get(c.Request.Context(), key).Int()
		if err != nil && err != redis.Nil {
			// Log error but allow request
			r.config.Logger.Error("Rate limiter Redis error", zap.Error(err))
			c.Next()
			return
		}

		if val >= r.config.RequestsPerMinute {
			r.config.Logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.Int("current", val),
				zap.Int("limit", r.config.RequestsPerMinute))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"message": "Too many requests, please try again later",
				"code": "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := r.redisClient.Pipeline()
		pipe.Incr(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, time.Minute)
		_, err = pipe.Exec(c.Request.Context())
		if err != nil {
			r.config.Logger.Error("Rate limiter pipeline error", zap.Error(err))
		}

		c.Next()
	})
}

// getRateLimitKey generates a rate limit key for the request
func (r *RateLimiter) getRateLimitKey(c *gin.Context) string {
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	
	// Create a unique key based on IP and User-Agent
	key := fmt.Sprintf("rate_limit:%s:%s", clientIP, userAgent)
	
	// If user is authenticated, use user ID instead
	if userID, exists := c.Get("userID"); exists {
		key = fmt.Sprintf("rate_limit:user:%v", userID)
	}
	
	return key
}

// Close closes the Redis connection
func (r *RateLimiter) Close() error {
	return r.redisClient.Close()
}

// InMemoryRateLimiter implements a simple in-memory rate limiter for development
type InMemoryRateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(limit int, window time.Duration) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Middleware returns the rate limiting middleware
func (r *InMemoryRateLimiter) Middleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		key := r.getRateLimitKey(c)
		now := time.Now()
		
		r.mutex.Lock()
		defer r.mutex.Unlock()
		
		// Remove old requests outside the window
		var validRequests []time.Time
		for _, reqTime := range r.requests[key] {
			if now.Sub(reqTime) < r.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		
		// Check if limit exceeded
		if len(validRequests) >= r.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"message": "Too many requests, please try again later",
				"code": "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		
		// Add current request
		validRequests = append(validRequests, now)
		r.requests[key] = validRequests
		
		c.Next()
	})
}

// getRateLimitKey generates a rate limit key for the request
func (r *InMemoryRateLimiter) getRateLimitKey(c *gin.Context) string {
	clientIP := c.ClientIP()
	
	// If user is authenticated, use user ID instead
	if userID, exists := c.Get("userID"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	
	return fmt.Sprintf("ip:%s", clientIP)
}