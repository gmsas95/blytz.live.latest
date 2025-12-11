package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz.live.latest/services/payment-service/internal/config"
	"github.com/gmsas95/blytz.live.latest/services/payment-service/internal/models"
)

// MockPaymentProcessor for testing
type MockPaymentProcessor struct {
	mock.Mock
}

func (m *MockPaymentProcessor) ProcessPayment(ctx context.Context, payment *models.Payment) (*models.PaymentResult, error) {
	args := m.Called(ctx, payment)
	return args.Get(0).(*models.PaymentResult), args.Error(1)
}

func (m *MockPaymentProcessor) RefundPayment(ctx context.Context, paymentID string, amount float64) (*models.PaymentResult, error) {
	args := m.Called(ctx, paymentID, amount)
	return args.Get(0).(*models.PaymentResult), args.Error(1)
}

func (m *MockPaymentProcessor) GetPaymentStatus(ctx context.Context, paymentID string) (*models.PaymentStatus, error) {
	args := m.Called(ctx, paymentID)
	return args.Get(0).(*models.PaymentStatus), args.Error(1)
}

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Payment{}, &models.PaymentMethod{}, &models.Transaction{})
	return db
}

func TestPaymentService_CreatePayment(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	t.Run("Valid payment creation", func(t *testing.T) {
		payment := &models.Payment{
			OrderID:        "order-123",
			UserID:         "user-123",
			Amount:         100.00,
			Currency:       "USD",
			PaymentMethod:  "credit_card",
			Status:         "pending",
			Description:    "Test payment",
		}

		err := service.CreatePayment(ctx, payment)
		assert.NoError(t, err)
		assert.NotEmpty(t, payment.ID)
		assert.Equal(t, "pending", payment.Status)
	})

	t.Run("Invalid payment - missing order ID", func(t *testing.T) {
		payment := &models.Payment{
			UserID:        "user-123",
			Amount:        100.00,
			Currency:      "USD",
			PaymentMethod: "credit_card",
			Status:        "pending",
		}

		err := service.CreatePayment(ctx, payment)
		assert.Error(t, err)
	})

	t.Run("Invalid payment - zero amount", func(t *testing.T) {
		payment := &models.Payment{
			OrderID:       "order-123",
			UserID:        "user-123",
			Amount:        0.00,
			Currency:      "USD",
			PaymentMethod: "credit_card",
			Status:        "pending",
		}

		err := service.CreatePayment(ctx, payment)
		assert.Error(t, err)
	})
}

func TestPaymentService_ProcessPayment(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	// Create a test payment
	payment := &models.Payment{
		OrderID:        "order-123",
		UserID:         "user-123",
		Amount:         100.00,
		Currency:       "USD",
		PaymentMethod:  "credit_card",
		Status:         "pending",
		Description:    "Test payment",
	}
	err := service.CreatePayment(ctx, payment)
	assert.NoError(t, err)

	t.Run("Successful payment processing", func(t *testing.T) {
		err := service.ProcessPayment(ctx, payment.ID)
		assert.NoError(t, err)

		// Check if payment status is updated
		updatedPayment, err := service.GetPayment(ctx, payment.ID)
		assert.NoError(t, err)
		assert.Equal(t, "completed", updatedPayment.Status)
		assert.NotEmpty(t, updatedPayment.TransactionID)
	})

	t.Run("Process non-existent payment", func(t *testing.T) {
		err := service.ProcessPayment(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("Process already completed payment", func(t *testing.T) {
		// Create and complete a payment
		completedPayment := &models.Payment{
			OrderID:        "order-456",
			UserID:         "user-456",
			Amount:         50.00,
			Currency:       "USD",
			PaymentMethod:  "credit_card",
			Status:         "completed",
			Description:    "Completed payment",
		}
		err := service.CreatePayment(ctx, completedPayment)
		assert.NoError(t, err)

		err = service.ProcessPayment(ctx, completedPayment.ID)
		assert.Error(t, err)
	})
}

func TestPaymentService_RefundPayment(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	// Create a completed payment
	payment := &models.Payment{
		OrderID:        "order-123",
		UserID:         "user-123",
		Amount:         100.00,
		Currency:       "USD",
		PaymentMethod:  "credit_card",
		Status:         "completed",
		Description:    "Test payment",
		TransactionID:  "txn-123",
	}
	err := service.CreatePayment(ctx, payment)
	assert.NoError(t, err)

	t.Run("Successful refund", func(t *testing.T) {
		refundAmount := 50.00
		err := service.RefundPayment(ctx, payment.ID, refundAmount)
		assert.NoError(t, err)

		// Check if refund transaction is created
		transactions, err := service.GetTransactions(ctx, payment.ID)
		assert.NoError(t, err)
		assert.Len(t, transactions, 1)
		assert.Equal(t, "refund", transactions[0].Type)
		assert.Equal(t, refundAmount, transactions[0].Amount)
	})

	t.Run("Refund amount greater than payment amount", func(t *testing.T) {
		err := service.RefundPayment(ctx, payment.ID, 150.00)
		assert.Error(t, err)
	})

	t.Run("Refund non-existent payment", func(t *testing.T) {
		err := service.RefundPayment(ctx, "non-existent-id", 50.00)
		assert.Error(t, err)
	})

	t.Run("Refund pending payment", func(t *testing.T) {
		pendingPayment := &models.Payment{
			OrderID:        "order-456",
			UserID:         "user-456",
			Amount:         75.00,
			Currency:       "USD",
			PaymentMethod:  "credit_card",
			Status:         "pending",
			Description:    "Pending payment",
		}
		err := service.CreatePayment(ctx, pendingPayment)
		assert.NoError(t, err)

		err = service.RefundPayment(ctx, pendingPayment.ID, 25.00)
		assert.Error(t, err)
	})
}

func TestPaymentService_GetPayment(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	// Create a test payment
	payment := &models.Payment{
		OrderID:        "order-123",
		UserID:         "user-123",
		Amount:         100.00,
		Currency:       "USD",
		PaymentMethod:  "credit_card",
		Status:         "pending",
		Description:    "Test payment",
	}
	err := service.CreatePayment(ctx, payment)
	assert.NoError(t, err)

	t.Run("Get existing payment", func(t *testing.T) {
		retrieved, err := service.GetPayment(ctx, payment.ID)
		assert.NoError(t, err)
		assert.Equal(t, payment.OrderID, retrieved.OrderID)
		assert.Equal(t, payment.Amount, retrieved.Amount)
	})

	t.Run("Get non-existent payment", func(t *testing.T) {
		_, err := service.GetPayment(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}

func TestPaymentService_GetUserPayments(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	userID := "user-123"

	// Create multiple payments for the user
	payment1 := &models.Payment{
		OrderID:        "order-1",
		UserID:         userID,
		Amount:         100.00,
		Currency:       "USD",
		PaymentMethod:  "credit_card",
		Status:         "completed",
		Description:    "Payment 1",
	}
	payment2 := &models.Payment{
		OrderID:        "order-2",
		UserID:         userID,
		Amount:         75.00,
		Currency:       "USD",
		PaymentMethod:  "paypal",
		Status:         "pending",
		Description:    "Payment 2",
	}

	err := service.CreatePayment(ctx, payment1)
	assert.NoError(t, err)
	err = service.CreatePayment(ctx, payment2)
	assert.NoError(t, err)

	// Create payment for different user
	otherPayment := &models.Payment{
		OrderID:        "order-3",
		UserID:         "other-user",
		Amount:         50.00,
		Currency:       "USD",
		PaymentMethod:  "credit_card",
		Status:         "completed",
		Description:    "Other user payment",
	}
	err = service.CreatePayment(ctx, otherPayment)
	assert.NoError(t, err)

	t.Run("Get user payments", func(t *testing.T) {
		payments, err := service.GetUserPayments(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, payments, 2)
	})

	t.Run("Get payments for non-existent user", func(t *testing.T) {
		payments, err := service.GetUserPayments(ctx, "non-existent-user")
		assert.NoError(t, err)
		assert.Len(t, payments, 0)
	})
}

func TestPaymentService_CreatePaymentMethod(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	t.Run("Valid payment method creation", func(t *testing.T) {
		paymentMethod := &models.PaymentMethod{
			UserID:      "user-123",
			Type:        "credit_card",
			Provider:    "stripe",
			LastFour:    "4242",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			IsDefault:   true,
		}

		err := service.CreatePaymentMethod(ctx, paymentMethod)
		assert.NoError(t, err)
		assert.NotEmpty(t, paymentMethod.ID)
	})

	t.Run("Invalid payment method - missing user ID", func(t *testing.T) {
		paymentMethod := &models.PaymentMethod{
			Type:        "credit_card",
			Provider:    "stripe",
			LastFour:    "4242",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			IsDefault:   true,
		}

		err := service.CreatePaymentMethod(ctx, paymentMethod)
		assert.Error(t, err)
	})
}

func TestPaymentService_GetUserPaymentMethods(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	userID := "user-123"

	// Create payment methods
	method1 := &models.PaymentMethod{
		UserID:      userID,
		Type:        "credit_card",
		Provider:    "stripe",
		LastFour:    "4242",
		ExpiryMonth: 12,
		ExpiryYear:  2025,
		IsDefault:   true,
	}
	method2 := &models.PaymentMethod{
		UserID:      userID,
		Type:        "paypal",
		Provider:    "paypal",
		IsDefault:   false,
	}

	err := service.CreatePaymentMethod(ctx, method1)
	assert.NoError(t, err)
	err = service.CreatePaymentMethod(ctx, method2)
	assert.NoError(t, err)

	t.Run("Get user payment methods", func(t *testing.T) {
		methods, err := service.GetUserPaymentMethods(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, methods, 2)
	})

	t.Run("Get payment methods for non-existent user", func(t *testing.T) {
		methods, err := service.GetUserPaymentMethods(ctx, "non-existent-user")
		assert.NoError(t, err)
		assert.Len(t, methods, 0)
	})
}

func TestPaymentService_ValidatePayment(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	t.Run("Valid payment validation", func(t *testing.T) {
		payment := &models.Payment{
			OrderID:        "order-123",
			UserID:         "user-123",
			Amount:         100.00,
			Currency:       "USD",
			PaymentMethod:  "credit_card",
			Status:         "pending",
			Description:    "Test payment",
		}

		err := service.ValidatePayment(ctx, payment)
		assert.NoError(t, err)
	})

	t.Run("Invalid payment - negative amount", func(t *testing.T) {
		payment := &models.Payment{
			OrderID:        "order-123",
			UserID:         "user-123",
			Amount:         -100.00,
			Currency:       "USD",
			PaymentMethod:  "credit_card",
			Status:         "pending",
			Description:    "Test payment",
		}

		err := service.ValidatePayment(ctx, payment)
		assert.Error(t, err)
	})

	t.Run("Invalid payment - unsupported currency", func(t *testing.T) {
		payment := &models.Payment{
			OrderID:        "order-123",
			UserID:         "user-123",
			Amount:         100.00,
			Currency:       "INVALID",
			PaymentMethod:  "credit_card",
			Status:         "pending",
			Description:    "Test payment",
		}

		err := service.ValidatePayment(ctx, payment)
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkPaymentService_CreatePayment(b *testing.B) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payment := &models.Payment{
			OrderID:        fmt.Sprintf("order-%d", i),
			UserID:         "user-123",
			Amount:         float64(100 + i),
			Currency:       "USD",
			PaymentMethod:  "credit_card",
			Status:         "pending",
			Description:    "Test payment",
		}
		service.CreatePayment(ctx, payment)
	}
}

func BenchmarkPaymentService_GetPayment(b *testing.B) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL:    ":memory:",
		StripeSecretKey: "test-secret",
		Environment:     "test",
	}

	service := NewPaymentService(db, cfg)
	ctx := context.Background()

	// Create a test payment
	payment := &models.Payment{
		OrderID:        "order-123",
		UserID:         "user-123",
		Amount:         100.00,
		Currency:       "USD",
		PaymentMethod:  "credit_card",
		Status:         "pending",
		Description:    "Test payment",
	}
	service.CreatePayment(ctx, payment)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetPayment(ctx, payment.ID)
	}
}