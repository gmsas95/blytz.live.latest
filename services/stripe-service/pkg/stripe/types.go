package stripe

import (
	"time"

	"github.com/stripe/stripe-go/v74"
)

// Connect account request types
type CreateConnectAccountRequest struct {
	SellerID       string                `json:"seller_id" validate:"required,uuid"`
	Type          string                `json:"type" validate:"required,oneof=express standard custom"`
	Country       string                `json:"country" validate:"required,len=2"`
	Email         string                `json:"email" validate:"required,email"`
	BusinessType  string                `json:"business_type" validate:"required,oneof=individual company"`
	Capabilities  []string              `json:"capabilities"`
	BusinessProfile *BusinessProfileRequest `json:"business_profile"`
	ExternalAccount *stripe.ExternalAccountParams `json:"external_account"`
	Person         *PersonRequest        `json:"person"`
	Company        *CompanyRequest       `json:"company"`
	Preferences    *AccountPreferencesRequest `json:"preferences"`
	Tools          *AccountToolsRequest `json:"tools"`
	Metadata       map[string]string      `json:"metadata"`
}

type BusinessProfileRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Website     string `json:"website" validate:"omitempty,url"`
	Email       string `json:"email" validate:"omitempty,email"`
	Phone       string `json:"phone" validate:"omitempty,e164"`
	Description string `json:"description" validate:"omitempty,max=500"`
	SupportPhone string `json:"support_phone" validate:"omitempty,e164"`
	SupportEmail string `json:"support_email" validate:"omitempty,email"`
}

type PersonRequest struct {
	FirstName string          `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string          `json:"last_name" validate:"required,min=2,max=50"`
	Email     string          `json:"email" validate:"required,email"`
	Phone     string          `json:"phone" validate:"omitempty,e164"`
	DOB       *DOBRequest    `json:"dob"`
	Address   *AddressRequest `json:"address"`
	Relationship *PersonRelationshipRequest `json:"relationship"`
	IDNumber  string          `json:"id_number"`
	SSNLast4 string          `json:"ssn_last_4"`
	Verification *PersonVerificationRequest `json:"verification"`
}

type DOBRequest struct {
	Day   int `json:"day" validate:"required,min=1,max=31"`
	Month int `json:"month" validate:"required,min=1,max=12"`
	Year  int `json:"year" validate:"required,min=1900,max=2006"`
}

type AddressRequest struct {
	Country     string `json:"country" validate:"required,len=2"`
	City        string `json:"city" validate:"required,min=2,max=100"`
	State       string `json:"state" validate:"required,min=2,max=100"`
	Line1       string `json:"line1" validate:"required,min=5,max=200"`
	Line2       string `json:"line2" validate:"omitempty,max=200"`
	PostalCode  string `json:"postal_code" validate:"required,min=3,max=20"`
}

type PersonRelationshipRequest struct {
	AccountOpener bool `json:"account_opener"`
	Director      bool `json:"director"`
	Owner         bool `json:"owner"`
	Executive     bool `json:"executive"`
	Representative bool `json:"representative"`
}

type PersonVerificationRequest struct {
	Document *VerificationDocumentRequest `json:"document"`
	AdditionalVerification *AdditionalVerificationRequest `json:"additional_verification"`
}

type VerificationDocumentRequest struct {
	Front string `json:"front"`
	Back  string `json:"back"`
}

type AdditionalVerificationRequest struct {
	Document *VerificationDocumentRequest `json:"document"`
}

type CompanyRequest struct {
	Name              string          `json:"name" validate:"required,min=2,max=100"`
	Phone             string          `json:"phone" validate:"omitempty,e164"`
	Email             string          `json:"email" validate:"omitempty,email"`
	TaxID             string          `json:"tax_id"`
	RegistrationNumber string          `json:"registration_number"`
	Address           *AddressRequest `json:"address"`
	RegistrationType  string          `json:"registration_type"`
}

type AccountPreferencesRequest struct {
	CardPayments *CardPaymentsRequest `json:"card_payments"`
	ACHCreditTransfer *ACHCreditTransferRequest `json:"ach_credit_transfer"`
}

type CardPaymentsRequest struct {
	StatementDescriptorPrefix string `json:"statement_descriptor_prefix"`
	DeclineOnAVCFailure string `json:"decline_on_avc_failure"`
}

type ACHCreditTransferRequest struct {
	AccountHolderName string `json:"account_holder_name"`
}

type AccountToolsRequest struct {
	CardIssuing *CardIssuingRequest `json:"card_issuing"`
	FinancialConnections *FinancialConnectionsRequest `json:"financial_connections"`
}

type CardIssuingRequest struct {
	StripeDashboard *CardIssuingStripeDashboardRequest `json:"stripe_dashboard"`
}

type CardIssuingStripeDashboardRequest struct {
	Features []string `json:"features"`
}

type FinancialConnectionsRequest struct {
	Permissions []string `json:"permissions"`
}

// Account link request
type CreateAccountLinkRequest struct {
	AccountID   string `json:"account_id" validate:"required"`
	RefreshURL  string `json:"refresh_url" validate:"required,url"`
	ReturnURL   string `json:"return_url" validate:"required,url"`
	Type        string `json:"type" validate:"required,oneof=account_onboarding account_update custom_account_verification"`
	Collection  string `json:"collection" validate:"omitempty,oneof=current_future currently_due"`
}

// Login link request
type CreateLoginLinkRequest struct {
	AccountID string `json:"account_id" validate:"required"`
	RedirectURL string `json:"redirect_url" validate:"omitempty,url"`
}

// Payment intent request types
type CreatePaymentIntentRequest struct {
	Amount                int64                    `json:"amount" validate:"required,min=50"`
	Currency              string                    `json:"currency" validate:"required,len=3"`
	CustomerID            string                    `json:"customer_id"`
	PaymentMethodID        string                    `json:"payment_method_id"`
	PaymentMethodTypes     []*string                  `json:"payment_method_types"`
	Description           string                    `json:"description" validate:"omitempty,max=500"`
	ConfirmationMethod     string                    `json:"confirmation_method" validate:"omitempty,oneof=manual automatic"`
	CaptureMethod          string                    `json:"capture_method" validate:"omitempty,oneof=automatic manual"`
	SetupFutureUsage       string                    `json:"setup_future_usage" validate:"omitempty,oneof=on_session off_session"`
	OffSession            bool                      `json:"off_session"`
	DestinationAccountID  string                    `json:"destination_account_id"`
	TransferGroup         string                    `json:"transfer_group"`
	ApplicationFeeAmount  int64                     `json:"application_fee_amount" validate:"omitempty,min=0"`
	Metadata              map[string]string          `json:"metadata"`
	ReceiptEmail          string                    `json:"receipt_email" validate:"omitempty,email"`
	Shipping              *ShippingDetailsRequest    `json:"shipping"`
	ReturnURL             string                    `json:"return_url" validate:"omitempty,url"`
	PaymentMethodOptions  *PaymentMethodOptionsRequest `json:"payment_method_options"`
	PaymentMethodData     *PaymentMethodDataRequest  `json:"payment_method_data"`
}

type ShippingDetailsRequest struct {
	Name    string                 `json:"name"`
	Address *AddressRequest         `json:"address"`
	Carrier string                 `json:"carrier"`
	TrackingNumber string            `json:"tracking_number"`
	Phone   string                 `json:"phone" validate:"omitempty,e164"`
}

type PaymentMethodOptionsRequest struct {
	Card       *CardPaymentMethodOptionsRequest       `json:"card"`
	GrabPay    *GrabPayPaymentMethodOptionsRequest   `json:"grabpay"`
	FPX        *FPXPaymentMethodOptionsRequest       `json:"fpx"`
	WechatPay  *WechatPayPaymentMethodOptionsRequest `json:"wechat_pay"`
}

type CardPaymentMethodOptionsRequest struct {
	RequestThreeDSecure string `json:"request_three_d_secure" validate:"omitempty,oneof=automatic any"`
	Network           string `json:"network" validate:"omitempty,oneof=amex cartes_bancaires diners discover interac jcb mastercard unionpay visa"`
	Installments      *CardInstallmentsRequest `json:"installments"`
}

type CardInstallmentsRequest struct {
	Plan        string `json:"plan"`
	Count        int64  `json:"count"`
	Interval     string `json:"interval"`
	InstanceID   string `json:"instance_id"`
}

type GrabPayPaymentMethodOptionsRequest struct {
	AutoCapture bool `json:"auto_capture"`
}

type FPXPaymentMethodOptionsRequest struct {
	Bank         string `json:"bank"`
	AccountHolder string `json:"account_holder"`
}

type WechatPayPaymentMethodOptionsRequest struct {
	Client        string `json:"client" validate:"omitempty,oneof=web mobile weapp"`
	AppID         string `json:"app_id"`
	Scene         string `json:"scene" validate:"omitempty"`
	SetupFuture   bool   `json:"setup_future"`
}

type PaymentMethodDataRequest struct {
	Type      string                    `json:"type" validate:"required"`
	Card      *CardPaymentMethodDataRequest      `json:"card"`
	GrabPay   *GrabPayPaymentMethodDataRequest   `json:"grabpay"`
	FPX       *FPXPaymentMethodDataRequest       `json:"fpx"`
	WechatPay *WechatPayPaymentMethodDataRequest `json:"wechat_pay"`
}

type CardPaymentMethodDataRequest struct {
	Number      string `json:"number" validate:"required,len=16"`
	ExpMonth    int64  `json:"exp_month" validate:"required,min=1,max=12"`
	ExpYear     int64  `json:"exp_year" validate:"required,min=2020,max=2100"`
	CVC         string `json:"cvc" validate:"required,len=3,4"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Address     *AddressRequest `json:"address"`
	Brand       string `json:"brand"`
	Fingerprint string `json:"fingerprint"`
}

type GrabPayPaymentMethodDataRequest struct {
	Country string `json:"country"`
}

type FPXPaymentMethodDataRequest struct {
	Bank         string `json:"bank" validate:"required"`
	AccountHolder string `json:"account_holder" validate:"required"`
}

type WechatPayPaymentMethodDataRequest struct {
	AppID       string `json:"app_id" validate:"required"`
	ClientType   string `json:"client_type" validate:"required,oneof=web mobile weapp"`
	Scene        string `json:"scene"`
	Package      string `json:"package"`
	Sign         string `json:"sign"`
	Timestamp    int64  `json:"timestamp"`
	NonceStr     string `json:"nonce_str"`
}

// Confirm payment intent request
type ConfirmPaymentIntentRequest struct {
	PaymentMethodID string `json:"payment_method_id"`
	ReturnURL       string `json:"return_url" validate:"omitempty,url"`
	PaymentMethodOptions *PaymentMethodOptionsRequest `json:"payment_method_options"`
}

// Checkout session request types
type CreateCheckoutSessionRequest struct {
	LineItems           []*stripe.CheckoutSessionLineItemParams `json:"line_items" validate:"required,min=1"`
	Mode                string                              `json:"mode" validate:"required,oneof=payment subscription setup"`
	CustomerID          string                              `json:"customer_id"`
	CustomerEmail       string                              `json:"customer_email" validate:"omitempty,email"`
	PaymentMethodTypes  []*string                           `json:"payment_method_types"`
	SuccessURL          string                              `json:"success_url" validate:"required,url"`
	CancelURL           string                              `json:"cancel_url" validate:"required,url"`
	DestinationAccountID string                              `json:"destination_account_id"`
	ApplicationFeeAmount int64                               `json:"application_fee_amount" validate:"omitempty,min=0"`
	Metadata            map[string]string                    `json:"metadata"`
	SubscriptionData    *stripe.CheckoutSessionSubscriptionDataParams `json:"subscription_data"`
	PaymentIntentData    *stripe.CheckoutSessionPaymentIntentDataParams `json:"payment_intent_data"`
	ShippingAddressCollection *stripe.CheckoutSessionShippingAddressCollectionParams `json:"shipping_address_collection"`
	ShippingOptions     []*stripe.CheckoutSessionShippingOptionParams `json:"shipping_options"`
	AllowPromotionCodes bool `json:"allow_promotion_codes"`
	BillingAddressCollection *stripe.CheckoutSessionBillingAddressCollectionParams `json:"billing_address_collection"`
	ClientReferenceID   string `json:"client_reference_id"`
	Locale              string `json:"locale"`
	SubmitType          string `json:"submit_type" validate:"omitempty,oneof=auto book pay"`
	ExpiresAt           time.Time `json:"expires_at"`
}

// Transfer request types
type CreateTransferRequest struct {
	Amount               int64             `json:"amount" validate:"required,min=100"`
	Currency             string            `json:"currency" validate:"required,len=3"`
	DestinationAccountID string            `json:"destination_account_id" validate:"required"`
	TransferGroup        string            `json:"transfer_group"`
	IDEMPOTENCYKey      string            `json:"idempotency_key"`
	Description          string            `json:"description" validate:"omitempty,max=500"`
	Metadata             map[string]string `json:"metadata"`
	SourceTransactionID   string            `json:"source_transaction_id"`
	SourceType           string            `json:"source_type" validate:"omitempty,oneof=card alipay ach_credit_transfer"`
}

// Payout request types
type CreatePayoutRequest struct {
	Amount             int64             `json:"amount" validate:"required,min=100"`
	Currency           string            `json:"currency" validate:"required,len=3"`
	DestinationAccountID string            `json:"destination_account_id"`
	Method             string            `json:"method" validate:"omitempty,oneof=instant standard"`
	IDEMPOTENCYKey      string            `json:"idempotency_key"`
	Description         string            `json:"description" validate:"omitempty,max=500"`
	Metadata            map[string]string `json:"metadata"`
	StatementIdentifier string            `json:"statement_identifier"`
}

// Refund request types
type CreateRefundRequest struct {
	PaymentIntentID string            `json:"payment_intent_id" validate:"required"`
	Amount         int64             `json:"amount" validate:"omitempty,min=50"`
	Currency       string            `json:"currency" validate:"omitempty,len=3"`
	Reason         string            `json:"reason" validate:"omitempty,oneof=duplicate fraudulent requested_by_customer expired_lost_card expired_uncaptured_failed_charge abandoned disputed_subscription unrecoverable_prevention risk_assessment"`
	IDEMPOTENCYKey string            `json:"idempotency_key"`
	Description    string            `json:"description" validate:"omitempty,max=500"`
	ReverseTransfer bool              `json:"reverse_transfer"`
	TransferID     string            `json:"transfer_id"`
	Metadata       map[string]string `json:"metadata"`
}

// List request types
type ListAccountsRequest struct {
	Limit        int64  `json:"limit" validate:"omitempty,min=1,max=100"`
	StartingAfter string  `json:"starting_after"`
	EndingBefore  string  `json:"ending_before"`
	Status       string  `json:"status" validate:"omitempty,oneof=new active rejected pending restricted"`
	Type         string  `json:"type" validate:"omitempty,oneof=express standard custom"`
}

type ListTransfersRequest struct {
	Limit                 int64  `json:"limit" validate:"omitempty,min=1,max=100"`
	StartingAfter         string  `json:"starting_after"`
	EndingBefore         string  `json:"ending_before"`
	DestinationAccountID  string  `json:"destination_account_id"`
	TransferGroup        string  `json:"transfer_group"`
	CreatedGTE           int64  `json:"created_gte"`
	CreatedLTE           int64  `json:"created_lte"`
}

// Response types
type ConnectAccountResponse struct {
	ID                   string                  `json:"id"`
	Type                 string                  `json:"type"`
	Country              string                  `json:"country"`
	Email                string                  `json:"email"`
	BusinessType         string                  `json:"business_type"`
	Status               string                  `json:"status"`
	ChargesEnabled       bool                    `json:"charges_enabled"`
	PayoutsEnabled       bool                    `json:"payouts_enabled"`
	Requirements         *stripe.Requirements     `json:"requirements"`
	Capabilities         map[string]interface{}    `json:"capabilities"`
	BusinessProfile      *stripe.BusinessProfile  `json:"business_profile"`
	Settings            *stripe.AccountSettings  `json:"settings"`
	ExternalAccounts     []*stripe.ExternalAccount `json:"external_accounts"`
	Links                map[string]string        `json:"links"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
}

type AccountLinkResponse struct {
	ID          string     `json:"id"`
	Object       string     `json:"object"`
	Created      int64      `json:"created"`
	ExpiresAt    int64      `json:"expires_at"`
	URL          string     `json:"url"`
	RefreshURL   string     `json:"refresh_url"`
	ReturnURL    string     `json:"return_url"`
	Type         string     `json:"type"`
	AccountID    string     `json:"account_id"`
}

type LoginLinkResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Created   int64  `json:"created"`
	URL       string `json:"url"`
	AccountID string `json:"account_id"`
}

type PaymentIntentResponse struct {
	ID                    string                     `json:"id"`
	Object                string                     `json:"object"`
	Amount                int64                      `json:"amount"`
	Currency              string                     `json:"currency"`
	Status                string                     `json:"status"`
	Description           string                     `json:"description"`
	CreatedAt             int64                      `json:"created_at"`
	ConfirmationMethod     string                     `json:"confirmation_method"`
	CaptureMethod         string                     `json:"capture_method"`
	SetupFutureUsage      string                     `json:"setup_future_usage"`
	PaymentMethodID       string                     `json:"payment_method_id"`
	PaymentMethod         *stripe.PaymentMethod      `json:"payment_method"`
	ChargeID              string                     `json:"charge_id"`
	ClientSecret          string                     `json:"client_secret"`
	LatestCharge          *stripe.Charge            `json:"latest_charge"`
	NextAction            *stripe.NextAction         `json:"next_action"`
	ApplicationFeeAmount  int64                      `json:"application_fee_amount"`
	TransferData          *stripe.TransferData       `json:"transfer_data"`
	ReceiptEmail         string                     `json:"receipt_email"`
	Shipping             *stripe.ShippingDetails     `json:"shipping"`
	Metadata             map[string]string          `json:"metadata"`
	CancellationReason    string                     `json:"cancellation_reason"`
	Review               *stripe.Review            `json:"review"`
	ProcessURL           string                     `json:"process_url"`
}

type CheckoutSessionResponse struct {
	ID                    string                        `json:"id"`
	Object                string                        `json:"object"`
	Created               int64                         `json:"created"`
	ExpiresAt             int64                         `json:"expires_at"`
	URL                   string                        `json:"url"`
	Mode                  string                        `json:"mode"`
	Status                string                        `json:"status"`
	CustomerID            string                        `json:"customer"`
	CustomerEmail         string                        `json:"customer_email"`
	LineItems             []*stripe.LineItem            `json:"line_items"`
	PaymentIntentID       string                        `json:"payment_intent_id"`
	PaymentStatus         string                        `json:"payment_status"`
	SuccessURL            string                        `json:"success_url"`
	CancelURL             string                        `json:"cancel_url"`
	SubmitType            string                        `json:"submit_type"`
	BillingAddressCollection *stripe.CheckoutSessionBillingAddressCollectionParams `json:"billing_address_collection"`
	ShippingAddressCollection *stripe.CheckoutSessionShippingAddressCollectionParams `json:"shipping_address_collection"`
	ShippingOptions       []*stripe.CheckoutSessionShippingOptionParams `json:"shipping_options"`
	PaymentMethodTypes    []string                      `json:"payment_method_types"`
	Metadata              map[string]string             `json:"metadata"`
	Locale                string                        `json:"locale"`
	ClientReferenceID     string                        `json:"client_reference_id"`
	CancelURL             string                        `json:"cancel_url"`
	TotalDetails          *stripe.TotalDetails         `json:"total_details"`
}

type TransferResponse struct {
	ID                     string                    `json:"id"`
	Object                 string                    `json:"object"`
	Amount                 int64                     `json:"amount"`
	Currency               string                    `json:"currency"`
	Destination            string                    `json:"destination"`
	DestinationPayment     string                    `json:"destination_payment"`
	TransferGroup          string                    `json:"transfer_group"`
	Livemode               bool                      `json:"livemode"`
	ReversalID            string                    `json:"reversal_id"`
	Reversals              []*stripe.Reversal        `json:"reversals"`
	Description           string                    `json:"description"`
	Metadata               map[string]string         `json:"metadata"`
	SourceTransaction      string                    `json:"source_transaction"`
	SourceType             string                    `json:"source_type"`
	TransferReversal      bool                      `json:"transfer_reversal"`
	SourceID              string                    `json:"source_id"`
	SourceType            string                    `json:"source_type"`
	Mode                  string                    `json:"mode"`
	Created               int64                     `json:"created"`
	ArrivalDate           int64                     `json:"arrival_date"`
	Date                  int64                     `json:"date"`
	AvailableOn           int64                     `json:"available_on"`
}

type PayoutResponse struct {
	ID                 string                `json:"id"`
	Object             string                `json:"object"`
	Amount             int64                 `json:"amount"`
	Currency           string                `json:"currency"`
	Destination        string                `json:"destination"`
	DestinationBalance string                `json:"destination_balance"`
	Livemode          bool                  `json:"livemode"`
	Method            string                `json:"method"`
	Status            string                `json:"status"`
	Type              string                `json:"type"`
	ArrivalDate       int64                 `json:"arrival_date"`
	Created           int64                 `json:"created"`
	StatementDescriptor string              `json:"statement_descriptor"`
	FailureCode       string                `json:"failure_code"`
	FailureMessage    string                `json:"failure_message"`
	OriginalPayout   *stripe.Payout        `json:"original_payout"`
	ReversedBy       *stripe.Payout        `json:"reversed_by"`
	Metadata          map[string]string     `json:"metadata"`
}

type RefundResponse struct {
	ID             string                    `json:"id"`
	Object         string                    `json:"object"`
	Amount         int64                     `json:"amount"`
	Currency       string                    `json:"currency"`
	PaymentIntent  string                    `json:"payment_intent"`
	ChargeID       string                    `json:"charge"`
	Reason         string                    `json:"reason"`
	ReceiptNumber  string                    `json:"receipt_number"`
	RefundedAmount int64                     `json:"refunded_amount"`
	Status         string                    `json:"status"`
	Description    string                    `json:"description"`
	Metadata       map[string]string         `json:"metadata"`
	Created        int64                     `json:"created"`
	BalanceTransaction *stripe.BalanceTransaction `json:"balance_transaction"`
}

// Error response type
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Code      string `json:"code"`
	Type      string `json:"type"`
	Param      string `json:"param"`
	StripeCode string `json:"stripe_code"`
}

// Success response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Timestamp int64    `json:"timestamp"`
}

// Pagination response
type PaginatedResponse struct {
	Success     bool        `json:"success"`
	Data        interface{} `json:"data"`
	Message     string      `json:"message"`
	Timestamp   int64       `json:"timestamp"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	CurrentPage int64  `json:"current_page"`
	TotalPages  int64  `json:"total_pages"`
	TotalItems  int64  `json:"total_items"`
	ItemsPerPage int64  `json:"items_per_page"`
	HasNextPage bool   `json:"has_next_page"`
	HasPrevPage bool   `json:"has_prev_page"`
}