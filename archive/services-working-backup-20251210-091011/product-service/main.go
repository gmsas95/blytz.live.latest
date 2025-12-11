package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Product struct
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	Condition   string    `json:"condition"` // "new", "used", "refurbished"
	Images      []string  `json:"images"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"` // "active", "inactive", "sold"
	SellerID    string    `json:"seller_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Request structs
type CreateProductRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	Condition   string    `json:"condition"`
	Images      []string  `json:"images"`
	Tags        []string  `json:"tags"`
	SellerID    string    `json:"seller_id"`
}

type UpdateProductRequest struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Price       *float64  `json:"price,omitempty"`
	Category    *string   `json:"category,omitempty"`
	Condition   *string   `json:"condition,omitempty"`
	Images      *[]string `json:"images,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
	Status      *string   `json:"status,omitempty"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Pagination struct
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Product service with in-memory storage
type ProductService struct {
	products []Product
	mu       sync.RWMutex
}

// New product service
func NewProductService() *ProductService {
	return &ProductService{
		products: []Product{
			{
				ID:          "product-1",
				Name:        "Vintage Camera",
				Description: "Beautiful vintage camera from the 1970s in excellent condition",
				Price:       299.99,
				Category:    "Electronics",
				Condition:   "used",
				Images:      []string{"https://picsum.photos/400/300?random=1"},
				Tags:        []string{"vintage", "camera", "photography"},
				Status:      "active",
				SellerID:    "seller-1",
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().Add(-24 * time.Hour),
			},
			{
				ID:          "product-2",
				Name:        "Designer Handbag",
				Description: "Authentic designer handbag, barely used, comes with dust bag",
				Price:       450.00,
				Category:    "Fashion",
				Condition:   "used",
				Images:      []string{"https://picsum.photos/400/300?random=2"},
				Tags:        []string{"designer", "handbag", "luxury"},
				Status:      "active",
				SellerID:    "seller-2",
				CreatedAt:   time.Now().Add(-12 * time.Hour),
				UpdatedAt:   time.Now().Add(-12 * time.Hour),
			},
			{
				ID:          "product-3",
				Name:        "Gaming Laptop",
				Description: "High-performance gaming laptop, RTX 3080, 32GB RAM, 1TB SSD",
				Price:       1899.99,
				Category:    "Electronics",
				Condition:   "new",
				Images:      []string{"https://picsum.photos/400/300?random=3"},
				Tags:        []string{"gaming", "laptop", "computer", "rtx"},
				Status:      "active",
				SellerID:    "seller-1",
				CreatedAt:   time.Now().Add(-6 * time.Hour),
				UpdatedAt:   time.Now().Add(-6 * time.Hour),
			},
			{
				ID:          "product-4",
				Name:        "Running Shoes",
				Description: "Professional running shoes, size 10, worn only once",
				Price:       89.99,
				Category:    "Sports",
				Condition:   "used",
				Images:      []string{"https://picsum.photos/400/300?random=4"},
				Tags:        []string{"running", "shoes", "sports", "athletic"},
				Status:      "active",
				SellerID:    "seller-3",
				CreatedAt:   time.Now().Add(-3 * time.Hour),
				UpdatedAt:   time.Now().Add(-3 * time.Hour),
			},
			{
				ID:          "product-5",
				Name:        "Smartphone",
				Description: "Latest smartphone, 256GB, excellent condition, original box",
				Price:       699.00,
				Category:    "Electronics",
				Condition:   "refurbished",
				Images:      []string{"https://picsum.photos/400/300?random=5"},
				Tags:        []string{"smartphone", "phone", "mobile", "refurbished"},
				Status:      "inactive",
				SellerID:    "seller-2",
				CreatedAt:   time.Now().Add(-1 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			},
		},
	}
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// JSON response helper
func writeJSONResponse(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// Parse pagination parameters
func parsePagination(r *http.Request) (int, int) {
	page := 1
	perPage := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if p := r.URL.Query().Get("per_page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	return page, perPage
}

// Paginate results
func paginate(items []Product, page, perPage int) ([]Product, Pagination) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Product{}, Pagination{page, perPage, total, totalPages}
	}
	if end > total {
		end = total
	}

	return items[start:end], Pagination{page, perPage, total, totalPages}
}

// Health check
func (s *ProductService) health(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Product service is healthy and working!",
		Data: map[string]interface{}{
			"service":   "product-service",
			"version":   "v2.0-working",
			"status":    "healthy",
			"timestamp": time.Now(),
			"products":  len(s.products),
		},
	})
}

// Get all products
func (s *ProductService) getAllProducts(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter active products only
	var activeProducts []Product
	for _, product := range s.products {
		if product.Status == "active" {
			activeProducts = append(activeProducts, product)
		}
	}

	// Apply search filter if provided
	search := strings.ToLower(r.URL.Query().Get("search"))
	if search != "" {
		var filteredProducts []Product
		for _, product := range activeProducts {
			if strings.Contains(strings.ToLower(product.Name), search) ||
				strings.Contains(strings.ToLower(product.Description), search) ||
				strings.Contains(strings.ToLower(product.Category), search) {
				filteredProducts = append(filteredProducts, product)
			}
		}
		activeProducts = filteredProducts
	}

	// Apply category filter if provided
	if category := strings.ToLower(r.URL.Query().Get("category")); category != "" {
		var filteredProducts []Product
		for _, product := range activeProducts {
			if strings.Contains(strings.ToLower(product.Category), category) {
				filteredProducts = append(filteredProducts, product)
			}
		}
		activeProducts = filteredProducts
	}

	// Apply condition filter if provided
	if condition := strings.ToLower(r.URL.Query().Get("condition")); condition != "" {
		var filteredProducts []Product
		for _, product := range activeProducts {
			if strings.Contains(strings.ToLower(product.Condition), condition) {
				filteredProducts = append(filteredProducts, product)
			}
		}
		activeProducts = filteredProducts
	}

	// Paginate results
	paginatedProducts, pagination := paginate(activeProducts, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Products retrieved successfully",
		Data: map[string]interface{}{
			"products":   paginatedProducts,
			"pagination": pagination,
		},
	})
}

// Get product by ID
func (s *ProductService) getProduct(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	if productID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Product ID is required",
			Error:   "No product ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, product := range s.products {
		if product.ID == productID {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Product retrieved successfully",
				Data: map[string]interface{}{
					"product": product,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Product not found",
		Error:   "Product with ID " + productID + " does not exist",
	})
}

// Create product
func (s *ProductService) createProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid product data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Name == "" || req.SellerID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Name and seller ID are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.Price <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Price must be greater than 0",
			Error:   "Invalid price",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create new product
	newProduct := Product{
		ID:          fmt.Sprintf("product-%d", time.Now().UnixNano()),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		Condition:   req.Condition,
		Images:      req.Images,
		Tags:        req.Tags,
		Status:      "active",
		SellerID:    req.SellerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.products = append(s.products, newProduct)

	log.Printf("Product created: %s (%s) by seller %s", newProduct.Name, newProduct.ID, newProduct.SellerID)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Product created successfully",
		Data: map[string]interface{}{
			"product": newProduct,
		},
	})
}

// Update product
func (s *ProductService) updateProduct(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	if productID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Product ID is required",
			Error:   "No product ID provided",
		})
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid product data",
			Error:   err.Error(),
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and update product
	for i := range s.products {
		if s.products[i].ID == productID {
			// Update fields if provided
			if req.Name != nil {
				s.products[i].Name = *req.Name
			}
			if req.Description != nil {
				s.products[i].Description = *req.Description
			}
			if req.Price != nil {
				if *req.Price <= 0 {
					writeJSONResponse(w, http.StatusBadRequest, Response{
						Success: false,
						Message: "Price must be greater than 0",
						Error:   "Invalid price",
					})
					return
				}
				s.products[i].Price = *req.Price
			}
			if req.Category != nil {
				s.products[i].Category = *req.Category
			}
			if req.Condition != nil {
				s.products[i].Condition = *req.Condition
			}
			if req.Images != nil {
				s.products[i].Images = *req.Images
			}
			if req.Tags != nil {
				s.products[i].Tags = *req.Tags
			}
			if req.Status != nil {
				s.products[i].Status = *req.Status
			}

			s.products[i].UpdatedAt = time.Now()

			log.Printf("Product updated: %s", productID)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Product updated successfully",
				Data: map[string]interface{}{
					"product": s.products[i],
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Product not found",
		Error:   "Product with ID " + productID + " does not exist",
	})
}

// Delete product
func (s *ProductService) deleteProduct(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	if productID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Product ID is required",
			Error:   "No product ID provided",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and delete product
	for i := range s.products {
		if s.products[i].ID == productID {
			// Remove product from slice
			s.products = append(s.products[:i], s.products[i+1:]...)

			log.Printf("Product deleted: %s", productID)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Product deleted successfully",
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Product not found",
		Error:   "Product with ID " + productID + " does not exist",
	})
}

// Get categories
func (s *ProductService) getCategories(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Extract unique categories
	categories := make(map[string]bool)
	for _, product := range s.products {
		if product.Status == "active" {
			categories[product.Category] = true
		}
	}

	// Convert to slice
	var categoryList []string
	for category := range categories {
		categoryList = append(categoryList, category)
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Categories retrieved successfully",
		Data: map[string]interface{}{
			"categories": categoryList,
			"count":      len(categoryList),
		},
	})
}

// Get seller products
func (s *ProductService) getSellerProducts(w http.ResponseWriter, r *http.Request) {
	sellerID := r.URL.Query().Get("seller_id")
	if sellerID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Seller ID is required",
			Error:   "No seller ID provided",
		})
		return
	}

	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter products by seller
	var sellerProducts []Product
	for _, product := range s.products {
		if product.SellerID == sellerID && product.Status == "active" {
			sellerProducts = append(sellerProducts, product)
		}
	}

	// Paginate results
	paginatedProducts, pagination := paginate(sellerProducts, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Seller products retrieved successfully",
		Data: map[string]interface{}{
			"products":   paginatedProducts,
			"pagination": pagination,
			"seller_id":  sellerID,
		},
	})
}

// Start product service
func main() {
	service := NewProductService()

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/products", corsMiddleware(service.getAllProducts))
	http.Handle("/api/v1/products/", corsMiddleware(service.getProduct)) // For GET by ID
	http.Handle("/api/v1/products/create", corsMiddleware(service.createProduct))
	http.Handle("/api/v1/products/update", corsMiddleware(service.updateProduct))
	http.Handle("/api/v1/products/delete", corsMiddleware(service.deleteProduct))
	http.Handle("/api/v1/products/categories", corsMiddleware(service.getCategories))
	http.Handle("/api/v1/products/seller", corsMiddleware(service.getSellerProducts))

	port := ":8086"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	fmt.Printf("🚀 PRODUCT SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)
	fmt.Printf("🛍️  Products list: http://localhost%s/api/v1/products\n", port)
	fmt.Printf("📝 Product details: http://localhost%s/api/v1/products/{id}\n", port)
	fmt.Printf("➕ Create product: http://localhost%s/api/v1/products/create\n", port)
	fmt.Printf("📦 Categories: http://localhost%s/api/v1/products/categories\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("📊 Total products: %d\n", len(service.products))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(port, nil))
}