package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/api/handlers"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/config"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/models"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/repository"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/services"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/api/routes"
)

type SearchIntegrationTestSuite struct {
	suite.Suite
	router        *gin.Engine
	searchService *services.SearchService
}

func (suite *SearchIntegrationTestSuite) SetupSuite() {
	// Initialize test database
	dbConfig := &config.DatabaseConfig{
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/blytz_test?sslmode=disable",
		Environment:  "test",
	}

	db, err := config.InitDB(dbConfig)
	require.NoError(suite.T(), err)

	// Run migrations
	err = config.MigrateDatabase(db)
	require.NoError(suite.T(), err)

	// Initialize repositories and services
	searchRepo := repository.NewSearchRepository(db)
	searchService := services.NewSearchService(searchRepo, nil, nil) // No Redis for tests

	// Initialize handlers and routes
	searchHandler := handlers.NewSearchHandler(searchService)
	suite.router = gin.New()
	routes.SetupRoutes(suite.router, searchHandler)
	suite.searchService = searchService
}

func (suite *SearchIntegrationTestSuite) TearDownSuite() {
	// Cleanup test database if needed
}

func (suite *SearchIntegrationTestSuite) TestSearchEndpoint() {
	tests := []struct {
		name           string
		query          string
		expectedStatus  int
		expectedFields []string
	}{
		{
			name:           "Valid search query",
			query:          "laptop",
			expectedStatus:  http.StatusOK,
			expectedFields: []string{"results", "total", "page", "limit", "query_time_ms"},
		},
		{
			name:           "Empty search query",
			query:          "",
			expectedStatus:  http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "Search with pagination",
			query:          "phone",
			expectedStatus:  http.StatusOK,
			expectedFields: []string{"results", "total", "page", "limit", "query_time_ms"},
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/search?q="+tt.query, nil)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for _, field := range tt.expectedFields {
				_, exists := response[field]
				assert.True(t, exists, "Expected field %s not found in response", field)
			}
		})
	}
}

func (suite *SearchIntegrationTestSuite) TestIndexEntityEndpoint() {
	// Test indexing a single entity
	indexReq := models.IndexRequest{
		EntityType:  "product",
		EntityID:    "test-product-1",
		Title:       "Test Laptop",
		Description: "A high-performance laptop for professionals",
		Tags:        []string{"electronics", "computer", "laptop"},
		Category:    "Electronics",
		SubCategory: "Laptops",
		Price:       999.99,
		Status:      "active",
		Location:    "New York",
		SellerID:    "seller-123",
		SellerName:  "Tech Store",
		Rating:      4.5,
		ReviewCount: 150,
		ViewCount:   1000,
		LikesCount:  50,
		IsActive:    true,
	}

	reqBody, _ := json.Marshal(indexReq)
	req, _ := http.NewRequest("POST", "/api/v1/index", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "message")
	assert.Contains(suite.T(), response, "entity_type")
	assert.Contains(suite.T(), response, "entity_id")
}

func (suite *SearchIntegrationTestSuite) TestBulkIndexEndpoint() {
	// Test bulk indexing
	entities := []models.IndexRequest{
		{
			EntityType:  "product",
			EntityID:    "bulk-product-1",
			Title:       "Bulk Test Product 1",
			Description: "First bulk test product",
			Category:    "Test",
			Price:       10.99,
			Status:      "active",
		},
		{
			EntityType:  "product",
			EntityID:    "bulk-product-2",
			Title:       "Bulk Test Product 2",
			Description: "Second bulk test product",
			Category:    "Test",
			Price:       20.99,
			Status:      "active",
		},
	}

	bulkReq := models.BulkIndexRequest{Entities: entities}
	reqBody, _ := json.Marshal(bulkReq)
	req, _ := http.NewRequest("POST", "/api/v1/index/bulk", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "message")
	assert.Contains(suite.T(), response, "count")
}

func (suite *SearchIntegrationTestSuite) TestDeleteFromIndexEndpoint() {
	// First index an entity
	indexReq := models.IndexRequest{
		EntityType: "product",
		EntityID:   "delete-test-1",
		Title:      "Product to Delete",
		Category:   "Test",
		Price:       15.99,
		Status:      "active",
	}

	// Index the entity
	err := suite.searchService.IndexEntity(context.Background(), &indexReq)
	require.NoError(suite.T(), err)

	// Wait a bit for indexing
	time.Sleep(100 * time.Millisecond)

	// Delete the entity
	deleteReq := models.DeleteIndexRequest{
		EntityType: "product",
		EntityID:   "delete-test-1",
	}

	reqBody, _ := json.Marshal(deleteReq)
	req, _ := http.NewRequest("DELETE", "/api/v1/index", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "message")
	assert.Contains(suite.T(), response, "entity_type")
	assert.Contains(suite.T(), response, "entity_id")
}

func (suite *SearchIntegrationTestSuite) TestSearchSuggestionsEndpoint() {
	// Test search suggestions
	req, _ := http.NewRequest("GET", "/api/v1/search/suggestions?q=laptop&limit=5", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "query")
	assert.Contains(suite.T(), response, "suggestions")
}

func (suite *SearchIntegrationTestSuite) TestSearchFacetsEndpoint() {
	// Test search facets
	req, _ := http.NewRequest("GET", "/api/v1/search/facet?category=Electronics", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Check for expected facet fields
	expectedFacetFields := []string{"categories", "sub_categories", "price_ranges", "locations", "ratings"}
	for _, field := range expectedFacetFields {
		assert.Contains(suite.T(), response, field, "Expected facet field %s not found", field)
	}
}

func (suite *SearchIntegrationTestSuite) TestPopularSearchesEndpoint() {
	// Test popular searches
	req, _ := http.NewRequest("GET", "/api/v1/search/popular?limit=10", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "popular_searches")
}

func (suite *SearchIntegrationTestSuite) TestHealthEndpoint() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "service")
	assert.Contains(suite.T(), response, "status")
	assert.Contains(suite.T(), response, "cache")
	assert.Contains(suite.T(), response, "timestamp")
}

func (suite *SearchIntegrationTestSuite) TestSearchByTypeEndpoint() {
	// Test search by specific entity type
	req, _ := http.NewRequest("GET", "/api/v1/search/type/product?q=laptop", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "results")
	assert.Contains(suite.T(), response, "total")
	assert.Contains(suite.T(), response, "page")
	assert.Contains(suite.T(), response, "limit")
}

func (suite *SearchIntegrationTestSuite) TestDeleteByTypeAndIDEndpoint() {
	// First index an entity
	indexReq := models.IndexRequest{
		EntityType: "product",
		EntityID:   "delete-by-id-test-1",
		Title:      "Product to Delete by ID",
		Category:   "Test",
		Price:       25.99,
		Status:      "active",
	}

	// Index the entity
	err := suite.searchService.IndexEntity(context.Background(), &indexReq)
	require.NoError(suite.T(), err)

	// Wait a bit for indexing
	time.Sleep(100 * time.Millisecond)

	// Delete the entity using type and ID endpoint
	req, _ := http.NewRequest("DELETE", "/api/v1/index/type/product/id/delete-by-id-test-1", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "message")
	assert.Contains(suite.T(), response, "entity_type")
	assert.Contains(suite.T(), response, "entity_id")
}

func (suite *SearchIntegrationTestSuite) TestAdminStatsEndpoint() {
	req, _ := http.NewRequest("GET", "/api/v1/admin/stats", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "stats")
}

func (suite *SearchIntegrationTestSuite) TestAdminStatusEndpoint() {
	req, _ := http.NewRequest("GET", "/api/v1/admin/status", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "service")
	assert.Contains(suite.T(), response, "status")
	assert.Contains(suite.T(), response, "cache")
	assert.Contains(suite.T(), response, "features")
	assert.Contains(suite.T(), response, "performance")
}

func TestSearchIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SearchIntegrationTestSuite))
}

// Benchmark tests
func BenchmarkSearchEndpoint(b *testing.B) {
	// Setup similar to integration test
	dbConfig := &config.DatabaseConfig{
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/blytz_test?sslmode=disable",
		Environment:  "test",
	}

	db, err := config.InitDB(dbConfig)
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}

	searchRepo := repository.NewSearchRepository(db)
	searchService := services.NewSearchService(searchRepo, nil, nil)
	searchHandler := handlers.NewSearchHandler(searchService)

	router := gin.New()
	routes.SetupRoutes(router, searchHandler)

	// Prepare test data
	testQuery := &models.SearchQuery{
		Query:  "laptop",
		Page:   1,
		Limit:  20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := searchService.Search(context.Background(), testQuery)
		if err != nil {
			b.Errorf("Search failed: %v", err)
		}
	}
}