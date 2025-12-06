package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz-mvp/services/payment-service/internal/models"
)

type PaymentService struct {
	db                *gorm.DB
	logger           *zap.Logger
	fiuuAPI          string
	fiuuMerchantID   string
	fiuuClientID     string
	fiuuClientSecret string
	fiuuSecretKey    string
	isSandbox        bool
}

func NewPaymentService(db *gorm.DB, logger *zap.Logger, fiuuAPI, fiuuMerchantID, fiuuClientID, fiuuClientSecret, fiuuSecretKey string, isSandbox bool) *PaymentService {
	return &PaymentService{
		db:                db,
		logger:           logger,
		fiuuAPI:          fiuuAPI,
		fiuuMerchantID:   fiuuMerchantID,
		fiuuClientID:     fiuuClientID,
		fiuuClientSecret: fiuuClientSecret,
		fiuuSecretKey:    fiuuSecretKey,
		isSandbox:        isSandbox,
	}
}

// === PAYMENT MANAGEMENT ===

// CreatePayment creates new payment with Fiuu integration
func (s *PaymentService) CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	s.logger.Info("💳 Payment Service: Creating payment", 
		zap.String("order_id", req.OrderID),
		zap.Float64("amount", req.Amount))

	// Validate request
	if err := s.validateCreatePaymentRequest(req); err != nil {
		return nil, err
	}

	// Check if payment already exists for this order
	var existingPayment models.Payment
	err := s.db.Where("order_id = ? AND status IN ?", 
		req.OrderID, []string{models.PaymentStatusPending, models.PaymentStatusProcessing}).First(&existingPayment).Error
	if err == nil {
		// Return existing payment
		response := s.buildPaymentResponse(&existingPayment)
		return response, nil
	}

	// Create local payment record
	payment := &models.Payment{
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Description:   req.Description,
		Status:        models.PaymentStatusPending,
		PaymentType:   req.PaymentType,
		PaymentMethod: req.PaymentMethod,
		CustomerName:  req.Customer.Name,
		CustomerEmail: req.Customer.Email,
		CustomerPhone: req.Customer.Phone,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		Metadata:      req.Metadata,
	}

	// Set expiry
	if req.ExpiryMinutes > 0 {
		payment.ExpiresAt = time.Now().Add(time.Duration(req.ExpiryMinutes) * time.Minute)
	}

	// Create payment with Fiuu
	fiuuReq := &models.FiuuCreatePaymentRequest{
		MerchantID:   s.fiuuMerchantID,
		OrderID:      req.OrderID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Description:   req.Description,
		Customer:      req.Customer,
		PaymentType:   req.PaymentType,
		ReturnURL:     req.ReturnURL,
		CallbackURL:   req.CallbackURL,
		SuccessURL:    req.SuccessURL,
		FailURL:       req.FailURL,
		ExpiryMinutes: req.ExpiryMinutes,
		Metadata:      req.Metadata,
	}

	fiuuResp, err := s.createFiuuPayment(ctx, fiuuReq)
	if err != nil {
		s.logger.Error("💳 Payment Service: Fiuu payment creation failed", zap.Error(err))
		return nil, fmt.Errorf("failed to create Fiuu payment: %w", err)
	}

	// Update payment with Fiuu response
	payment.PaymentID = fiuuResp.PaymentID
	payment.TransactionID = fiuuResp.TransactionID
	payment.PaymentURL = fiuuResp.PaymentURL
	payment.QRCode = fiuuResp.QRCode

	// Parse expiry time
	if fiuuResp.ExpiresAt != "" {
		if expiryTime, err := time.Parse("2006-01-02T15:04:05Z", fiuuResp.ExpiresAt); err == nil {
			payment.ExpiresAt = expiryTime
		}
	}

	// Save payment in database
	err = s.db.Create(payment).Error
	if err != nil {
		s.logger.Error("💳 Payment Service: Failed to save payment", zap.Error(err))
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	s.logger.Info("💳 Payment Service: Payment created successfully", 
		zap.String("payment_id", payment.ID),
		zap.String("fiuu_payment_id", payment.PaymentID))

	return s.buildPaymentResponse(payment), nil
}

// GetPayment retrieves payment by ID
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	s.logger.Info("💳 Payment Service: Getting payment", zap.String("payment_id", paymentID))

	var payment models.Payment
	err := s.db.Where("id = ?", paymentID).First(&payment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, err
	}

	s.logger.Info("💳 Payment Service: Payment retrieved successfully", zap.String("payment_id", paymentID))
	return &payment, nil
}

// UpdatePaymentStatus updates payment status from Fiuu
func (s *PaymentService) UpdatePaymentStatus(ctx context.Context, paymentID string) (*models.Payment, error) {
	s.logger.Info("💳 Payment Service: Updating payment status", zap.String("payment_id", paymentID))

	// Get payment from database
	var payment models.Payment
	err := s.db.Where("id = ?", paymentID).First(&payment).Error
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Check payment status from Fiuu
	fiuuResp, err := s.getFiuuPaymentStatus(ctx, payment.PaymentID)
	if err != nil {
		s.logger.Error("💳 Payment Service: Failed to get Fiuu payment status", zap.Error(err))
		return nil, fmt.Errorf("failed to get Fiuu payment status: %w", err)
	}

	// Update payment status based on Fiuu response
	updated := false
	now := time.Now()

	if fiuuResp.Status == "success" && payment.Status != models.PaymentStatusSuccess {
		payment.Status = models.PaymentStatusSuccess
		payment.PaidAt = &now
		payment.PaymentMethod = fiuuResp.PaymentMethod
		updated = true

		// Update order status (would call Order Service)
		s.logger.Info("💳 Payment Service: Payment successful", 
			zap.String("payment_id", paymentID),
			zap.String("order_id", payment.OrderID))

	} else if fiuuResp.Status == "failed" && payment.Status != models.PaymentStatusFailed {
		payment.Status = models.PaymentStatusFailed
		payment.FailedAt = &now
		payment.FailureReason = fiuuResp.FailureReason
		updated = true

		// Update order status (would call Order Service)
		s.logger.Warn("💳 Payment Service: Payment failed", 
			zap.String("payment_id", paymentID),
			zap.String("failure_reason", fiuuResp.FailureReason))

	} else if fiuuResp.Status == "cancelled" && payment.Status != models.PaymentStatusCancelled {
		payment.Status = models.PaymentStatusCancelled
		payment.CancelledAt = &now
		updated = true
	}

	if updated {
		err = s.db.Save(&payment).Error
		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}
	}

	s.logger.Info("💳 Payment Service: Payment status updated", 
		zap.String("payment_id", paymentID),
		zap.String("status", payment.Status))

	return &payment, nil
}

// ProcessWebhook processes webhook from Fiuu
func (s *PaymentService) ProcessWebhook(ctx context.Context, event string, payload []byte, signature string) error {
	s.logger.Info("💳 Payment Service: Processing webhook", 
		zap.String("event", event),
		zap.String("signature", signature))

	// Verify webhook signature
	if !s.verifyWebhookSignature(payload, signature) {
		s.logger.Error("💳 Payment Service: Invalid webhook signature")
		return fmt.Errorf("invalid webhook signature")
	}

	// Parse webhook based on event type
	switch event {
	case "payment.success":
		var webhook models.FiuuWebhookPaymentSuccess
		if err := json.Unmarshal(payload, &webhook); err != nil {
			return fmt.Errorf("failed to parse success webhook: %w", err)
		}
		return s.processSuccessWebhook(ctx, &webhook)

	case "payment.failed":
		var webhook models.FiuuWebhookPaymentFailed
		if err := json.Unmarshal(payload, &webhook); err != nil {
			return fmt.Errorf("failed to parse failed webhook: %w", err)
		}
		return s.processFailedWebhook(ctx, &webhook)

	case "payment.refunded":
		var webhook models.FiuuWebhookPaymentRefunded
		if err := json.Unmarshal(payload, &webhook); err != nil {
			return fmt.Errorf("failed to parse refunded webhook: %w", err)
		}
		return s.processRefundedWebhook(ctx, &webhook)

	default:
		s.logger.Warn("💳 Payment Service: Unknown webhook event", zap.String("event", event))
		return fmt.Errorf("unknown webhook event: %s", event)
	}
}

// RefundPayment creates refund with Fiuu
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string, req *models.RefundPaymentRequest) (*models.RefundResponse, error) {
	s.logger.Info("💳 Payment Service: Creating refund", 
		zap.String("payment_id", paymentID),
		zap.Float64("refund_amount", req.RefundAmount))

	// Get payment from database
	var payment models.Payment
	err := s.db.Where("id = ?", paymentID).First(&payment).Error
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Check if payment can be refunded
	if payment.Status != models.PaymentStatusSuccess {
		return nil, fmt.Errorf("payment cannot be refunded - status: %s", payment.Status)
	}

	// Check refund amount
	if req.RefundAmount > payment.Amount-payment.RefundAmount {
		return nil, fmt.Errorf("refund amount cannot exceed remaining amount")
	}

	// Create refund with Fiuu
	fiuuReq := &models.FiuuRefundRequest{
		PaymentID:    payment.PaymentID,
		RefundAmount: req.RefundAmount,
		Reason:       req.Reason,
		Reference:    req.Reference,
	}

	fiuuResp, err := s.createFiuuRefund(ctx, fiuuReq)
	if err != nil {
		s.logger.Error("💳 Payment Service: Fiuu refund creation failed", zap.Error(err))
		return nil, fmt.Errorf("failed to create Fiuu refund: %w", err)
	}

	// Create local refund record
	refund := &models.PaymentRefund{
		PaymentID:    payment.ID,
		RefundID:     fiuuResp.RefundID,
		RefundAmount: req.RefundAmount,
		Reason:       req.Reason,
		Status:       models.PaymentStatusProcessing,
		Reference:    req.Reference,
	}

	err = s.db.Create(refund).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create refund record: %w", err)
	}

	// Update payment refund amount
	payment.RefundAmount += req.RefundAmount
	if payment.RefundAmount >= payment.Amount {
		payment.Status = models.PaymentStatusRefunded
		payment.RefundedAt = &[]time.Time{time.Now()}[0]
		payment.RefundReason = req.Reason
	}

	err = s.db.Save(&payment).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	response := &models.RefundResponse{
		Refund:              *refund,
		EstimatedCompletion: fiuuResp.EstimatedCompletion,
		NetRefundAmount:     fiuuResp.NetRefundAmount,
		ProcessingFee:       fiuuResp.ProcessingFee,
	}

	s.logger.Info("💳 Payment Service: Refund created successfully", 
		zap.String("refund_id", refund.ID),
		zap.String("fiuu_refund_id", fiuuResp.RefundID))

	return response, nil
}

// GetPaymentMethods gets available payment methods from Fiuu
func (s *PaymentService) GetPaymentMethods(ctx context.Context) (*models.PaymentMethodsResponse, error) {
	s.logger.Info("💳 Payment Service: Getting payment methods")

	fiuuResp, err := s.getFiuuPaymentMethods(ctx)
	if err != nil {
		s.logger.Error("💳 Payment Service: Failed to get Fiuu payment methods", zap.Error(err))
		return nil, fmt.Errorf("failed to get Fiuu payment methods: %w", err)
	}

	response := &models.PaymentMethodsResponse{
		Methods:  fiuuResp.Methods,
		Total:    len(fiuuResp.Methods),
		Currency: fiuuResp.Currency,
	}

	s.logger.Info("💳 Payment Service: Payment methods retrieved successfully", 
		zap.Int("total", response.Total))

	return response, nil
}

// SearchPayments searches payments with filters
func (s *PaymentService) SearchPayments(ctx context.Context, req *models.SearchPaymentsRequest) (*models.PaymentsResponse, error) {
	s.logger.Info("💳 Payment Service: Searching payments", zap.String("query", req.Query))

	// Build query
	query := s.db.Model(&models.Payment{})

	// Apply filters
	if req.Query != "" {
		query = query.Where("order_id ILIKE ? OR description ILIKE ?", 
			"%"+req.Query+"%", "%"+req.Query+"%")
	}
	if req.OrderID != "" {
		query = query.Where("order_id = ?", req.OrderID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.PaymentType != "" {
		query = query.Where("payment_type = ?", req.PaymentType)
	}
	if req.PaymentMethod != "" {
		query = query.Where("payment_method = ?", req.PaymentMethod)
	}
	if req.CustomerEmail != "" {
		query = query.Where("customer_email = ?", req.CustomerEmail)
	}
	if req.CustomerPhone != "" {
		query = query.Where("customer_phone = ?", req.CustomerPhone)
	}
	if req.MinAmount > 0 {
		query = query.Where("amount >= ?", req.MinAmount)
	}
	if req.MaxAmount > 0 {
		query = query.Where("amount <= ?", req.MaxAmount)
	}
	if req.StartDate != "" {
		// Parse start date (simplified)
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		// Parse end date (simplified)
		query = query.Where("created_at <= ?", req.EndDate)
	}

	// Apply sorting
	sortBy := "created_at"
	if req.SortBy == "amount" {
		sortBy = "amount"
	} else if req.SortBy == "status" {
		sortBy = "status"
	}

	sortOrder := "DESC"
	if req.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(sortBy + " " + sortOrder)

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var payments []models.Payment
	err := query.Limit(limit).Offset(offset).Find(&payments).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrevious := page > 1

	response := &models.PaymentsResponse{
		Payments:    payments,
		Total:       total,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}

	s.logger.Info("💳 Payment Service: Payments searched successfully", 
		zap.Int64("total", total),
		zap.Int("count", len(payments)))

	return response, nil
}

// === FIUU API INTEGRATION ===

// createFiuuPayment creates payment with Fiuu API
func (s *PaymentService) createFiuuPayment(ctx context.Context, req *models.FiuuCreatePaymentRequest) (*models.FiuuCreatePaymentResponse, error) {
	s.logger.Info("💳 Payment Service: Creating Fiuu payment", 
		zap.String("order_id", req.OrderID))

	url := fmt.Sprintf("%s/v1/payment/create", s.fiuuAPI)
	
	// Add merchant ID to request
	req.MerchantID = s.fiuuMerchantID

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.getAccessToken())
	httpReq.Header.Set("X-Merchant-ID", s.fiuuMerchantID)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fiuu API error: %d - %s", resp.StatusCode, string(body))
	}

	var response models.FiuuCreatePaymentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.logger.Info("💳 Payment Service: Fiuu payment created successfully", 
		zap.String("fiuu_payment_id", response.PaymentID))

	return &response, nil
}

// getFiuuPaymentStatus gets payment status from Fiuu API
func (s *PaymentService) getFiuuPaymentStatus(ctx context.Context, paymentID string) (*models.FiuuPaymentStatusResponse, error) {
	s.logger.Info("💳 Payment Service: Getting Fiuu payment status", 
		zap.String("payment_id", paymentID))

	url := fmt.Sprintf("%s/v1/payment/status/%s", s.fiuuAPI, paymentID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+s.getAccessToken())
	httpReq.Header.Set("X-Merchant-ID", s.fiuuMerchantID)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fiuu API error: %d - %s", resp.StatusCode, string(body))
	}

	var response models.FiuuPaymentStatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// createFiuuRefund creates refund with Fiuu API
func (s *PaymentService) createFiuuRefund(ctx context.Context, req *models.FiuuRefundRequest) (*models.FiuuRefundResponse, error) {
	s.logger.Info("💳 Payment Service: Creating Fiuu refund", 
		zap.String("payment_id", req.PaymentID))

	url := fmt.Sprintf("%s/v1/payment/refund", s.fiuuAPI)

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.getAccessToken())
	httpReq.Header.Set("X-Merchant-ID", s.fiuuMerchantID)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fiuu API error: %d - %s", resp.StatusCode, string(body))
	}

	var response models.FiuuRefundResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.logger.Info("💳 Payment Service: Fiuu refund created successfully", 
		zap.String("fiuu_refund_id", response.RefundID))

	return &response, nil
}

// getFiuuPaymentMethods gets payment methods from Fiuu API
func (s *PaymentService) getFiuuPaymentMethods(ctx context.Context) (*models.FiuuPaymentMethodsResponse, error) {
	s.logger.Info("💳 Payment Service: Getting Fiuu payment methods")

	url := fmt.Sprintf("%s/v1/payment/methods", s.fiuuAPI)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+s.getAccessToken())
	httpReq.Header.Set("X-Merchant-ID", s.fiuuMerchantID)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fiuu API error: %d - %s", resp.StatusCode, string(body))
	}

	var response models.FiuuPaymentMethodsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// === HELPER METHODS ===

// buildPaymentResponse builds payment response
func (s *PaymentService) buildPaymentResponse(payment *models.Payment) *models.PaymentResponse {
	response := &models.PaymentResponse{
		Payment:        *payment,
		PaymentURL:     payment.PaymentURL,
		QRCode:         payment.QRCode,
		TimeRemaining:  payment.GetTimeRemaining(),
		IsExpired:      payment.IsExpired(),
		CanRetry:       payment.CanRetry(),
		AvailableMethods: []models.PaymentMethod{}, // Would populate from API
	}

	return response
}

// verifyWebhookSignature verifies webhook signature
func (s *PaymentService) verifyWebhookSignature(payload []byte, signature string) bool {
	h := hmac.New(sha256.New, []byte(s.fiuuSecretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// getAccessToken gets access token (simplified - would use OAuth2 flow)
func (s *PaymentService) getAccessToken() string {
	// In production, this would implement OAuth2 client credentials flow
	// For now, return client ID as token
	return s.fiuuClientID
}

// processSuccessWebhook processes payment success webhook
func (s *PaymentService) processSuccessWebhook(ctx context.Context, webhook *models.FiuuWebhookPaymentSuccess) error {
	s.logger.Info("💳 Payment Service: Processing success webhook", 
		zap.String("payment_id", webhook.PaymentID))

	// Log webhook
	webhookLog := &models.WebhookLog{
		PaymentID:  webhook.PaymentID,
		Event:      webhook.Event,
		Payload:     string(marshalWebhook(webhook)),
		Signature:  webhook.Signature,
		IsVerified: true,
		Processed:  false,
	}
	s.db.Create(webhookLog)

	// Find local payment
	var payment models.Payment
	err := s.db.Where("payment_id = ?", webhook.PaymentID).First(&payment).Error
	if err != nil {
		s.logger.Error("💳 Payment Service: Payment not found for webhook", 
			zap.String("fiuu_payment_id", webhook.PaymentID))
		return nil // Don't return error to avoid webhook retries
	}

	// Update payment status
	if payment.Status != models.PaymentStatusSuccess {
		payment.Status = models.PaymentStatusSuccess
		payment.TransactionID = webhook.TransactionID
		payment.PaymentMethod = webhook.PaymentMethod
		
		now := time.Now()
		payment.PaidAt = &now

		err = s.db.Save(&payment).Error
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		// Update order status (would call Order Service)
		s.logger.Info("💳 Payment Service: Payment updated from webhook", 
			zap.String("payment_id", payment.ID),
			zap.String("order_id", payment.OrderID))
	}

	// Mark webhook as processed
	webhookLog.Processed = true
	s.db.Save(webhookLog)

	return nil
}

// processFailedWebhook processes payment failed webhook
func (s *PaymentService) processFailedWebhook(ctx context.Context, webhook *models.FiuuWebhookPaymentFailed) error {
	s.logger.Info("💳 Payment Service: Processing failed webhook", 
		zap.String("payment_id", webhook.PaymentID))

	// Log webhook
	webhookLog := &models.WebhookLog{
		PaymentID:  webhook.PaymentID,
		Event:      webhook.Event,
		Payload:     string(marshalWebhook(webhook)),
		Signature:  webhook.Signature,
		IsVerified: true,
		Processed:  false,
	}
	s.db.Create(webhookLog)

	// Find local payment
	var payment models.Payment
	err := s.db.Where("payment_id = ?", webhook.PaymentID).First(&payment).Error
	if err != nil {
		s.logger.Error("💳 Payment Service: Payment not found for webhook", 
			zap.String("fiuu_payment_id", webhook.PaymentID))
		return nil
	}

	// Update payment status
	if payment.Status != models.PaymentStatusFailed {
		payment.Status = models.PaymentStatusFailed
		payment.FailureReason = webhook.FailureReason
		
		now := time.Now()
		payment.FailedAt = &now

		err = s.db.Save(&payment).Error
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		// Update order status (would call Order Service)
		s.logger.Info("💳 Payment Service: Payment updated from webhook", 
			zap.String("payment_id", payment.ID),
			zap.String("order_id", payment.OrderID))
	}

	// Mark webhook as processed
	webhookLog.Processed = true
	s.db.Save(webhookLog)

	return nil
}

// processRefundedWebhook processes payment refunded webhook
func (s *PaymentService) processRefundedWebhook(ctx context.Context, webhook *models.FiuuWebhookPaymentRefunded) error {
	s.logger.Info("💳 Payment Service: Processing refunded webhook", 
		zap.String("payment_id", webhook.PaymentID),
		zap.String("refund_id", webhook.RefundID))

	// Log webhook
	webhookLog := &models.WebhookLog{
		PaymentID:  webhook.PaymentID,
		Event:      webhook.Event,
		Payload:     string(marshalWebhook(webhook)),
		Signature:  webhook.Signature,
		IsVerified: true,
		Processed:  false,
	}
	s.db.Create(webhookLog)

	// Find local payment
	var payment models.Payment
	err := s.db.Where("payment_id = ?", webhook.PaymentID).First(&payment).Error
	if err != nil {
		s.logger.Error("💳 Payment Service: Payment not found for webhook", 
			zap.String("fiuu_payment_id", webhook.PaymentID))
		return nil
	}

	// Find local refund
	var refund models.PaymentRefund
	err = s.db.Where("refund_id = ?", webhook.RefundID).First(&refund).Error
	if err == nil {
		// Update refund status
		refund.Status = models.PaymentStatusSuccess
		now := time.Now()
		refund.ProcessedAt = &now
		s.db.Save(&refund)

		// Update payment
		payment.RefundAmount += webhook.RefundAmount
		if payment.RefundAmount >= payment.Amount {
			payment.Status = models.PaymentStatusRefunded
			payment.RefundedAt = &now
		}

		err = s.db.Save(&payment).Error
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}

	// Mark webhook as processed
	webhookLog.Processed = true
	s.db.Save(webhookLog)

	return nil
}

// marshalWebhook helper to marshal webhook to JSON
func marshalWebhook(webhook interface{}) []byte {
	data, _ := json.Marshal(webhook)
	return data
}

// validateCreatePaymentRequest validates payment creation request
func (s *PaymentService) validateCreatePaymentRequest(req *models.CreatePaymentRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if req.Currency != "MYR" {
		return fmt.Errorf("currency must be MYR")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.PaymentType == "" {
		return fmt.Errorf("payment type is required")
	}
	if req.PaymentMethod == "" {
		return fmt.Errorf("payment method is required")
	}
	if req.Customer.Email == "" {
		return fmt.Errorf("customer email is required")
	}
	if req.Customer.Phone == "" {
		return fmt.Errorf("customer phone is required")
	}
	if req.ReturnURL == "" {
		return fmt.Errorf("return URL is required")
	}
	if req.CallbackURL == "" {
		return fmt.Errorf("callback URL is required")
	}
	return nil
}