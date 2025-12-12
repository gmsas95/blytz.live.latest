package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz-mvp/services/product-service/internal/config"
	"github.com/gmsas95/blytz-mvp/services/product-service/internal/models"
)

// MockProductRepository for testing
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetAll(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*models.Product, error) {
	args := m.Called(ctx, limit, offset, filters)
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*models.Product, error) {
	args := m.Called(ctx, categoryID, limit, offset)
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateStock(ctx context.Context, productID string, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Product{}, &models.Category{}, &models.ProductImage{})
	return db
}

func TestProductService_CreateProduct(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	t.Run("Valid product creation", func(t *testing.T) {
		product := &models.Product{
			Name:         "Test Product",
			Description:  "Test Description",
			Price:        99.99,
			CategoryID:   "cat-123",
			SellerID:     "seller-123",
			Stock:        10,
			Status:       "active",
			Condition:    "new",
			Weight:       1.5,
			Dimensions:   "10x8x2",
		}

		err := service.CreateProduct(ctx, product)
		assert.NoError(t, err)
		assert.NotEmpty(t, product.ID)
		assert.Equal(t, "active", product.Status)
	})

	t.Run("Invalid product - missing name", func(t *testing.T) {
		product := &models.Product{
			Description: "Test Description",
			Price:       99.99,
			CategoryID:  "cat-123",
			SellerID:    "seller-123",
			Stock:       10,
			Status:      "active",
		}

		err := service.CreateProduct(ctx, product)
		assert.Error(t, err)
	})

	t.Run("Invalid product - negative price", func(t *testing.T) {
		product := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       -99.99,
			CategoryID:  "cat-123",
			SellerID:    "seller-123",
			Stock:       10,
			Status:      "active",
		}

		err := service.CreateProduct(ctx, product)
		assert.Error(t, err)
	})

	t.Run("Invalid product - negative stock", func(t *testing.T) {
		product := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			CategoryID:  "cat-123",
			SellerID:    "seller-123",
			Stock:       -5,
			Status:      "active",
		}

		err := service.CreateProduct(ctx, product)
		assert.Error(t, err)
	})
}

func TestProductService_GetProduct(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	// Create a test product
	product := &models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		CategoryID:   "cat-123",
		SellerID:     "seller-123",
		Stock:        10,
		Status:       "active",
		Condition:    "new",
	}
	err := service.CreateProduct(ctx, product)
	assert.NoError(t, err)

	t.Run("Get existing product", func(t *testing.T) {
		retrieved, err := service.GetProduct(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, product.Name, retrieved.Name)
		assert.Equal(t, product.Price, retrieved.Price)
	})

	t.Run("Get non-existent product", func(t *testing.T) {
		_, err := service.GetProduct(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	// Create a test product
	product := &models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		CategoryID:   "cat-123",
		SellerID:     "seller-123",
		Stock:        10,
		Status:       "active",
		Condition:    "new",
	}
	err := service.CreateProduct(ctx, product)
	assert.NoError(t, err)

	t.Run("Update product successfully", func(t *testing.T) {
		product.Name = "Updated Product"
		product.Price = 149.99
		product.Stock = 15

		err := service.UpdateProduct(ctx, product)
		assert.NoError(t, err)

		updated, err := service.GetProduct(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Product", updated.Name)
		assert.Equal(t, 149.99, updated.Price)
		assert.Equal(t, 15, updated.Stock)
	})

	t.Run("Update non-existent product", func(t *testing.T) {
		nonExistentProduct := &models.Product{
			ID:          "non-existent-id",
			Name:        "Non-existent Product",
			Description: "Description",
			Price:       99.99,
			CategoryID:  "cat-123",
			SellerID:    "seller-123",
			Stock:       10,
			Status:      "active",
		}

		err := service.UpdateProduct(ctx, nonExistentProduct)
		assert.Error(t, err)
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	// Create a test product
	product := &models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		CategoryID:   "cat-123",
		SellerID:     "seller-123",
		Stock:        10,
		Status:       "active",
		Condition:    "new",
	}
	err := service.CreateProduct(ctx, product)
	assert.NoError(t, err)

	t.Run("Delete existing product", func(t *testing.T) {
		err := service.DeleteProduct(ctx, product.ID)
		assert.NoError(t, err)

		// Verify product is deleted
		_, err = service.GetProduct(ctx, product.ID)
		assert.Error(t, err)
	})

	t.Run("Delete non-existent product", func(t *testing.T) {
		err := service.DeleteProduct(ctx, "non-existent-id")
		assert.Error(t, err)
	})
}

func TestProductService_SearchProducts(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	// Create test products
	products := []*models.Product{
		{
			Name:        "iPhone 13",
			Description: "Latest iPhone model",
			Price:       999.99,
			CategoryID:  "cat-electronics",
			SellerID:    "seller-123",
			Stock:       10,
			Status:      "active",
		},
		{
			Name:        "Samsung Galaxy",
			Description: "Android smartphone",
			Price:       799.99,
			CategoryID:  "cat-electronics",
			SellerID:    "seller-123",
			Stock:       15,
			Status:      "active",
		},
		{
			Name:        "Laptop Pro",
			Description: "Professional laptop",
			Price:       1299.99,
			CategoryID:  "cat-electronics",
			SellerID:    "seller-456",
			Stock:       5,
			Status:      "active",
		},
	}

	for _, product := range products {
		err := service.CreateProduct(ctx, product)
		assert.NoError(t, err)
	}

	t.Run("Search by name", func(t *testing.T) {
		results, err := service.SearchProducts(ctx, "iPhone", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "iPhone 13", results[0].Name)
	})

	t.Run("Search by description", func(t *testing.T) {
		results, err := service.SearchProducts(ctx, "Android", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Samsung Galaxy", results[0].Name)
	})

	t.Run("Search with no results", func(t *testing.T) {
		results, err := service.SearchProducts(ctx, "NonExistent", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

func TestProductService_GetProductsByCategory(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	// Create test products in different categories
	electronicsProduct := &models.Product{
		Name:        "iPhone 13",
		Description: "Latest iPhone model",
		Price:       999.99,
		CategoryID:  "cat-electronics",
		SellerID:    "seller-123",
		Stock:       10,
		Status:      "active",
	}

	clothingProduct := &models.Product{
		Name:        "T-Shirt",
		Description: "Cotton t-shirt",
		Price:       19.99,
		CategoryID:  "cat-clothing",
		SellerID:    "seller-123",
		Stock:       50,
		Status:      "active",
	}

	err := service.CreateProduct(ctx, electronicsProduct)
	assert.NoError(t, err)
	err = service.CreateProduct(ctx, clothingProduct)
	assert.NoError(t, err)

	t.Run("Get products by category", func(t *testing.T) {
		products, err := service.GetProductsByCategory(ctx, "cat-electronics", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, products, 1)
		assert.Equal(t, "iPhone 13", products[0].Name)
	})

	t.Run("Get products from empty category", func(t *testing.T) {
		products, err := service.GetProductsByCategory(ctx, "cat-empty", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, products, 0)
	})
}

func TestProductService_UpdateStock(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	// Create a test product
	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		CategoryID:  "cat-123",
		SellerID:    "seller-123",
		Stock:       10,
		Status:      "active",
	}
	err := service.CreateProduct(ctx, product)
	assert.NoError(t, err)

	t.Run("Increase stock", func(t *testing.T) {
		err := service.UpdateStock(ctx, product.ID, 5)
		assert.NoError(t, err)

		updated, err := service.GetProduct(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, 15, updated.Stock)
	})

	t.Run("Decrease stock", func(t *testing.T) {
		err := service.UpdateStock(ctx, product.ID, -3)
		assert.NoError(t, err)

		updated, err := service.GetProduct(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, 12, updated.Stock)
	})

	t.Run("Update stock for non-existent product", func(t *testing.T) {
		err := service.UpdateStock(ctx, "non-existent-id", 5)
		assert.Error(t, err)
	})

	t.Run("Stock below zero", func(t *testing.T) {
		err := service.UpdateStock(ctx, product.ID, -20)
		assert.Error(t, err)
	})
}

func TestProductService_ValidateProduct(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	t.Run("Valid product validation", func(t *testing.T) {
		product := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			CategoryID:  "cat-123",
			SellerID:    "seller-123",
			Stock:       10,
			Status:      "active",
		}

		err := service.ValidateProduct(ctx, product)
		assert.NoError(t, err)
	})

	t.Run("Invalid product - empty name", func(t *testing.T) {
		product := &models.Product{
			Name:        "",
			Description: "Test Description",
			Price:       99.99,
			CategoryID:  "cat-123",
			SellerID:    "seller-123",
			Stock:       10,
			Status:      "active",
		}

		err := service.ValidateProduct(ctx, product)
		assert.Error(t, err)
	})

	t.Run("Invalid product - negative price", func(t *testing.T) {
		product := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       -99.99,
			CategoryID:  "cat-123",
			SellerID:    "seller-123",
			Stock:       10,
			Status:      "active",
		}

		err := service.ValidateProduct(ctx, product)
		assert.Error(t, err)
	})

	t.Run("Invalid product - invalid status", func(t *testing.T) {
		product := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			CategoryID:  "cat-123",
			SellerID:    "seller-123",
			Stock:       10,
			Status:      "invalid-status",
		}

		err := service.ValidateProduct(ctx, product)
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkProductService_CreateProduct(b *testing.B) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		product := &models.Product{
			Name:         fmt.Sprintf("Test Product %d", i),
			Description:  "Test Description",
			Price:        float64(99 + i),
			CategoryID:   "cat-123",
			SellerID:     "seller-123",
			Stock:        10,
			Status:       "active",
			Condition:    "new",
		}
		service.CreateProduct(ctx, product)
	}
}

func BenchmarkProductService_SearchProducts(b *testing.B) {
	db := setupTestDB()
	cfg := &config.Config{
		DatabaseURL: ":memory:",
		JWTSecret:   "test-secret",
	}

	service := NewProductService(db, cfg)
	ctx := context.Background()

	// Create test products
	for i := 0; i < 100; i++ {
		product := &models.Product{
			Name:         fmt.Sprintf("Product %d", i),
			Description:  "Test Description",
			Price:        float64(99 + i),
			CategoryID:   "cat-123",
			SellerID:     "seller-123",
			Stock:        10,
			Status:       "active",
		}
		service.CreateProduct(ctx, product)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SearchProducts(ctx, "Product", 10, 0)
	}
}