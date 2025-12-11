package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/dtos"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/services"
	"github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	stripe "github.com/stripe/stripe-go/v84"
	"go.uber.org/zap"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentService *services.PaymentService
	logger         *zap.Logger
}

// NewPaymentHandler creates a new payment handler instance
func NewPaymentHandler(paymentService *services.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

// CreatePaymentIntent creates a new payment intent
func (h *PaymentHandler) CreatePaymentIntent(c *gin.Context) {
	var req dtos.CreatePaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for create payment intent",
			zap.Error(err),
		)
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := dtos.ValidateCreatePaymentIntentRequest(&req); len(validationErrors) > 0 {
		utils.SendValidationErrorResponse(c, validationErrors)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("User ID not found in context")
		utils.SendErrorResponse(c, errors.NewAuthenticationError("USER_NOT_FOUND", "User authentication required"))
		return
	}

	// Override user ID from request with authenticated user ID
	req.UserID = userID.(string)

	// Create service request
	serviceReq := &services.CreatePaymentIntentRequest{
		Amount:                req.Amount,
		Currency:              req.Currency,
		CustomerID:            req.CustomerID,
		UserID:                req.UserID,
		AuctionID:             req.AuctionID,
		ConnectedAccountID:    req.ConnectedAccountID,
		PaymentMethodTypes:    req.PaymentMethodTypes,
		Metadata:              req.Metadata,
	}

	// Create payment intent
	paymentIntent, err := h.paymentService.CreatePaymentIntent(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to create payment intent",
			zap.Error(err),
			zap.String("user_id", req.UserID),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_INTENT_CREATION_FAILED", "Failed to create payment intent", err.Error()))
		return
	}

	// Get the Stripe payment intent to include client secret
	stripePI, err := h.paymentService.GetPaymentStatus(c.Request.Context(), paymentIntent.StripePaymentIntentID)
	if err != nil {
		h.logger.Error("Failed to retrieve payment intent for client secret",
			zap.Error(err),
			zap.String("payment_intent_id", paymentIntent.StripePaymentIntentID),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_INTENT_RETRIEVAL_FAILED", "Failed to retrieve payment intent", err.Error()))
		return
	}

	// Prepare next action if present
	var nextAction *dtos.NextAction
	if stripePI.Status == "requires_action" && stripePI.Metadata != "" {
		// In a real implementation, you would parse the Stripe payment intent's NextAction
		// For now, we'll create a placeholder
		nextAction = &dtos.NextAction{
			Type: "use_stripe_sdk",
		}
	}

	// Convert to response DTO
	response := dtos.ToPaymentIntentResponse(paymentIntent, "", nextAction)

	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// ConfirmPayment confirms a payment
func (h *PaymentHandler) ConfirmPayment(c *gin.Context) {
	var req dtos.ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for confirm payment",
			zap.Error(err),
		)
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := dtos.ValidateConfirmPaymentRequest(&req); len(validationErrors) > 0 {
		utils.SendValidationErrorResponse(c, validationErrors)
		return
	}

	// Create service request
	serviceReq := &services.ConfirmPaymentRequest{
		PaymentIntentID: req.PaymentIntentID,
		PaymentMethodID: req.PaymentMethodID,
	}

	// Confirm payment
	paymentIntent, err := h.paymentService.ConfirmPayment(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to confirm payment",
			zap.Error(err),
			zap.String("payment_intent_id", req.PaymentIntentID),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_CONFIRMATION_FAILED", "Failed to confirm payment", err.Error()))
		return
	}

	// Prepare next action if present
	var nextAction *dtos.NextAction
	if paymentIntent.Status == "requires_action" {
		// In a real implementation, you would parse the Stripe payment intent's NextAction
		nextAction = &dtos.NextAction{
			Type: "use_stripe_sdk",
		}
	}

	// Convert to response DTO
	response := dtos.ToPaymentStatusResponse(paymentIntent, "", nextAction)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetPaymentStatus retrieves the status of a payment
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	paymentIntentID := c.Param("id")
	if paymentIntentID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("MISSING_PAYMENT_INTENT_ID", "Payment Intent ID is required"))
		return
	}

	// Get payment status
	paymentIntent, err := h.paymentService.GetPaymentStatus(c.Request.Context(), paymentIntentID)
	if err != nil {
		h.logger.Error("Failed to get payment status",
			zap.Error(err),
			zap.String("payment_intent_id", paymentIntentID),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_STATUS_RETRIEVAL_FAILED", "Failed to get payment status", err.Error()))
		return
	}

	// Convert to response DTO
	response := dtos.ToPaymentStatusResponse(paymentIntent, "", nil)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// RefundPayment processes a refund for a payment
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentIntentID := c.Param("id")
	if paymentIntentID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("MISSING_PAYMENT_INTENT_ID", "Payment Intent ID is required"))
		return
	}

	var req dtos.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for refund payment",
			zap.Error(err),
		)
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Set payment intent ID from URL parameter
	req.PaymentIntentID = paymentIntentID

	// Validate request
	if validationErrors := dtos.ValidateRefundRequest(&req); len(validationErrors) > 0 {
		utils.SendValidationErrorResponse(c, validationErrors)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("User ID not found in context")
		utils.SendErrorResponse(c, errors.NewAuthenticationError("USER_NOT_FOUND", "User authentication required"))
		return
	}

	// Create refund request for service
	refundReq := &services.RefundPaymentRequest{
		PaymentIntentID: req.PaymentIntentID,
		Amount:          req.Amount,
		Reason:          req.Reason,
		RefundedBy:      userID.(string),
	}

	// Process refund
	stripeRefund, err := h.paymentService.RefundPayment(c.Request.Context(), refundReq)
	if err != nil {
		h.logger.Error("Failed to process refund",
			zap.Error(err),
			zap.String("payment_intent_id", req.PaymentIntentID),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("REFUND_PROCESSING_FAILED", "Failed to process refund", err.Error()))
		return
	}

	// Convert to response DTO
	response := &dtos.RefundResponse{
		ID:            stripeRefund.ID,
		Amount:        stripeRefund.Amount,
		Currency:      string(stripeRefund.Currency),
		PaymentIntent: stripeRefund.PaymentIntent.ID,
		Status:        string(stripeRefund.Status),
		Reason:        string(stripeRefund.Reason),
		Metadata:      stripeRefund.Metadata,
		CreatedAt:     time.Unix(stripeRefund.Created, 0),
	}

	// Use stripe import to avoid unused import warning
	_ = stripe.RefundReason("")

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetPaymentHistory retrieves payment history for a user
func (h *PaymentHandler) GetPaymentHistory(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("User ID not found in context")
		utils.SendErrorResponse(c, errors.NewAuthenticationError("USER_NOT_FOUND", "User authentication required"))
		return
	}

	// Get pagination parameters
	page, perPage := utils.GetPaginationParams(c)

	// Get payment history
	payments, total, err := h.paymentService.GetPaymentHistory(c.Request.Context(), userID.(string), page, perPage)
	if err != nil {
		h.logger.Error("Failed to get payment history",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_HISTORY_RETRIEVAL_FAILED", "Failed to get payment history", err.Error()))
		return
	}

	// Convert to response DTO
	response := dtos.ToPaymentHistoryResponse(payments, page, perPage, total)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// AdminGetPaymentHistory retrieves payment history for all users (admin only)
func (h *PaymentHandler) AdminGetPaymentHistory(c *gin.Context) {
	// Get user ID from query parameters (optional)
	userID := c.Query("user_id")

	// Get pagination parameters
	page, perPage := utils.GetPaginationParams(c)

	var payments []*models.PaymentIntent
	var total int64
	var err error

	if userID != "" {
		// Get payment history for specific user
		payments, total, err = h.paymentService.GetPaymentHistory(c.Request.Context(), userID, page, perPage)
	} else {
		// Get all payment history (admin only)
		// This would require an additional method in the service
		// For now, we'll return an error
		utils.SendErrorResponse(c, errors.NewAuthorizationError("INSUFFICIENT_PERMISSIONS", "Admin access required to view all payment history"))
		return
	}

	if err != nil {
		h.logger.Error("Failed to get payment history",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_HISTORY_RETRIEVAL_FAILED", "Failed to get payment history", err.Error()))
		return
	}

	// Convert to response DTO
	response := dtos.ToPaymentHistoryResponse(payments, page, perPage, total)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// WebhookHandler handles Stripe webhooks
func (h *PaymentHandler) WebhookHandler(c *gin.Context) {
	// Get the raw request body
	body, err := c.GetRawData()
	if err != nil {
		h.logger.Error("Failed to read webhook request body",
			zap.Error(err),
		)
		utils.SendErrorResponse(c, errors.NewValidationError("WEBHOOK_BODY_READ_FAILED", "Failed to read webhook request body"))
		return
	}

	// Get Stripe signature from headers
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		h.logger.Error("Missing Stripe signature in webhook request")
		utils.SendErrorResponse(c, errors.NewValidationError("MISSING_STRIPE_SIGNATURE", "Stripe signature is required"))
		return
	}

	// In a real implementation, you would verify the webhook signature
	// and process the event accordingly
	// For now, we'll just log the event

	h.logger.Info("Received Stripe webhook",
		zap.String("signature", signature),
		zap.String("body", string(body)),
	)

	// Respond to acknowledge receipt of the webhook
	c.Status(http.StatusOK)
}