package services_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/services"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v84"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	return db, mock
}

func TestConnectedAccountService_CreateConnectedAccount(t *testing.T) {
	db, sqlMock := setupTestDB(t)
	logger := zaptest.NewLogger(t)
	cfg := &config.StripeConfig{}

	mockStripeClient := new(mocks.MockStripeClient)

	service := services.NewConnectedAccountService(db, mockStripeClient, cfg, logger)

	req := &services.CreateConnectedAccountRequest{
		UserID:   "user_123",
		Email:    "test@example.com",
		Country:  "US",
		Username: "testuser",
	}

	expectedStripeAccount := &stripe.Account{
		ID:             "acct_123",
		Type:           stripe.AccountTypeExpress,
		ChargesEnabled: false,
		PayoutsEnabled: false,
		Capabilities:   &stripe.AccountCapabilities{},
	}

	// 1. Check existing account (Expect query)
	// Query: SELECT * FROM "connected_accounts" WHERE user_id = $1 ...
	// Query: SELECT * FROM "connected_accounts" WHERE user_id = $1 ... LIMIT $2
	sqlMock.ExpectQuery(`SELECT \* FROM "connected_accounts" WHERE user_id = \$1 ORDER BY "connected_accounts"."id" LIMIT \$2`).
		WithArgs("user_123", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// 2. Stripe API call
	mockStripeClient.On("CreateConnectedAccount",
		mock.Anything,
		mock.MatchedBy(func(params *services.ConnectedAccountParams) bool {
			return params.UserID == req.UserID && params.Email == req.Email
		}),
	).Return(expectedStripeAccount, nil)

	// 3. Create account in DB (Expect insert)
	// INSERT INTO "connected_accounts" ...
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(`INSERT INTO "connected_accounts"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("uuid-123"))
	sqlMock.ExpectCommit()

	account, err := service.CreateConnectedAccount(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, account)
	assert.Equal(t, "acct_123", account.StripeAccountID)
	assert.Equal(t, "user_123", account.UserID)

	// Verify expectations
	mockStripeClient.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestConnectedAccountService_GetAccountBalance(t *testing.T) {
	db, sqlMock := setupTestDB(t)
	logger := zaptest.NewLogger(t)
	cfg := &config.StripeConfig{}

	mockStripeClient := new(mocks.MockStripeClient)
	service := services.NewConnectedAccountService(db, mockStripeClient, cfg, logger)

	accountID := "account_uuid_123"
	stripeAccountID := "acct_stripe_123"

	// Mock GetConnectedAccount query
	sqlMock.ExpectQuery(`SELECT \* FROM "connected_accounts" WHERE id = \$1 ORDER BY "connected_accounts"."id" LIMIT \$2`).
		WithArgs(accountID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "stripe_account_id", "country"}).
			AddRow(accountID, "user_123", stripeAccountID, "US"))

	// Mock Stripe GetAccount call (called by GetConnectedAccount)
	mockStripeClient.On("GetAccount", mock.Anything, stripeAccountID).
		Return(&stripe.Account{
			ID:             stripeAccountID,
			ChargesEnabled: true,
			PayoutsEnabled: true,
		}, nil)

	// Mock Stripe GetBalance call
	expectedBalance := &stripe.Balance{
		Available: []*stripe.BalanceAmount{
			{Amount: 10000, Currency: "usd"},
		},
		Pending: []*stripe.BalanceAmount{
			{Amount: 500, Currency: "usd"},
		},
	}

	mockStripeClient.On("GetBalance", mock.Anything, stripeAccountID).
		Return(expectedBalance, nil)

	balance, err := service.GetAccountBalance(context.Background(), accountID)

	require.NoError(t, err)
	require.NotNil(t, balance)
	assert.Equal(t, 1, len(balance.Available))
	assert.Equal(t, int64(10000), balance.Available[0].Amount)

	mockStripeClient.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestConnectedAccountService_GetAccountTransactions(t *testing.T) {
	db, sqlMock := setupTestDB(t)
	logger := zaptest.NewLogger(t)
	cfg := &config.StripeConfig{}

	mockStripeClient := new(mocks.MockStripeClient)
	service := services.NewConnectedAccountService(db, mockStripeClient, cfg, logger)

	accountID := "account_uuid_123"
	stripeAccountID := "acct_stripe_123"

	// Mock GetConnectedAccount query
	sqlMock.ExpectQuery(`SELECT \* FROM "connected_accounts" WHERE id = \$1 ORDER BY "connected_accounts"."id" LIMIT \$2`).
		WithArgs(accountID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "stripe_account_id", "country"}).
			AddRow(accountID, "user_123", stripeAccountID, "US"))

	// Mock Stripe GetAccount call (called by GetConnectedAccount)
	mockStripeClient.On("GetAccount", mock.Anything, stripeAccountID).
		Return(&stripe.Account{
			ID:             stripeAccountID,
			ChargesEnabled: true,
			PayoutsEnabled: true,
		}, nil)

	// Mock Stripe ListBalanceTransactions call
	expectedTransactions := []*stripe.BalanceTransaction{
		{
			ID:       "txn_123",
			Amount:   5000,
			Currency: "usd",
			Type:     "payment",
			Created:  1609459200,
		},
		{
			ID:       "txn_456",
			Amount:   -200,
			Currency: "usd",
			Type:     "fee",
			Created:  1609459300,
		},
	}

	params := &services.BalanceTransactionListParams{
		Limit: 10,
	}

	mockStripeClient.On("ListBalanceTransactions", mock.Anything, mock.MatchedBy(func(p *services.BalanceTransactionListParams) bool {
		return p.AccountID == stripeAccountID && p.Limit == 10
	})).Return(expectedTransactions, nil)

	transactions, err := service.GetAccountTransactions(context.Background(), accountID, params)

	require.NoError(t, err)
	require.NotNil(t, transactions)
	assert.Equal(t, 2, len(transactions))
	assert.Equal(t, "txn_123", transactions[0].ID)
	assert.Equal(t, int64(5000), transactions[0].Amount)

	mockStripeClient.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
