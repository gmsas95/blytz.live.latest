package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	shared_utils "github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
	shared_errors "github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
	"github.com/gmsas95/blytz.live.latest/services/product-service/internal/models"
)

// ProductService handles product business logic
type ProductService struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewProductService creates a new product service
func NewProductService(db *sql.DB, logger *zap.Logger) *ProductService {
	return &ProductService{
		db:     db,
		logger: logger,
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(req *models.CreateProductRequest) (*models.Product, error) {
	// Convert images array to JSON
	imagesJSON, _ := json.Marshal(req.Images)
	tagsJSON, _ := json.Marshal(req.Tags)

	query := `
		INSERT INTO products (
			name, description, price, currency, image_url, images, 
			stock, category, subcategory, tags, status, is_featured, is_active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING id, product_id, created_at, updated_at
	`

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Currency:    req.Currency,
		ImageURL:    req.ImageURL,
		Stock:       req.Stock,
		Category:    req.Category,
		Subcategory: req.Subcategory,
		Status:      models.ProductStatusActive,
		IsFeatured:  false,
		IsActive:    true,
	}

	err := s.db.QueryRow(
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Currency,
		product.ImageURL,
		string(imagesJSON),
		product.Stock,
		product.Category,
		product.Subcategory,
		string(tagsJSON),
		product.Status,
		product.IsFeatured,
		product.IsActive,
	).Scan(&product.ID, &product.ProductID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("CREATE_PRODUCT_FAILED", "Failed to create product")
	}

	// Set the images and tags in the product
	product.SetImagesArray(req.Images)
	product.SetTagsArray(req.Tags)

	return product, nil
}

// GetProductByID retrieves a product by ID
func (s *ProductService) GetProductByID(id uint) (*models.Product, error) {
	query := `
		SELECT id, product_id, name, description, price, currency, image_url, images,
			   seller_id, seller_name, stock, reserved, status, is_featured, is_active,
			   category, subcategory, tags, metadata, created_at, updated_at
		FROM products 
		WHERE id = $1 AND deleted_at IS NULL
	`

	product := &models.Product{}
	var imagesJSON, tagsJSON, metadataJSON sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&product.ID,
		&product.ProductID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Currency,
		&product.ImageURL,
		&imagesJSON,
		&product.SellerID,
		&product.SellerName,
		&product.Stock,
		&product.Reserved,
		&product.Status,
		&product.IsFeatured,
		&product.IsActive,
		&product.Category,
		&product.Subcategory,
		&tagsJSON,
		&metadataJSON,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, shared_errors.NewNotFoundError("PRODUCT_NOT_FOUND", "Product not found")
		}
		s.logger.Error("Failed to get product", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("GET_PRODUCT_FAILED", "Failed to get product")
	}

	// Parse JSON fields
	if imagesJSON.Valid {
		product.Images = imagesJSON.String
	}
	if tagsJSON.Valid {
		product.Tags = tagsJSON.String
	}
	if metadataJSON.Valid {
		product.Metadata = metadataJSON.String
	}

	return product, nil
}

// GetProductByProductID retrieves a product by ProductID
func (s *ProductService) GetProductByProductID(productID string) (*models.Product, error) {
	query := `
		SELECT id, product_id, name, description, price, currency, image_url, images,
			   seller_id, seller_name, stock, reserved, status, is_featured, is_active,
			   category, subcategory, tags, metadata, created_at, updated_at
		FROM products 
		WHERE product_id = $1 AND deleted_at IS NULL
	`

	product := &models.Product{}
	var imagesJSON, tagsJSON, metadataJSON sql.NullString

	err := s.db.QueryRow(query, productID).Scan(
		&product.ID,
		&product.ProductID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Currency,
		&product.ImageURL,
		&imagesJSON,
		&product.SellerID,
		&product.SellerName,
		&product.Stock,
		&product.Reserved,
		&product.Status,
		&product.IsFeatured,
		&product.IsActive,
		&product.Category,
		&product.Subcategory,
		&tagsJSON,
		&metadataJSON,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, shared_errors.NewNotFoundError("PRODUCT_NOT_FOUND", "Product not found")
		}
		s.logger.Error("Failed to get product by ProductID", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("GET_PRODUCT_FAILED", "Failed to get product")
	}

	// Parse JSON fields
	if imagesJSON.Valid {
		product.Images = imagesJSON.String
	}
	if tagsJSON.Valid {
		product.Tags = tagsJSON.String
	}
	if metadataJSON.Valid {
		product.Metadata = metadataJSON.String
	}

	return product, nil
}

// ListProducts retrieves a paginated list of products
func (s *ProductService) ListProducts(page, pageSize int, filter *models.ProductFilter) (*models.ProductListResponse, error) {
	offset := (page - 1) * pageSize

	// Build WHERE clause
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.Category != "" {
			whereClause += fmt.Sprintf(" AND category = $%d", argIndex)
			args = append(args, filter.Category)
			argIndex++
		}
		if filter.Subcategory != "" {
			whereClause += fmt.Sprintf(" AND subcategory = $%d", argIndex)
			args = append(args, filter.Subcategory)
			argIndex++
		}
		if filter.MinPrice > 0 {
			whereClause += fmt.Sprintf(" AND price >= $%d", argIndex)
			args = append(args, filter.MinPrice)
			argIndex++
		}
		if filter.MaxPrice > 0 {
			whereClause += fmt.Sprintf(" AND price <= $%d", argIndex)
			args = append(args, filter.MaxPrice)
			argIndex++
		}
		if filter.SellerID != "" {
			whereClause += fmt.Sprintf(" AND seller_id = $%d", argIndex)
			args = append(args, filter.SellerID)
			argIndex++
		}
		if filter.Status != "" {
			whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, filter.Status)
			argIndex++
		}
		if filter.IsFeatured != nil {
			whereClause += fmt.Sprintf(" AND is_featured = $%d", argIndex)
			args = append(args, *filter.IsFeatured)
			argIndex++
		}
		if filter.Search != "" {
			whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1)
			args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
			argIndex += 2
		}
		if len(filter.Tags) > 0 {
			for _, tag := range filter.Tags {
				whereClause += fmt.Sprintf(" AND tags LIKE $%d", argIndex)
				args = append(args, "%"+tag+"%")
				argIndex++
			}
		}
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM products " + whereClause
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		s.logger.Error("Failed to count products", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("COUNT_PRODUCTS_FAILED", "Failed to count products")
	}

	// Get products
	query := `
		SELECT id, product_id, name, description, price, currency, image_url, images,
			   seller_id, seller_name, stock, reserved, status, is_featured, is_active,
			   category, subcategory, tags, metadata, created_at, updated_at
		FROM products 
	` + whereClause + `
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d
	`

	// Add pagination args
	args = append(args, pageSize, offset)
	query = fmt.Sprintf(query, argIndex, argIndex+1)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.logger.Error("Failed to list products", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("LIST_PRODUCTS_FAILED", "Failed to list products")
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		var imagesJSON, tagsJSON, metadataJSON sql.NullString

		err := rows.Scan(
			&product.ID,
			&product.ProductID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Currency,
			&product.ImageURL,
			&imagesJSON,
			&product.SellerID,
			&product.SellerName,
			&product.Stock,
			&product.Reserved,
			&product.Status,
			&product.IsFeatured,
			&product.IsActive,
			&product.Category,
			&product.Subcategory,
			&tagsJSON,
			&metadataJSON,
			&product.CreatedAt,
			&product.UpdatedAt,
		)

		if err != nil {
			s.logger.Error("Failed to scan product", zap.Error(err))
			return nil, shared_errors.NewDatabaseError("SCAN_PRODUCT_FAILED", "Failed to scan product")
		}

		// Parse JSON fields
		if imagesJSON.Valid {
			product.Images = imagesJSON.String
		}
		if tagsJSON.Valid {
			product.Tags = tagsJSON.String
		}
		if metadataJSON.Valid {
			product.Metadata = metadataJSON.String
		}

		products = append(products, product)
	}

	return &models.ProductListResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  int64(offset+pageSize) < total,
	}, nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(id uint, req *models.UpdateProductRequest) (*models.Product, error) {
	// First, get the existing product
	existingProduct, err := s.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	// Build update query
	setClause := ""
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		setClause += fmt.Sprintf("name = $%d, ", argIndex)
		args = append(args, req.Name)
		argIndex++
		existingProduct.Name = req.Name
	}
	if req.Description != "" {
		setClause += fmt.Sprintf("description = $%d, ", argIndex)
		args = append(args, req.Description)
		argIndex++
		existingProduct.Description = req.Description
	}
	if req.Price > 0 {
		setClause += fmt.Sprintf("price = $%d, ", argIndex)
		args = append(args, req.Price)
		argIndex++
		existingProduct.Price = req.Price
	}
	if req.Currency != "" {
		setClause += fmt.Sprintf("currency = $%d, ", argIndex)
		args = append(args, req.Currency)
		argIndex++
		existingProduct.Currency = req.Currency
	}
	if req.ImageURL != "" {
		setClause += fmt.Sprintf("image_url = $%d, ", argIndex)
		args = append(args, req.ImageURL)
		argIndex++
		existingProduct.ImageURL = req.ImageURL
	}
	if req.Stock >= 0 {
		setClause += fmt.Sprintf("stock = $%d, ", argIndex)
		args = append(args, req.Stock)
		argIndex++
		existingProduct.Stock = req.Stock
	}
	if req.Category != "" {
		setClause += fmt.Sprintf("category = $%d, ", argIndex)
		args = append(args, req.Category)
		argIndex++
		existingProduct.Category = req.Category
	}
	if req.Subcategory != "" {
		setClause += fmt.Sprintf("subcategory = $%d, ", argIndex)
		args = append(args, req.Subcategory)
		argIndex++
		existingProduct.Subcategory = req.Subcategory
	}
	if req.Status != "" {
		setClause += fmt.Sprintf("status = $%d, ", argIndex)
		args = append(args, req.Status)
		argIndex++
		existingProduct.Status = req.Status
	}

	setClause += fmt.Sprintf("is_featured = $%d, ", argIndex)
	args = append(args, req.IsFeatured)
	argIndex++
	existingProduct.IsFeatured = req.IsFeatured

	// Handle images array update
	if req.Images != nil {
		imagesJSON, _ := json.Marshal(req.Images)
		setClause += fmt.Sprintf("images = $%d, ", argIndex)
		args = append(args, string(imagesJSON))
		argIndex++
		existingProduct.SetImagesArray(req.Images)
	}

	// Handle tags array update
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		setClause += fmt.Sprintf("tags = $%d, ", argIndex)
		args = append(args, string(tagsJSON))
		argIndex++
		existingProduct.SetTagsArray(req.Tags)
	}

	// Add updated_at
	setClause += fmt.Sprintf("updated_at = NOW()")

	// Add WHERE clause
	setClause += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	query := "UPDATE products SET " + setClause

	_, err = s.db.Exec(query, args...)
	if err != nil {
		s.logger.Error("Failed to update product", zap.Error(err))
		return nil, shared_errors.NewDatabaseError("UPDATE_PRODUCT_FAILED", "Failed to update product")
	}

	// Refresh product data
	return s.GetProductByID(id)
}

// DeleteProduct soft deletes a product
func (s *ProductService) DeleteProduct(id uint) error {
	query := "UPDATE products SET deleted_at = NOW() WHERE id = $1"
	
	result, err := s.db.Exec(query, id)
	if err != nil {
		s.logger.Error("Failed to delete product", zap.Error(err))
		return shared_errors.NewDatabaseError("DELETE_PRODUCT_FAILED", "Failed to delete product")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", zap.Error(err))
		return shared_errors.NewDatabaseError("DELETE_PRODUCT_FAILED", "Failed to delete product")
	}

	if rowsAffected == 0 {
		return shared_errors.NewNotFoundError("PRODUCT_NOT_FOUND", "Product not found")
	}

	return nil
}

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	service *ProductService
	logger  *zap.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(service *ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

// CreateProduct handles product creation
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request_body": "Invalid request format: " + err.Error(),
		})
		return
	}

	product, err := h.service.CreateProduct(&req)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	response := models.ProductResponse{
		Product:   *product,
		Available: product.GetAvailable(),
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusCreated, "Product created successfully", response)
}

// GetProduct handles getting a product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"id": "Invalid product ID format",
		})
		return
	}

	product, err := h.service.GetProductByID(uint(id))
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	response := models.ProductResponse{
		Product:   *product,
		Available: product.GetAvailable(),
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, response)
}

// GetProductByProductID handles getting a product by ProductID
func (h *ProductHandler) GetProductByProductID(c *gin.Context) {
	productID := c.Param("product_id")
	if productID == "" {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"product_id": "Product ID is required",
		})
		return
	}

	product, err := h.service.GetProductByProductID(productID)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	response := models.ProductResponse{
		Product:   *product,
		Available: product.GetAvailable(),
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, response)
}

// ListProducts handles listing products with pagination and filtering
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, perPage := shared_utils.GetPaginationParams(c)

	// Build filter from query parameters
	filter := &models.ProductFilter{
		Category:    c.Query("category"),
		Subcategory: c.Query("subcategory"),
		SellerID:    c.Query("seller_id"),
		Status:      c.Query("status"),
		Search:      c.Query("search"),
	}

	// Parse price range
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseInt(minPriceStr, 10, 64); err == nil {
			filter.MinPrice = minPrice
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseInt(maxPriceStr, 10, 64); err == nil {
			filter.MaxPrice = maxPrice
		}
	}

	// Parse featured flag
	if featuredStr := c.Query("is_featured"); featuredStr != "" {
		if featured, err := strconv.ParseBool(featuredStr); err == nil {
			filter.IsFeatured = &featured
		}
	}

	// Parse tags
	if tagsStr := c.Query("tags"); tagsStr != "" {
		filter.Tags = strings.Split(tagsStr, ",")
		for i, tag := range filter.Tags {
			filter.Tags[i] = strings.TrimSpace(tag)
		}
	}

	result, err := h.service.ListProducts(page, perPage, filter)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	pagination := shared_utils.CalculatePagination(page, perPage, result.Total)
	shared_utils.SendPaginatedResponse(c, http.StatusOK, result.Products, pagination)
}

// UpdateProduct handles updating a product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"id": "Invalid product ID format",
		})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"request_body": "Invalid request format: " + err.Error(),
		})
		return
	}

	product, err := h.service.UpdateProduct(uint(id), &req)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	response := models.ProductResponse{
		Product:   *product,
		Available: product.GetAvailable(),
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Product updated successfully", response)
}

// DeleteProduct handles deleting a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		shared_utils.SendValidationErrorResponse(c, map[string]string{
			"id": "Invalid product ID format",
		})
		return
	}

	if err := h.service.DeleteProduct(uint(id)); err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponseWithMessage(c, http.StatusOK, "Product deleted successfully", nil)
}

// HealthCheck handles health check requests
func (h *ProductHandler) HealthCheck(c *gin.Context) {
	status := shared_utils.NewHealthStatus("ok")
	status.AddService("database", "ok", "Database connected successfully")
	
	shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"service":  "product-service",
		"version":  "v1.0.0",
		"status":   status,
		"time":     time.Now(),
	})
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize logger
	logger, err := shared_utils.NewDevelopmentLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/products_db?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Database connected successfully")

	// Run migrations
	if err := runMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	logger.Info("Database migrations completed")

	// Initialize services
	productService := NewProductService(db, logger)

	// Initialize handlers
	productHandler := NewProductHandler(productService, logger)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware using shared package
	router.Use(shared_utils.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", productHandler.HealthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Product routes
		products := v1.Group("/products")
		{
			products.GET("", productHandler.ListProducts)                    // List all products
			products.GET("/:id", productHandler.GetProduct)                 // Get product by ID
			products.GET("/by-product-id/:product_id", productHandler.GetProductByProductID) // Get product by ProductID
			products.POST("", productHandler.CreateProduct)                 // Create product
			products.PUT("/:id", productHandler.UpdateProduct)              // Update product
			products.DELETE("/:id", productHandler.DeleteProduct)           // Delete product
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086" // Use port 8086 as per docker-compose.yml
	}

	logger.Info("Product Service starting",
		zap.String("port", port),
		zap.String("database", "PostgreSQL"),
		zap.String("version", "v1.0.0"),
	)

	fmt.Printf("🚀 Product Service starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("📦 Products endpoint: http://localhost:%s/api/v1/products\n", port)
	fmt.Printf("🗄️  Database: PostgreSQL\n")
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// runMigrations creates the products table if it doesn't exist
func runMigrations(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			product_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL,
			currency VARCHAR(3) NOT NULL DEFAULT 'USD',
			image_url TEXT,
			images JSONB DEFAULT '[]'::jsonb,
			seller_id UUID,
			seller_name VARCHAR(255),
			stock INTEGER DEFAULT 0 NOT NULL,
			reserved INTEGER DEFAULT 0 NOT NULL,
			status VARCHAR(50) DEFAULT 'active' NOT NULL,
			is_featured BOOLEAN DEFAULT false NOT NULL,
			is_active BOOLEAN DEFAULT true NOT NULL,
			category VARCHAR(100),
			subcategory VARCHAR(100),
			tags JSONB DEFAULT '[]'::jsonb,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
			deleted_at TIMESTAMP WITH TIME ZONE
		);

		CREATE INDEX IF NOT EXISTS idx_products_product_id ON products(product_id);
		CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
		CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
		CREATE INDEX IF NOT EXISTS idx_products_featured ON products(is_featured);
		CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
		CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock);
		CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);

		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_products_updated_at ON products;
		CREATE TRIGGER update_products_updated_at 
			BEFORE UPDATE ON products 
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := db.Exec(query)
	return err
}