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

// TransferHandler handles transfer-related HTTP requests
type TransferHandler struct {
	transferService *services.TransferService
	logger          *zap.Logger
}

// NewTransferHandler creates a new transfer handler
func NewTransferHandler(transferService *services.TransferService, logger *zap.Logger) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
		logger:          logger,
	}
}

// CreateTransfer creates a new transfer
func (h *TransferHandler) CreateTransfer(c *gin.Context) {
	var req dtos.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request body"))
		return
	}

	// Validate request
	if errors := dtos.ValidateCreateTransferRequest(&req); len(errors) > 0 {
		utils.SendValidationErrorResponse(c, errors)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// For now, we'll use the authenticated user ID as the requested by
	// In a real implementation, you might want to check if the user has permission
	// to create transfers for the specified connected account

	// Create transfer
	transfer, err := h.transferService.CreateTransfer(c.Request.Context(), &services.CreateTransferRequest{
		Amount:               req.Amount,
		Currency:             req.Currency,
		ConnectedAccountID:    req.ConnectedAccountID,
		TransferGroup:        req.TransferGroup,
		SourceTransaction:     req.SourceTransaction,
		DestinationPaymentID:  req.DestinationPaymentID,
		FeePercent:           req.FeePercent,
		Metadata:             req.Metadata,
	})

	if err != nil {
		h.logger.Error("Failed to create transfer",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("connected_account_id", req.ConnectedAccountID),
		)
		utils.SendErrorResponse(c, errors.NewInternalError("TRANSFER_ERROR", "Failed to create transfer"))
		return
	}

	// Convert to response DTO
	response := dtos.ToTransferResponse(transfer)

	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// CreateBatchTransfers creates multiple transfers in a batch
func (h *TransferHandler) CreateBatchTransfers(c *gin.Context) {
	var req dtos.CreateBatchTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request body"))
		return
	}

	// Validate that at least one transfer is provided
	if len(req.Transfers) == 0 {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "At least one transfer is required"))
		return
	}

	// Validate each transfer
	for i, transfer := range req.Transfers {
		if errors := dtos.ValidateCreateTransferRequest(&transfer); len(errors) > 0 {
			utils.SendValidationErrorResponse(c, map[string]string{
				"transfer_" + strconv.Itoa(i): "Invalid transfer data",
			})
			return
		}
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Convert to service requests
	serviceReqs := make([]*services.CreateTransferRequest, len(req.Transfers))
	for i, transfer := range req.Transfers {
		serviceReqs[i] = &services.CreateTransferRequest{
			Amount:               transfer.Amount,
			Currency:             transfer.Currency,
			ConnectedAccountID:    transfer.ConnectedAccountID,
			TransferGroup:        transfer.TransferGroup,
			SourceTransaction:     transfer.SourceTransaction,
			DestinationPaymentID:  transfer.DestinationPaymentID,
			FeePercent:           transfer.FeePercent,
			Metadata:             transfer.Metadata,
		}
	}

	// Create batch transfers
	transfers, err := h.transferService.CreateBatchTransfers(c.Request.Context(), serviceReqs)
	if err != nil {
		h.logger.Error("Failed to create batch transfers",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.Int("transfer_count", len(serviceReqs)),
		)
		utils.SendErrorResponse(c, errors.NewInternalError("BATCH_TRANSFER_ERROR", "Failed to create batch transfers"))
		return
	}

	// Convert to response DTOs
	responses := make([]dtos.TransferResponse, len(transfers))
	for i, transfer := range transfers {
		responses[i] = *dtos.ToTransferResponse(transfer)
	}

	utils.SendSuccessResponse(c, http.StatusCreated, responses)
}

// GetTransfer retrieves a transfer by ID
func (h *TransferHandler) GetTransfer(c *gin.Context) {
	transferID := c.Param("id")
	if transferID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_ID", "Transfer ID is required"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get transfer
	transfer, err := h.transferService.GetTransfer(c.Request.Context(), transferID)
	if err != nil {
		h.logger.Error("Failed to get transfer",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("transfer_id", transferID),
		)
		if err.Error() == "transfer not found" {
			utils.SendErrorResponse(c, errors.NewNotFoundError("TRANSFER_NOT_FOUND", "Transfer not found"))
		} else {
			utils.SendErrorResponse(c, errors.NewInternalError("GET_TRANSFER_ERROR", "Failed to retrieve transfer"))
		}
		return
	}

	// Convert to response DTO
	response := dtos.ToTransferResponse(transfer)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetTransferHistory retrieves transfer history for an account
func (h *TransferHandler) GetTransferHistory(c *gin.Context) {
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

	// Get transfer history
	transfers, total, err := h.transferService.GetTransferHistory(c.Request.Context(), accountID, page, perPage)
	if err != nil {
		h.logger.Error("Failed to get transfer history",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("account_id", accountID),
		)
		utils.SendErrorResponse(c, errors.NewInternalError("GET_TRANSFER_HISTORY_ERROR", "Failed to retrieve transfer history"))
		return
	}

	// Convert to response DTO
	response := dtos.ToTransferHistoryResponse(transfers, page, perPage, total)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// ReverseTransfer reverses a transfer
func (h *TransferHandler) ReverseTransfer(c *gin.Context) {
	transferID := c.Param("id")
	if transferID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_ID", "Transfer ID is required"))
		return
	}

	var req dtos.ReverseTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request body"))
		return
	}

	// Validate request
	if errors := dtos.ValidateReverseTransferRequest(&req); len(errors) > 0 {
		utils.SendValidationErrorResponse(c, errors)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, errors.NewAuthenticationError("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Reverse transfer
	transfer, err := h.transferService.ReverseTransfer(c.Request.Context(), transferID, req.Amount)
	if err != nil {
		h.logger.Error("Failed to reverse transfer",
			zap.Error(err),
			zap.String("user_id", userID.(string)),
			zap.String("transfer_id", transferID),
			zap.Float64("amount", req.Amount),
		)
		if err.Error() == "transfer not found" {
			utils.SendErrorResponse(c, errors.NewNotFoundError("TRANSFER_NOT_FOUND", "Transfer not found"))
		} else if len(err.Error()) >= len("transfer cannot be reversed") && err.Error()[:len("transfer cannot be reversed")] == "transfer cannot be reversed" {
			utils.SendErrorResponse(c, errors.NewValidationError("INVALID_STATUS", err.Error()))
		} else {
			utils.SendErrorResponse(c, errors.NewInternalError("REVERSE_TRANSFER_ERROR", "Failed to reverse transfer"))
		}
		return
	}

	// Convert to response DTO
	response := dtos.ToTransferResponse(transfer)

	utils.SendSuccessResponse(c, http.StatusOK, response)
}