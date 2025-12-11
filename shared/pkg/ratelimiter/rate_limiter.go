package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter interface defines rate limiting behavior
type RateLimiter interface {
	Allow(key string) bool
	AllowN(key string, n int) bool
	Reset(key string)
}

// Config holds rate limiter configuration
type Config struct {
	RequestsPerSecond float64
	BurstSize        int
	RedisURL         string
	WindowDuration   time.Duration
	Logger           *zap.Logger
}

// TokenBucket implements token bucket rate limiting
type TokenBucket struct {
	capacity   int64
	tokens     int64
	refillRate int64
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tokensToAdd := int64(elapsed * float64(tb.refillRate))

	tb.tokens += tokensToAdd
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

// RedisRateLimiter implements distributed rate limiting using Redis
type RedisRateLimiter struct {
	client     *redis.Client
	config     Config
	buckets    map[string]*TokenBucket
	mutex      sync.RWMutex
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(config Config) (*RedisRateLimiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: "",
		DB:       2, // Use DB 2 for rate limiting
		PoolSize: 10,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		if config.Logger != nil {
			config.Logger.Warn("Redis connection failed for rate limiter, using local buckets", zap.Error(err))
		}
		// Fall back to local token buckets
		return &RedisRateLimiter{
			client:  client,
			config:  config,
			buckets: make(map[string]*TokenBucket),
		}, nil
	}

	return &RedisRateLimiter{
		client:  client,
		config:  config,
		buckets: make(map[string]*TokenBucket),
	}, nil
}

// Allow checks if a request is allowed for the given key
func (rl *RedisRateLimiter) Allow(key string) bool {
	// Try Redis first if available
	if rl.client != nil {
		allowed, err := rl.allowRedis(key)
		if err == nil {
			return allowed
		}
		// Fall back to local buckets if Redis fails
		if rl.config.Logger != nil {
			rl.config.Logger.Warn("Redis rate limiting failed, using local bucket", zap.Error(err))
		}
	}

	// Use local token bucket as fallback
	return rl.allowLocal(key)
}

// AllowN checks if N requests are allowed for the given key
func (rl *RedisRateLimiter) AllowN(key string, n int) bool {
	if rl.client != nil {
		allowed, err := rl.allowNRedis(key, n)
		if err == nil {
			return allowed
		}
	}

	return rl.allowNLocal(key, n)
}

// Reset resets the rate limiter for the given key
func (rl *RedisRateLimiter) Reset(key string) {
	if rl.client != nil {
		ctx := context.Background()
		rl.client.Del(ctx, rl.getRedisKey(key))
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	delete(rl.buckets, key)
}

// allowRedis implements Redis-based rate limiting
func (rl *RedisRateLimiter) allowRedis(key string) (bool, error) {
	ctx := context.Background()
	redisKey := rl.getRedisKey(key)

	// Use Redis INCR with expiration for sliding window
	current, err := rl.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	// Set expiration on first request
	if current == 1 {
		rl.client.Expire(ctx, redisKey, rl.config.WindowDuration)
	}

	// Check if under limit
	return current <= int64(rl.config.BurstSize), nil
}

// allowNRedis implements Redis-based rate limiting for N requests
func (rl *RedisRateLimiter) allowNRedis(key string, n int) (bool, error) {
	ctx := context.Background()
	redisKey := rl.getRedisKey(key)

	// Use Redis INCRBY for atomic increment
	current, err := rl.client.IncrBy(ctx, redisKey, int64(n)).Result()
	if err != nil {
		return false, err
	}

	// Set expiration on first request
	if current == int64(n) {
		rl.client.Expire(ctx, redisKey, rl.config.WindowDuration)
	}

	// Check if under limit
	return current <= int64(rl.config.BurstSize), nil
}

// allowLocal implements local token bucket rate limiting
func (rl *RedisRateLimiter) allowLocal(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = NewTokenBucket(
			int64(rl.config.BurstSize),
			int64(rl.config.RequestsPerSecond),
		)
		rl.buckets[key] = bucket
	}

	return bucket.Allow()
}

// allowNLocal implements local token bucket rate limiting for N requests
func (rl *RedisRateLimiter) allowNLocal(key string, n int) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = NewTokenBucket(
			int64(rl.config.BurstSize),
			int64(rl.config.RequestsPerSecond),
		)
		rl.buckets[key] = bucket
	}

	return bucket.Allow()
}

// getRedisKey generates Redis key for rate limiting
func (rl *RedisRateLimiter) getRedisKey(key string) string {
	return fmt.Sprintf("rate_limit:%s", key)
}

// GinMiddleware creates a Gin middleware for rate limiting
func (rl *RedisRateLimiter) GinMiddleware(keyExtractor func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyExtractor(c)
		if key == "" {
			key = "anonymous"
		}

		if !rl.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Maximum %d requests per %v allowed", 
					rl.config.BurstSize, rl.config.WindowDuration),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DefaultKeyExtractor extracts key from IP address
func DefaultKeyExtractor(c *gin.Context) string {
	return c.ClientIP()
}

// UserKeyExtractor extracts key from user ID (if authenticated)
func UserKeyExtractor(c *gin.Context) string {
	// Try to get user ID from context (should be set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return fmt.Sprintf("user:%s", id)
		}
	}
	// Fall back to IP
	return DefaultKeyExtractor(c)
}

// APIKeyExtractor extracts key from API key
func APIKeyExtractor(c *gin.Context) string {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return fmt.Sprintf("api_key:%s", apiKey)
	}
	return DefaultKeyExtractor(c)
}

// NewDefaultConfig creates a default rate limiter configuration
func NewDefaultConfig() Config {
	return Config{
		RequestsPerSecond: 10.0,  // 10 requests per second
		BurstSize:        100,    // Allow burst of 100 requests
		RedisURL:         "localhost:6379",
		WindowDuration:    time.Minute, // 1 minute window
	}
}

// NewStrictConfig creates a strict rate limiter configuration for high-load scenarios
func NewStrictConfig() Config {
	return Config{
		RequestsPerSecond: 5.0,   // 5 requests per second
		BurstSize:        50,     // Allow burst of 50 requests
		RedisURL:         "localhost:6379",
		WindowDuration:    time.Minute, // 1 minute window
	}
}

// NewRelaxedConfig creates a relaxed rate limiter configuration for internal services
func NewRelaxedConfig() Config {
	return Config{
		RequestsPerSecond: 100.0, // 100 requests per second
		BurstSize:        1000,   // Allow burst of 1000 requests
		RedisURL:         "localhost:6379",
		WindowDuration:    time.Minute, // 1 minute window
	}
}