package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/handlers"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/routes"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/database"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/services"
	"github.com/gmsas95/blytz-mvp/shared/pkg/auth"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Response struct for API responses
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// New stripe service with optimized database and Redis
func NewStripeService() (*database.Database, *services.ConnectedAccountService, *handlers.ConnectedAccountHandler, *services.PaymentService, *handlers.PaymentHandler, *services.TransferService, *handlers.TransferHandler, *services.PayoutService, *handlers.PayoutHandler, *services.WebhookService, *handlers.WebhookHandler, *handlers.PaymentMethodHandler, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize database
	dbConfig := &database.Config{
		Host:            cfg.PostgresHost,
		Port:            cfg.PostgresPort,
		User:            cfg.PostgresUser,
		Password:        cfg.PostgresPassword,
		DBName:          cfg.PostgresDB,
		SSLMode:         "disable",
		MaxOpenConns:    50,
		MaxIdleConns:    25,
		ConnMaxLifetime: 300 * time.Second,
		ConnMaxIdleTime: 60 * time.Second,
	}

	db, err := database.NewDatabase(dbConfig, logger)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run auto-migration
	if err := db.AutoMigrate(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to run auto-migration: %w", err)
	}

	// Initialize Stripe client
	stripeClient, err := services.NewStripeClient(&config.StripeConfig{
		SecretKey:      cfg.SecretKey,
		PublishableKey: cfg.PublishableKey,
		Environment:    cfg.Environment,
	})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to initialize Stripe client: %w", err)
	}

	// Initialize connected account service
	connectedAccountService := services.NewConnectedAccountService(
		db.GetDB(),
		stripeClient,
		cfg,
		logger,
	)

	// Initialize connected account handler
	connectedAccountHandler := handlers.NewConnectedAccountHandler(
		connectedAccountService,
		logger,
	)

	// Initialize payment service
	paymentService := services.NewPaymentService(
		db.GetDB(),
		stripeClient,
		cfg,
		logger,
	)

	// Initialize payment handler
	paymentHandler := handlers.NewPaymentHandler(
		paymentService,
		logger,
	)

	// Initialize transfer service
	transferService := services.NewTransferService(
		db.GetDB(),
		stripeClient,
		cfg,
		logger,
	)

	// Initialize transfer handler
	transferHandler := handlers.NewTransferHandler(
		transferService,
		logger,
	)

	// Initialize payout service
	payoutService := services.NewPayoutService(
		db.GetDB(),
		stripeClient,
		cfg,
		logger,
	)

	// Initialize payout handler
	payoutHandler := handlers.NewPayoutHandler(
		payoutService,
		logger,
	)

	// Initialize webhook service
	webhookService := services.NewWebhookService(
		db.GetDB(),
		stripeClient,
		cfg,
		logger,
	)

	// Initialize webhook handler
	webhookHandler := handlers.NewWebhookHandler(
		webhookService,
		logger,
	)

	// Initialize payment method handler
	paymentMethodHandler := handlers.NewPaymentMethodHandler(paymentService, stripeClient, logger)

	return db, connectedAccountService, connectedAccountHandler, paymentService, paymentHandler, transferService, transferHandler, payoutService, payoutHandler, webhookService, webhookHandler, paymentMethodHandler, nil
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Health check with database status
func health(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database health
		healthData := db.HealthCheck()

		// Get statistics from database
		var connectedAccounts, totalTransfers, totalPayouts, totalWebhookEvents int64
		db.GetDB().Raw("SELECT COUNT(*) FROM connected_accounts").Scan(&connectedAccounts)
		db.GetDB().Raw("SELECT COUNT(*) FROM transfers").Scan(&totalTransfers)
		db.GetDB().Raw("SELECT COUNT(*) FROM payouts").Scan(&totalPayouts)
		db.GetDB().Raw("SELECT COUNT(*) FROM stripe_events").Scan(&totalWebhookEvents)

		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Stripe service is healthy and working!",
			Data: map[string]interface{}{
				"service":              "stripe-service",
				"version":              "v1.0",
				"status":               healthData["status"],
				"timestamp":            time.Now(),
				"database_status":      healthData,
				"connected_accounts":   connectedAccounts,
				"total_transfers":      totalTransfers,
				"total_payouts":        totalPayouts,
				"total_webhook_events": totalWebhookEvents,
			},
		})
	}
}

// Start stripe service
func main() {
	// Initialize service components
	db, connectedAccountService, connectedAccountHandler, paymentService, paymentHandler, transferService, transferHandler, payoutService, payoutHandler, webhookService, webhookHandler, paymentMethodHandler, err := NewStripeService()
	if err != nil {
		log.Fatal("Failed to initialize stripe service:", err)
	}
	defer db.Close()

	// Use services to avoid unused variable warnings
	_ = connectedAccountService
	_ = paymentService
	_ = transferService
	_ = payoutService
	_ = webhookService
	_ = paymentMethodHandler

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Setup Gin router with performance optimizations
	gin.SetMode(gin.ReleaseMode)
	r := routes.SetupRouter(connectedAccountHandler, paymentHandler, transferHandler, payoutHandler, webhookHandler, paymentMethodHandler)

	// Initialize auth client
	authClient := auth.NewAuthClient(cfg.AuthServiceURL)

	// Setup middleware
	authMiddleware := auth.GinAuthMiddleware(authClient)
	adminAuthMiddleware := func(c *gin.Context) {
		// Check if user is admin
		userRole, exists := c.Get("userRole")
		if !exists || userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", health(db))

		// Connected account routes
		routes.ConnectedAccountRoutes(v1, connectedAccountHandler, authMiddleware)
		routes.ConnectedAccountAdminRoutes(v1, connectedAccountHandler, adminAuthMiddleware)

		// Payment routes
		routes.PaymentRoutes(v1, paymentHandler, authMiddleware)
		routes.PaymentAdminRoutes(v1, paymentHandler, adminAuthMiddleware)

		// Transfer and payout routes
		routes.TransferPayoutRoutes(v1, transferHandler, payoutHandler, authMiddleware, adminAuthMiddleware)

		// Webhook routes
		routes.WebhookRoutes(v1, webhookHandler)
		routes.WebhookAdminRoutes(v1, webhookHandler, adminAuthMiddleware)

		// Payment method routes
		routes.PaymentMethodRoutes(v1, paymentMethodHandler, authMiddleware)
		routes.PaymentMethodAdminRoutes(v1, paymentMethodHandler, adminAuthMiddleware)
	}

	port := getEnv("PORT", "8095")

	fmt.Printf("🚀 STRIPE SERVICE - v1.0\n")
	fmt.Printf("📊 Health check: http://localhost:%s/api/v1/health\n", port)
	fmt.Printf("💳 Payment processing with Stripe enabled\n")
	fmt.Printf("🔗 Connected Account Management: http://localhost:%s/api/v1/accounts\n", port)
	fmt.Printf("💰 Payment Processing: http://localhost:%s/api/v1/payments\n", port)
	fmt.Printf("💸 Transfer Management: http://localhost:%s/api/v1/transfers\n", port)
	fmt.Printf("💵 Payout Management: http://localhost:%s/api/v1/payouts\n", port)
	fmt.Printf("🔗 Webhook Endpoint: http://localhost:%s/api/v1/webhooks/stripe\n", port)
	fmt.Printf("💳 Payment Methods: http://localhost:%s/api/v1/payment-methods\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("🗄️  Database: PostgreSQL with optimized connection pools\n")
	fmt.Printf("🎯 Status: Ready to process payments, transfers, payouts, and webhooks!\n")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
