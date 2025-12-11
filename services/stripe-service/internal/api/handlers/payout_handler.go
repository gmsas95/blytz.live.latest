package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/dtos"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/services"
	"github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	"go.uber.org/zap"
)

// PayoutHandler handles payout-related HTTP requests
type PayoutHandler struct {
	payoutService *services.PayoutService
	logger        *zap.Logger
}

// NewPayoutHandler creates a new payout handler
func NewPayoutHandler(payoutService *services.PayoutService, logger *zap.Logger) *PayoutHandler {
	return &PayoutHandler{
		payoutService: payoutService,
		logger:        logger,
	}
}

// CreatePayout creates a new manual payout
func (h *PayoutHandler) CreatePayout(c *gin.Context) {
	var req dtos.CreatePayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request body"))
		return
	}

	// Validate request
	if errors := dtos.ValidateCreatePayoutRequest(&req); len(errors) > 0 {
		utils.SendValidationErrorResponse(c, errors)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Create payout
	payout, err := h.payoutService.CreatePayout(c.Request.Context(), &services.CreatePayoutRequest{
		Amount:               req.Amount,
		Currency:             req.Currency,
		ConnectedAccountID:    req.ConnectedAccountID,
		StatementDescriptor:   req.StatementDescriptor,
		RequestedBy:          req.RequestedBy,
		Metadata:             req.Metadata,
	})

	if err != nil {
		h.logger.Error("Failed to create payout",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("connected_account_id", req.ConnectedAccountID),
		)
		utils.SendErrorResponse(c, errors.NewInternalError("PAYOUT_ERROR", "Failed to create payout"))
		return
	}

	// Convert to response DTO
	response := dtos.ToPayoutResponse(payout)

	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// CreateScheduledPayout creates a new scheduled payout
func (h *PayoutHandler) CreateScheduledPayout(c *gin.Context) {
	var req dtos.CreateScheduledPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request body"))
		return
	}

	// Validate request
	if errors := dtos.ValidateCreateScheduledPayoutRequest(&req); len(errors) > 0 {
		utils.SendValidationErrorResponse(c, errors)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Create scheduled payout
	payout, err := h.payoutService.CreateScheduledPayout(c.Request.Context(), &services.CreateScheduledPayoutRequest{
		Amount:               req.Amount,
		Currency:             req.Currency,
		ConnectedAccountID:    req.ConnectedAccountID,
		Schedule:             req.Schedule,
		Frequency:            req.Frequency,
		RequestedBy:          req.RequestedBy,
		Metadata:             req.Metadata,
	})

	if err != nil {
		h.logger.Error("Failed to create scheduled payout",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("connected_account_id", req.ConnectedAccountID),
		)
		utils.SendErrorResponse(c, errors.NewInternalError("SCHEDULED_PAYOUT_ERROR", "Failed to create scheduled payout"))
		return
	}

	// Convert to response DTO
	response := dtos.ToPayoutResponse(payout)

	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// GetPayout retrieves a payout by ID
func (h *PayoutHandler) GetPayout(c *gin.Context) {
	payoutID := c.Param("id")
	if payoutID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_ID", "Payout ID is required"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get payout
	payout, err := h.payoutService.GetPayout(c.Request.Context(), payoutID)
	if err != nil {
		h.logger.Error("Failed to get payout",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("payout_id", payoutID),
		)
		if err.Error() == "payout not found" {
			utils.SendErrorResponse(c, errors.NewNotFoundError("PAYOUT_NOT_FOUND", "Payout not found"))
		} else {
			utils.SendErrorResponse(c, errors.NewInternalError("GET_PAYOUT_ERROR", "Failed to retrieve payout"))
		}
		return
	}

	// Convert to response DTO
	response := dtos.ToPayoutResponse(payout)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetPayoutHistory retrieves payout history for an account
func (h *PayoutHandler) GetPayoutHistory(c *gin.Context) {
	accountID := c.Param("account_id")
	if accountID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_ID", "Account ID is required"))
		return
	}

	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 10
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get payout history
	payouts, total, err := h.payoutService.GetPayoutHistory(c.Request.Context(), accountID, page, perPage)
	if err != nil {
		h.logger.Error("Failed to get payout history",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("account_id", accountID),
		)
		utils.SendErrorResponse(c, errors.NewInternalError("GET_PAYOUT_HISTORY_ERROR", "Failed to retrieve payout history"))
		return
	}

	// Convert to response DTO
	response := dtos.ToPayoutHistoryResponse(payouts, page, perPage, total)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// CancelPayout cancels a pending payout
func (h *PayoutHandler) CancelPayout(c *gin.Context) {
	payoutID := c.Param("id")
	if payoutID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_ID", "Payout ID is required"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Cancel payout
	payout, err := h.payoutService.CancelPayout(c.Request.Context(), payoutID)
	if err != nil {
		h.logger.Error("Failed to cancel payout",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("payout_id", payoutID),
		)
		if err.Error() == "payout not found" {
			utils.SendErrorResponse(c, errors.NewNotFoundError("PAYOUT_NOT_FOUND", "Payout not found"))
		} else if len(err.Error()) >= len("payout cannot be cancelled") && err.Error()[:len("payout cannot be cancelled")] == "payout cannot be cancelled" {
			utils.SendErrorResponse(c, errors.NewValidationError("INVALID_STATUS", err.Error()))
		} else {
			utils.SendErrorResponse(c, errors.NewInternalError("CANCEL_PAYOUT_ERROR", "Failed to cancel payout"))
		}
		return
	}

	// Convert to response DTO
	response := dtos.ToPayoutResponse(payout)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetAvailableBalance checks available balance for payouts
func (h *PayoutHandler) GetAvailableBalance(c *gin.Context) {
	accountID := c.Param("account_id")
	if accountID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_ID", "Account ID is required"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get available balance
	balance, err := h.payoutService.GetAvailableBalance(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get available balance",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("account_id", accountID),
		)
		if err.Error() == "connected account not found" {
			utils.SendErrorResponse(c, errors.NewNotFoundError("ACCOUNT_NOT_FOUND", "Connected account not found"))
		} else {
			utils.SendErrorResponse(c, errors.NewInternalError("GET_BALANCE_ERROR", "Failed to retrieve balance"))
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, balance)
}