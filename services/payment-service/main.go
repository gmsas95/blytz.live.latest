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

// Payment struct
type Payment struct {
	ID              string    `json:"id"`
	OrderID         string    `json:"order_id"`
	UserID          string    `json:"user_id"`
	SellerID        string    `json:"seller_id"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	PaymentMethod   string    `json:"payment_method"` // "stripe", "paypal", "fuu", "bank_transfer"
	PaymentType     string    `json:"payment_type"`    // "credit_card", "debit_card", "paypal", "online_banking"
	Provider        string    `json:"provider"`        // Payment provider name
	ProviderTxID    string    `json:"provider_tx_id"`  // Transaction ID from provider
	Status          string    `json:"status"`          // "pending", "processing", "completed", "failed", "refunded", "cancelled"
	FailureReason   *string   `json:"failure_reason,omitempty"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	RefundedAt      *time.Time `json:"refunded_at,omitempty"`
	RefundedAmount   float64   `json:"refunded_amount"`
	Fee             float64   `json:"fee"`           // Processing fee
	NetAmount       float64   `json:"net_amount"`    // Amount after fees
	Description     string    `json:"description"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PaymentMethod struct
type PaymentMethod struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Type        string `json:"type"`        // "stripe_card", "paypal_account", "fuu_wallet", "bank_account"
	Provider    string `json:"provider"`    // "stripe", "paypal", "fuu", "bank"
	LastFour    string `json:"last_four,omitempty"`
	Brand       string `json:"brand,omitempty"`        // Card brand (Visa, Mastercard, etc.)
	ExpiryMonth string `json:"expiry_month,omitempty"`
	ExpiryYear  string `json:"expiry_year,omitempty"`
	Email       string `json:"email,omitempty"`       // For PayPal/Fiuu
	IsDefault   bool   `json:"is_default"`
	Status      string `json:"status"` // "active", "inactive", "expired"
	CreatedAt   time.Time `json:"created_at"`
}

// Request structs
type CreatePaymentRequest struct {
	OrderID       string                 `json:"order_id"`
	UserID        string                 `json:"user_id"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	PaymentMethod string                 `json:"payment_method"`
	PaymentType   string                 `json:"payment_type"`
	Description   string                 `json:"description"`
	PaymentMethodID string                `json:"payment_method_id,omitempty"`
	SavePaymentMethod bool               `json:"save_payment_method,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type RefundPaymentRequest struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
}

type AddPaymentMethodRequest struct {
	UserID   string `json:"user_id"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
	Token    string `json:"token"` // Payment token from client
	IsDefault bool `json:"is_default"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Payment service with in-memory storage
type PaymentService struct {
	payments       []Payment
	paymentMethods []PaymentMethod
	mu             sync.RWMutex
}

// New payment service
func NewPaymentService() *PaymentService {
	return &PaymentService{
		payments: []Payment{
			{
				ID:              "payment-1",
				OrderID:         "order-1",
				UserID:          "user-1",
				SellerID:        "seller-1",
				Amount:          450.00,
				Currency:        "USD",
				PaymentMethod:   "stripe",
				PaymentType:     "credit_card",
				Provider:        "stripe",
				ProviderTxID:    "ch_1234567890abcdef",
				Status:          "completed",
				ProcessedAt:     &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
				RefundedAmount:   0.00,
				Fee:             13.50, // 3% Stripe fee
				NetAmount:       436.50,
				Description:     "Payment for order #order-1",
				Metadata: map[string]interface{}{
					"order_type": "auction_win",
					"product_name": "Vintage Camera",
				},
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			{
				ID:              "payment-2",
				OrderID:         "order-2",
				UserID:          "user-2",
				SellerID:        "seller-2",
				Amount:          450.00,
				Currency:        "USD",
				PaymentMethod:   "paypal",
				PaymentType:     "paypal",
				Provider:        "paypal",
				ProviderTxID:    "PAYID-9876543210fedcba",
				Status:          "completed",
				ProcessedAt:     &[]time.Time{time.Now().Add(-3 * time.Hour)}[0],
				RefundedAmount:   0.00,
				Fee:             18.00, // 4% PayPal fee
				NetAmount:       432.00,
				Description:     "Payment for order #order-2",
				Metadata: map[string]interface{}{
					"order_type": "buy_now",
					"product_name": "Designer Handbag",
				},
				CreatedAt: time.Now().Add(-4 * time.Hour),
				UpdatedAt: time.Now().Add(-3 * time.Hour),
			},
			{
				ID:              "payment-3",
				OrderID:         "order-3",
				UserID:          "user-1",
				SellerID:        "seller-1",
				Amount:          2100.00,
				Currency:        "USD",
				PaymentMethod:   "fuu",
				PaymentType:     "online_banking",
				Provider:        "fuu",
				ProviderTxID:    "FIUU-20241207-00001",
				Status:          "pending",
				RefundedAmount:   0.00,
				Fee:             42.00, // 2% Fiuu fee
				NetAmount:       2058.00,
				Description:     "Payment for order #order-3",
				Metadata: map[string]interface{}{
					"order_type": "auction_win",
					"product_name": "Gaming Laptop",
				},
				CreatedAt: time.Now().Add(-30 * time.Minute),
				UpdatedAt: time.Now().Add(-15 * time.Minute),
			},
		},
		paymentMethods: []PaymentMethod{
			{
				ID:          "pm-1",
				UserID:      "user-1",
				Type:        "stripe_card",
				Provider:    "stripe",
				LastFour:    "4242",
				Brand:       "Visa",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				IsDefault:   true,
				Status:      "active",
				CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			},
			{
				ID:       "pm-2",
				UserID:   "user-2",
				Type:     "paypal_account",
				Provider: "paypal",
				Email:    "user2@example.com",
				IsDefault: true,
				Status:   "active",
				CreatedAt: time.Now().Add(-15 * 24 * time.Hour),
			},
			{
				ID:       "pm-3",
				UserID:   "user-1",
				Type:     "fuu_wallet",
				Provider: "fuu",
				Email:    "user1@fuu.com",
				IsDefault: false,
				Status:   "active",
				CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			},
		},
	}
}

// Calculate processing fee
func calculateFee(amount float64, provider string) float64 {
	switch provider {
	case "stripe":
		return amount * 0.03 // 3% Stripe fee
	case "paypal":
		return amount * 0.04 // 4% PayPal fee
	case "fuu":
		return amount * 0.02 // 2% Fiuu fee
	case "bank_transfer":
		return 5.00 // Fixed $5 fee for bank transfers
	default:
		return amount * 0.025 // Default 2.5%
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
func paginatePayments(items []Payment, page, perPage int) ([]Payment, map[string]interface{}) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Payment{}, map[string]interface{}{
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
func (s *PaymentService) health(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pendingPayments := 0
	completedPayments := 0
	totalVolume := 0.0
	totalFees := 0.0

	for _, payment := range s.payments {
		switch payment.Status {
		case "pending":
			pendingPayments++
		case "completed":
			completedPayments++
		}
		totalVolume += payment.Amount
		totalFees += payment.Fee
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Payment service is healthy and working!",
		Data: map[string]interface{}{
			"service":           "payment-service",
			"version":           "v2.0-working",
			"status":            "healthy",
			"timestamp":         time.Now(),
			"total_payments":    len(s.payments),
			"pending_payments":  pendingPayments,
			"completed_payments": completedPayments,
			"total_volume":      totalVolume,
			"total_fees":        totalFees,
			"payment_methods":   len(s.paymentMethods),
		},
	})
}

// Get all payments
func (s *PaymentService) getAllPayments(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply user filter if provided
	var filteredPayments []Payment
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		for _, payment := range s.payments {
			if payment.UserID == userID {
				filteredPayments = append(filteredPayments, payment)
			}
		}
	} else if sellerID := r.URL.Query().Get("seller_id"); sellerID != "" {
		for _, payment := range s.payments {
			if payment.SellerID == sellerID {
				filteredPayments = append(filteredPayments, payment)
			}
		}
	} else if orderID := r.URL.Query().Get("order_id"); orderID != "" {
		for _, payment := range s.payments {
			if payment.OrderID == orderID {
				filteredPayments = append(filteredPayments, payment)
			}
		}
	} else {
		filteredPayments = s.payments
	}

	// Apply status filter if provided
	if status := strings.ToLower(r.URL.Query().Get("status")); status != "" {
		var statusFilteredPayments []Payment
		for _, payment := range filteredPayments {
			if strings.Contains(strings.ToLower(payment.Status), status) {
				statusFilteredPayments = append(statusFilteredPayments, payment)
			}
		}
		filteredPayments = statusFilteredPayments
	}

	// Apply payment method filter if provided
	if method := strings.ToLower(r.URL.Query().Get("payment_method")); method != "" {
		var methodFilteredPayments []Payment
		for _, payment := range filteredPayments {
			if strings.Contains(strings.ToLower(payment.PaymentMethod), method) {
				methodFilteredPayments = append(methodFilteredPayments, payment)
			}
		}
		filteredPayments = methodFilteredPayments
	}

	// Sort payments by creation date (newest first)
	for i := 0; i < len(filteredPayments)-1; i++ {
		for j := i + 1; j < len(filteredPayments); j++ {
			if filteredPayments[i].CreatedAt.Before(filteredPayments[j].CreatedAt) {
				filteredPayments[i], filteredPayments[j] = filteredPayments[j], filteredPayments[i]
			}
		}
	}

	// Paginate results
	paginatedPayments, pagination := paginatePayments(filteredPayments, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Payments retrieved successfully",
		Data: map[string]interface{}{
			"payments":   paginatedPayments,
			"pagination": pagination,
		},
	})
}

// Get payment by ID
func (s *PaymentService) getPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := strings.TrimPrefix(r.URL.Path, "/api/v1/payments/")
	if paymentID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Payment ID is required",
			Error:   "No payment ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, payment := range s.payments {
		if payment.ID == paymentID {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Payment retrieved successfully",
				Data: map[string]interface{}{
					"payment": payment,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Payment not found",
		Error:   "Payment with ID " + paymentID + " does not exist",
	})
}

// Create payment
func (s *PaymentService) createPayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid payment data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.OrderID == "" || req.UserID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Order ID and User ID are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.Amount <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Amount must be greater than 0",
			Error:   "Invalid amount",
		})
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Validate payment method
	validMethods := map[string]bool{
		"stripe":        true,
		"paypal":        true,
		"fuu":           true,
		"bank_transfer":  true,
	}
	if !validMethods[req.PaymentMethod] {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid payment method",
			Error:   "Payment method must be one of: stripe, paypal, fuu, bank_transfer",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate fees
	fee := calculateFee(req.Amount, req.PaymentMethod)
	netAmount := req.Amount - fee

	// Create new payment
	newPayment := Payment{
		ID:              fmt.Sprintf("payment-%d", time.Now().UnixNano()),
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		SellerID:        "seller-1", // In real app, get from order
		Amount:          req.Amount,
		Currency:        req.Currency,
		PaymentMethod:   req.PaymentMethod,
		PaymentType:     req.PaymentType,
		Provider:        req.PaymentMethod,
		ProviderTxID:    fmt.Sprintf("%s-%d", strings.ToUpper(req.PaymentMethod), time.Now().UnixNano()),
		Status:          "pending",
		RefundedAmount:   0.00,
		Fee:             fee,
		NetAmount:       netAmount,
		Description:     req.Description,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	s.payments = append(s.payments, newPayment)

	// Simulate payment processing
	go func(paymentID string) {
		time.Sleep(2 * time.Second) // Simulate processing time
		
		s.mu.Lock()
		defer s.mu.Unlock()

		for i := range s.payments {
			if s.payments[i].ID == paymentID {
				// 90% success rate
				if time.Now().UnixNano()%10 != 0 {
					s.payments[i].Status = "completed"
					s.payments[i].ProcessedAt = &[]time.Time{time.Now()}[0]
				} else {
					s.payments[i].Status = "failed"
					s.payments[i].FailureReason = &[]string{"Insufficient funds"}[0]
				}
				s.payments[i].UpdatedAt = time.Now()
				break
			}
		}
	}(newPayment.ID)

	log.Printf("Payment created: %s for order %s (%.2f %s)", newPayment.ID, newPayment.OrderID, newPayment.Amount, newPayment.Currency)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Payment created successfully",
		Data: map[string]interface{}{
			"payment": newPayment,
		},
	})
}

// Refund payment
func (s *PaymentService) refundPayment(w http.ResponseWriter, r *http.Request) {
	var req RefundPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid refund data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.PaymentID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Payment ID is required",
			Error:   "No payment ID provided",
		})
		return
	}

	if req.Amount <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Refund amount must be greater than 0",
			Error:   "Invalid refund amount",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find payment
	for i := range s.payments {
		if s.payments[i].ID == req.PaymentID {
			if s.payments[i].Status != "completed" {
				writeJSONResponse(w, http.StatusBadRequest, Response{
					Success: false,
					Message: "Payment cannot be refunded",
					Error:   "Only completed payments can be refunded",
				})
				return
			}

			if req.Amount > s.payments[i].Amount-s.payments[i].RefundedAmount {
				writeJSONResponse(w, http.StatusBadRequest, Response{
					Success: false,
					Message: "Refund amount exceeds available amount",
					Error:   "Refund amount too high",
				})
				return
			}

			// Process refund
			s.payments[i].RefundedAmount += req.Amount
			if s.payments[i].RefundedAmount >= s.payments[i].Amount {
				s.payments[i].Status = "refunded"
				s.payments[i].RefundedAt = &[]time.Time{time.Now()}[0]
			}
			s.payments[i].UpdatedAt = time.Now()

			log.Printf("Payment refunded: %s (%.2f %s)", req.PaymentID, req.Amount, s.payments[i].Currency)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Payment refunded successfully",
				Data: map[string]interface{}{
					"payment":        s.payments[i],
					"refund_amount":  req.Amount,
					"reason":         req.Reason,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Payment not found",
		Error:   "Payment with ID " + req.PaymentID + " does not exist",
	})
}

// Get user payment methods
func (s *PaymentService) getUserPaymentMethods(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID is required",
			Error:   "No user ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get payment methods for user
	var userPaymentMethods []PaymentMethod
	for _, method := range s.paymentMethods {
		if method.UserID == userID && method.Status == "active" {
			userPaymentMethods = append(userPaymentMethods, method)
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Payment methods retrieved successfully",
		Data: map[string]interface{}{
			"payment_methods": userPaymentMethods,
			"count":           len(userPaymentMethods),
			"user_id":         userID,
		},
	})
}

// Get payment statistics
func (s *PaymentService) getPaymentStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]float64{
		"total_payments":     0,
		"completed_payments": 0,
		"failed_payments":    0,
		"refunded_payments":  0,
		"total_volume":       0,
		"total_fees":         0,
		"total_refunded":     0,
	}

	for _, payment := range s.payments {
		stats["total_payments"]++
		stats["total_volume"] += payment.Amount
		stats["total_fees"] += payment.Fee
		stats["total_refunded"] += payment.RefundedAmount

		switch payment.Status {
		case "completed":
			stats["completed_payments"]++
		case "failed":
			stats["failed_payments"]++
		case "refunded":
			stats["refunded_payments"]++
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Payment statistics retrieved successfully",
		Data: map[string]interface{}{
			"stats": stats,
		},
	})
}

// Start payment service
func main() {
	service := NewPaymentService()

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/payments", corsMiddleware(service.getAllPayments))
	http.Handle("/api/v1/payments/", corsMiddleware(service.getPayment)) // For GET by ID
	http.Handle("/api/v1/payments/create", corsMiddleware(service.createPayment))
	http.Handle("/api/v1/payments/refund", corsMiddleware(service.refundPayment))
	http.Handle("/api/v1/payments/methods", corsMiddleware(service.getUserPaymentMethods))
	http.Handle("/api/v1/payments/stats", corsMiddleware(service.getPaymentStats))

	port := ":8089"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	fmt.Printf("🚀 PAYMENT SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)
	fmt.Printf("💳 Payments list: http://localhost%s/api/v1/payments\n", port)
	fmt.Printf("📝 Payment details: http://localhost%s/api/v1/payments/{id}\n", port)
	fmt.Printf("➕ Create payment: http://localhost%s/api/v1/payments/create\n", port)
	fmt.Printf("💰 Refund payment: http://localhost%s/api/v1/payments/refund\n", port)
	fmt.Printf("🏦 Payment methods: http://localhost%s/api/v1/payments/methods?user_id={id}\n", port)
	fmt.Printf("📈 Payment stats: http://localhost%s/api/v1/payments/stats\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("💳 Total payments: %d\n", len(service.payments))
	fmt.Printf("🏦 Payment methods: %d\n", len(service.paymentMethods))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(port, nil))
}