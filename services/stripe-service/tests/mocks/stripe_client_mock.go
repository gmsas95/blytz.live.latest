package mocks

import (
	"context"

	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/services"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v84"
)

// MockStripeClient is a mock of the StripeClient
type MockStripeClient struct {
	mock.Mock
}

func (m *MockStripeClient) CreatePaymentIntent(ctx context.Context, params *services.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentIntent), args.Error(1)
}

func (m *MockStripeClient) CreateConnectedAccount(ctx context.Context, params *services.ConnectedAccountParams) (*stripe.Account, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Account), args.Error(1)
}

func (m *MockStripeClient) GetAccount(ctx context.Context, accountID string) (*stripe.Account, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Account), args.Error(1)
}

func (m *MockStripeClient) CreateTransfer(ctx context.Context, params *services.TransferParams) (*stripe.Transfer, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Transfer), args.Error(1)
}

func (m *MockStripeClient) GetTransfer(ctx context.Context, transferID string) (*stripe.Transfer, error) {
	args := m.Called(ctx, transferID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Transfer), args.Error(1)
}

func (m *MockStripeClient) ReverseTransfer(ctx context.Context, transferID string, amount int64) (*stripe.TransferReversal, error) {
	args := m.Called(ctx, transferID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.TransferReversal), args.Error(1)
}

func (m *MockStripeClient) CreateAccountLink(ctx context.Context, accountID, refreshURL, returnURL string) (*stripe.AccountLink, error) {
	args := m.Called(ctx, accountID, refreshURL, returnURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.AccountLink), args.Error(1)
}

func (m *MockStripeClient) CreateLoginLink(ctx context.Context, accountID string) (*stripe.LoginLink, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.LoginLink), args.Error(1)
}

func (m *MockStripeClient) ListTransfers(ctx context.Context, destination string, limit int64) ([]*stripe.Transfer, error) {
	args := m.Called(ctx, destination, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*stripe.Transfer), args.Error(1)
}

func (m *MockStripeClient) CreatePayout(ctx context.Context, params *services.PayoutParams) (*stripe.Payout, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Payout), args.Error(1)
}

func (m *MockStripeClient) GetPayout(ctx context.Context, payoutID string) (*stripe.Payout, error) {
	args := m.Called(ctx, payoutID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Payout), args.Error(1)
}

func (m *MockStripeClient) CancelPayout(ctx context.Context, payoutID string) (*stripe.Payout, error) {
	args := m.Called(ctx, payoutID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Payout), args.Error(1)
}

func (m *MockStripeClient) ListPayouts(ctx context.Context, destination string, limit int64) ([]*stripe.Payout, error) {
	args := m.Called(ctx, destination, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*stripe.Payout), args.Error(1)
}

func (m *MockStripeClient) GetBalance(ctx context.Context, accountID string) (*stripe.Balance, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Balance), args.Error(1)
}

func (m *MockStripeClient) ListBalanceTransactions(ctx context.Context, params *services.BalanceTransactionListParams) ([]*stripe.BalanceTransaction, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*stripe.BalanceTransaction), args.Error(1)
}

func (m *MockStripeClient) GetPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
	args := m.Called(ctx, paymentIntentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentIntent), args.Error(1)
}

func (m *MockStripeClient) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (*stripe.PaymentIntent, error) {
	args := m.Called(ctx, paymentIntentID, paymentMethodID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentIntent), args.Error(1)
}

func (m *MockStripeClient) CreateCustomPaymentMethod(ctx context.Context, params *services.CustomPaymentMethodParams) (*stripe.PaymentMethod, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentMethod), args.Error(1)
}

func (m *MockStripeClient) CreateMbWayPaymentMethod(ctx context.Context, params *services.MbWayPaymentMethodParams) (*stripe.PaymentMethod, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentMethod), args.Error(1)
}

func (m *MockStripeClient) CreateTWINTPaymentMethod(ctx context.Context, params *services.TWINTPaymentMethodParams) (*stripe.PaymentMethod, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentMethod), args.Error(1)
}

func (m *MockStripeClient) CreateCryptoPaymentMethod(ctx context.Context, params *services.CryptoPaymentMethodParams) (*stripe.PaymentMethod, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentMethod), args.Error(1)
}

func (m *MockStripeClient) GetPaymentMethodConfigs() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockStripeClient) ValidatePaymentMethodType(methodType, currency string) error {
	args := m.Called(methodType, currency)
	return args.Error(0)
}
