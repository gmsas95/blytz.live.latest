# 🎯 Product Service Implementation Plan

## 📍 **Current Status: Ready to Implement**

**Service**: Product Service (Port 8082)  
**Architecture**: ✅ Complete  
**Business Logic**: ❌ Missing (20%)  
**Priority**: 🔴 **HIGH** - **START HERE**

---

## 🏗️ **What's Already Implemented**

### **✅ Foundation Complete**
```go
// ✅ Database Models (internal/models/models.go)
type Product struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       int64   `json:"price"`
    CategoryID  string  `json:"category_id"`
    ImageURL    string  `json:"image_url"`
    Stock       int     `json:"stock"`
    Status      string  `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// ✅ Service Structure (internal/services/product.go)
// ✅ Handler Structure (internal/api/handlers/product.go)
// ✅ Router Setup (internal/api/router.go)
// ✅ Database Initialization (cmd/main.go)
// ✅ Docker Configuration (Dockerfile)
```

### **❌ What's Missing**
```go
// ❌ Business Logic Functions:
- CreateProduct()
- GetProduct() 
- UpdateProduct()
- DeleteProduct()
- ListProducts()
- SearchProducts()

// ❌ API Handler Functions:
- CreateProduct(c *gin.Context)
- GetProduct(c *gin.Context)
- UpdateProduct(c *gin.Context)
- DeleteProduct(c *gin.Context)
- ListProducts(c *gin.Context)
- SearchProducts(c *gin.Context)

// ❌ Route Definitions:
- POST   /products
- GET    /products/:id
- PUT    /products/:id
- DELETE /products/:id
- GET    /products
- GET    /products/search
```

---

## 🎯 **Implementation Plan (Week 1-2)**

### **📅 Week 1: Core CRUD Operations**

#### **Day 1-2: Service Layer Implementation**
**File**: `internal/services/product.go`

```go
package services

import (
    "context"
    "errors"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "go.uber.org/zap"
    "gorm.io/gorm"
    
    "github.com/gmsas95/blytz-mvp/services/product-service/internal/models"
    "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
)

type ProductService struct {
    db     *gorm.DB
    logger *zap.Logger
}

func NewProductService(db *gorm.DB, logger *zap.Logger) *ProductService {
    return &ProductService{
        db:     db,
        logger: logger,
    }
}

// CREATE PRODUCT
func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
    s.logger.Info("Creating product", zap.String("name", product.Name))
    
    // Validate input
    if product.Name == "" {
        return nil, errors.ErrInvalidProductName
    }
    if product.Price <= 0 {
        return nil, errors.ErrInvalidPrice
    }
    if product.Stock < 0 {
        return nil, errors.ErrInvalidStock
    }
    
    // Generate ID
    product.ID = uuid.New().String()
    product.Status = models.ProductStatusActive
    product.CreatedAt = time.Now()
    product.UpdatedAt = time.Now()
    
    // Save to database
    if err := s.db.WithContext(ctx).Create(product).Error; err != nil {
        s.logger.Error("Failed to create product", zap.Error(err))
        return nil, fmt.Errorf("failed to create product: %w", err)
    }
    
    s.logger.Info("Product created successfully", zap.String("id", product.ID))
    return product, nil
}

// GET PRODUCT BY ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
    s.logger.Info("Getting product", zap.String("id", id))
    
    var product models.Product
    if err := s.db.WithContext(ctx).Where("id = ?", id).First(&product).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.ErrProductNotFound
        }
        s.logger.Error("Failed to get product", zap.Error(err))
        return nil, fmt.Errorf("failed to get product: %w", err)
    }
    
    return &product, nil
}

// UPDATE PRODUCT
func (s *ProductService) UpdateProduct(ctx context.Context, id string, product *models.Product) (*models.Product, error) {
    s.logger.Info("Updating product", zap.String("id", id))
    
    // Check if product exists
    existing, err := s.GetProduct(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Validate input
    if product.Name != "" && product.Name != existing.Name {
        existing.Name = product.Name
    }
    if product.Price > 0 && product.Price != existing.Price {
        existing.Price = product.Price
    }
    if product.Description != "" {
        existing.Description = product.Description
    }
    if product.CategoryID != "" {
        existing.CategoryID = product.CategoryID
    }
    if product.ImageURL != "" {
        existing.ImageURL = product.ImageURL
    }
    if product.Stock >= 0 {
        existing.Stock = product.Stock
    }
    
    existing.UpdatedAt = time.Now()
    
    // Save to database
    if err := s.db.WithContext(ctx).Save(existing).Error; err != nil {
        s.logger.Error("Failed to update product", zap.Error(err))
        return nil, fmt.Errorf("failed to update product: %w", err)
    }
    
    s.logger.Info("Product updated successfully", zap.String("id", id))
    return existing, nil
}

// DELETE PRODUCT
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
    s.logger.Info("Deleting product", zap.String("id", id))
    
    // Check if product exists
    _, err := s.GetProduct(ctx, id)
    if err != nil {
        return err
    }
    
    // Soft delete
    if err := s.db.WithContext(ctx).Delete(&models.Product{}, "id = ?", id).Error; err != nil {
        s.logger.Error("Failed to delete product", zap.Error(err))
        return fmt.Errorf("failed to delete product: %w", err)
    }
    
    s.logger.Info("Product deleted successfully", zap.String("id", id))
    return nil
}

// LIST PRODUCTS
func (s *ProductService) ListProducts(ctx context.Context, filter *ProductFilter) ([]*models.Product, int64, error) {
    s.logger.Info("Listing products")
    
    var products []*models.Product
    var total int64
    
    query := s.db.WithContext(ctx).Model(&models.Product{})
    
    // Apply filters
    if filter.CategoryID != "" {
        query = query.Where("category_id = ?", filter.CategoryID)
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    if filter.MinPrice > 0 {
        query = query.Where("price >= ?", filter.MinPrice)
    }
    if filter.MaxPrice > 0 {
        query = query.Where("price <= ?", filter.MaxPrice)
    }
    
    // Count total
    if err := query.Count(&total).Error; err != nil {
        s.logger.Error("Failed to count products", zap.Error(err))
        return nil, 0, fmt.Errorf("failed to count products: %w", err)
    }
    
    // Apply pagination
    if filter.Limit > 0 {
        query = query.Limit(filter.Limit)
    }
    if filter.Offset > 0 {
        query = query.Offset(filter.Offset)
    }
    
    // Get products
    if err := query.Order("created_at DESC").Find(&products).Error; err != nil {
        s.logger.Error("Failed to list products", zap.Error(err))
        return nil, 0, fmt.Errorf("failed to list products: %w", err)
    }
    
    return products, total, nil
}

// SEARCH PRODUCTS
func (s *ProductService) SearchProducts(ctx context.Context, query string, limit int) ([]*models.Product, error) {
    s.logger.Info("Searching products", zap.String("query", query))
    
    var products []*models.Product
    
    // Search in name and description
    searchQuery := "%" + query + "%"
    if err := s.db.WithContext(ctx).
        Where("name ILIKE ? OR description ILIKE ?", searchQuery, searchQuery).
        Where("status = ?", models.ProductStatusActive).
        Order("created_at DESC").
        Limit(limit).
        Find(&products).Error; err != nil {
        s.logger.Error("Failed to search products", zap.Error(err))
        return nil, fmt.Errorf("failed to search products: %w", err)
    }
    
    return products, nil
}

// FILTER STRUCTURE
type ProductFilter struct {
    CategoryID string `json:"category_id"`
    Status     string `json:"status"`
    MinPrice   int64  `json:"min_price"`
    MaxPrice   int64  `json:"max_price"`
    Limit      int    `json:"limit"`
    Offset     int    `json:"offset"`
}
```

#### **Day 3-4: API Handler Implementation**
**File**: `internal/api/handlers/product.go`

```go
package handlers

import (
    "strconv"
    "strings"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    "github.com/gmsas95/blytz-mvp/services/product-service/internal/models"
    "github.com/gmsas95/blytz-mvp/services/product-service/internal/services"
    "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
    "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
)

type ProductHandler struct {
    productService *services.ProductService
    logger         *zap.Logger
}

func NewProductHandler(productService *services.ProductService, logger *zap.Logger) *ProductHandler {
    return &ProductHandler{
        productService: productService,
        logger:         logger,
    }
}

// CREATE PRODUCT
func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var req models.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, errors.ErrInvalidRequestBody)
        return
    }
    
    // Create product model
    product := &models.Product{
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        CategoryID:  req.CategoryID,
        ImageURL:    req.ImageURL,
        Stock:       req.Stock,
    }
    
    // Create product
    created, err := h.productService.CreateProduct(c.Request.Context(), product)
    if err != nil {
        h.logger.Error("Failed to create product", zap.Error(err))
        utils.ErrorResponse(c, err)
        return
    }
    
    utils.SuccessResponse(c, mapProductToResponse(created))
}

// GET PRODUCT BY ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        utils.ErrorResponse(c, errors.ErrInvalidProductID)
        return
    }
    
    product, err := h.productService.GetProduct(c.Request.Context(), id)
    if err != nil {
        h.logger.Error("Failed to get product", zap.Error(err))
        utils.ErrorResponse(c, err)
        return
    }
    
    utils.SuccessResponse(c, mapProductToResponse(product))
}

// UPDATE PRODUCT
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        utils.ErrorResponse(c, errors.ErrInvalidProductID)
        return
    }
    
    var req models.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, errors.ErrInvalidRequestBody)
        return
    }
    
    // Create product model
    product := &models.Product{
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        CategoryID:  req.CategoryID,
        ImageURL:    req.ImageURL,
        Stock:       req.Stock,
    }
    
    // Update product
    updated, err := h.productService.UpdateProduct(c.Request.Context(), id, product)
    if err != nil {
        h.logger.Error("Failed to update product", zap.Error(err))
        utils.ErrorResponse(c, err)
        return
    }
    
    utils.SuccessResponse(c, mapProductToResponse(updated))
}

// DELETE PRODUCT
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        utils.ErrorResponse(c, errors.ErrInvalidProductID)
        return
    }
    
    if err := h.productService.DeleteProduct(c.Request.Context(), id); err != nil {
        h.logger.Error("Failed to delete product", zap.Error(err))
        utils.ErrorResponse(c, err)
        return
    }
    
    utils.SuccessResponse(c, gin.H{"message": "Product deleted successfully"})
}

// LIST PRODUCTS
func (h *ProductHandler) ListProducts(c *gin.Context) {
    // Parse query parameters
    filter := &services.ProductFilter{}
    
    if categoryID := c.Query("category_id"); categoryID != "" {
        filter.CategoryID = categoryID
    }
    if status := c.Query("status"); status != "" {
        filter.Status = status
    }
    if minPrice := c.Query("min_price"); minPrice != "" {
        if price, err := strconv.ParseInt(minPrice, 10, 64); err == nil {
            filter.MinPrice = price
        }
    }
    if maxPrice := c.Query("max_price"); maxPrice != "" {
        if price, err := strconv.ParseInt(maxPrice, 10, 64); err == nil {
            filter.MaxPrice = price
        }
    }
    if limit := c.Query("limit"); limit != "" {
        if l, err := strconv.Atoi(limit); err == nil {
            filter.Limit = l
        }
    }
    if offset := c.Query("offset"); offset != "" {
        if o, err := strconv.Atoi(offset); err == nil {
            filter.Offset = o
        }
    }
    
    // Get products
    products, total, err := h.productService.ListProducts(c.Request.Context(), filter)
    if err != nil {
        h.logger.Error("Failed to list products", zap.Error(err))
        utils.ErrorResponse(c, err)
        return
    }
    
    // Map to response
    productResponses := make([]*models.ProductResponse, len(products))
    for i, product := range products {
        productResponses[i] = mapProductToResponse(product)
    }
    
    utils.SuccessResponse(c, gin.H{
        "products": productResponses,
        "total":    total,
        "limit":    filter.Limit,
        "offset":   filter.Offset,
    })
}

// SEARCH PRODUCTS
func (h *ProductHandler) SearchProducts(c *gin.Context) {
    query := strings.TrimSpace(c.Query("q"))
    if query == "" {
        utils.ErrorResponse(c, errors.ErrInvalidSearchQuery)
        return
    }
    
    limit := 20 // default
    if l := c.Query("limit"); l != "" {
        if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
            limit = parsed
        }
    }
    
    products, err := h.productService.SearchProducts(c.Request.Context(), query, limit)
    if err != nil {
        h.logger.Error("Failed to search products", zap.Error(err))
        utils.ErrorResponse(c, err)
        return
    }
    
    // Map to response
    productResponses := make([]*models.ProductResponse, len(products))
    for i, product := range products {
        productResponses[i] = mapProductToResponse(product)
    }
    
    utils.SuccessResponse(c, gin.H{
        "products": productResponses,
        "query":    query,
        "total":    len(productResponses),
    })
}

// HELPER: Map product to response
func mapProductToResponse(product *models.Product) *models.ProductResponse {
    return &models.ProductResponse{
        ID:          product.ID,
        Name:        product.Name,
        Description: product.Description,
        Price:       product.Price,
        CategoryID:  product.CategoryID,
        ImageURL:    product.ImageURL,
        Stock:       product.Stock,
        Status:      product.Status,
        CreatedAt:   product.CreatedAt.Format(time.RFC3339),
        UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
    }
}
```

#### **Day 5: Router Implementation**
**File**: `internal/api/router.go`

```go
package api

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/gmsas95/blytz-mvp/services/product-service/internal/api/handlers"
    "github.com/gmsas95/blytz-mvp/services/product-service/internal/config"
    "github.com/gmsas95/blytz-mvp/services/product-service/internal/models"
    "github.com/gmsas95/blytz-mvp/services/product-service/internal/services"
    "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
    "go.uber.org/zap"
)

func SetupRouter(db *gorm.DB, logger *zap.Logger, cfg *config.Config) *gin.Engine {
    router := gin.Default()
    
    // Add middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(utils.CorrelationMiddleware(logger))
    
    // Add CORS
    router.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.Status(http.StatusNoContent)
            return
        }
        c.Next()
    })
    
    // Initialize services
    productService := services.NewProductService(db, logger)
    
    // Initialize handlers
    productHandler := handlers.NewProductHandler(productService, logger)
    
    // Health endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status":    "ok",
            "service":   "product-service",
            "timestamp": time.Now().UTC(),
        })
    })
    
    // Product routes
    api := router.Group("/api")
    {
        // Public routes
        public := api.Group("/public")
        {
            public.GET("/products", productHandler.ListProducts)
            public.GET("/products/search", productHandler.SearchProducts)
            public.GET("/products/:id", productHandler.GetProduct)
        }
        
        // Protected routes (add auth middleware later)
        protected := api.Group("")
        {
            protected.POST("/products", productHandler.CreateProduct)
            protected.PUT("/products/:id", productHandler.UpdateProduct)
            protected.DELETE("/products/:id", productHandler.DeleteProduct)
        }
    }
    
    return router
}
```

### **📅 Week 2: Enhancement & Testing**

#### **Day 6-7: Model Updates**
**File**: `internal/models/models.go`

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// Product Status Constants
type ProductStatus string

const (
    ProductStatusActive   ProductStatus = "active"
    ProductStatusInactive ProductStatus = "inactive"
    ProductStatusDraft    ProductStatus = "draft"
    ProductStatusArchived ProductStatus = "archived"
)

// Product Model
type Product struct {
    ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Name        string         `json:"name" gorm:"not null;index"`
    Description string         `json:"description" gorm:"type:text"`
    Price       int64          `json:"price" gorm:"not null"` // Price in cents
    CategoryID  string         `json:"category_id" gorm:"index"`
    ImageURL    string         `json:"image_url"`
    Images      []ProductImage `json:"images" gorm:"foreignKey:ProductID"`
    Stock       int            `json:"stock" gorm:"default:0"`
    Status      ProductStatus  `json:"status" gorm:"default:'active';index"`
    Tags        string         `json:"tags" gorm:"type:text"`
    SKU         string         `json:"sku" gorm:"uniqueIndex"`
    Weight      float64        `json:"weight"`
    Dimensions  string         `json:"dimensions"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Product Image Model
type ProductImage struct {
    ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    ProductID  string    `json:"product_id" gorm:"not null;index"`
    ImageURL   string    `json:"image_url" gorm:"not null"`
    IsMain     bool      `json:"is_main" gorm:"default:false"`
    SortOrder  int       `json:"sort_order" gorm:"default:0"`
    CreatedAt  time.Time `json:"created_at"`
    Product    Product   `json:"product" gorm:"foreignKey:ProductID"`
}

// Request/Response Models
type CreateProductRequest struct {
    Name        string  `json:"name" binding:"required,min=1,max=255"`
    Description string  `json:"description" binding:"max=2000"`
    Price       int64   `json:"price" binding:"required,min=1"`
    CategoryID  string  `json:"category_id" binding:"required"`
    ImageURL    string  `json:"image_url"`
    Stock       int     `json:"stock" binding:"min=0"`
    Tags        string  `json:"tags"`
    SKU         string  `json:"sku" binding:"max=100"`
    Weight      float64 `json:"weight" binding:"min=0"`
    Dimensions  string  `json:"dimensions" binding:"max=100"`
}

type UpdateProductRequest struct {
    Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
    Description *string `json:"description" binding:"omitempty,max=2000"`
    Price       *int64  `json:"price" binding:"omitempty,min=1"`
    CategoryID  *string `json:"category_id"`
    ImageURL    *string `json:"image_url"`
    Stock       *int    `json:"stock" binding:"omitempty,min=0"`
    Status      *string `json:"status"`
    Tags        *string `json:"tags"`
    Weight      *float64 `json:"weight" binding:"omitempty,min=0"`
    Dimensions  *string `json:"dimensions" binding:"omitempty,max=100"`
}

type ProductResponse struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       int64   `json:"price"`
    CategoryID  string  `json:"category_id"`
    ImageURL    string  `json:"image_url"`
    Stock       int     `json:"stock"`
    Status      string  `json:"status"`
    Tags        string  `json:"tags"`
    SKU         string  `json:"sku"`
    Weight      float64 `json:"weight"`
    Dimensions  string  `json:"dimensions"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

type ProductListResponse struct {
    Products []*ProductResponse `json:"products"`
    Total    int64             `json:"total"`
    Limit    int               `json:"limit"`
    Offset   int               `json:"offset"`
}
```

---

## 🧪 **Testing Plan**

### **📋 Manual Testing Commands**

#### **1. Health Check**
```bash
curl http://localhost:8082/health
```

#### **2. Create Product**
```bash
curl -X POST http://localhost:8082/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "description": "A great test product",
    "price": 2999,
    "category_id": "cat-123",
    "stock": 100,
    "sku": "TEST-001"
  }'
```

#### **3. Get Product**
```bash
curl http://localhost:8082/api/public/products/[product-id]
```

#### **4. List Products**
```bash
curl http://localhost:8082/api/public/products
curl http://localhost:8082/api/public/products?limit=10&offset=0
```

#### **5. Search Products**
```bash
curl "http://localhost:8082/api/public/products/search?q=test"
```

#### **6. Update Product**
```bash
curl -X PUT http://localhost:8082/api/products/[product-id] \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Test Product",
    "price": 3499,
    "stock": 50
  }'
```

#### **7. Delete Product**
```bash
curl -X DELETE http://localhost:8082/api/products/[product-id]
```

#### **8. Gateway Routing Test**
```bash
# Test through API Gateway
curl http://localhost:8080/api/v1/products
curl http://localhost:8080/api/v1/products/search?q=test
```

---

## ✅ **Acceptance Criteria**

### **✅ Core Functionality**
- [ ] Create product via API → Product saved in database
- [ ] Get product by ID → Product returned correctly
- [ ] Update product via API → Product updated in database
- [ ] Delete product via API → Product soft deleted
- [ ] List products → Paginated list returned
- [ ] Search products → Relevant results returned

### **✅ API Quality**
- [ ] Input validation → Proper error responses
- [ ] Missing product → 404 error returned
- [ ] Invalid data → 400 error returned
- [ ] Response format → Consistent JSON structure
- [ ] Status codes → Correct HTTP status codes

### **✅ Integration**
- [ ] Gateway routing → `/api/v1/products` works through Gateway
- [ ] Database connection → PostgreSQL integration working
- [ ] Health endpoint → `/health` returns service status
- [ ] Docker container → Service starts and runs correctly

### **✅ Performance**
- [ ] Response time → <200ms for simple requests
- [ ] Database queries → Efficient with proper indexing
- [ ] Error handling → Graceful failure with logging

---

## 🎯 **Ready to Start!**

### **📁 Working Directory**
```bash
cd /home/sas/blytzmvp-clean/services/product-service
```

### **📝 Files to Edit/Implement:**
1. `internal/services/product.go` - **START HERE**
2. `internal/api/handlers/product.go` - **NEXT**
3. `internal/api/router.go` - **THEN**
4. `internal/models/models.go` - **FINALLY**

### **🚀 Implementation Order:**
1. **Day 1-2**: Service layer (CRUD operations)
2. **Day 3-4**: API handlers (HTTP endpoints)
3. **Day 5**: Router setup and routes
4. **Day 6-7**: Model updates and enhancements
5. **Day 8-10**: Testing and bug fixes

---

## 📞 **Need Help?**

**During implementation, ask me about:**
- **GORM queries** - Database operations
- **API patterns** - Go/Gin best practices
- **Error handling** - Proper error responses
- **Validation** - Input validation strategies
- **Testing** - Debug API endpoints

**Share your code with me for review and suggestions!** 🎯

**Ready to implement Product Service?** 🚀