package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/models"
)

// MockRepository for testing
type MockAuctionRepository struct {
	mock.Mock
}

func (m *MockAuctionRepository) Create(ctx context.Context, auction *models.Auction) error {
	args := m.Called(ctx, auction)
	return args.Error(0)
}

func (m *MockAuctionRepository) GetByID(ctx context.Context, id string) (*models.Auction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Auction), args.Error(1)
}

func (m *MockAuctionRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Auction, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Auction), args.Error(1)
}

func (m *MockAuctionRepository) GetActive(ctx context.Context) ([]*models.Auction, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Auction), args.Error(1)
}

func (m *MockAuctionRepository) Update(ctx context.Context, auction *models.Auction) error {
	args := m.Called(ctx, auction)
	return args.Error(0)
}

func (m *MockAuctionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAuctionRepository) PlaceBid(ctx context.Context, bid *models.Bid) error {
	args := m.Called(ctx, bid)
	return args.Error(0)
}

func (m *MockAuctionRepository) GetBids(ctx context.Context, auctionID string) ([]*models.Bid, error) {
	args := m.Called(ctx, auctionID)
	return args.Get(0).([]*models.Bid), args.Error(1)
}

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Auction{}, &models.Bid{})
	return db
}

func TestAuctionService_CreateAuction(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewAuctionService(db, cfg)
	ctx := context.Background()

	t.Run("Valid auction creation", func(t *testing.T) {
		auction := &models.Auction{
			Title:         "Test Auction",
			Description:   "Test Description",
			StartingPrice: 100.00,
			CurrentPrice:  100.00,
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(24 * time.Hour),
			Status:        "active",
			SellerID:      "seller-123",
		}

		err := service.CreateAuction(ctx, auction)
		assert.NoError(t, err)
		assert.NotEmpty(t, auction.ID)
	})

	t.Run("Invalid auction - missing title", func(t *testing.T) {
		auction := &models.Auction{
			Description:   "Test Description",
			StartingPrice: 100.00,
			CurrentPrice:  100.00,
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(24 * time.Hour),
			Status:        "active",
			SellerID:      "seller-123",
		}

		err := service.CreateAuction(ctx, auction)
		assert.Error(t, err)
	})

	t.Run("Invalid auction - end time before start time", func(t *testing.T) {
		auction := &models.Auction{
			Title:         "Test Auction",
			Description:   "Test Description",
			StartingPrice: 100.00,
			CurrentPrice:  100.00,
			StartTime:     time.Now().Add(24 * time.Hour),
			EndTime:       time.Now(),
			Status:        "active",
			SellerID:      "seller-123",
		}

		err := service.CreateAuction(ctx, auction)
		assert.Error(t, err)
	})
}

func TestAuctionService_GetAuction(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewAuctionService(db, cfg)
	ctx := context.Background()

	// Create a test auction first
	auction := &models.Auction{
		Title:         "Test Auction",
		Description:   "Test Description",
		StartingPrice: 100.00,
		CurrentPrice:  100.00,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        "active",
		SellerID:      "seller-123",
	}
	err := service.CreateAuction(ctx, auction)
	assert.NoError(t, err)

	t.Run("Get existing auction", func(t *testing.T) {
		retrieved, err := service.GetAuction(ctx, auction.ID)
		assert.NoError(t, err)
		assert.Equal(t, auction.Title, retrieved.Title)
		assert.Equal(t, auction.Description, retrieved.Description)
	})

	t.Run("Get non-existent auction", func(t *testing.T) {
		_, err := service.GetAuction(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}

func TestAuctionService_PlaceBid(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewAuctionService(db, cfg)
	ctx := context.Background()

	// Create a test auction
	auction := &models.Auction{
		Title:         "Test Auction",
		Description:   "Test Description",
		StartingPrice: 100.00,
		CurrentPrice:  100.00,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        "active",
		SellerID:      "seller-123",
	}
	err := service.CreateAuction(ctx, auction)
	assert.NoError(t, err)

	t.Run("Valid bid placement", func(t *testing.T) {
		bid := &models.Bid{
			AuctionID: auction.ID,
			BidderID:  "bidder-123",
			Amount:    150.00,
		}

		err := service.PlaceBid(ctx, bid)
		assert.NoError(t, err)

		// Check if auction current price is updated
		updatedAuction, err := service.GetAuction(ctx, auction.ID)
		assert.NoError(t, err)
		assert.Equal(t, 150.00, updatedAuction.CurrentPrice)
	})

	t.Run("Bid lower than current price", func(t *testing.T) {
		bid := &models.Bid{
			AuctionID: auction.ID,
			BidderID:  "bidder-456",
			Amount:    120.00, // Lower than current price of 150.00
		}

		err := service.PlaceBid(ctx, bid)
		assert.Error(t, err)
	})

	t.Run("Bid on inactive auction", func(t *testing.T) {
		// Create an inactive auction
		inactiveAuction := &models.Auction{
			Title:         "Inactive Auction",
			Description:   "Test Description",
			StartingPrice: 100.00,
			CurrentPrice:  100.00,
			StartTime:     time.Now().Add(-48 * time.Hour),
			EndTime:       time.Now().Add(-24 * time.Hour),
			Status:        "ended",
			SellerID:      "seller-123",
		}
		err := service.CreateAuction(ctx, inactiveAuction)
		assert.NoError(t, err)

		bid := &models.Bid{
			AuctionID: inactiveAuction.ID,
			BidderID:  "bidder-789",
			Amount:    200.00,
		}

		err = service.PlaceBid(ctx, bid)
		assert.Error(t, err)
	})
}

func TestAuctionService_GetActiveAuctions(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewAuctionService(db, cfg)
	ctx := context.Background()

	// Create test auctions
	now := time.Now()
	activeAuction := &models.Auction{
		Title:         "Active Auction",
		Description:   "Test Description",
		StartingPrice: 100.00,
		CurrentPrice:  100.00,
		StartTime:     now.Add(-1 * time.Hour),
		EndTime:       now.Add(23 * time.Hour),
		Status:        "active",
		SellerID:      "seller-123",
	}

	inactiveAuction := &models.Auction{
		Title:         "Inactive Auction",
		Description:   "Test Description",
		StartingPrice: 100.00,
		CurrentPrice:  100.00,
		StartTime:     now.Add(-48 * time.Hour),
		EndTime:       now.Add(-24 * time.Hour),
		Status:        "ended",
		SellerID:      "seller-123",
	}

	err := service.CreateAuction(ctx, activeAuction)
	assert.NoError(t, err)
	err = service.CreateAuction(ctx, inactiveAuction)
	assert.NoError(t, err)

	t.Run("Get active auctions", func(t *testing.T) {
		activeAuctions, err := service.GetActiveAuctions(ctx)
		assert.NoError(t, err)
		assert.Len(t, activeAuctions, 1)
		assert.Equal(t, activeAuction.Title, activeAuctions[0].Title)
	})
}

func TestAuctionService_UpdateAuctionStatus(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewAuctionService(db, cfg)
	ctx := context.Background()

	// Create a test auction
	auction := &models.Auction{
		Title:         "Test Auction",
		Description:   "Test Description",
		StartingPrice: 100.00,
		CurrentPrice:  100.00,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        "active",
		SellerID:      "seller-123",
	}
	err := service.CreateAuction(ctx, auction)
	assert.NoError(t, err)

	t.Run("Update auction status", func(t *testing.T) {
		err := service.UpdateAuctionStatus(ctx, auction.ID, "ended")
		assert.NoError(t, err)

		updatedAuction, err := service.GetAuction(ctx, auction.ID)
		assert.NoError(t, err)
		assert.Equal(t, "ended", updatedAuction.Status)
	})

	t.Run("Update non-existent auction status", func(t *testing.T) {
		err := service.UpdateAuctionStatus(ctx, "non-existent-id", "ended")
		assert.Error(t, err)
	})
}

func TestAuctionService_GetBids(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewAuctionService(db, cfg)
	ctx := context.Background()

	// Create a test auction
	auction := &models.Auction{
		Title:         "Test Auction",
		Description:   "Test Description",
		StartingPrice: 100.00,
		CurrentPrice:  100.00,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        "active",
		SellerID:      "seller-123",
	}
	err := service.CreateAuction(ctx, auction)
	assert.NoError(t, err)

	// Place some bids
	bid1 := &models.Bid{
		AuctionID: auction.ID,
		BidderID:  "bidder-123",
		Amount:    150.00,
	}
	bid2 := &models.Bid{
		AuctionID: auction.ID,
		BidderID:  "bidder-456",
		Amount:    200.00,
	}

	err = service.PlaceBid(ctx, bid1)
	assert.NoError(t, err)
	err = service.PlaceBid(ctx, bid2)
	assert.NoError(t, err)

	t.Run("Get bids for auction", func(t *testing.T) {
		bids, err := service.GetBids(ctx, auction.ID)
		assert.NoError(t, err)
		assert.Len(t, bids, 2)
	})

	t.Run("Get bids for non-existent auction", func(t *testing.T) {
		_, err := service.GetBids(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkAuctionService_CreateAuction(b *testing.B) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewAuctionService(db, cfg)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auction := &models.Auction{
			Title:         fmt.Sprintf("Test Auction %d", i),
			Description:   "Test Description",
			StartingPrice: 100.00,
			CurrentPrice:  100.00,
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(24 * time.Hour),
			Status:        "active",
			SellerID:      "seller-123",
		}
		service.CreateAuction(ctx, auction)
	}
}

func BenchmarkAuctionService_PlaceBid(b *testing.B) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewAuctionService(db, cfg)
	ctx := context.Background()

	// Create a test auction
	auction := &models.Auction{
		Title:         "Test Auction",
		Description:   "Test Description",
		StartingPrice: 100.00,
		CurrentPrice:  100.00,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        "active",
		SellerID:      "seller-123",
	}
	service.CreateAuction(ctx, auction)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bid := &models.Bid{
			AuctionID: auction.ID,
			BidderID:  fmt.Sprintf("bidder-%d", i),
			Amount:    float64(100 + i),
		}
		service.PlaceBid(ctx, bid)
	}
}