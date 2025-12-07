package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Order struct
type Order struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	SellerID       string    `json:"seller_id"`
	ProductID      string    `json:"product_id"`
	AuctionID      *string   `json:"auction_id,omitempty"`
	ProductName    string    `json:"product_name"`
	ProductImage   string    `json:"product_image"`
	Quantity       int       `json:"quantity"`
	UnitPrice      float64   `json:"unit_price"`
	TotalAmount    float64   `json:"total_amount"`
	OrderType      string    `json:"order_type"` // "buy_now", "auction_win", "direct_purchase"
	Status         string    `json:"status"` // "pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "refunded"
	PaymentStatus  string    `json:"payment_status"` // "pending", "paid", "failed", "refunded"
	PaymentID      *string   `json:"payment_id,omitempty"`
	ShippingAddress Address   `json:"shipping_address"`
	TrackingNumber *string   `json:"tracking_number,omitempty"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	DeliveryDate   *time.Time `json:"delivery_date,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OrderItem struct
type OrderItem struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

// Address struct
type Address struct {
	Street    string `json:"street"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`
	Country   string `json:"country"`
	Phone     string `json:"phone"`
}

// Request structs
type CreateOrderRequest struct {
	UserID           string  `json:"user_id"`
	ProductID        string  `json:"product_id"`
	AuctionID        *string `json:"auction_id,omitempty"`
	Quantity         int     `json:"quantity"`
	OrderType        string  `json:"order_type"`
	UnitPrice        float64 `json:"unit_price"`
	ShippingAddress  Address `json:"shipping_address"`
}

type UpdateOrderRequest struct {
	Status          *string    `json:"status,omitempty"`
	PaymentStatus   *string    `json:"payment_status,omitempty"`
	PaymentID       *string    `json:"payment_id,omitempty"`
	TrackingNumber  *string    `json:"tracking_number,omitempty"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	DeliveryDate    *time.Time `json:"delivery_date,omitempty"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Order service with in-memory storage
type OrderService struct {
	orders []Order
	mu     sync.RWMutex
}

// New order service
func NewOrderService() *OrderService {
	return &OrderService{
		orders: []Order{
			{
				ID:               "order-1",
				UserID:           "user-1",
				SellerID:         "seller-1",
				ProductID:        "product-1",
				ProductName:      "Vintage Camera",
				ProductImage:     "https://picsum.photos/400/300?random=1",
				Quantity:         1,
				UnitPrice:        450.00,
				TotalAmount:      450.00,
				OrderType:        "auction_win",
				Status:           "processing",
				PaymentStatus:    "paid",
				PaymentID:        &[]string{"payment-123"}[0],
				ShippingAddress: Address{
					Street:  "123 Main St",
					City:    "New York",
					State:   "NY",
					ZipCode: "10001",
					Country: "USA",
					Phone:   "+1-555-0123",
				},
				TrackingNumber: &[]string{"TRACK123456"}[0],
				EstimatedDelivery: &[]time.Time{time.Now().Add(3 * 24 * time.Hour)}[0],
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			},
			{
				ID:               "order-2",
				UserID:           "user-2",
				SellerID:         "seller-2",
				ProductID:        "product-2",
				ProductName:      "Designer Handbag",
				ProductImage:     "https://picsum.photos/400/300?random=2",
				Quantity:         1,
				UnitPrice:        450.00,
				TotalAmount:      450.00,
				OrderType:        "buy_now",
				Status:           "shipped",
				PaymentStatus:    "paid",
				PaymentID:        &[]string{"payment-456"}[0],
				ShippingAddress: Address{
					Street:  "456 Oak Ave",
					City:    "Los Angeles",
					State:   "CA",
					ZipCode: "90210",
					Country: "USA",
					Phone:   "+1-555-0456",
				},
				TrackingNumber: &[]string{"TRACK789012"}[0],
				EstimatedDelivery: &[]time.Time{time.Now().Add(2 * 24 * time.Hour)}[0],
				CreatedAt: time.Now().Add(-24 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			{
				ID:               "order-3",
				UserID:           "user-1",
				SellerID:         "seller-1",
				ProductID:        "product-3",
				AuctionID:        &[]string{"auction-3"}[0],
				ProductName:      "Gaming Laptop",
				ProductImage:     "https://picsum.photos/400/300?random=3",
				Quantity:         1,
				UnitPrice:        2100.00,
				TotalAmount:      2100.00,
				OrderType:        "auction_win",
				Status:           "confirmed",
				PaymentStatus:    "pending",
				ShippingAddress: Address{
					Street:  "789 Pine Rd",
					City:    "Chicago",
					State:   "IL",
					ZipCode: "60601",
					Country: "USA",
					Phone:   "+1-555-0789",
				},
				CreatedAt: time.Now().Add(-30 * time.Minute),
				UpdatedAt: time.Now().Add(-15 * time.Minute),
			},
		},
	}
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// JSON response helper
func writeJSONResponse(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// Parse pagination parameters
func parsePagination(r *http.Request) (int, int) {
	page := 1
	perPage := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if p := r.URL.Query().Get("per_page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	return page, perPage
}

// Paginate results
func paginateOrders(items []Order, page, perPage int) ([]Order, map[string]interface{}) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Order{}, map[string]interface{}{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
		}
	}
	if end > total {
		end = total
	}

	return items[start:end], map[string]interface{}{
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	}
}

// Health check
func (s *OrderService) health(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pendingOrders := 0
	paidOrders := 0
	shippedOrders := 0

	for _, order := range s.orders {
		switch order.Status {
		case "pending", "confirmed":
			pendingOrders++
		case "processing", "shipped":
			shippedOrders++
		}
		if order.PaymentStatus == "paid" {
			paidOrders++
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Order service is healthy and working!",
		Data: map[string]interface{}{
			"service":        "order-service",
			"version":        "v2.0-working",
			"status":         "healthy",
			"timestamp":      time.Now(),
			"total_orders":   len(s.orders),
			"pending_orders": pendingOrders,
			"paid_orders":    paidOrders,
			"shipped_orders": shippedOrders,
		},
	})
}

// Get all orders
func (s *OrderService) getAllOrders(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply user filter if provided
	var filteredOrders []Order
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		for _, order := range s.orders {
			if order.UserID == userID {
				filteredOrders = append(filteredOrders, order)
			}
		}
	} else if sellerID := r.URL.Query().Get("seller_id"); sellerID != "" {
		for _, order := range s.orders {
			if order.SellerID == sellerID {
				filteredOrders = append(filteredOrders, order)
			}
		}
	} else {
		filteredOrders = s.orders
	}

	// Apply status filter if provided
	if status := strings.ToLower(r.URL.Query().Get("status")); status != "" {
		var statusFilteredOrders []Order
		for _, order := range filteredOrders {
			if strings.Contains(strings.ToLower(order.Status), status) {
				statusFilteredOrders = append(statusFilteredOrders, order)
			}
		}
		filteredOrders = statusFilteredOrders
	}

	// Apply payment status filter if provided
	if paymentStatus := strings.ToLower(r.URL.Query().Get("payment_status")); paymentStatus != "" {
		var paymentFilteredOrders []Order
		for _, order := range filteredOrders {
			if strings.Contains(strings.ToLower(order.PaymentStatus), paymentStatus) {
				paymentFilteredOrders = append(paymentFilteredOrders, order)
			}
		}
		filteredOrders = paymentFilteredOrders
	}

	// Apply order type filter if provided
	if orderType := strings.ToLower(r.URL.Query().Get("order_type")); orderType != "" {
		var typeFilteredOrders []Order
		for _, order := range filteredOrders {
			if strings.Contains(strings.ToLower(order.OrderType), orderType) {
				typeFilteredOrders = append(typeFilteredOrders, order)
			}
		}
		filteredOrders = typeFilteredOrders
	}

	// Sort orders by creation date (newest first)
	for i := 0; i < len(filteredOrders)-1; i++ {
		for j := i + 1; j < len(filteredOrders); j++ {
			if filteredOrders[i].CreatedAt.Before(filteredOrders[j].CreatedAt) {
				filteredOrders[i], filteredOrders[j] = filteredOrders[j], filteredOrders[i]
			}
		}
	}

	// Paginate results
	paginatedOrders, pagination := paginateOrders(filteredOrders, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Orders retrieved successfully",
		Data: map[string]interface{}{
			"orders":    paginatedOrders,
			"pagination": pagination,
		},
	})
}

// Get order by ID
func (s *OrderService) getOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	if orderID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Order ID is required",
			Error:   "No order ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, order := range s.orders {
		if order.ID == orderID {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Order retrieved successfully",
				Data: map[string]interface{}{
					"order": order,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Order not found",
		Error:   "Order with ID " + orderID + " does not exist",
	})
}

// Create order
func (s *OrderService) createOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid order data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.UserID == "" || req.ProductID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID and Product ID are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.Quantity <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Quantity must be greater than 0",
			Error:   "Invalid quantity",
		})
		return
	}

	if req.UnitPrice <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Unit price must be greater than 0",
			Error:   "Invalid unit price",
		})
		return
	}

	if req.ShippingAddress.Street == "" || req.ShippingAddress.City == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Shipping address is incomplete",
			Error:   "Missing shipping address fields",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate total amount
	totalAmount := float64(req.Quantity) * req.UnitPrice

	// Create new order
	newOrder := Order{
		ID:              fmt.Sprintf("order-%d", time.Now().UnixNano()),
		UserID:          req.UserID,
		SellerID:        "seller-1", // In real app, get from product
		ProductID:       req.ProductID,
		AuctionID:       req.AuctionID,
		ProductName:     "Sample Product", // In real app, get from product
		ProductImage:    "https://picsum.photos/400/300?random=order",
		Quantity:        req.Quantity,
		UnitPrice:       req.UnitPrice,
		TotalAmount:     totalAmount,
		OrderType:       req.OrderType,
		Status:          "pending",
		PaymentStatus:   "pending",
		ShippingAddress: req.ShippingAddress,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	s.orders = append(s.orders, newOrder)

	log.Printf("Order created: %s for user %s", newOrder.ID, newOrder.UserID)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Order created successfully",
		Data: map[string]interface{}{
			"order": newOrder,
		},
	})
}

// Update order
func (s *OrderService) updateOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	if orderID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Order ID is required",
			Error:   "No order ID provided",
		})
		return
	}

	var req UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid order data",
			Error:   err.Error(),
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and update order
	for i := range s.orders {
		if s.orders[i].ID == orderID {
			// Update fields if provided
			if req.Status != nil {
				s.orders[i].Status = *req.Status
			}
			if req.PaymentStatus != nil {
				s.orders[i].PaymentStatus = *req.PaymentStatus
			}
			if req.PaymentID != nil {
				s.orders[i].PaymentID = req.PaymentID
			}
			if req.TrackingNumber != nil {
				s.orders[i].TrackingNumber = req.TrackingNumber
			}
			if req.EstimatedDelivery != nil {
				s.orders[i].EstimatedDelivery = req.EstimatedDelivery
			}
			if req.DeliveryDate != nil {
				s.orders[i].DeliveryDate = req.DeliveryDate
			}

			s.orders[i].UpdatedAt = time.Now()

			log.Printf("Order updated: %s", orderID)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Order updated successfully",
				Data: map[string]interface{}{
					"order": s.orders[i],
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Order not found",
		Error:   "Order with ID " + orderID + " does not exist",
	})
}

// Cancel order
func (s *OrderService) cancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	if orderID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Order ID is required",
			Error:   "No order ID provided",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and cancel order
	for i := range s.orders {
		if s.orders[i].ID == orderID {
			if s.orders[i].Status == "shipped" || s.orders[i].Status == "delivered" {
				writeJSONResponse(w, http.StatusBadRequest, Response{
					Success: false,
					Message: "Cannot cancel shipped or delivered order",
					Error:   "Order cannot be cancelled in current status",
				})
				return
			}

			s.orders[i].Status = "cancelled"
			s.orders[i].UpdatedAt = time.Now()

			log.Printf("Order cancelled: %s", orderID)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Order cancelled successfully",
				Data: map[string]interface{}{
					"order": s.orders[i],
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Order not found",
		Error:   "Order with ID " + orderID + " does not exist",
	})
}

// Get user orders
func (s *OrderService) getUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID is required",
			Error:   "No user ID provided",
		})
		return
	}

	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter orders by user
	var userOrders []Order
	for _, order := range s.orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}

	// Sort orders by creation date (newest first)
	for i := 0; i < len(userOrders)-1; i++ {
		for j := i + 1; j < len(userOrders); j++ {
			if userOrders[i].CreatedAt.Before(userOrders[j].CreatedAt) {
				userOrders[i], userOrders[j] = userOrders[j], userOrders[i]
			}
		}
	}

	// Paginate results
	paginatedOrders, pagination := paginateOrders(userOrders, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "User orders retrieved successfully",
		Data: map[string]interface{}{
			"orders":    paginatedOrders,
			"pagination": pagination,
			"user_id":   userID,
		},
	})
}

// Get seller orders
func (s *OrderService) getSellerOrders(w http.ResponseWriter, r *http.Request) {
	sellerID := r.URL.Query().Get("seller_id")
	if sellerID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Seller ID is required",
			Error:   "No seller ID provided",
		})
		return
	}

	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter orders by seller
	var sellerOrders []Order
	for _, order := range s.orders {
		if order.SellerID == sellerID {
			sellerOrders = append(sellerOrders, order)
		}
	}

	// Sort orders by creation date (newest first)
	for i := 0; i < len(sellerOrders)-1; i++ {
		for j := i + 1; j < len(sellerOrders); j++ {
			if sellerOrders[i].CreatedAt.Before(sellerOrders[j].CreatedAt) {
				sellerOrders[i], sellerOrders[j] = sellerOrders[j], sellerOrders[i]
			}
		}
	}

	// Paginate results
	paginatedOrders, pagination := paginateOrders(sellerOrders, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Seller orders retrieved successfully",
		Data: map[string]interface{}{
			"orders":    paginatedOrders,
			"pagination": pagination,
			"seller_id": sellerID,
		},
	})
}

// Start order service
func main() {
	service := NewOrderService()

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/orders", corsMiddleware(service.getAllOrders))
	http.Handle("/api/v1/orders/", corsMiddleware(service.getOrder)) // For GET by ID and update
	http.Handle("/api/v1/orders/create", corsMiddleware(service.createOrder))
	http.Handle("/api/v1/orders/cancel", corsMiddleware(service.cancelOrder))
	http.Handle("/api/v1/orders/user", corsMiddleware(service.getUserOrders))
	http.Handle("/api/v1/orders/seller", corsMiddleware(service.getSellerOrders))

	port := ":8088"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	fmt.Printf("🚀 ORDER SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)
	fmt.Printf("📦 Orders list: http://localhost%s/api/v1/orders\n", port)
	fmt.Printf("📝 Order details: http://localhost%s/api/v1/orders/{id}\n", port)
	fmt.Printf("➕ Create order: http://localhost%s/api/v1/orders/create\n", port)
	fmt.Printf("🛍️  User orders: http://localhost%s/api/v1/orders/user?user_id={id}\n", port)
	fmt.Printf("👨‍💼 Seller orders: http://localhost%s/api/v1/orders/seller?seller_id={id}\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("📊 Total orders: %d\n", len(service.orders))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(port, nil))
}