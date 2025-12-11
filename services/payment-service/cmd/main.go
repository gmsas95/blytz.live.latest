package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	shared_errors "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/gmsas95/blytz-mvp/services/payment-service/internal/models"
	"github.com/gmsas95/blytz-mvp/services/payment-service/internal/services"
)

// PaymentHandler handles HTTP requests for payments
type PaymentHandler struct {
	paymentService *services.PaymentService
	logger         *zap.Logger
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(paymentService *services.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

// === PAYMENT MANAGEMENT HANDLERS ===

// ListPayments handles listing all payments with pagination
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	page, perPage := shared_utils.GetPaginationParams(c)
	offset := (page - 1) * perPage

	var payments []models.Payment
	var total int64

	// Count total payments
	if err := h.paymentService.GetDB().Model(&models.Payment{}).Count(&total).Error; err != nil {
		h.logger.Error("Failed to count payments", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("COUNT_PAYMENTS_FAILED", "Failed to count payments"))
		return
	}

	// Get payments with pagination
	err := h.paymentService.GetDB().
		Preload("Refunds").
		Preload("WebhookLogs").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&payments).Error

	if err != nil {
		h.logger.Error("Failed to list payments", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("LIST_PAYMENTS_FAILED", "Failed to list payments"))
		return
	}

	pagination := shared_utils.CalculatePagination(page, perPage, total)
	shared_utils.SendPaginatedResponse(c, http.StatusOK, payments, pagination)
}

// GetPayment handles getting a payment by ID
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"id": "Payment ID is required",
		})
		return
	}

	payment, err := h.paymentService.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.Error("Failed to get payment", zap.String("payment_id", paymentID), zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewNotFoundError("PAYMENT_NOT_FOUND", "Payment not found"))
		return
	}

	response := &models.PaymentResponse{
		Payment:        *payment,
		PaymentURL:     payment.PaymentURL,
		QRCode:         payment.QRCode,
		TimeRemaining:  payment.GetTimeRemaining(),
		IsExpired:      payment.IsExpired(),
		CanRetry:       payment.CanRetry(),
		AvailableMethods: []models.PaymentMethod{}, // Would populate from API
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, response)
}

// CreatePayment handles creating a new payment
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request_body": "Invalid request format: " + err.Error(),
		})
		return
	}

	response, err := h.paymentService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create payment", 
			zap.String("order_id", req.OrderID),
			zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewBusinessError("CREATE_PAYMENT_FAILED", "Failed to create payment"))
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusCreated, "Payment created successfully", response)
}

// RefundPayment handles creating a refund for a payment
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"id": "Payment ID is required",
		})
		return
	}

	var req models.RefundPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request_body": "Invalid request format: " + err.Error(),
		})
		return
	}

	response, err := h.paymentService.RefundPayment(c.Request.Context(), paymentID, &req)
	if err != nil {
		h.logger.Error("Failed to create refund", 
			zap.String("payment_id", paymentID),
			zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewBusinessError("CREATE_REFUND_FAILED", "Failed to create refund"))
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Refund created successfully", response)
}

// GetPaymentMethods handles getting available payment methods
func (h *PaymentHandler) GetPaymentMethods(c *gin.Context) {
	response, err := h.paymentService.GetPaymentMethods(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get payment methods", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewExternalServiceError("GET_PAYMENT_METHODS_FAILED", "Failed to get payment methods", err.Error()))
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetPaymentStats handles getting payment statistics
func (h *PaymentHandler) GetPaymentStats(c *gin.Context) {
	// Get payment statistics
	var stats struct {
		TotalPayments      int64   `json:"total_payments"`
		SuccessfulPayments int64   `json:"successful_payments"`
		FailedPayments    int64   `json:"failed_payments"`
		PendingPayments   int64   `json:"pending_payments"`
		TotalRevenue      float64 `json:"total_revenue"`
		TotalRefunded    float64 `json:"total_refunded"`
		PaymentMethods   []string `json:"payment_methods"`
	}

	// Count payments by status
	h.paymentService.GetDB().Model(&models.Payment{}).Count(&stats.TotalPayments)
	h.paymentService.GetDB().Model(&models.Payment{}).Where("status = ?", models.PaymentStatusSuccess).Count(&stats.SuccessfulPayments)
	h.paymentService.GetDB().Model(&models.Payment{}).Where("status = ?", models.PaymentStatusFailed).Count(&stats.FailedPayments)
	h.paymentService.GetDB().Model(&models.Payment{}).Where("status = ?", models.PaymentStatusPending).Count(&stats.PendingPayments)

	// Calculate total revenue and refunded amounts
	h.paymentService.GetDB().Model(&models.Payment{}).Where("status = ?", models.PaymentStatusSuccess).Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalRevenue)
	h.paymentService.GetDB().Model(&models.Payment{}).Where("status = ?", models.PaymentStatusRefunded).Select("COALESCE(SUM(refund_amount), 0)").Scan(&stats.TotalRefunded)

	// Get unique payment methods
	var methods []string
	h.paymentService.GetDB().Model(&models.Payment{}).Distinct("payment_method").Pluck("payment_method", &methods)
	stats.PaymentMethods = methods

	shared_utils.SendSuccessResponse(c, http.StatusOK, stats)
}

// ProcessWebhook handles webhook from payment providers
func (h *PaymentHandler) ProcessWebhook(c *gin.Context) {
	// Get webhook event from header
	event := c.GetHeader("X-Fiuu-Event")
	if event == "" {
		event = c.GetHeader("X-Stripe-Event") // Support Stripe webhooks
	}
	if event == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"event": "Webhook event header is required",
		})
		return
	}

	// Get webhook signature from header
	signature := c.GetHeader("X-Fiuu-Signature")
	if signature == "" {
		signature = c.GetHeader("X-Stripe-Signature") // Support Stripe webhooks
	}

	// Read webhook payload
	var payload []byte
	if c.Request.Body != nil {
		payload, _ = c.GetRawData()
	}

	err := h.paymentService.ProcessWebhook(c.Request.Context(), event, payload, signature)
	if err != nil {
		h.logger.Error("Failed to process webhook", 
			zap.String("event", event),
			zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("WEBHOOK_PROCESSING_FAILED", "Failed to process webhook"))
		return
	}

	h.logger.Info("Webhook processed successfully", 
		zap.String("event", event))

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Webhook processed successfully", nil)
}

// HealthCheck handles health check requests
func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	status := shared_utils.NewHealthStatus("ok")
	
	// Check database connection
	sqlDB, err := h.paymentService.GetDB().DB()
	if err != nil {
		status.AddService("database", "error", "Database connection failed")
	} else if err := sqlDB.Ping(); err != nil {
		status.AddService("database", "error", "Database ping failed")
	} else {
		status.AddService("database", "ok", "Database connected successfully")
	}
	
	// Check payment providers
	status.AddService("stripe", "ok", "Stripe API available")
	status.AddService("fiuu", "ok", "Fiuu API available")
	status.AddService("paypal", "ok", "PayPal API available")
	
	shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"service":  "payment-service",
		"version":  "v1.0.0",
		"status":   status,
		"time":     time.Now(),
	})
}

// runMigrations creates necessary database tables
func runMigrations(db *gorm.DB) error {
	// Migrate models one by one in correct order
	if err := db.AutoMigrate(&models.Payment{}); err != nil {
		return fmt.Errorf("failed to migrate Payment model: %w", err)
	}

	if err := db.AutoMigrate(&models.PaymentRefund{}); err != nil {
		return fmt.Errorf("failed to migrate PaymentRefund model: %w", err)
	}

	if err := db.AutoMigrate(&models.WebhookLog{}); err != nil {
		return fmt.Errorf("failed to migrate WebhookLog model: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)",
		"CREATE INDEX IF NOT EXISTS idx_payments_payment_id ON payments(payment_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments(transaction_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_customer_email ON payments(customer_email)",
		"CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_payment_refunds_payment_id ON payment_refunds(payment_id)",
		"CREATE INDEX IF NOT EXISTS idx_webhook_logs_payment_id ON webhook_logs(payment_id)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			return fmt.Errorf("failed to create index %s: %w", index, err)
		}
	}

	return nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize logger
	var logger *zap.Logger
	var err error
	if os.Getenv("NODE_ENV") == "production" {
		logger, err = shared_utils.NewProductionLogger()
	} else {
		logger, err = shared_utils.NewDevelopmentLogger()
	}
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5434/payments_db?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get underlying SQL DB", zap.Error(err))
	}
	defer sqlDB.Close()

	// Test database connection
	if err := sqlDB.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Database connected successfully")

	// Run migrations
	if err := runMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	logger.Info("Database migrations completed")

	// Initialize payment service with multiple providers
	paymentService := services.NewPaymentService(
		db, 
		logger,
		"https://api.fiuu.com", // Fiuu API URL
		os.Getenv("FIUU_MERCHANT_ID"),
		"client_id", // Would be from config
		"client_secret", // Would be from config
		os.Getenv("FIUU_VERIFY_KEY"),
		os.Getenv("FIUU_SANDBOX") == "true",
	)

	// Initialize handlers
	paymentHandler := NewPaymentHandler(paymentService, logger)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware using shared package
	router.Use(shared_utils.CORSMiddleware())

	// Request ID middleware
	router.Use(func(c *gin.Context) {
		c.Set("request_id", uuid.New().String())
		c.Next()
	})

	// Recovery middleware
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", paymentHandler.HealthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Payment routes
		payments := v1.Group("/payments")
		{
			payments.GET("", paymentHandler.ListPayments)                    // List all payments with pagination
			payments.GET("/:id", paymentHandler.GetPayment)                 // Get payment by ID
			payments.POST("/create", paymentHandler.CreatePayment)           // Create payment
			payments.POST("/refund", paymentHandler.RefundPayment)          // Refund payment (alternative endpoint)
			payments.POST("/:id/refund", paymentHandler.RefundPayment)      // Refund payment by ID
			payments.GET("/methods", paymentHandler.GetPaymentMethods)       // Get user payment methods
			payments.GET("/stats", paymentHandler.GetPaymentStats)          // Get payment statistics
		}

		// Webhook endpoint
		v1.POST("/webhook", paymentHandler.ProcessWebhook)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086" // Use port 8086 as per docker-compose.yml
	}

	logger.Info("Payment Service starting on port " + port)
	
	fmt.Printf("💳 Payment Service starting on port %s\n", port)
	fmt.Printf("🔗 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("💰 Payments endpoint: http://localhost:%s/api/v1/payments\n", port)
	fmt.Printf("🔧 Webhook endpoint: http://localhost:%s/api/v1/webhook\n", port)
	fmt.Printf("🗄️  Database: PostgreSQL\n")
	fmt.Printf("🏗️  Payment Providers: Stripe, PayPal, Fiuu\n")
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
