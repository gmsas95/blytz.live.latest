package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
	"github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
)

// ServiceConfig represents a microservice configuration
type ServiceConfig struct {
	Name      string   `json:"name"`
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	Instances []string `json:"instances"`
	Healthy   bool     `json:"healthy"`
}

// ServiceRegistry manages service discovery and health checking
type ServiceRegistry struct {
	services map[string]*ServiceConfig
	mu       sync.RWMutex
	logger   *zap.Logger
}

// LoadBalancer implements round-robin load balancing
type LoadBalancer struct {
	services map[string]int // service -> current index
	mu       sync.RWMutex
}

// RateLimiter implements simple rate limiting
type RateLimiter struct {
	clients map[string][]time.Time
	mu      sync.RWMutex
	limit   int
	window  time.Duration
}

// Gateway represents the API gateway
type Gateway struct {
	registry     *ServiceRegistry
	loadBalancer *LoadBalancer
	rateLimiter  *RateLimiter
	logger       *zap.Logger
	httpClient   *http.Client
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(logger *zap.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*ServiceConfig),
		logger:   logger,
	}
}

// RegisterService registers a new service
func (sr *ServiceRegistry) RegisterService(name, host string, port int) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instance := fmt.Sprintf("%s:%d", host, port)
	sr.services[name] = &ServiceConfig{
		Name:      name,
		Host:      host,
		Port:      port,
		Instances: []string{instance},
		Healthy:   true,
	}

	sr.logger.Info("Service registered",
		zap.String("service", name),
		zap.String("instance", instance),
	)
}

// GetHealthyInstance returns a healthy instance using round-robin
func (sr *ServiceRegistry) GetHealthyInstance(serviceName string) (string, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	service, exists := sr.services[serviceName]
	if !exists {
		return "", errors.NewNotFoundError("SERVICE_NOT_FOUND", fmt.Sprintf("Service %s not found", serviceName))
	}

	if !service.Healthy || len(service.Instances) == 0 {
		return "", errors.NewExternalServiceError("SERVICE_UNHEALTHY", fmt.Sprintf("Service %s is unhealthy", serviceName), "")
	}

	return service.Instances[0], nil
}

// CheckHealth performs health checks on all registered services
func (sr *ServiceRegistry) CheckHealth() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for name, service := range sr.services {
		for _, instance := range service.Instances {
			healthURL := fmt.Sprintf("http://%s/health", instance)
			resp, err := client.Get(healthURL)
			
			wasHealthy := service.Healthy
			service.Healthy = (err == nil && resp != nil && resp.StatusCode == http.StatusOK)
			
			if resp != nil {
				resp.Body.Close()
			}

			if wasHealthy != service.Healthy {
				status := "unhealthy"
				if service.Healthy {
					status = "healthy"
				}
				sr.logger.Info("Service health status changed",
					zap.String("service", name),
					zap.String("instance", instance),
					zap.String("status", status),
				)
			}
		}
	}
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		services: make(map[string]int),
	}
}

// GetNextInstance returns the next instance using round-robin
func (lb *LoadBalancer) GetNextInstance(serviceName string, instances []string) string {
	if len(instances) == 0 {
		return ""
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	currentIndex := lb.services[serviceName]
	nextIndex := (currentIndex + 1) % len(instances)
	lb.services[serviceName] = nextIndex

	return instances[nextIndex]
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
}

// Allow checks if a request is allowed based on rate limiting
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	// Clean old requests
	if requests, exists := rl.clients[clientID]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.clients[clientID] = validRequests
	}

	// Check if under limit
	if len(rl.clients[clientID]) >= rl.limit {
		return false
	}

	// Add current request
	rl.clients[clientID] = append(rl.clients[clientID], now)
	return true
}

// NewGateway creates a new API gateway
func NewGateway(logger *zap.Logger) *Gateway {
	return &Gateway{
		registry:     NewServiceRegistry(logger),
		loadBalancer: NewLoadBalancer(),
		rateLimiter:  NewRateLimiter(100, time.Minute), // 100 requests per minute
		logger:       logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetupServices configures all microservices
func (g *Gateway) SetupServices() {
	// Register all services based on docker-compose.yml
	g.registry.RegisterService("auth-service", "auth-service", 8084)
	g.registry.RegisterService("product-service", "product-service", 8082)
	g.registry.RegisterService("auction-service", "auction-service", 8083)
	g.registry.RegisterService("order-service", "order-service", 8085)
	g.registry.RegisterService("payment-service", "payment-service", 8086)
	g.registry.RegisterService("chat-service", "chat-service", 8088)
	g.registry.RegisterService("logistics-service", "logistics-service", 8087)
	g.registry.RegisterService("livekit-service", "livekit-service", 8089)
}

// RequestIDMiddleware adds a unique request ID to each request
func (g *Gateway) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RateLimitMiddleware applies rate limiting to requests
func (g *Gateway) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()
		
		if !g.rateLimiter.Allow(clientID) {
			utils.SendErrorResponse(c, errors.NewNetworkError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded"))
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// LoggingMiddleware logs request information
func (g *Gateway) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after processing
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		requestID, _ := c.Get("request_id")

		if raw != "" {
			path = path + "?" + raw
		}

		g.logger.Info("Request processed",
			zap.String("request_id", requestID.(string)),
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		)
	}
}

// ProxyMiddleware creates a reverse proxy for the specified service
func (g *Gateway) ProxyMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get healthy instance
		instance, err := g.registry.GetHealthyInstance(serviceName)
		if err != nil {
			g.logger.Error("Failed to get healthy instance",
				zap.String("service", serviceName),
				zap.Error(err),
			)
			utils.SendErrorResponse(c, err)
			c.Abort()
			return
		}

		// Create target URL
		target, err := url.Parse(fmt.Sprintf("http://%s", instance))
		if err != nil {
			g.logger.Error("Failed to parse target URL",
				zap.String("service", serviceName),
				zap.String("instance", instance),
				zap.Error(err),
			)
			utils.SendErrorResponse(c, errors.NewInternalError("PROXY_ERROR", "Failed to create proxy"))
			c.Abort()
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Modify request to include service info
		proxy.Director = func(req *http.Request) {
			req.Host = target.Host
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			
			// Add custom headers
			if requestID, exists := c.Get("request_id"); exists {
				req.Header.Set("X-Request-ID", requestID.(string))
			}
			req.Header.Set("X-Forwarded-For", c.ClientIP())
			req.Header.Set("X-Forwarded-Proto", c.Request.URL.Scheme)
			req.Header.Set("X-Forwarded-Host", c.Request.Host)
		}

		// Handle proxy errors
		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			g.logger.Error("Proxy error",
				zap.String("service", serviceName),
				zap.String("instance", instance),
				zap.Error(err),
			)
			utils.SendErrorResponse(c, errors.NewExternalServiceError("PROXY_ERROR", "Service proxy error", err.Error()))
		}

		// Serve the request
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthHandler returns the health status of the gateway and all services
func (g *Gateway) HealthHandler(c *gin.Context) {
	health := utils.NewHealthStatus("ok")
	health.AddService("gateway", "healthy", "Gateway is running")

	// Check all services
	g.registry.CheckHealth()

	g.registry.mu.RLock()
	defer g.registry.mu.RUnlock()

	for name, service := range g.registry.services {
		status := "unhealthy"
		message := "Service is not responding"
		if service.Healthy {
			status = "healthy"
			message = "Service is responding normally"
		}
		health.AddService(name, status, message)
	}

	utils.SendSuccessResponse(c, http.StatusOK, health)
}

// ServiceDiscoveryHandler returns the list of registered services
func (g *Gateway) ServiceDiscoveryHandler(c *gin.Context) {
	g.registry.mu.RLock()
	defer g.registry.mu.RUnlock()

	services := make(map[string]interface{})
	for name, service := range g.registry.services {
		services[name] = gin.H{
			"name":      service.Name,
			"instances": service.Instances,
			"healthy":   service.Healthy,
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"services": services,
		"total":    len(services),
	})
}

// SetupRouter configures the gin router with all routes and middleware
func (g *Gateway) SetupRouter() *gin.Engine {
	// Set gin mode
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(g.LoggingMiddleware())
	router.Use(gin.Recovery())
	router.Use(utils.CORSMiddleware())
	router.Use(g.RequestIDMiddleware())
	router.Use(g.RateLimitMiddleware())

	// Health check endpoint
	router.GET("/health", g.HealthHandler)
	router.GET("/health/services", g.ServiceDiscoveryHandler)

	// API routes with service proxies
	api := router.Group("/api/v1")
	{
		// Auth service routes
		auth := api.Group("/auth")
		{
			auth.Any("/*path", g.ProxyMiddleware("auth-service"))
		}

		// Product service routes
		product := api.Group("/products")
		{
			product.Any("/*path", g.ProxyMiddleware("product-service"))
		}

		// Auction service routes
		auction := api.Group("/auctions")
		{
			auction.Any("/*path", g.ProxyMiddleware("auction-service"))
		}

		// Order service routes
		order := api.Group("/orders")
		{
			order.Any("/*path", g.ProxyMiddleware("order-service"))
		}

		// Payment service routes
		payment := api.Group("/payments")
		{
			payment.Any("/*path", g.ProxyMiddleware("payment-service"))
		}

		// Chat service routes
		chat := api.Group("/chat")
		{
			chat.Any("/*path", g.ProxyMiddleware("chat-service"))
		}

		// Logistics service routes
		logistics := api.Group("/logistics")
		{
			logistics.Any("/*path", g.ProxyMiddleware("logistics-service"))
		}

		// LiveKit service routes
		livekit := api.Group("/livekit")
		{
			livekit.Any("/*path", g.ProxyMiddleware("livekit-service"))
		}
	}

	// Legacy routes for backward compatibility
	router.Any("/auth/*path", g.ProxyMiddleware("auth-service"))
	router.Any("/products/*path", g.ProxyMiddleware("product-service"))
	router.Any("/auctions/*path", g.ProxyMiddleware("auction-service"))
	router.Any("/orders/*path", g.ProxyMiddleware("order-service"))
	router.Any("/payments/*path", g.ProxyMiddleware("payment-service"))
	router.Any("/chat/*path", g.ProxyMiddleware("chat-service"))
	router.Any("/logistics/*path", g.ProxyMiddleware("logistics-service"))
	router.Any("/livekit/*path", g.ProxyMiddleware("livekit-service"))

	return router
}

// StartHealthChecker starts the background health checker
func (g *Gateway) StartHealthChecker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			g.logger.Info("Health checker stopped")
			return
		case <-ticker.C:
			g.registry.CheckHealth()
		}
	}
}

func main() {
	// Initialize logger
	logger, err := utils.NewProductionLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting API Gateway")

	// Create gateway
	gateway := NewGateway(logger)

	// Setup services
	gateway.SetupServices()

	// Setup router
	router := gateway.SetupRouter()

	// Get port from environment or use default
	port := "8080"
	if envPort := utils.GetEnv("PORT", "8080"); envPort != "" {
		port = envPort
	}

	// Start health checker in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go gateway.StartHealthChecker(ctx)

	// Start server
	logger.Info("API Gateway starting",
		zap.String("port", port),
		zap.String("mode", gin.Mode()),
	)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	
	logger.Info("Shutting down API Gateway")
	
	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}
	
	logger.Info("API Gateway stopped")
}