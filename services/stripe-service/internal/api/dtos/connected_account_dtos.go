package dtos

import (
	"time"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/models"
)

// CreateConnectedAccountRequest represents a request to create a connected account
type CreateConnectedAccountRequest struct {
	UserID      string `json:"user_id" binding:"required" validate:"required,uuid"`
	Email       string `json:"email" binding:"required,email" validate:"required,email"`
	Country     string `json:"country" binding:"required,len=2" validate:"required,len=2"`
	Username    string `json:"username" validate:"omitempty,max=100"`
	AccountType string `json:"account_type" validate:"omitempty,oneof=express custom standard"`
}

// UpdateConnectedAccountRequest represents a request to update a connected account
type UpdateConnectedAccountRequest struct {
	Email   string `json:"email" binding:"omitempty,email" validate:"omitempty,email"`
	Country string `json:"country" binding:"omitempty,len=2" validate:"omitempty,len=2"`
}

// ConnectedAccountResponse represents a connected account response
type ConnectedAccountResponse struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	StripeAccountID    string    `json:"stripe_account_id"`
	AccountType        string    `json:"account_type"`
	Country            string    `json:"country"`
	Currency           string    `json:"currency"`
	VerificationStatus string    `json:"verification_status"`
	ChargesEnabled     bool      `json:"charges_enabled"`
	PayoutsEnabled     bool      `json:"payouts_enabled"`
	Metadata           string    `json:"metadata,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// OnboardingLinkRequest represents a request to get an onboarding link
type OnboardingLinkRequest struct {
	RefreshURL string `json:"refresh_url" binding:"required,url" validate:"required,url"`
	ReturnURL  string `json:"return_url" binding:"required,url" validate:"required,url"`
}

// OnboardingLinkResponse represents an onboarding link response
type OnboardingLinkResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginLinkResponse represents a login link response
type LoginLinkResponse struct {
	URL string `json:"url"`
}

// ConnectedAccountListRequest represents a request to list connected accounts
type ConnectedAccountListRequest struct {
	Page        int    `form:"page" validate:"omitempty,min=1"`
	PerPage     int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	UserID      string `form:"user_id" validate:"omitempty,uuid"`
	Status      string `form:"status" validate:"omitempty,oneof=pending verified restricted"`
	AccountType string `form:"account_type" validate:"omitempty,oneof=express custom standard"`
	Country     string `form:"country" validate:"omitempty,len=2"`
}

// ConnectedAccountListResponse represents a list of connected accounts
type ConnectedAccountListResponse struct {
	Accounts []ConnectedAccountResponse `json:"accounts"`
	Total    int64                      `json:"total"`
}

// AccountStatusRequest represents a request to update account status
type AccountStatusRequest struct {
	Status string `json:"status" binding:"required" validate:"required,oneof=active inactive suspended"`
}

// AccountVerificationRequest represents a request to trigger account verification
type AccountVerificationRequest struct {
	IPAddress string `json:"ip_address" validate:"omitempty,ip"`
	UserAgent string `json:"user_agent" validate:"omitempty,max=500"`
}

// AccountVerificationResponse represents a verification response
type AccountVerificationResponse struct {
	Status          string                 `json:"status"`
	Requirements    map[string]interface{} `json:"requirements,omitempty"`
	CurrentDeadline *time.Time             `json:"current_deadline,omitempty"`
	FutureDeadline  *time.Time             `json:"future_deadline,omitempty"`
}

// PayoutSettingsRequest represents a request to update payout settings
type PayoutSettingsRequest struct {
	DestinationBankAccount string  `json:"destination_bank_account" binding:"required"`
	StatementDescriptor    string  `json:"statement_descriptor" validate:"omitempty,max=22"`
	MinimumPayoutAmount    float64 `json:"minimum_payout_amount" validate:"omitempty,min=0"`
	PayoutSchedule         string  `json:"payout_schedule" validate:"omitempty,oneof=daily weekly monthly manual"`
}

// PayoutSettingsResponse represents payout settings response
type PayoutSettingsResponse struct {
	DestinationBankAccount string     `json:"destination_bank_account"`
	StatementDescriptor    string     `json:"statement_descriptor,omitempty"`
	MinimumPayoutAmount    float64    `json:"minimum_payout_amount"`
	PayoutSchedule         string     `json:"payout_schedule"`
	NextPayoutDate         *time.Time `json:"next_payout_date,omitempty"`
}

// AccountCapabilitiesResponse represents account capabilities
type AccountCapabilitiesResponse struct {
	CardPayments    bool `json:"card_payments"`
	Transfers       bool `json:"transfers"`
	LegacyPayments  bool `json:"legacy_payments"`
	BankDebits      bool `json:"bank_debits"`
	PlatformPayouts bool `json:"platform_payouts"`
}

// AccountRequirementsResponse represents account requirements
type AccountRequirementsResponse struct {
	CurrentlyDue    []string `json:"currently_due"`
	EventuallyDue   []string `json:"eventually_due"`
	PastDue         []string `json:"past_due"`
	DisabledReason  string   `json:"disabled_reason,omitempty"`
	CurrentDeadline *string  `json:"current_deadline,omitempty"`
	FutureDeadline  *string  `json:"future_deadline,omitempty"`
}

// AccountBalanceResponse represents account balance
type AccountBalanceResponse struct {
	Available []BalanceItem `json:"available"`
	Pending   []BalanceItem `json:"pending"`
}

// BalanceItem represents a balance item
type BalanceItem struct {
	Amount      int64            `json:"amount"`
	Currency    string           `json:"currency"`
	SourceTypes map[string]int64 `json:"source_types,omitempty"`
}

// AccountTransactionRequest represents a request to list account transactions
type AccountTransactionRequest struct {
	AccountID string `form:"account_id" validate:"required,uuid"`
	Page      int    `form:"page" validate:"omitempty,min=1"`
	PerPage   int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Type      string `form:"type" validate:"omitempty,oneof=payment payout transfer fee"`
	Status    string `form:"status" validate:"omitempty,oneof=succeeded pending failed canceled"`
	StartDate string `form:"start_date" validate:"omitempty,datetime=2006-01-02"`
	EndDate   string `form:"end_date" validate:"omitempty,datetime=2006-01-02"`
}

// AccountTransactionResponse represents a transaction response
type AccountTransactionResponse struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Amount      int64      `json:"amount"`
	Currency    string     `json:"currency"`
	Status      string     `json:"status"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	AvailableAt *time.Time `json:"available_at,omitempty"`
}

// AccountTransactionListResponse represents a list of transactions
type AccountTransactionListResponse struct {
	Transactions []AccountTransactionResponse `json:"transactions"`
	Total        int64                        `json:"total"`
}

// Helper functions to convert between models and DTOs

// ToConnectedAccountResponse converts a ConnectedAccount model to a response DTO
func ToConnectedAccountResponse(account *models.ConnectedAccount) ConnectedAccountResponse {
	return ConnectedAccountResponse{
		ID:                 account.ID,
		UserID:             account.UserID,
		StripeAccountID:    account.StripeAccountID,
		AccountType:        account.AccountType,
		Country:            account.Country,
		Currency:           account.Currency,
		VerificationStatus: account.VerificationStatus,
		ChargesEnabled:     account.ChargesEnabled,
		PayoutsEnabled:     account.PayoutsEnabled,
		Metadata:           account.Metadata,
		CreatedAt:          account.CreatedAt,
		UpdatedAt:          account.UpdatedAt,
	}
}

// ToConnectedAccountListResponse converts a list of ConnectedAccount models to a response DTO
func ToConnectedAccountListResponse(accounts []*models.ConnectedAccount, total int64) ConnectedAccountListResponse {
	accountResponses := make([]ConnectedAccountResponse, len(accounts))
	for i, account := range accounts {
		accountResponses[i] = ToConnectedAccountResponse(account)
	}

	return ConnectedAccountListResponse{
		Accounts: accountResponses,
		Total:    total,
	}
}

// ToConnectedAccountResponseList converts a slice of ConnectedAccount models to response DTOs
func ToConnectedAccountResponseList(accounts []*models.ConnectedAccount) []ConnectedAccountResponse {
	accountResponses := make([]ConnectedAccountResponse, len(accounts))
	for i, account := range accounts {
		accountResponses[i] = ToConnectedAccountResponse(account)
	}
	return accountResponses
}
