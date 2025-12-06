package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
func NewRateLimiter(config RateLimiterConfig) (*RateLimiter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.RedisURL,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RateLimiter{
		redisClient: rdb,
		config:      config,
	}, nil
}

// RateLimit creates a Gin middleware for rate limiting
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx := context.Background()

		// Get current request count
		current, err := rl.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			rl.config.Logger.Error("Failed to get rate limit count", zap.Error(err))
			c.Next()
			return
		}

		// Check if rate limit exceeded
		if current >= rl.config.RequestsPerMinute {
			rl.config.Logger.Warn("Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.Int("requests", current))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Increment request count with expiration
		pipe := rl.redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)

		if _, err := pipe.Exec(ctx); err != nil {
			rl.config.Logger.Error("Failed to increment rate limit", zap.Error(err))
		}

		c.Next()
	}
}

// RateLimitByPath creates rate limiting by specific path patterns
func (rl *RateLimiter) RateLimitByPath(pathLimits map[string]int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.Request.URL.Path

		// Find matching path limit
		limit := rl.config.RequestsPerMinute // default limit
		for pattern, limitValue := range pathLimits {
			if strings.Contains(path, pattern) {
				limit = limitValue
				break
			}
		}

		key := fmt.Sprintf("rate_limit:%s:%s", clientIP, path)
		ctx := context.Background()

		// Get current request count
		current, err := rl.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			rl.config.Logger.Error("Failed to get rate limit count", zap.Error(err))
			c.Next()
			return
		}

		// Check if rate limit exceeded
		if current >= limit {
			rl.config.Logger.Warn("Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.String("path", path),
				zap.Int("requests", current),
				zap.Int("limit", limit))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Increment request count with expiration
		pipe := rl.redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)

		if _, err := pipe.Exec(ctx); err != nil {
			rl.config.Logger.Error("Failed to increment rate limit", zap.Error(err))
		}

		c.Next()
	}
}

// GetRateLimitHeaders returns rate limit headers for responses
func GetRateLimitHeaders(limit, remaining, reset int) map[string]string {
	return map[string]string{
		"X-RateLimit-Limit":     strconv.Itoa(limit),
		"X-RateLimit-Remaining": strconv.Itoa(remaining),
		"X-RateLimit-Reset":     strconv.Itoa(reset),
	}
}
