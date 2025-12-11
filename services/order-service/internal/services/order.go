package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz.live.latest/services/order-service/internal/config"
	"github.com/gmsas95/blytz.live.latest/services/order-service/internal/models"
	"github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
)

type OrderService struct {
	db     *gorm.DB
	logger *zap.Logger
	config *config.Config
}

func NewOrderService(db *gorm.DB, logger *zap.Logger, config *config.Config) *OrderService {
	return &OrderService{
		db:     db,
		logger: logger,
		config: config,
	}
}

// GetDB returns database connection for use by handlers
func (s *OrderService) GetDB() *gorm.DB {
	return s.db
}

// === CART OPERATIONS ===

// GetOrCreateCart gets user's cart or creates empty one
func (s *OrderService) GetOrCreateCart(ctx context.Context, userID string) (*models.Cart, error) {
	var cart models.Cart
	err := s.db.Preload("Items").Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			cart = models.Cart{
				UserID:    userID,
				Items:     []models.CartItem{},
				Total:     0,
				ItemCount: 0,
			}
			if createErr := s.db.Create(&cart).Error; createErr != nil {
				s.logger.Error("Failed to create cart", zap.Error(createErr))
				return nil, fmt.Errorf("failed to create cart")
			}
			return &cart, nil
		}
		s.logger.Error("Failed to get cart", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get cart")
	}
	return &cart, nil
}

// AddToCart adds item to cart
func (s *OrderService) AddToCart(ctx context.Context, userID string, req *AddToCartRequest) (*models.Cart, error) {
	s.logger.Info("Adding to cart", zap.String("user_id", userID), zap.String("product_id", req.ProductID))

	// Get or create cart
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if item already exists in cart
	var existingItem models.CartItem
	err = s.db.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).First(&existingItem).Error
	
	if err == nil {
		// Update existing item quantity
		newQuantity := existingItem.Quantity + req.Quantity
		existingItem.Quantity = newQuantity
		existingItem.Total = existingItem.Price * int64(newQuantity)
		
		if updateErr := s.db.Save(&existingItem).Error; updateErr != nil {
			s.logger.Error("Failed to update cart item", zap.Error(updateErr))
			return nil, fmt.Errorf("failed to update cart item")
		}
	} else if err == gorm.ErrRecordNotFound {
		// Add new item
		cartItem := models.CartItem{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			AuctionID: req.AuctionID,
			Quantity:  req.Quantity,
			Price:     req.Price,
			Total:     req.Price * int64(req.Quantity),
		}
		
		if createErr := s.db.Create(&cartItem).Error; createErr != nil {
			s.logger.Error("Failed to create cart item", zap.Error(createErr))
			return nil, fmt.Errorf("failed to add item to cart")
		}
		
		cart.Items = append(cart.Items, cartItem)
	} else {
		s.logger.Error("Failed to check cart item", zap.Error(err))
		return nil, fmt.Errorf("failed to add item to cart")
	}

	// Recalculate cart totals
	if err := s.recalculateCartTotals(cart.ID); err != nil {
		return nil, err
	}

	// Return updated cart
	return s.GetOrCreateCart(ctx, userID)
}

// UpdateCartItem updates quantity of item in cart
func (s *OrderService) UpdateCartItem(ctx context.Context, userID string, itemID string, quantity int) (*models.Cart, error) {
	if quantity <= 0 {
		return s.RemoveFromCart(ctx, userID, itemID)
	}

	// Get cart
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update item
	var cartItem models.CartItem
	err = s.db.Where("id = ? AND cart_id = ?", itemID, cart.ID).First(&cartItem).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("cart item not found")
		}
		return nil, fmt.Errorf("failed to update cart item")
	}

	cartItem.Quantity = quantity
	cartItem.Total = cartItem.Price * int64(quantity)

	if updateErr := s.db.Save(&cartItem).Error; updateErr != nil {
		s.logger.Error("Failed to update cart item", zap.Error(updateErr))
		return nil, fmt.Errorf("failed to update cart item")
	}

	// Recalculate cart totals
	if err := s.recalculateCartTotals(cart.ID); err != nil {
		return nil, err
	}

	return s.GetOrCreateCart(ctx, userID)
}

// RemoveFromCart removes item from cart
func (s *OrderService) RemoveFromCart(ctx context.Context, userID string, itemID string) (*models.Cart, error) {
	// Get cart
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Delete item
	err = s.db.Where("id = ? AND cart_id = ?", itemID, cart.ID).Delete(&models.CartItem{}).Error
	if err != nil {
		s.logger.Error("Failed to remove cart item", zap.Error(err))
		return nil, fmt.Errorf("failed to remove cart item")
	}

	// Recalculate cart totals
	if err := s.recalculateCartTotals(cart.ID); err != nil {
		return nil, err
	}

	return s.GetOrCreateCart(ctx, userID)
}

// ClearCart clears all items from user's cart
func (s *OrderService) ClearCart(ctx context.Context, userID string) error {
	// Get cart
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	// Delete all items
	err = s.db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error
	if err != nil {
		s.logger.Error("Failed to clear cart", zap.Error(err))
		return fmt.Errorf("failed to clear cart")
	}

	// Update cart totals
	cart.Total = 0
	cart.ItemCount = 0
	if updateErr := s.db.Save(cart).Error; updateErr != nil {
		s.logger.Error("Failed to update cart totals", zap.Error(updateErr))
		return fmt.Errorf("failed to update cart")
	}

	return nil
}

// === ORDER OPERATIONS ===

// CreateOrder creates new order from cart or direct request
func (s *OrderService) CreateOrder(ctx context.Context, userID string, req *CreateOrderRequest) (*models.Order, error) {
	s.logger.Info("Creating order", zap.String("user_id", userID), zap.String("product_id", req.ProductID))

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if this is auction order vs regular order
	var order *models.Order
	var err error

	if req.AuctionID != nil && *req.AuctionID != "" {
		order, err = s.createAuctionOrder(ctx, userID, req)
	} else {
		order, err = s.createRegularOrder(ctx, userID, req)
	}

	if err != nil {
		return nil, err
	}

	// Reserve stock if needed
	if err := s.reserveStock(order.ProductID, order.Quantity); err != nil {
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}

	return order, nil
}

// createRegularOrder creates regular product order
func (s *OrderService) createRegularOrder(ctx context.Context, userID string, req *CreateOrderRequest) (*models.Order, error) {
	// Calculate total amount
	totalAmount := req.Price * int64(req.Quantity)

	// Create order
	order := &models.Order{
		UserID:        userID,
		AuctionID:     req.AuctionID,
		ProductID:     req.ProductID,
		ProductName:   req.ProductName,
		ProductImage:  req.ProductImage,
		Quantity:      req.Quantity,
		Price:         req.Price,
		TotalAmount:   totalAmount,
		Currency:      req.Currency,
		Status:        string(models.OrderStatusPending),
		PaymentStatus: string(models.PaymentStatusPending),
		PaymentMethod: req.PaymentMethod,
		ShippingAddress: models.Address{
			Name:        req.ShippingAddress.Name,
			Street:      req.ShippingAddress.Street,
			City:        req.ShippingAddress.City,
			State:       req.ShippingAddress.State,
			PostalCode:  req.ShippingAddress.PostalCode,
			Country:     req.ShippingAddress.Country,
			PhoneNumber: req.ShippingAddress.PhoneNumber,
		},
		BillingAddress: models.Address{
			Name:        req.BillingAddress.Name,
			Street:      req.BillingAddress.Street,
			City:        req.BillingAddress.City,
			State:       req.BillingAddress.State,
			PostalCode:  req.BillingAddress.PostalCode,
			Country:     req.BillingAddress.Country,
			PhoneNumber: req.BillingAddress.PhoneNumber,
		},
		Notes: req.Notes,
	}

	if createErr := s.db.Create(order).Error; createErr != nil {
		s.logger.Error("Failed to create order", zap.Error(createErr))
		return nil, fmt.Errorf("failed to create order")
	}

	return order, nil
}

// createAuctionOrder creates auction winner order
func (s *OrderService) createAuctionOrder(ctx context.Context, userID string, req *CreateOrderRequest) (*models.Order, error) {
	// For auction orders, validate auction win status
	// This would typically involve checking with auction service
	// For now, create similar to regular order
	
	return s.createRegularOrder(ctx, userID, req)
}

// GetOrder gets order by ID for user
func (s *OrderService) GetOrder(ctx context.Context, userID, orderID string) (*models.Order, error) {
	var order models.Order
	err := s.db.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		s.logger.Error("Failed to get order", zap.String("order_id", orderID), zap.Error(err))
		return nil, fmt.Errorf("failed to get order")
	}
	return &order, nil
}

// GetUserOrders gets all orders for user with pagination
func (s *OrderService) GetUserOrders(ctx context.Context, userID string, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	offset := (page - 1) * pageSize

	// Count total orders
	if err := s.db.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count orders", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count orders")
	}

	// Get orders with pagination
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		s.logger.Error("Failed to get user orders", zap.String("user_id", userID), zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get orders")
	}

	return orders, total, nil
}

// UpdateOrderStatus updates order status
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	// Validate status transition
	if !s.isValidOrderStatusTransition(status) {
		return fmt.Errorf("invalid order status")
	}

	err := s.db.Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("status", string(status)).Error

	if err != nil {
		s.logger.Error("Failed to update order status", zap.String("order_id", orderID), zap.Error(err))
		return fmt.Errorf("failed to update order status")
	}

	return nil
}

// UpdatePaymentStatus updates payment status
func (s *OrderService) UpdatePaymentStatus(ctx context.Context, orderID string, status models.PaymentStatus) error {
	err := s.db.Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("payment_status", string(status)).Error

	if err != nil {
		s.logger.Error("Failed to update payment status", zap.String("order_id", orderID), zap.Error(err))
		return fmt.Errorf("failed to update payment status")
	}

	return nil
}

// === HELPER METHODS ===

// recalculateCartTotals recalculates cart total and item count
func (s *OrderService) recalculateCartTotals(cartID string) error {
	var cartItems []models.CartItem
	err := s.db.Where("cart_id = ?", cartID).Find(&cartItems).Error
	if err != nil {
		return fmt.Errorf("failed to get cart items for recalculation")
	}

	var total int64
	var itemCount int

	for _, item := range cartItems {
		total += item.Total
		itemCount += item.Quantity
	}

	// Update cart totals
	err = s.db.Model(&models.Cart{}).
		Where("id = ?", cartID).
		Updates(map[string]interface{}{
			"total":      total,
			"item_count": itemCount,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to update cart totals")
	}

	return nil
}

// reserveStock reserves stock for order (would integrate with product service)
func (s *OrderService) reserveStock(productID string, quantity int) error {
	// This would integrate with product service
	// For now, just log the reservation
	s.logger.Info("Stock reservation", 
		zap.String("product_id", productID),
		zap.Int("quantity", quantity))
	return nil
}

// isValidOrderStatusTransition validates order status transitions
func (s *OrderService) isValidOrderStatusTransition(status models.OrderStatus) bool {
	validStatuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusProcessing,
		models.OrderStatusConfirmed,
		models.OrderStatusShipped,
		models.OrderStatusDelivered,
		models.OrderStatusCancelled,
		models.OrderStatusRefunded,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// === REQUEST TYPES ===

// AddToCartRequest represents add to cart request
type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	AuctionID *string `json:"auction_id,omitempty"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Price     int64  `json:"price" binding:"required,min=0"`
}

// CreateOrderRequest represents order creation request
type CreateOrderRequest struct {
	ProductID       string        `json:"product_id" binding:"required"`
	AuctionID       *string       `json:"auction_id,omitempty"`
	ProductName     string        `json:"product_name" binding:"required"`
	ProductImage    string        `json:"product_image,omitempty"`
	Quantity        int           `json:"quantity" binding:"required,min=1"`
	Price           int64         `json:"price" binding:"required,min=0"`
	Currency        string        `json:"currency" binding:"required,len=3"`
	PaymentMethod   string        `json:"payment_method,omitempty"`
	ShippingAddress AddressRequest `json:"shipping_address" binding:"required"`
	BillingAddress  AddressRequest `json:"billing_address" binding:"required"`
	Notes           string        `json:"notes,omitempty"`
}

// AddressRequest represents address request
type AddressRequest struct {
	Name        string `json:"name" binding:"required"`
	Street      string `json:"street" binding:"required"`
	City        string `json:"city" binding:"required"`
	State       string `json:"state" binding:"required"`
	PostalCode  string `json:"postal_code" binding:"required"`
	Country     string `json:"country" binding:"required"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// Validate validates CreateOrderRequest
func (r *CreateOrderRequest) Validate() error {
	if r.ProductID == "" {
		return errors.New("product_id is required")
	}
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if r.Price < 0 {
		return errors.New("price cannot be negative")
	}
	return nil
}

// Validate validates AddToCartRequest
func (r *AddToCartRequest) Validate() error {
	if r.ProductID == "" {
		return errors.New("product_id is required")
	}
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if r.Price < 0 {
		return errors.New("price cannot be negative")
	}
	return nil
}