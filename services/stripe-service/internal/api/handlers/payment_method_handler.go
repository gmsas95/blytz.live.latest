package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/dtos"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/services"
	"github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	"go.uber.org/zap"
)

// PaymentMethodHandler handles payment method related HTTP requests
type PaymentMethodHandler struct {
	paymentService *services.PaymentService
	stripeClient   *services.StripeClient
	logger         *zap.Logger
}

// NewPaymentMethodHandler creates a new payment method handler
func NewPaymentMethodHandler(paymentService *services.PaymentService, stripeClient *services.StripeClient, logger *zap.Logger) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		paymentService: paymentService,
		stripeClient:   stripeClient,
		logger:         logger,
	}
}

// GetPaymentMethodConfigs returns available payment method configurations
func (h *PaymentMethodHandler) GetPaymentMethodConfigs(c *gin.Context) {
	h.logger.Info("Getting payment method configurations")

	configs := models.GetDefaultPaymentMethodConfigs()

	// Get currency filter if provided
	currency := c.Query("currency")
	if currency != "" {
		// Filter configs by currency
		filteredConfigs := make(map[models.PaymentMethodType]models.PaymentMethodConfig)
		for methodType, config := range configs {
			for _, supportedCurrency := range config.SupportedCurrencies {
				if supportedCurrency == currency {
					filteredConfigs[methodType] = config
					break
				}
			}
		}
		configs = filteredConfigs
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"payment_methods": configs,
		"currency":        currency,
		"count":           len(configs),
	})
}

// CreateCustomPaymentMethod creates a custom payment method
func (h *PaymentMethodHandler) CreateCustomPaymentMethod(c *gin.Context) {
	var req dtos.CreateCustomPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for create custom payment method",
			zap.Error(err),
		)
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := dtos.ValidateCreateCustomPaymentMethodRequest(&req); len(validationErrors) > 0 {
		utils.SendValidationErrorResponse(c, validationErrors)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Validate payment method type and currency
	if err := h.stripeClient.ValidatePaymentMethodType(req.Type, req.Currency); err != nil {
		h.logger.Error("Payment method validation failed",
			zap.Error(err),
			zap.String("type", req.Type),
			zap.String("currency", req.Currency),
		)
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_PAYMENT_METHOD", err.Error()))
		return
	}

	// Create custom payment method parameters
	stripeParams := &services.CustomPaymentMethodParams{
		Type:       req.Type,
		Provider:   req.Provider,
		CustomerID: req.CustomerID,
		Data:       req.Data,
		Metadata:   req.Metadata,
	}

	// Add user ID to metadata
	if stripeParams.Metadata == nil {
		stripeParams.Metadata = make(map[string]string)
	}
	stripeParams.Metadata["user_id"] = userID.(string)

	// Create payment method with Stripe
	pm, err := h.stripeClient.CreateCustomPaymentMethod(c.Request.Context(), stripeParams)
	if err != nil {
		h.logger.Error("Failed to create custom payment method",
			zap.Error(err),
			zap.String("type", req.Type),
			zap.String("provider", req.Provider),
			zap.String("user_id", userID.(string)),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_METHOD_CREATION_FAILED", "Failed to create custom payment method", err.Error()))
		return
	}

	h.logger.Info("Custom payment method created successfully",
		zap.String("payment_method_id", pm.ID),
		zap.String("type", req.Type),
		zap.String("provider", req.Provider),
	)

	response := dtos.ToPaymentMethodResponse(pm, "custom")
	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// CreateMbWayPaymentMethod creates an MbWay payment method
func (h *PaymentMethodHandler) CreateMbWayPaymentMethod(c *gin.Context) {
	var req dtos.CreateMbWayPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for create mbway payment method",
			zap.Error(err),
		)
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := dtos.ValidateCreateMbWayPaymentMethodRequest(&req); len(validationErrors) > 0 {
		utils.SendValidationErrorResponse(c, validationErrors)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Validate payment method type and currency
	if err := h.stripeClient.ValidatePaymentMethodType("mbway", req.Currency); err != nil {
		h.logger.Error("Payment method validation failed",
			zap.Error(err),
			zap.String("phone_number", req.PhoneNumber),
			zap.String("country_code", req.CountryCode),
		)
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_PAYMENT_METHOD", err.Error()))
		return
	}

	// Create MbWay payment method parameters
	stripeParams := &services.MbWayPaymentMethodParams{
		PhoneNumber: req.PhoneNumber,
		CountryCode: req.CountryCode,
		CustomerID:  req.CustomerID,
		Metadata:    req.Metadata,
	}

	// Add user ID to metadata
	if stripeParams.Metadata == nil {
		stripeParams.Metadata = make(map[string]string)
	}
	stripeParams.Metadata["user_id"] = userID.(string)

	// Create payment method with Stripe
	pm, err := h.stripeClient.CreateMbWayPaymentMethod(c.Request.Context(), stripeParams)
	if err != nil {
		h.logger.Error("Failed to create MbWay payment method",
			zap.Error(err),
			zap.String("phone_number", req.PhoneNumber),
			zap.String("country_code", req.CountryCode),
			zap.String("user_id", userID.(string)),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_METHOD_CREATION_FAILED", "Failed to create mbway payment method", err.Error()))
		return
	}

	h.logger.Info("MbWay payment method created successfully",
		zap.String("payment_method_id", pm.ID),
		zap.String("phone_number", req.PhoneNumber),
		zap.String("country_code", req.CountryCode),
	)

	response := dtos.ToPaymentMethodResponse(pm, "mbway")
	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// CreateTWINTPaymentMethod creates a TWINT payment method
func (h *PaymentMethodHandler) CreateTWINTPaymentMethod(c *gin.Context) {
	var req dtos.CreateTWINTPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for create twint payment method",
			zap.Error(err),
		)
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := dtos.ValidateCreateTWINTPaymentMethodRequest(&req); len(validationErrors) > 0 {
		utils.SendValidationErrorResponse(c, validationErrors)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Validate payment method type and currency
	if err := h.stripeClient.ValidatePaymentMethodType("twint", req.Currency); err != nil {
		h.logger.Error("Payment method validation failed",
			zap.Error(err),
			zap.String("customer_id", req.CustomerID),
		)
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_PAYMENT_METHOD", err.Error()))
		return
	}

	// Create TWINT payment method parameters
	stripeParams := &services.TWINTPaymentMethodParams{
		QRCode:     req.QRCode,
		DeviceID:    req.DeviceID,
		MerchantID:  req.MerchantID,
		CustomerID:  req.CustomerID,
		Metadata:    req.Metadata,
	}

	// Add user ID to metadata
	if stripeParams.Metadata == nil {
		stripeParams.Metadata = make(map[string]string)
	}
	stripeParams.Metadata["user_id"] = userID.(string)

	// Create payment method with Stripe
	pm, err := h.stripeClient.CreateTWINTPaymentMethod(c.Request.Context(), stripeParams)
	if err != nil {
		h.logger.Error("Failed to create TWINT payment method",
			zap.Error(err),
			zap.String("customer_id", req.CustomerID),
			zap.String("user_id", userID.(string)),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_METHOD_CREATION_FAILED", "Failed to create twint payment method", err.Error()))
		return
	}

	h.logger.Info("TWINT payment method created successfully",
		zap.String("payment_method_id", pm.ID),
		zap.String("customer_id", req.CustomerID),
	)

	response := dtos.ToPaymentMethodResponse(pm, "twint")
	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// CreateCryptoPaymentMethod creates a crypto payment method
func (h *PaymentMethodHandler) CreateCryptoPaymentMethod(c *gin.Context) {
	var req dtos.CreateCryptoPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for create crypto payment method",
			zap.Error(err),
		)
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := dtos.ValidateCreateCryptoPaymentMethodRequest(&req); len(validationErrors) > 0 {
		utils.SendValidationErrorResponse(c, validationErrors)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Validate payment method type and currency
	if err := h.stripeClient.ValidatePaymentMethodType("crypto", req.Currency); err != nil {
		h.logger.Error("Payment method validation failed",
			zap.Error(err),
			zap.String("crypto_type", req.CryptoType),
			zap.String("wallet_address", req.WalletAddress),
		)
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_PAYMENT_METHOD", err.Error()))
		return
	}

	// Create crypto payment method parameters
	stripeParams := &services.CryptoPaymentMethodParams{
		CryptoType:    req.CryptoType,
		WalletAddress: req.WalletAddress,
		Network:       req.Network,
		CustomerID:    req.CustomerID,
		Metadata:      req.Metadata,
	}

	// Add user ID to metadata
	if stripeParams.Metadata == nil {
		stripeParams.Metadata = make(map[string]string)
	}
	stripeParams.Metadata["user_id"] = userID.(string)

	// Create payment method with Stripe
	pm, err := h.stripeClient.CreateCryptoPaymentMethod(c.Request.Context(), stripeParams)
	if err != nil {
		h.logger.Error("Failed to create crypto payment method",
			zap.Error(err),
			zap.String("crypto_type", req.CryptoType),
			zap.String("wallet_address", req.WalletAddress),
			zap.String("user_id", userID.(string)),
		)
		utils.SendErrorResponse(c, errors.NewExternalServiceError("PAYMENT_METHOD_CREATION_FAILED", "Failed to create crypto payment method", err.Error()))
		return
	}

	h.logger.Info("Crypto payment method created successfully",
		zap.String("payment_method_id", pm.ID),
		zap.String("crypto_type", req.CryptoType),
		zap.String("wallet_address", req.WalletAddress),
	)

	response := dtos.ToPaymentMethodResponse(pm, "crypto")
	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// GetPaymentMethod retrieves a payment method by ID
func (h *PaymentMethodHandler) GetPaymentMethod(c *gin.Context) {
	paymentMethodID := c.Param("id")
	if paymentMethodID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("MISSING_PAYMENT_METHOD_ID", "Payment Method ID is required"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	h.logger.Info("Getting payment method",
		zap.String("payment_method_id", paymentMethodID),
		zap.String("user_id", userID.(string)),
	)

	// This would typically retrieve the payment method from Stripe
	// For now, we'll return a placeholder response
	response := gin.H{
		"id":     paymentMethodID,
		"type":   "unknown",
		"status": "active",
		"user_id": userID.(string),
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// ListAllPaymentMethods lists all payment methods (admin only)
func (h *PaymentMethodHandler) ListAllPaymentMethods(c *gin.Context) {
	h.logger.Info("Listing all payment methods (admin)")

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get all payment method configurations
	configs := models.GetDefaultPaymentMethodConfigs()

	// Convert to response format
	response := gin.H{
		"payment_methods": configs,
		"count":           len(configs),
		"user_id":        userID.(string),
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// UpdatePaymentMethodStatus updates a payment method status (admin only)
func (h *PaymentMethodHandler) UpdatePaymentMethodStatus(c *gin.Context) {
	paymentMethodID := c.Param("id")
	if paymentMethodID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("MISSING_PAYMENT_METHOD_ID", "Payment Method ID is required"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive pending"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for update payment method status",
			zap.Error(err),
		)
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	h.logger.Info("Updating payment method status",
		zap.String("payment_method_id", paymentMethodID),
		zap.String("status", req.Status),
		zap.String("user_id", userID.(string)),
	)

	// This would typically update the payment method status in Stripe
	// For now, we'll return a success response
	response := gin.H{
		"id":     paymentMethodID,
		"status":  req.Status,
		"user_id": userID.(string),
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// ListPaymentMethods lists payment methods for a user
func (h *PaymentMethodHandler) ListPaymentMethods(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	paymentMethodType := c.Query("type")

	h.logger.Info("Listing payment methods",
		zap.String("user_id", userID.(string)),
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("type", paymentMethodType),
	)

	// This would typically retrieve payment methods from Stripe
	// For now, we'll return a placeholder response
	response := gin.H{
		"payment_methods": []gin.H{},
		"pagination": gin.H{
			"page":  page,
			"limit":  limit,
			"total": 0,
		},
		"user_id": userID.(string),
		"type":    paymentMethodType,
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// DeletePaymentMethod deletes a payment method
func (h *PaymentMethodHandler) DeletePaymentMethod(c *gin.Context) {
	paymentMethodID := c.Param("id")
	if paymentMethodID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("MISSING_PAYMENT_METHOD_ID", "Payment Method ID is required"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	h.logger.Info("Deleting payment method",
		zap.String("payment_method_id", paymentMethodID),
		zap.String("user_id", userID.(string)),
	)

	// This would typically delete the payment method from Stripe
	// For now, we'll return a placeholder response
	response := gin.H{
		"id":      paymentMethodID,
		"deleted": true,
		"user_id": userID.(string),
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}