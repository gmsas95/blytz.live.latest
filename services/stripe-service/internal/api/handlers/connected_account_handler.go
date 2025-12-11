package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/api/dtos"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/services"
	"github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	"go.uber.org/zap"
)

// ConnectedAccountHandler handles connected account related requests
type ConnectedAccountHandler struct {
	connectedAccountService *services.ConnectedAccountService
	logger                  *zap.Logger
}

// NewConnectedAccountHandler creates a new connected account handler
func NewConnectedAccountHandler(connectedAccountService *services.ConnectedAccountService, logger *zap.Logger) *ConnectedAccountHandler {
	return &ConnectedAccountHandler{
		connectedAccountService: connectedAccountService,
		logger:                  logger,
	}
}

// CreateConnectedAccount creates a new connected account
func (h *ConnectedAccountHandler) CreateConnectedAccount(c *gin.Context) {
	var req dtos.CreateConnectedAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, &errors.AppError{
			Type:    errors.AuthenticationError,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	// Use authenticated user ID if not provided in request
	if req.UserID == "" {
		req.UserID = userID.(string)
	} else if req.UserID != userID.(string) {
		// User can only create account for themselves
		utils.SendErrorResponse(c, &errors.AppError{
			Type:    errors.AuthorizationError,
			Code:    "FORBIDDEN",
			Message: "Cannot create account for another user",
		})
		return
	}

	// Create connected account
	account, err := h.connectedAccountService.CreateConnectedAccount(c.Request.Context(), &services.CreateConnectedAccountRequest{
		UserID:      req.UserID,
		Email:       req.Email,
		Country:     req.Country,
		Username:    req.Username,
		AccountType: req.AccountType,
	})
	if err != nil {
		h.logger.Error("Failed to create connected account", zap.Error(err))
		utils.SendErrorResponse(c, err)
		return
	}

	response := dtos.ToConnectedAccountResponse(account)
	utils.SendSuccessResponse(c, http.StatusCreated, response)
}

// GetConnectedAccount retrieves a connected account by ID
func (h *ConnectedAccountHandler) GetConnectedAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		utils.SendValidationErrorResponse(c, map[string]string{"id": "Account ID is required"})
		return
	}

	account, err := h.connectedAccountService.GetConnectedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Check if user owns this account
	userID, exists := c.Get("userID")
	if exists && userID.(string) != account.UserID {
		utils.SendErrorResponse(c, errors.NewAuthorizationError("FORBIDDEN", "Access denied to this account"))
		return
	}

	response := dtos.ToConnectedAccountResponse(account)
	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetConnectedAccountByUserID retrieves a connected account by user ID
func (h *ConnectedAccountHandler) GetConnectedAccountByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		utils.SendValidationErrorResponse(c, map[string]string{"user_id": "User ID is required"})
		return
	}

	// Check if user is requesting their own account
	authUserID, exists := c.Get("userID")
	if exists && authUserID.(string) != userID {
		utils.SendErrorResponse(c, errors.NewAuthorizationError("FORBIDDEN", "Access denied to this account"))
		return
	}

	account, err := h.connectedAccountService.GetConnectedAccountByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get connected account by user ID", zap.Error(err), zap.String("user_id", userID))
		utils.SendErrorResponse(c, err)
		return
	}

	response := dtos.ToConnectedAccountResponse(account)
	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// UpdateConnectedAccount updates a connected account
func (h *ConnectedAccountHandler) UpdateConnectedAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		utils.SendValidationErrorResponse(c, map[string]string{"id": "Account ID is required"})
		return
	}

	var req dtos.UpdateConnectedAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Get existing account to check ownership
	existingAccount, err := h.connectedAccountService.GetConnectedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get existing connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Check if user owns this account
	userID, exists := c.Get("userID")
	if !exists || userID.(string) != existingAccount.UserID {
		utils.SendErrorResponse(c, errors.NewAuthorizationError("FORBIDDEN", "Access denied to this account"))
		return
	}

	// Update connected account
	account, err := h.connectedAccountService.UpdateConnectedAccount(c.Request.Context(), accountID, &services.UpdateConnectedAccountRequest{
		Email:   req.Email,
		Country: req.Country,
	})
	if err != nil {
		h.logger.Error("Failed to update connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	response := dtos.ToConnectedAccountResponse(account)
	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// DeleteConnectedAccount deactivates a connected account
func (h *ConnectedAccountHandler) DeleteConnectedAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		utils.SendValidationErrorResponse(c, map[string]string{"id": "Account ID is required"})
		return
	}

	// Get existing account to check ownership
	existingAccount, err := h.connectedAccountService.GetConnectedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get existing connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Check if user owns this account
	userID, exists := c.Get("userID")
	if !exists || userID.(string) != existingAccount.UserID {
		utils.SendErrorResponse(c, errors.NewAuthorizationError("FORBIDDEN", "Access denied to this account"))
		return
	}

	// Delete connected account
	err = h.connectedAccountService.DeleteConnectedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to delete connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Connected account deleted successfully", nil)
}

// GetAccountOnboardingLink generates an onboarding URL for a connected account
func (h *ConnectedAccountHandler) GetAccountOnboardingLink(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		utils.SendValidationErrorResponse(c, map[string]string{"id": "Account ID is required"})
		return
	}

	var req dtos.OnboardingLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Get existing account to check ownership
	existingAccount, err := h.connectedAccountService.GetConnectedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get existing connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Check if user owns this account
	userID, exists := c.Get("userID")
	if !exists || userID.(string) != existingAccount.UserID {
		utils.SendErrorResponse(c, errors.NewAuthorizationError("FORBIDDEN", "Access denied to this account"))
		return
	}

	// Generate onboarding link
	url, err := h.connectedAccountService.GetAccountOnboardingLink(c.Request.Context(), accountID, req.RefreshURL, req.ReturnURL)
	if err != nil {
		h.logger.Error("Failed to generate onboarding link", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Account links expire in 30 days
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	response := dtos.OnboardingLinkResponse{
		URL:       url,
		ExpiresAt: expiresAt,
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetAccountLoginLink generates a dashboard login URL for a connected account
func (h *ConnectedAccountHandler) GetAccountLoginLink(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		utils.SendValidationErrorResponse(c, map[string]string{"id": "Account ID is required"})
		return
	}

	// Get existing account to check ownership
	existingAccount, err := h.connectedAccountService.GetConnectedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get existing connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Check if user owns this account
	userID, exists := c.Get("userID")
	if !exists || userID.(string) != existingAccount.UserID {
		utils.SendErrorResponse(c, errors.NewAuthorizationError("FORBIDDEN", "Access denied to this account"))
		return
	}

	// Generate login link
	url, err := h.connectedAccountService.GetAccountLoginLink(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to generate login link", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	response := dtos.LoginLinkResponse{
		URL: url,
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// ListConnectedAccounts lists connected accounts with pagination and filtering
func (h *ConnectedAccountHandler) ListConnectedAccounts(c *gin.Context) {
	var req dtos.ConnectedAccountListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Set default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}

	// If user is not admin, only show their own accounts
	userID, exists := c.Get("userID")
	if exists {
		userRole, _ := c.Get("userRole")
		if userRole != "admin" {
			req.UserID = userID.(string)
		}
	}

	// Get accounts from service
	accounts, total, err := h.connectedAccountService.ListConnectedAccounts(c.Request.Context(), req.Page, req.PerPage, req.UserID)
	if err != nil {
		h.logger.Error("Failed to list connected accounts", zap.Error(err))
		utils.SendErrorResponse(c, err)
		return
	}

	response := dtos.ConnectedAccountListResponse{
		Accounts: dtos.ToConnectedAccountResponseList(accounts),
		Total:    total,
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetAccountBalance retrieves balance for a connected account
func (h *ConnectedAccountHandler) GetAccountBalance(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		utils.SendValidationErrorResponse(c, map[string]string{"id": "Account ID is required"})
		return
	}

	// Get existing account to check ownership
	existingAccount, err := h.connectedAccountService.GetConnectedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get existing connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Check if user owns this account
	userID, exists := c.Get("userID")
	if !exists || userID.(string) != existingAccount.UserID {
		utils.SendErrorResponse(c, errors.NewAuthorizationError("FORBIDDEN", "Access denied to this account"))
		return
	}

	// Get balance from service
	balance, err := h.connectedAccountService.GetAccountBalance(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get account balance", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Map to response
	available := make([]dtos.BalanceItem, len(balance.Available))
	for i, b := range balance.Available {
		sourceTypes := make(map[string]int64)
		for k, v := range b.SourceTypes {
			sourceTypes[string(k)] = v
		}
		available[i] = dtos.BalanceItem{
			Amount:      b.Amount,
			Currency:    string(b.Currency),
			SourceTypes: sourceTypes,
		}
	}

	pending := make([]dtos.BalanceItem, len(balance.Pending))
	for i, b := range balance.Pending {
		sourceTypes := make(map[string]int64)
		for k, v := range b.SourceTypes {
			sourceTypes[string(k)] = v
		}
		pending[i] = dtos.BalanceItem{
			Amount:      b.Amount,
			Currency:    string(b.Currency),
			SourceTypes: sourceTypes,
		}
	}

	response := dtos.AccountBalanceResponse{
		Available: available,
		Pending:   pending,
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetAccountTransactions retrieves transactions for a connected account
func (h *ConnectedAccountHandler) GetAccountTransactions(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		utils.SendValidationErrorResponse(c, map[string]string{"id": "Account ID is required"})
		return
	}

	var req dtos.AccountTransactionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.SendValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Set account ID from URL parameter
	req.AccountID = accountID

	// Set default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}

	// Get existing account to check ownership
	existingAccount, err := h.connectedAccountService.GetConnectedAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get existing connected account", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Check if user owns this account
	userID, exists := c.Get("userID")
	if !exists || userID.(string) != existingAccount.UserID {
		utils.SendErrorResponse(c, errors.NewAuthorizationError("FORBIDDEN", "Access denied to this account"))
		return
	}

	// Prepare params
	params := &services.BalanceTransactionListParams{
		Limit:  int64(req.PerPage),
		Type:   req.Type,
		Status: req.Status,
	}

	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			params.CreatedStart = t.Unix()
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			// Add 23:59:59 to end date to include the full day
			params.CreatedEnd = t.Add(24*time.Hour - time.Second).Unix()
		}
	}

	// Get transactions from service
	transactions, err := h.connectedAccountService.GetAccountTransactions(c.Request.Context(), accountID, params)
	if err != nil {
		h.logger.Error("Failed to get account transactions", zap.Error(err), zap.String("account_id", accountID))
		utils.SendErrorResponse(c, err)
		return
	}

	// Map to response
	txResponses := make([]dtos.AccountTransactionResponse, len(transactions))
	for i, tx := range transactions {
		var availableAt *time.Time
		if tx.AvailableOn > 0 {
			t := time.Unix(tx.AvailableOn, 0)
			availableAt = &t
		}

		txResponses[i] = dtos.AccountTransactionResponse{
			ID:          tx.ID,
			Type:        string(tx.Type),
			Amount:      tx.Amount,
			Currency:    string(tx.Currency),
			Status:      string(tx.Status),
			Description: tx.Description,
			CreatedAt:   time.Unix(tx.Created, 0),
			AvailableAt: availableAt,
		}
	}

	response := dtos.AccountTransactionListResponse{
		Transactions: txResponses,
		Total:        int64(len(txResponses)), // Stripe doesn't return total count for list endpoints
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}
