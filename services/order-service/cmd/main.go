package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	shared_utils "github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
	shared_errors "github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
	"github.com/gmsas95/blytz.live.latest/services/order-service/internal/models"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(db *gorm.DB, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		db:     db,
		logger:  logger,
	}
}

// ListOrders handles listing all orders with pagination
func (h *OrderHandler) ListOrders(c *gin.Context) {
	page, perPage := shared_utils.GetPaginationParams(c)
	offset := (page - 1) * perPage

	var orders []models.Order
	var total int64

	// Count total orders
	if err := h.db.Model(&models.Order{}).Count(&total).Error; err != nil {
		h.logger.Error("Failed to count orders", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("COUNT_ORDERS_FAILED", "Failed to count orders"))
		return
	}

	// Get orders with pagination
	err := h.db.Preload("OrderItems").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		h.logger.Error("Failed to list orders", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("LIST_ORDERS_FAILED", "Failed to list orders"))
		return
	}

	pagination := shared_utils.CalculatePagination(page, perPage, total)
	shared_utils.SendPaginatedResponse(c, http.StatusOK, orders, pagination)
}

// GetOrder handles getting an order by ID
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"id": "Order ID is required",
		})
		return
	}

	var order models.Order
	err := h.db.Preload("OrderItems").Where("id = ?", orderID).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			shared_utils.SendErrorResponse(c, shared_errors.NewNotFoundError("ORDER_NOT_FOUND", "Order not found"))
			return
		}
		h.logger.Error("Failed to get order", zap.String("order_id", orderID), zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("GET_ORDER_FAILED", "Failed to get order"))
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, order)
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	ProductID       string                `json:"product_id" binding:"required"`
	ProductName     string                `json:"product_name" binding:"required"`
	ProductImage    string                `json:"product_image,omitempty"`
	Quantity        int                   `json:"quantity" binding:"required,min=1"`
	Price           int64                 `json:"price" binding:"required,min=0"`
	Currency        string                `json:"currency,omitempty"`
	ShippingAddress models.Address        `json:"shipping_address" binding:"required"`
	BillingAddress  models.Address        `json:"billing_address,omitempty"`
	Notes           string                `json:"notes,omitempty"`
	PaymentMethod   string                `json:"payment_method,omitempty"`
	AuctionID       string                `json:"auction_id,omitempty"`
}

// CreateOrder handles creating a new order
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request_body": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Get user ID from context (would be set by auth middleware in production)
	userID, exists := c.Get("user_id")
	if !exists {
		// For now, use a dummy user ID (in real app, this would come from authentication)
		userID = c.GetHeader("X-User-ID")
		if userID == "" {
			userID = "dummy-user-id" // Default for testing
		}
	}

	// Set default currency if not provided
	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Calculate total amount
	totalAmount := req.Price * int64(req.Quantity)

	// Create order
	order := models.Order{
		UserID:          userID.(string),
		ProductID:       req.ProductID,
		ProductName:     req.ProductName,
		ProductImage:    req.ProductImage,
		Quantity:        req.Quantity,
		Price:           req.Price,
		TotalAmount:     totalAmount,
		Currency:        req.Currency,
		Status:          string(models.OrderStatusPending),
		PaymentStatus:   string(models.PaymentStatusPending),
		PaymentMethod:   req.PaymentMethod,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		Notes:           req.Notes,
	}

	// Set auction ID if provided
	if req.AuctionID != "" {
		order.AuctionID = &req.AuctionID
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create order
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create order", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("CREATE_ORDER_FAILED", "Failed to create order"))
		return
	}

	// Create order item
	orderItem := models.OrderItem{
		OrderID:     order.ID,
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
		Quantity:    req.Quantity,
		Price:       req.Price,
		TotalPrice:  totalAmount,
	}

	if err := tx.Create(&orderItem).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create order item", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("CREATE_ORDER_ITEM_FAILED", "Failed to create order item"))
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		h.logger.Error("Failed to commit transaction", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("TRANSACTION_FAILED", "Failed to complete order creation"))
		return
	}

	// Reload order with items
	h.db.Preload("OrderItems").First(&order, order.ID)

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusCreated, "Order created successfully", order)
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateOrderStatus handles updating order status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"id": "Order ID is required",
		})
		return
	}

	var req UpdateOrderStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request_body": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate status transition
	var order models.Order
	if err := h.db.Where("id = ?", orderID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			shared_utils.SendErrorResponse(c, shared_errors.NewNotFoundError("ORDER_NOT_FOUND", "Order not found"))
			return
		}
		h.logger.Error("Failed to get order", zap.String("order_id", orderID), zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("GET_ORDER_FAILED", "Failed to get order"))
		return
	}

	// Check if status transition is valid
	if !h.isValidStatusTransition(models.OrderStatus(order.Status), models.OrderStatus(req.Status)) {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_STATUS_TRANSITION", fmt.Sprintf("Cannot transition from %s to %s", order.Status, req.Status)))
		return
	}

	// Update status
	err := h.db.Model(&order).Update("status", req.Status).Error
	if err != nil {
		h.logger.Error("Failed to update order status", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("UPDATE_ORDER_STATUS_FAILED", "Failed to update order status"))
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Order status updated successfully", nil)
}

// isValidStatusTransition checks if the status transition is valid
func (h *OrderHandler) isValidStatusTransition(currentStatus, newStatus models.OrderStatus) bool {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending:   {models.OrderStatusProcessing, models.OrderStatusConfirmed, models.OrderStatusCancelled},
		models.OrderStatusProcessing: {models.OrderStatusConfirmed, models.OrderStatusShipped, models.OrderStatusCancelled},
		models.OrderStatusConfirmed:  {models.OrderStatusShipped, models.OrderStatusCancelled},
		models.OrderStatusShipped:    {models.OrderStatusDelivered},
		models.OrderStatusDelivered:  {}, // Final state
		models.OrderStatusCancelled:  {models.OrderStatusRefunded},
		models.OrderStatusRefunded:   {}, // Final state
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

// CancelOrder handles cancelling an order
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"id": "Order ID is required",
		})
		return
	}

	// Get current order status
	var order models.Order
	if err := h.db.Where("id = ?", orderID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			shared_utils.SendErrorResponse(c, shared_errors.NewNotFoundError("ORDER_NOT_FOUND", "Order not found"))
			return
		}
		h.logger.Error("Failed to get order", zap.String("order_id", orderID), zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("GET_ORDER_FAILED", "Failed to get order"))
		return
	}

	// Check if order can be cancelled
	if order.Status == string(models.OrderStatusDelivered) || order.Status == string(models.OrderStatusCancelled) || order.Status == string(models.OrderStatusRefunded) {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("ORDER_CANNOT_BE_CANCELLED", "Order cannot be cancelled in current status"))
		return
	}

	// Update status to cancelled
	err := h.db.Model(&order).Update("status", string(models.OrderStatusCancelled)).Error
	if err != nil {
		h.logger.Error("Failed to cancel order", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("CANCEL_ORDER_FAILED", "Failed to cancel order"))
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Order cancelled successfully", nil)
}

// GetUserOrders handles getting orders for a specific user
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"user_id": "User ID is required",
		})
		return
	}

	page, perPage := shared_utils.GetPaginationParams(c)
	offset := (page - 1) * perPage

	var orders []models.Order
	var total int64

	// Count total orders for user
	if err := h.db.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		h.logger.Error("Failed to count user orders", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("COUNT_USER_ORDERS_FAILED", "Failed to count user orders"))
		return
	}

	// Get orders with pagination
	err := h.db.Preload("OrderItems").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		h.logger.Error("Failed to get user orders", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("GET_USER_ORDERS_FAILED", "Failed to get user orders"))
		return
	}

	pagination := shared_utils.CalculatePagination(page, perPage, total)
	shared_utils.SendPaginatedResponse(c, http.StatusOK, orders, pagination)
}

// GetSellerOrders handles getting orders for a specific seller
func (h *OrderHandler) GetSellerOrders(c *gin.Context) {
	sellerID := c.Param("seller_id")
	if sellerID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"seller_id": "Seller ID is required",
		})
		return
	}

	page, perPage := shared_utils.GetPaginationParams(c)
	offset := (page - 1) * perPage

	var orders []models.Order
	var total int64

	// Count total orders for seller (this is a simplified approach - in production you'd join with products table)
	if err := h.db.Model(&models.Order{}).Count(&total).Error; err != nil {
		h.logger.Error("Failed to count seller orders", zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("COUNT_SELLER_ORDERS_FAILED", "Failed to count seller orders"))
		return
	}

	// Get orders with pagination
	err := h.db.Preload("OrderItems").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		h.logger.Error("Failed to get seller orders", zap.String("seller_id", sellerID), zap.Error(err))
		shared_utils.SendErrorResponse(c, shared_errors.NewDatabaseError("GET_SELLER_ORDERS_FAILED", "Failed to get seller orders"))
		return
	}

	pagination := shared_utils.CalculatePagination(page, perPage, total)
	shared_utils.SendPaginatedResponse(c, http.StatusOK, orders, pagination)
}

// HealthCheck handles health check requests
func (h *OrderHandler) HealthCheck(c *gin.Context) {
	status := shared_utils.NewHealthStatus("ok")
	
	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		status.AddService("database", "error", "Database connection failed")
	} else if err := sqlDB.Ping(); err != nil {
		status.AddService("database", "error", "Database ping failed")
	} else {
		status.AddService("database", "ok", "Database connected successfully")
	}
	
	shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"service":  "order-service",
		"version":  "v1.0.0",
		"status":   status,
		"time":     time.Now(),
	})
}

// runMigrations creates the necessary database tables
func runMigrations(db *gorm.DB) error {
	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.Order{},
		&models.OrderItem{},
		&models.Cart{},
		&models.CartItem{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_orders_product_id ON orders(product_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_cart_user_id ON carts(user_id)",
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
	logger, err := shared_utils.NewDevelopmentLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5434/orders_db?sslmode=disable"
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

	// Initialize handlers
	orderHandler := NewOrderHandler(db, logger)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware using shared package
	router.Use(shared_utils.CORSMiddleware())

	// Request ID middleware
	router.Use(func(c *gin.Context) {
		c.Set("request_id", uuid.New().String())
		c.Next()
	})

	// Health check endpoint
	router.GET("/health", orderHandler.HealthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Order routes
		orders := v1.Group("/orders")
		{
			orders.GET("", orderHandler.ListOrders)                    // List all orders (admin)
			orders.GET("/:id", orderHandler.GetOrder)                 // Get order by ID
			orders.POST("", orderHandler.CreateOrder)                  // Create order
			orders.PUT("/:id", orderHandler.UpdateOrderStatus)        // Update order status
			orders.POST("/cancel/:id", orderHandler.CancelOrder)      // Cancel order
			orders.GET("/user/:user_id", orderHandler.GetUserOrders)   // Get user's orders
			orders.GET("/seller/:seller_id", orderHandler.GetSellerOrders) // Get seller's orders
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085" // Use port 8085 as per docker-compose.yml
	}

	logger.Info("Order Service starting",
		zap.String("port", port),
		zap.String("database", "PostgreSQL"),
		zap.String("version", "v1.0.0"),
	)

	fmt.Printf("🚀 Order Service starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("📦 Orders endpoint: http://localhost:%s/api/v1/orders\n", port)
	fmt.Printf("🗄️  Database: PostgreSQL\n")
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}