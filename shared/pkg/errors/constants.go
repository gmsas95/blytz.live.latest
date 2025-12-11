package errors

// Common error codes and messages
const (
	// Validation error constants
	ErrInvalidRequestBody     = "INVALID_REQUEST_BODY"
	ErrInvalidEmail          = "INVALID_EMAIL"
	ErrInvalidPassword       = "INVALID_PASSWORD"
	ErrInvalidPhoneNumber    = "INVALID_PHONE_NUMBER"
	ErrInvalidUUID           = "INVALID_UUID"
	ErrInvalidJSON           = "INVALID_JSON"
	ErrMissingRequiredField  = "MISSING_REQUIRED_FIELD"
	ErrFieldTooShort         = "FIELD_TOO_SHORT"
	ErrFieldTooLong          = "FIELD_TOO_LONG"
	ErrInvalidFormat         = "INVALID_FORMAT"
	ErrInvalidDateRange      = "INVALID_DATE_RANGE"

	// Authentication error constants
	ErrUnauthorized          = "UNAUTHORIZED"
	ErrInvalidCredentials    = "INVALID_CREDENTIALS"
	ErrTokenExpired          = "TOKEN_EXPIRED"
	ErrTokenInvalid          = "TOKEN_INVALID"
	ErrTokenMissing          = "TOKEN_MISSING"
	ErrInvalidRefreshToken   = "INVALID_REFRESH_TOKEN"
	ErrSessionExpired        = "SESSION_EXPIRED"

	// Authorization error constants
	ErrForbidden             = "FORBIDDEN"
	ErrInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrAccountSuspended      = "ACCOUNT_SUSPENDED"
	ErrAccountNotVerified    = "ACCOUNT_NOT_VERIFIED"

	// Not found error constants
	ErrUserNotFound          = "USER_NOT_FOUND"
	ErrProductNotFound       = "PRODUCT_NOT_FOUND"
	ErrOrderNotFound         = "ORDER_NOT_FOUND"
	ErrAuctionNotFound       = "AUCTION_NOT_FOUND"
	ErrCategoryNotFound      = "CATEGORY_NOT_FOUND"
	ErrResourceNotFound      = "RESOURCE_NOT_FOUND"

	// Conflict error constants
	ErrEmailAlreadyExists    = "EMAIL_ALREADY_EXISTS"
	ErrUserAlreadyExists     = "USER_ALREADY_EXISTS"
	ErrProductAlreadyExists  = "PRODUCT_ALREADY_EXISTS"
	ErrResourceConflict      = "RESOURCE_CONFLICT"
	ErrDuplicateEntry        = "DUPLICATE_ENTRY"

	// Business error constants
	ErrInsufficientStock     = "INSUFFICIENT_STOCK"
	ErrAuctionEnded          = "AUCTION_ENDED"
	ErrAuctionNotStarted     = "AUCTION_NOT_STARTED"
	ErrBidTooLow             = "BID_TOO_LOW"
	ErrInvalidBidAmount      = "INVALID_BID_AMOUNT"
	ErrPaymentRequired       = "PAYMENT_REQUIRED"
	ErrPaymentFailed         = "PAYMENT_FAILED"
	ErrOrderCannotBeCancelled = "ORDER_CANNOT_BE_CANCELLED"
	ErrInvalidOrderStatus    = "INVALID_ORDER_STATUS"

	// External service error constants
	ErrStripeError           = "STRIPE_ERROR"
	ErrEmailServiceError     = "EMAIL_SERVICE_ERROR"
	ErrSMSServiceError       = "SMS_SERVICE_ERROR"
	ErrLiveKitError          = "LIVEKIT_ERROR"
	ErrPaymentGatewayError   = "PAYMENT_GATEWAY_ERROR"

	// Database error constants
	ErrDatabaseConnection    = "DATABASE_CONNECTION"
	ErrDatabaseQuery         = "DATABASE_QUERY"
	ErrDatabaseTransaction   = "DATABASE_TRANSACTION"
	ErrRecordNotFound        = "RECORD_NOT_FOUND"
	ErrDuplicateKey          = "DUPLICATE_KEY"

	// Network error constants
	ErrNetworkTimeout        = "NETWORK_TIMEOUT"
	ErrServiceUnavailable    = "SERVICE_UNAVAILABLE"
	ErrConnectionFailed      = "CONNECTION_FAILED"
	ErrRateLimitExceeded     = "RATE_LIMIT_EXCEEDED"

	// Internal error constants
	ErrInternalServer        = "INTERNAL_SERVER_ERROR"
	ErrConfiguration         = "CONFIGURATION_ERROR"
	ErrFileOperation         = "FILE_OPERATION_ERROR"
	ErrSerialization         = "SERIALIZATION_ERROR"
	ErrParsing               = "PARSING_ERROR"
)

// Error message templates
var (
	// Validation messages
	MsgInvalidRequestBody     = "Invalid request body format"
	MsgInvalidEmail          = "Invalid email address format"
	MsgInvalidPassword       = "Password must be at least 8 characters long"
	MsgInvalidPhoneNumber    = "Invalid phone number format"
	MsgInvalidUUID           = "Invalid UUID format"
	MsgInvalidJSON           = "Invalid JSON format"
	MsgMissingRequiredField  = "Required field is missing"
	MsgFieldTooShort         = "Field value is too short"
	MsgFieldTooLong          = "Field value is too long"
	MsgInvalidFormat         = "Invalid field format"
	MsgInvalidDateRange      = "Invalid date range"

	// Authentication messages
	MsgUnauthorized          = "Authentication required"
	MsgInvalidCredentials    = "Invalid email or password"
	MsgTokenExpired          = "Authentication token has expired"
	MsgTokenInvalid          = "Invalid authentication token"
	MsgTokenMissing          = "Authentication token is required"
	MsgInvalidRefreshToken   = "Invalid refresh token"
	MsgSessionExpired        = "User session has expired"

	// Authorization messages
	MsgForbidden             = "Access to this resource is forbidden"
	MsgInsufficientPermissions = "You don't have permission to perform this action"
	MsgAccountSuspended      = "Your account has been suspended"
	MsgAccountNotVerified    = "Your account has not been verified"

	// Not found messages
	MsgUserNotFound          = "User not found"
	MsgProductNotFound       = "Product not found"
	MsgOrderNotFound         = "Order not found"
	MsgAuctionNotFound       = "Auction not found"
	MsgCategoryNotFound      = "Category not found"
	MsgResourceNotFound      = "Resource not found"

	// Conflict messages
	MsgEmailAlreadyExists    = "Email address is already registered"
	MsgUserAlreadyExists     = "User already exists"
	MsgProductAlreadyExists  = "Product already exists"
	MsgResourceConflict      = "Resource conflict detected"
	MsgDuplicateEntry        = "Duplicate entry detected"

	// Business messages
	MsgInsufficientStock     = "Product is out of stock"
	MsgAuctionEnded          = "Auction has already ended"
	MsgAuctionNotStarted     = "Auction has not started yet"
	MsgBidTooLow             = "Bid amount is too low"
	MsgInvalidBidAmount      = "Invalid bid amount"
	MsgPaymentRequired       = "Payment is required for this action"
	MsgPaymentFailed         = "Payment processing failed"
	MsgOrderCannotBeCancelled = "Order cannot be cancelled in current status"
	MsgInvalidOrderStatus    = "Invalid order status for this operation"

	// External service messages
	MsgStripeError           = "Payment service error"
	MsgEmailServiceError     = "Email service error"
	MsgSMSServiceError       = "SMS service error"
	MsgLiveKitError          = "Video streaming service error"
	MsgPaymentGatewayError   = "Payment gateway error"

	// Database messages
	MsgDatabaseConnection    = "Database connection error"
	MsgDatabaseQuery         = "Database query error"
	MsgDatabaseTransaction   = "Database transaction error"
	MsgRecordNotFound        = "Record not found in database"
	MsgDuplicateKey          = "Duplicate key constraint violation"

	// Network messages
	MsgNetworkTimeout        = "Network request timeout"
	MsgServiceUnavailable    = "Service is temporarily unavailable"
	MsgConnectionFailed      = "Failed to establish connection"
	MsgRateLimitExceeded     = "Rate limit exceeded"

	// Internal messages
	MsgInternalServer        = "Internal server error"
	MsgConfiguration         = "Configuration error"
	MsgFileOperation         = "File operation error"
	MsgSerialization         = "Data serialization error"
	MsgParsing               = "Data parsing error"
)

// Predefined common errors for quick usage
var (
	ErrInvalidRequestBodyError     = NewValidationError(ErrInvalidRequestBody, MsgInvalidRequestBody)
	ErrInvalidEmailError          = NewValidationError(ErrInvalidEmail, MsgInvalidEmail)
	ErrInvalidPasswordError       = NewValidationError(ErrInvalidPassword, MsgInvalidPassword)
	ErrUnauthorizedError          = NewAuthenticationError(ErrUnauthorized, MsgUnauthorized)
	ErrInvalidCredentialsError    = NewAuthenticationError(ErrInvalidCredentials, MsgInvalidCredentials)
	ErrTokenExpiredError          = NewAuthenticationError(ErrTokenExpired, MsgTokenExpired)
	ErrForbiddenError             = NewAuthorizationError(ErrForbidden, MsgForbidden)
	ErrUserNotFoundError          = NewNotFoundError(ErrUserNotFound, MsgUserNotFound)
	ErrProductNotFoundError       = NewNotFoundError(ErrProductNotFound, MsgProductNotFound)
	ErrEmailAlreadyExistsError    = NewConflictError(ErrEmailAlreadyExists, MsgEmailAlreadyExists)
	ErrInsufficientStockError     = NewBusinessError(ErrInsufficientStock, MsgInsufficientStock)
	ErrAuctionEndedError          = NewBusinessError(ErrAuctionEnded, MsgAuctionEnded)
	ErrPaymentFailedError         = NewBusinessError(ErrPaymentFailed, MsgPaymentFailed)
	ErrStripeErrorError           = NewExternalServiceError(ErrStripeError, MsgStripeError, "")
	ErrDatabaseConnectionError    = NewDatabaseError(ErrDatabaseConnection, MsgDatabaseConnection)
	ErrNetworkTimeoutError        = NewNetworkError(ErrNetworkTimeout, MsgNetworkTimeout)
	ErrInternalServerError        = NewInternalError(ErrInternalServer, MsgInternalServer)
)