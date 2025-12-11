package services

import (
	"context"
	"encoding/json"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
	"github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/account"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConnectedAccountService provides business logic for connected accounts
type ConnectedAccountService struct {
	db           *gorm.DB
	stripeClient StripeClientInterface
	config       *config.StripeConfig
	logger       *zap.Logger
}

// NewConnectedAccountService creates a new connected account service
func NewConnectedAccountService(db *gorm.DB, stripeClient StripeClientInterface, config *config.StripeConfig, logger *zap.Logger) *ConnectedAccountService {
	return &ConnectedAccountService{
		db:           db,
		stripeClient: stripeClient,
		config:       config,
		logger:       logger,
	}
}

// CreateConnectedAccount creates a new Stripe Connect account
func (s *ConnectedAccountService) CreateConnectedAccount(ctx context.Context, req *CreateConnectedAccountRequest) (*models.ConnectedAccount, error) {
	s.logger.Info("Creating connected account",
		zap.String("user_id", req.UserID),
		zap.String("email", req.Email),
		zap.String("country", req.Country),
	)

	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check if user already has a connected account
	var existingAccount models.ConnectedAccount
	if err := s.db.Where("user_id = ?", req.UserID).First(&existingAccount).Error; err == nil {
		return nil, errors.NewConflictError("USER_ALREADY_HAS_ACCOUNT", "User already has a connected account")
	} else if err != gorm.ErrRecordNotFound {
		return nil, errors.NewDatabaseError("DB_ERROR", "Failed to check existing account")
	}

	// Create Stripe account
	stripeParams := &ConnectedAccountParams{
		UserID:   req.UserID,
		Email:    req.Email,
		Country:  req.Country,
		Username: req.Username,
	}

	stripeAccount, err := s.stripeClient.CreateConnectedAccount(ctx, stripeParams)
	if err != nil {
		return nil, errors.NewExternalServiceError("STRIPE_ERROR", "Failed to create Stripe account", err.Error())
	}

	// Create local record
	connectedAccount := &models.ConnectedAccount{
		UserID:             req.UserID,
		StripeAccountID:    stripeAccount.ID,
		AccountType:        string(stripeAccount.Type),
		Country:            req.Country,
		Currency:           s.getDefaultCurrency(req.Country),
		VerificationStatus: "pending", // Default status for new accounts
		ChargesEnabled:     stripeAccount.ChargesEnabled,
		PayoutsEnabled:     stripeAccount.PayoutsEnabled,
	}

	// Store capabilities and requirements as JSON
	if stripeAccount.Capabilities != nil {
		capabilitiesJSON, _ := json.Marshal(stripeAccount.Capabilities)
		connectedAccount.Metadata = string(capabilitiesJSON)
	}

	if err := s.db.Create(connectedAccount).Error; err != nil {
		return nil, errors.NewDatabaseError("DB_ERROR", "Failed to create connected account record")
	}

	s.logger.Info("Connected account created successfully",
		zap.String("account_id", connectedAccount.ID),
		zap.String("stripe_account_id", stripeAccount.ID),
		zap.String("user_id", req.UserID),
	)

	return connectedAccount, nil
}

// GetConnectedAccount retrieves a connected account by ID
func (s *ConnectedAccountService) GetConnectedAccount(ctx context.Context, accountID string) (*models.ConnectedAccount, error) {
	s.logger.Info("Retrieving connected account", zap.String("account_id", accountID))

	var connectedAccount models.ConnectedAccount
	if err := s.db.Where("id = ?", accountID).First(&connectedAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("ACCOUNT_NOT_FOUND", "Connected account not found")
		}
		return nil, errors.NewDatabaseError("DB_ERROR", "Failed to retrieve connected account")
	}

	// Get latest data from Stripe
	stripeAccount, err := s.stripeClient.GetAccount(ctx, connectedAccount.StripeAccountID)
	if err != nil {
		s.logger.Error("Failed to get Stripe account", zap.Error(err), zap.String("stripe_account_id", connectedAccount.StripeAccountID))
		// Continue with local data if Stripe is unavailable
	} else {
		// Update local record with latest Stripe data
		connectedAccount.VerificationStatus = "pending" // Default status
		connectedAccount.ChargesEnabled = stripeAccount.ChargesEnabled
		connectedAccount.PayoutsEnabled = stripeAccount.PayoutsEnabled

		if stripeAccount.Capabilities != nil {
			capabilitiesJSON, _ := json.Marshal(stripeAccount.Capabilities)
			connectedAccount.Metadata = string(capabilitiesJSON)
		}

		s.db.Save(&connectedAccount)
	}

	return &connectedAccount, nil
}

// GetConnectedAccountByUserID retrieves a connected account by user ID
func (s *ConnectedAccountService) GetConnectedAccountByUserID(ctx context.Context, userID string) (*models.ConnectedAccount, error) {
	s.logger.Info("Retrieving connected account by user ID", zap.String("user_id", userID))

	var connectedAccount models.ConnectedAccount
	if err := s.db.Where("user_id = ?", userID).First(&connectedAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("ACCOUNT_NOT_FOUND", "Connected account not found for user")
		}
		return nil, errors.NewDatabaseError("DB_ERROR", "Failed to retrieve connected account")
	}

	// Get latest data from Stripe
	stripeAccount, err := s.stripeClient.GetAccount(ctx, connectedAccount.StripeAccountID)
	if err != nil {
		s.logger.Error("Failed to get Stripe account", zap.Error(err), zap.String("stripe_account_id", connectedAccount.StripeAccountID))
		// Continue with local data if Stripe is unavailable
	} else {
		// Update local record with latest Stripe data
		connectedAccount.VerificationStatus = "pending" // Default status
		connectedAccount.ChargesEnabled = stripeAccount.ChargesEnabled
		connectedAccount.PayoutsEnabled = stripeAccount.PayoutsEnabled

		if stripeAccount.Capabilities != nil {
			capabilitiesJSON, _ := json.Marshal(stripeAccount.Capabilities)
			connectedAccount.Metadata = string(capabilitiesJSON)
		}

		s.db.Save(&connectedAccount)
	}

	return &connectedAccount, nil
}

// UpdateConnectedAccount updates a connected account
func (s *ConnectedAccountService) UpdateConnectedAccount(ctx context.Context, accountID string, req *UpdateConnectedAccountRequest) (*models.ConnectedAccount, error) {
	s.logger.Info("Updating connected account", zap.String("account_id", accountID))

	// Get existing account
	connectedAccount, err := s.GetConnectedAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Update Stripe account if needed
	if req.Email != "" || req.Country != "" {
		stripeParams := &stripe.AccountParams{
			Email: stripe.String(req.Email),
		}

		if req.Country != "" {
			stripeParams.Country = stripe.String(req.Country)
		}

		_, err := account.Update(connectedAccount.StripeAccountID, stripeParams)
		if err != nil {
			return nil, errors.NewExternalServiceError("STRIPE_ERROR", "Failed to update Stripe account", err.Error())
		}
	}

	// Update local record
	updates := make(map[string]interface{})
	if req.Email != "" {
		// Note: Email is stored in Stripe, not locally
	}
	if req.Country != "" {
		updates["country"] = req.Country
	}

	if len(updates) > 0 {
		if err := s.db.Model(connectedAccount).Updates(updates).Error; err != nil {
			return nil, errors.NewDatabaseError("DB_ERROR", "Failed to update connected account")
		}
	}

	// Refresh account data
	return s.GetConnectedAccount(ctx, accountID)
}

// DeleteConnectedAccount deactivates a connected account
func (s *ConnectedAccountService) DeleteConnectedAccount(ctx context.Context, accountID string) error {
	s.logger.Info("Deleting connected account", zap.String("account_id", accountID))

	// Get existing account
	connectedAccount, err := s.GetConnectedAccount(ctx, accountID)
	if err != nil {
		return err
	}

	// Delete from Stripe (this actually deactivates the account)
	_, err = account.Del(connectedAccount.StripeAccountID, nil)
	if err != nil {
		return errors.NewExternalServiceError("STRIPE_ERROR", "Failed to delete Stripe account", err.Error())
	}

	// Soft delete from local database
	if err := s.db.Delete(connectedAccount).Error; err != nil {
		return errors.NewDatabaseError("DB_ERROR", "Failed to delete connected account")
	}

	s.logger.Info("Connected account deleted successfully", zap.String("account_id", accountID))
	return nil
}

// GetAccountOnboardingLink generates an onboarding URL for a connected account
func (s *ConnectedAccountService) GetAccountOnboardingLink(ctx context.Context, accountID string, refreshURL, returnURL string) (string, error) {
	s.logger.Info("Generating onboarding link", zap.String("account_id", accountID))

	// Get existing account
	connectedAccount, err := s.GetConnectedAccount(ctx, accountID)
	if err != nil {
		return "", err
	}

	// Create account link with Stripe
	accountLink, err := s.stripeClient.CreateAccountLink(ctx, connectedAccount.StripeAccountID, refreshURL, returnURL)
	if err != nil {
		return "", errors.NewExternalServiceError("STRIPE_ERROR", "Failed to create onboarding link", err.Error())
	}

	s.logger.Info("Onboarding link generated successfully",
		zap.String("account_id", accountID),
	)

	return accountLink.URL, nil
}

// GetAccountLoginLink generates a dashboard login URL for a connected account
func (s *ConnectedAccountService) GetAccountLoginLink(ctx context.Context, accountID string) (string, error) {
	s.logger.Info("Generating login link", zap.String("account_id", accountID))

	// Get existing account
	connectedAccount, err := s.GetConnectedAccount(ctx, accountID)
	if err != nil {
		return "", err
	}

	// Create login link with Stripe
	loginLink, err := s.stripeClient.CreateLoginLink(ctx, connectedAccount.StripeAccountID)
	if err != nil {
		return "", errors.NewExternalServiceError("STRIPE_ERROR", "Failed to create login link", err.Error())
	}

	s.logger.Info("Login link generated successfully",
		zap.String("account_id", accountID),
	)

	return loginLink.URL, nil
}

// ListConnectedAccounts retrieves a list of connected accounts with pagination and optional user ID filter
func (s *ConnectedAccountService) ListConnectedAccounts(ctx context.Context, page, perPage int, userID string) ([]*models.ConnectedAccount, int64, error) {
	s.logger.Info("Listing connected accounts",
		zap.Int("page", page),
		zap.Int("per_page", perPage),
		zap.String("user_id", userID),
	)

	var accounts []*models.ConnectedAccount
	var total int64

	// Calculate offset
	offset := (page - 1) * perPage

	// Build query
	query := s.db.Model(&models.ConnectedAccount{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("DB_ERROR", "Failed to count connected accounts")
	}

	// Get accounts with pagination
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&accounts).Error; err != nil {
		return nil, 0, errors.NewDatabaseError("DB_ERROR", "Failed to list connected accounts")
	}

	s.logger.Info("Connected accounts listed successfully",
		zap.Int("count", len(accounts)),
		zap.Int64("total", total),
	)

	return accounts, total, nil
}

// GetAccountBalance retrieves the balance for a connected account
func (s *ConnectedAccountService) GetAccountBalance(ctx context.Context, accountID string) (*stripe.Balance, error) {
	s.logger.Info("Getting account balance", zap.String("account_id", accountID))

	// Get connected account to find Stripe Account ID
	connectedAccount, err := s.GetConnectedAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Get balance from Stripe
	balance, err := s.stripeClient.GetBalance(ctx, connectedAccount.StripeAccountID)
	if err != nil {
		s.logger.Error("Failed to get account balance",
			zap.Error(err),
			zap.String("account_id", accountID),
			zap.String("stripe_account_id", connectedAccount.StripeAccountID),
		)
		return nil, errors.NewExternalServiceError("STRIPE_ERROR", "Failed to retrieve account balance", err.Error())
	}

	return balance, nil
}

// GetAccountTransactions retrieves transactions for a connected account
func (s *ConnectedAccountService) GetAccountTransactions(ctx context.Context, accountID string, params *BalanceTransactionListParams) ([]*stripe.BalanceTransaction, error) {
	s.logger.Info("Getting account transactions", zap.String("account_id", accountID))

	// Get connected account to find Stripe Account ID
	connectedAccount, err := s.GetConnectedAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Update params with Stripe Account ID
	if params == nil {
		params = &BalanceTransactionListParams{}
	}
	params.AccountID = connectedAccount.StripeAccountID

	// Get transactions from Stripe
	transactions, err := s.stripeClient.ListBalanceTransactions(ctx, params)
	if err != nil {
		s.logger.Error("Failed to get account transactions",
			zap.Error(err),
			zap.String("account_id", accountID),
			zap.String("stripe_account_id", connectedAccount.StripeAccountID),
		)
		return nil, errors.NewExternalServiceError("STRIPE_ERROR", "Failed to retrieve account transactions", err.Error())
	}

	return transactions, nil
}

// Helper methods

func (s *ConnectedAccountService) validateCreateRequest(req *CreateConnectedAccountRequest) error {
	if req.UserID == "" {
		return errors.NewValidationError("INVALID_USER_ID", "User ID is required")
	}

	if req.Email == "" {
		return errors.NewValidationError("INVALID_EMAIL", "Email is required")
	}

	if req.Country == "" {
		return errors.NewValidationError("INVALID_COUNTRY", "Country is required")
	}

	if len(req.Country) != 2 {
		return errors.NewValidationError("INVALID_COUNTRY", "Country must be a 2-letter ISO code")
	}

	// Validate account type
	if req.AccountType != "" && req.AccountType != "express" && req.AccountType != "custom" && req.AccountType != "standard" {
		return errors.NewValidationError("INVALID_ACCOUNT_TYPE", "Account type must be express, custom, or standard")
	}

	return nil
}

func (s *ConnectedAccountService) getDefaultCurrency(country string) string {
	// Default currency mapping based on country
	currencyMap := map[string]string{
		"US": "usd",
		"CA": "cad",
		"GB": "gbp",
		"EU": "eur",
		"AU": "aud",
		"JP": "jpy",
		"SG": "sgd",
		"HK": "hkd",
		"MY": "myr",
	}

	if currency, exists := currencyMap[country]; exists {
		return currency
	}

	return "usd" // Default to USD
}

// Request structs

// CreateConnectedAccountRequest represents a request to create a connected account
type CreateConnectedAccountRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Country     string `json:"country" binding:"required,len=2"`
	Username    string `json:"username"`
	AccountType string `json:"account_type"`
}

// UpdateConnectedAccountRequest represents a request to update a connected account
type UpdateConnectedAccountRequest struct {
	Email   string `json:"email" binding:"omitempty,email"`
	Country string `json:"country" binding:"omitempty,len=2"`
}
