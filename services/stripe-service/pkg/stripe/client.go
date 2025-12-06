package stripe

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
	"github.com/stripe/stripe-go/v74/accountlink"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/loginlink"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/payout"
	"github.com/stripe/stripe-go/v74/refund"
	"github.com/stripe/stripe-go/v74/transfer"
	"github.com/stripe/stripe-go/v74/webhook"
)

// StripeClient wraps Stripe Go SDK for marketplace operations
type StripeClient struct {
	config      *Config
	backend     stripe.Backend
	accountID   string // Platform account ID
	connectID    string // Connect client ID
}

// Config holds Stripe configuration
type Config struct {
	SecretKey              string `mapstructure:"secret_key"`
	PublishableKey         string `mapstructure:"publishable_key"`
	WebhookSecret         string `mapstructure:"webhook_secret"`
	PlatformAccountID     string `mapstructure:"platform_account_id"`
	ConnectClientID        string `mapstructure:"connect_client_id"`
	ApplicationFeePercent  float64 `mapstructure:"application_fee_percent"`
	MinApplicationFee     int64   `mapstructure:"min_application_fee"`
	DefaultCurrency        string  `mapstructure:"default_currency"`
	WebhookEndpoint       string  `mapstructure:"webhook_endpoint"`
	SuccessURL           string  `mapstructure:"success_url"`
	CancelURL            string  `mapstructure:"cancel_url"`
}

// NewStripeClient creates a new Stripe client
func NewStripeClient(config *Config) *StripeClient {
	// Set Stripe API key
	stripe.Key = config.SecretKey
	
	// Create backend
	backend := &stripe.Backend{
		URL:        stripe.APIURL,
		HTTPClient: stripe.NewBackends(nil).API.HTTPClient,
	}

	return &StripeClient{
		config:    config,
		backend:   backend,
		accountID: config.PlatformAccountID,
		connectID:  config.ConnectClientID,
	}
}

// CreateConnectAccount creates a new Connect account for a seller
func (c *StripeClient) CreateConnectAccount(ctx context.Context, req *CreateConnectAccountRequest) (*stripe.Account, error) {
	params := &stripe.AccountParams{
		Type:        stripe.String(req.Type),
		Country:     stripe.String(req.Country),
		Email:       stripe.String(req.Email),
		BusinessType: stripe.String(req.BusinessType),
		Capabilities: map[string]*stripe.AccountCapabilityParams{
			"transfers":     {Status: stripe.String("active")},
			"card_payments": {Status: stripe.String("active")},
			"acss_debit":   {Status: stripe.String("inactive")}, // Canadian debit
			"affirm":        {Status: stripe.String("inactive")},
			"afterpay_clearpay": {Status: stripe.String("inactive")},
			"alipay":       {Status: stripe.String("inactive")},
			"au_becs_debit": {Status: stripe.String("inactive")},
			"bacs_debit":   {Status: stripe.String("inactive")},
			"bancontact":   {Status: stripe.String("inactive")},
			"boleto":       {Status: stripe.String("inactive")},
			"eps":          {Status: stripe.String("inactive")},
			"fpx":          {Status: stripe.String("active")}, // Malaysian
			"giropay":      {Status: stripe.String("inactive")},
			"grabpay":      {Status: stripe.String("active")}, // Malaysian
			"ideal":        {Status: stripe.String("inactive")},
			"klarna":       {Status: stripe.String("inactive")},
			"konbini":      {Status: stripe.String("inactive")},
			"link":         {Status: stripe.String("inactive")},
			"oxxo":         {Status: stripe.String("inactive")},
			"p24":          {Status: stripe.String("inactive")},
			"sepa_debit":   {Status: stripe.String("inactive")},
			"sofort":       {Status: stripe.String("inactive")},
			"us_bank_account": {Status: stripe.String("inactive")},
			"wechat_pay":    {Status: stripe.String("active")}, // Malaysian
		},
		Metadata: map[string]string{
			"seller_id":   req.SellerID,
			"platform":    "blytz",
			"created_at":  time.Now().Format(time.RFC3339),
		},
	}

	// Add business profile if provided
	if req.BusinessProfile != nil {
		params.BusinessProfile = &stripe.AccountBusinessProfileParams{
			Name:        stripe.String(req.BusinessProfile.Name),
			Website:     stripe.String(req.BusinessProfile.Website),
			Email:       stripe.String(req.BusinessProfile.Email),
			Phone:       stripe.String(req.BusinessProfile.Phone),
			Description: stripe.String(req.BusinessProfile.Description),
		}
	}

	// Add external account for payouts
	if req.ExternalAccount != nil {
		params.ExternalAccount = req.ExternalAccount
	}

	// Add company information if company type
	if req.Type == "company" && req.Company != nil {
		params.Company = &stripe.AccountCompanyParams{
			Name:         stripe.String(req.Company.Name),
			Phone:        stripe.String(req.Company.Phone),
			TaxID:        stripe.String(req.Company.TaxID),
			Email:        stripe.String(req.Company.Email),
			Address: &stripe.AddressParams{
				Country:  stripe.String(req.Company.Address.Country),
				City:     stripe.String(req.Company.Address.City),
				State:    stripe.String(req.Company.Address.State),
				Line1:    stripe.String(req.Company.Address.Line1),
				Line2:    stripe.String(req.Company.Address.Line2),
				PostalCode: stripe.String(req.Company.Address.PostalCode),
			},
			RegistrationNumber: stripe.String(req.Company.RegistrationNumber),
		}
	}

	// Add person information if individual type
	if req.Type == "individual" && req.Person != nil {
		params.Person = &stripe.AccountPersonParams{
			FirstName:     stripe.String(req.Person.FirstName),
			LastName:      stripe.String(req.Person.LastName),
			Email:         stripe.String(req.Person.Email),
			Phone:         stripe.String(req.Person.Phone),
			Relationship: &stripe.AccountPersonRelationshipParams{
				AccountOpener: stripe.Bool(true),
				Owner:         stripe.Bool(true),
				Director:      stripe.Bool(false),
				Executive:     stripe.Bool(false),
			},
			Address: &stripe.AddressParams{
				Country:  stripe.String(req.Person.Address.Country),
				City:     stripe.String(req.Person.Address.City),
				State:    stripe.String(req.Person.Address.State),
				Line1:    stripe.String(req.Person.Address.Line1),
				Line2:    stripe.String(req.Person.Address.Line2),
				PostalCode: stripe.String(req.Person.Address.PostalCode),
			},
		}
	}

	account, err := stripe.New(c.backend).Accounts.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Connect account: %w", err)
	}

	return account, nil
}

// CreateAccountLink creates an account link for onboarding
func (c *StripeClient) CreateAccountLink(ctx context.Context, req *CreateAccountLinkRequest) (*stripe.AccountLink, error) {
	params := &stripe.AccountLinkParams{
		Account:    stripe.String(req.AccountID),
		RefreshURL: stripe.String(req.RefreshURL),
		ReturnURL:  stripe.String(req.ReturnURL),
		Type:       stripe.String(req.Type),
	}

	accountLink, err := stripe.New(c.backend).AccountLinks.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account link: %w", err)
	}

	return accountLink, nil
}

// CreateLoginLink creates a login link for existing accounts
func (c *StripeClient) CreateLoginLink(ctx context.Context, req *CreateLoginLinkRequest) (*stripe.LoginLink, error) {
	params := &stripe.LoginLinkParams{
		Account: stripe.String(req.AccountID),
	}

	loginLink, err := stripe.New(c.backend).LoginLinks.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create login link: %w", err)
	}

	return loginLink, nil
}

// CreatePaymentIntent creates a payment intent for marketplace
func (c *StripeClient) CreatePaymentIntent(ctx context.Context, req *CreatePaymentIntentRequest) (*stripe.PaymentIntent, error) {
	// Calculate platform fee
	platformFee := int64(float64(req.Amount) * c.config.ApplicationFeePercent)
	if platformFee < c.config.MinApplicationFee {
		platformFee = c.config.MinApplicationFee
	}

	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(req.Amount),
		Currency:           stripe.String(req.Currency),
		ConfirmationMethod:  stripe.String("manual"),
		CaptureMethod:       stripe.String("automatic"),
		Description:        stripe.String(req.Description),
		Metadata:           req.Metadata,
	}

	// Add transfer data for marketplace
	if req.DestinationAccountID != "" {
		transferAmount := req.Amount - platformFee
		params.TransferData = &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(req.DestinationAccountID),
			Amount:      stripe.Int64(transferAmount),
		}
		params.ApplicationFeeAmount = stripe.Int64(platformFee)
	}

	// Add payment method types
	if len(req.PaymentMethodTypes) > 0 {
		params.PaymentMethodTypes = req.PaymentMethodTypes
	} else {
		// Default payment methods for Malaysia
		params.PaymentMethodTypes = stripe.StringSlice([]string{
			"card",
			"grabpay",
			"fpx",
			"wechat_pay",
		})
	}

	// Add customer ID if provided
	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}

	// Add receipt email
	if req.ReceiptEmail != "" {
		params.ReceiptEmail = stripe.String(req.ReceiptEmail)
	}

	paymentIntent, err := stripe.New(c.backend).PaymentIntents.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return paymentIntent, nil
}

// ConfirmPaymentIntent confirms a payment intent
func (c *StripeClient) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, req *ConfirmPaymentIntentRequest) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentConfirmParams{
		PaymentMethod: stripe.String(req.PaymentMethodID),
	}

	if req.ReturnURL != "" {
		params.ReturnURL = stripe.String(req.ReturnURL)
	}

	paymentIntent, err := stripe.New(c.backend).PaymentIntents.Confirm(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm payment intent: %w", err)
	}

	return paymentIntent, nil
}

// CreateCheckoutSession creates a checkout session for marketplace
func (c *StripeClient) CreateCheckoutSession(ctx context.Context, req *CreateCheckoutSessionRequest) (*stripe.CheckoutSession, error) {
	// Calculate platform fee
	totalAmount := req.LineItems[0].PriceData.UnitAmount
	platformFee := int64(float64(totalAmount) * c.config.ApplicationFeePercent)
	if platformFee < c.config.MinApplicationFee {
		platformFee = c.config.MinApplicationFee
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: req.PaymentMethodTypes,
		LineItems:          req.LineItems,
		Mode:               stripe.String(req.Mode),
		SuccessURL:         stripe.String(req.SuccessURL),
		CancelURL:          stripe.String(req.CancelURL),
		Metadata:           req.Metadata,
	}

	// Add customer information
	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}

	if req.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(req.CustomerEmail)
	}

	// Add transfer data for marketplace
	if req.DestinationAccountID != "" {
		transferAmount := totalAmount - platformFee
		params.TransferData = &stripe.CheckoutSessionTransferDataParams{
			Destination: stripe.String(req.DestinationAccountID),
			Amount:      stripe.Int64(transferAmount),
		}
		params.PaymentIntentData = &stripe.CheckoutSessionPaymentIntentDataParams{
			ApplicationFeeAmount: stripe.Int64(platformFee),
		}
	}

	// Add subscription data if needed
	if req.SubscriptionData != nil {
		params.SubscriptionData = req.SubscriptionData
	}

	// Add shipping information
	if req.ShippingAddressCollection != nil {
		params.ShippingAddressCollection = req.ShippingAddressCollection
	}

	if req.ShippingOptions != nil {
		params.ShippingOptions = req.ShippingOptions
	}

	session, err := stripe.New(c.backend).CheckoutSessions.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return session, nil
}

// CreateTransfer creates a transfer from platform to seller
func (c *StripeClient) CreateTransfer(ctx context.Context, req *CreateTransferRequest) (*stripe.Transfer, error) {
	params := &stripe.TransferParams{
		Amount:        stripe.Int64(req.Amount),
		Currency:      stripe.String(req.Currency),
		Destination:   stripe.String(req.DestinationAccountID),
		Description:   stripe.String(req.Description),
		Metadata:      req.Metadata,
	}

	// Add source transaction if linked to payment
	if req.SourceTransactionID != "" {
		params.SourceTransaction = &stripe.TransferSourceTransactionParams{
			Type:   stripe.String("payment_intent"),
			ID:     stripe.String(req.SourceTransactionID),
		}
	}

	transfer, err := stripe.New(c.backend).Transfers.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}

	return transfer, nil
}

// CreatePayout creates a payout from seller to bank account
func (c *StripeClient) CreatePayout(ctx context.Context, req *CreatePayoutRequest) (*stripe.Payout, error) {
	params := &stripe.PayoutParams{
		Amount:        stripe.Int64(req.Amount),
		Currency:      stripe.String(req.Currency),
		Destination:   stripe.String(req.DestinationAccountID),
		Description:   stripe.String(req.Description),
		Method:        stripe.String(req.Method),
		Metadata:      req.Metadata,
	}

	payout, err := stripe.New(c.backend).Payouts.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payout: %w", err)
	}

	return payout, nil
}

// CreateRefund creates a refund for a payment
func (c *StripeClient) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*stripe.Refund, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentIntentID),
		Amount:        stripe.Int64(req.Amount),
		Reason:        stripe.String(req.Reason),
		Metadata:      req.Metadata,
	}

	// Add transfer reversal for marketplace
	if req.ReverseTransfer && req.TransferID != "" {
		params.ReverseTransfer = stripe.Bool(true)
		params.Transfer = stripe.String(req.TransferID)
	}

	refund, err := stripe.New(c.backend).Refunds.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return refund, nil
}

// GetAccount retrieves a Connect account
func (c *StripeClient) GetAccount(ctx context.Context, accountID string) (*stripe.Account, error) {
	account, err := stripe.New(c.backend).Accounts.Get(accountID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Connect account: %w", err)
	}

	return account, nil
}

// GetPaymentIntent retrieves a payment intent
func (c *StripeClient) GetPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
	paymentIntent, err := stripe.New(c.backend).PaymentIntents.Get(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	return paymentIntent, nil
}

// GetCheckoutSession retrieves a checkout session
func (c *StripeClient) GetCheckoutSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
	session, err := stripe.New(c.backend).CheckoutSessions.Get(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkout session: %w", err)
	}

	return session, nil
}

// GetTransfer retrieves a transfer
func (c *StripeClient) GetTransfer(ctx context.Context, transferID string) (*stripe.Transfer, error) {
	transfer, err := stripe.New(c.backend).Transfers.Get(transferID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}

	return transfer, nil
}

// GetPayout retrieves a payout
func (c *StripeClient) GetPayout(ctx context.Context, payoutID string) (*stripe.Payout, error) {
	payout, err := stripe.New(c.backend).Payouts.Get(payoutID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payout: %w", err)
	}

	return payout, nil
}

// ListAccounts lists Connect accounts
func (c *StripeClient) ListAccounts(ctx context.Context, req *ListAccountsRequest) ([]*stripe.Account, string, error) {
	params := &stripe.AccountListParams{
		Limit: stripe.Int64(req.Limit),
	}

	if req.StartingAfter != "" {
		params.StartingAfter = stripe.String(req.StartingAfter)
	}

	if req.EndingBefore != "" {
		params.EndingBefore = stripe.String(req.EndingBefore)
	}

	// Filter by status
	if req.Status != "" {
		params.Filters = append(params.Filters, stripe.Filter{
			Key:   "status",
			Value: req.Status,
		})
	}

	// Filter by type
	if req.Type != "" {
		params.Filters = append(params.Filters, stripe.Filter{
			Key:   "type",
			Value: req.Type,
		})
	}

	iter := stripe.New(c.backend).Accounts.List(params)
	var accounts []*stripe.Account
	for iter.Next() {
		accounts = append(accounts, iter.Account())
	}

	if err := iter.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to list Connect accounts: %w", err)
	}

	var nextToken string
	if iter.Meta() != nil && iter.Meta().Next != "" {
		nextToken = iter.Meta().Next
	}

	return accounts, nextToken, nil
}

// ListTransfers lists transfers
func (c *StripeClient) ListTransfers(ctx context.Context, req *ListTransfersRequest) ([]*stripe.Transfer, string, error) {
	params := &stripe.TransferListParams{
		Limit: stripe.Int64(req.Limit),
	}

	if req.StartingAfter != "" {
		params.StartingAfter = stripe.String(req.StartingAfter)
	}

	if req.EndingBefore != "" {
		params.EndingBefore = stripe.String(req.EndingBefore)
	}

	// Filter by destination account
	if req.DestinationAccountID != "" {
		params.Filters = append(params.Filters, stripe.Filter{
			Key:   "destination",
			Value: req.DestinationAccountID,
		})
	}

	// Filter by date range
	if req.CreatedGTE != 0 {
		params.Filters = append(params.Filters, stripe.Filter{
			Key:   "created[gte]",
			Value: req.CreatedGTE,
		})
	}

	if req.CreatedLTE != 0 {
		params.Filters = append(params.Filters, stripe.Filter{
			Key:   "created[lte]",
			Value: req.CreatedLTE,
		})
	}

	iter := stripe.New(c.backend).Transfers.List(params)
	var transfers []*stripe.Transfer
	for iter.Next() {
		transfers = append(transfers, iter.Transfer())
	}

	if err := iter.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to list transfers: %w", err)
	}

	var nextToken string
	if iter.Meta() != nil && iter.Meta().Next != "" {
		nextToken = iter.Meta().Next
	}

	return transfers, nextToken, nil
}

// ConstructWebhookEvent constructs a webhook event from the raw payload
func (c *StripeClient) ConstructWebhookEvent(payload []byte, stripeSignature string) (stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, stripeSignature, c.config.WebhookSecret)
	if err != nil {
		return stripe.Event{}, fmt.Errorf("failed to construct webhook event: %w", err)
	}

	return event, nil
}

// CalculateApplicationFee calculates platform fee
func (c *StripeClient) CalculateApplicationFee(amount int64) int64 {
	fee := int64(float64(amount) * c.config.ApplicationFeePercent)
	if fee < c.config.MinApplicationFee {
		fee = c.config.MinApplicationFee
	}
	return fee
}

// CalculateSellerPayout calculates seller payout amount
func (c *StripeClient) CalculateSellerPayout(amount int64) int64 {
	fee := c.CalculateApplicationFee(amount)
	return amount - fee
}

// IsWebhookEventValid validates webhook event
func (c *StripeClient) IsWebhookEventValid(event stripe.Event, expectedType string) bool {
	// Check event type
	if event.Type != expectedType {
		return false
	}

	// Check created timestamp (should be within last 15 minutes)
	if time.Now().Unix()-event.Created > 900 {
		return false
	}

	return true
}

// GetSupportedPaymentMethods returns supported payment methods for Malaysia
func (c *StripeClient) GetSupportedPaymentMethods() []string {
	return []string{
		"card",
		"grabpay",
		"fpx",
		"wechat_pay",
	}
}

// GetSupportedCurrencies returns supported currencies
func (c *StripeClient) GetSupportedCurrencies() []string {
	return []string{
		"usd",  // US Dollar
		"eur",  // Euro
		"gbp",  // British Pound
		"jpy",  // Japanese Yen
		"sgd",  // Singapore Dollar
		"thb",  // Thai Baht
		"myr",  // Malaysian Ringgit
		"hkd",  // Hong Kong Dollar
		"aud",  // Australian Dollar
		"cad",  // Canadian Dollar
		"chf",  // Swiss Franc
	}
}

// ValidatePaymentMethod validates payment method
func (c *StripeClient) ValidatePaymentMethod(paymentMethod string) bool {
	supportedMethods := c.GetSupportedPaymentMethods()
	for _, method := range supportedMethods {
		if strings.EqualFold(method, paymentMethod) {
			return true
		}
	}
	return false
}

// ValidateCurrency validates currency
func (c *StripeClient) ValidateCurrency(currency string) bool {
	supportedCurrencies := c.GetSupportedCurrencies()
	for _, supported := range supportedCurrencies {
		if strings.EqualFold(supported, currency) {
			return true
		}
	}
	return false
}

// GetConfig returns the client configuration
func (c *StripeClient) GetConfig() *Config {
	return c.config
}