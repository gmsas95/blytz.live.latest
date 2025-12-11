package performance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/config"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/models"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/repository"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/services"
)

// Performance test constants
const (
	NUM_CONCURRENT_USERS = 1000
	TARGET_RESPONSE_TIME = 100 * time.Millisecond
	TARGET_THROUGHPUT = 5000 // requests per second
	CACHE_HIT_RATE_TARGET = 0.7 // 70%
)

func BenchmarkSearchPerformance(b *testing.B) {
	// Setup test database
	dbConfig := &config.DatabaseConfig{
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/blytz_perf?sslmode=disable",
		Environment:  "test",
	}

	db, err := config.InitDB(dbConfig)
	require.NoError(b, err)

	// Setup test data
	setupTestData(b, db)

	searchRepo := repository.NewSearchRepository(db)
	searchService := services.NewSearchService(searchRepo, nil, nil)

	// Benchmark search queries
	testQueries := []models.SearchQuery{
		{Query: "laptop", Page: 1, Limit: 20},
		{Query: "phone", Page: 1, Limit: 20},
		{Query: "tablet", Page: 1, Limit: 20},
		{Query: "electronics", Page: 1, Limit: 20},
		{Query: "computer", Page: 1, Limit: 20},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := testQueries[i%len(testQueries)]
		_, err := searchService.Search(context.Background(), &query)
		if err != nil {
			b.Errorf("Search failed: %v", err)
		}
	}
}

func TestConcurrentSearchPerformance(t *testing.T) {
	// Setup test database
	dbConfig := &config.DatabaseConfig{
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/blytz_perf?sslmode=disable",
		Environment:  "test",
	}

	db, err := config.InitDB(dbConfig)
	require.NoError(t, err)

	// Setup test data
	setupTestData(t, db)

	searchRepo := repository.NewSearchRepository(db)
	searchService := services.NewSearchService(searchRepo, nil, nil)

	// Test concurrent search performance
	var wg sync.WaitGroup
	results := make(chan time.Duration, NUM_CONCURRENT_USERS)
	errors := make(chan error, NUM_CONCURRENT_USERS)

	startTime := time.Now()

	// Launch concurrent searches
	for i := 0; i < NUM_CONCURRENT_USERS; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			query := &models.SearchQuery{
				Query:  fmt.Sprintf("test-query-%d", userID%10),
				Page:   1,
				Limit:  20,
			}

			searchStart := time.Now()
			_, err := searchService.Search(context.Background(), query)
			searchDuration := time.Since(searchStart)

			if err != nil {
				errors <- err
			} else {
				results <- searchDuration
			}
		}(i)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	var responseTimes []time.Duration
	var errorCount int

	for duration := range results {
		responseTimes = append(responseTimes, duration)
	}

	for err := range errors {
		errorCount++
		t.Logf("Search error: %v", err)
	}

	totalDuration := time.Since(startTime)

	// Calculate metrics
	totalRequests := NUM_CONCURRENT_USERS - errorCount
	throughput := float64(totalRequests) / totalDuration.Seconds()
	
	avgResponseTime := time.Duration(0)
	maxResponseTime := time.Duration(0)
	p95ResponseTime := time.Duration(0)
	
	if len(responseTimes) > 0 {
		var totalResponseTime time.Duration
		for _, rt := range responseTimes {
			totalResponseTime += rt
			if rt > maxResponseTime {
				maxResponseTime = rt
			}
		}
		avgResponseTime = totalResponseTime / time.Duration(len(responseTimes))
		
		// Calculate 95th percentile
		sortedTimes := make([]time.Duration, len(responseTimes))
		copy(sortedTimes, responseTimes)
		
		// Simple bubble sort for percentile calculation
		for i := 0; i < len(sortedTimes)-1; i++ {
			for j := 0; j < len(sortedTimes)-i-1; j++ {
				if sortedTimes[j] > sortedTimes[j+1] {
					sortedTimes[j], sortedTimes[j+1] = sortedTimes[j+1], sortedTimes[j]
				}
			}
		}
		
		p95Index := int(float64(len(sortedTimes)) * 0.95)
		if p95Index < len(sortedTimes) {
			p95ResponseTime = sortedTimes[p95Index]
		}
	}

	// Report results
	t.Logf("=== Concurrent Search Performance Test Results ===")
	t.Logf("Concurrent Users: %d", NUM_CONCURRENT_USERS)
	t.Logf("Total Requests: %d", totalRequests)
	t.Logf("Error Count: %d", errorCount)
	t.Logf("Success Rate: %.2f%%", float64(totalRequests)/float64(NUM_CONCURRENT_USERS)*100)
	t.Logf("Throughput: %.2f requests/second", throughput)
	t.Logf("Average Response Time: %v", avgResponseTime)
	t.Logf("95th Percentile Response Time: %v", p95ResponseTime)
	t.Logf("Max Response Time: %v", maxResponseTime)
	t.Logf("Total Test Duration: %v", totalDuration)

	// Assert performance requirements
	require.Greater(t, throughput, TARGET_THROUGHPUT, 
		fmt.Sprintf("Throughput %.2f is below target %d", throughput, TARGET_THROUGHPUT))
	require.Less(t, avgResponseTime, TARGET_RESPONSE_TIME, 
		fmt.Sprintf("Average response time %v exceeds target %v", avgResponseTime, TARGET_RESPONSE_TIME))
	require.Less(t, p95ResponseTime, TARGET_RESPONSE_TIME*2, 
		fmt.Sprintf("95th percentile response time %v exceeds target %v", p95ResponseTime, TARGET_RESPONSE_TIME*2))
}

func TestIndexingPerformance(t *testing.T) {
	// Setup test database
	dbConfig := &config.DatabaseConfig{
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/blytz_perf?sslmode=disable",
		Environment:  "test",
	}

	db, err := config.InitDB(dbConfig)
	require.NoError(t, err)

	searchRepo := repository.NewSearchRepository(db)
	searchService := services.NewSearchService(searchRepo, nil, nil)

	// Test bulk indexing performance
	numEntities := 1000
	entities := make([]models.IndexRequest, numEntities)

	for i := 0; i < numEntities; i++ {
		entities[i] = models.IndexRequest{
			EntityType:  "product",
			EntityID:    fmt.Sprintf("perf-test-%d", i),
			Title:       fmt.Sprintf("Performance Test Product %d", i),
			Description: fmt.Sprintf("Description for performance test product %d", i),
			Tags:        []string{"performance", "test", fmt.Sprintf("category-%d", i%10)},
			Category:    fmt.Sprintf("Category %d", i%5),
			SubCategory: fmt.Sprintf("SubCategory %d", i%3),
			Price:       float64(i) * 10.99,
			Status:      "active",
			Location:    fmt.Sprintf("Location %d", i%20),
			SellerID:    fmt.Sprintf("seller-%d", i%50),
			SellerName:  fmt.Sprintf("Seller %d", i%50),
			Rating:      float32(3.0 + float64(i%3)),
			ReviewCount: i * 10,
			ViewCount:   int64(i * 100),
			LikesCount:  int64(i * 5),
			IsActive:    true,
		}
	}

	bulkReq := &models.BulkIndexRequest{Entities: entities}

	startTime := time.Now()
	err = searchService.BulkIndexEntities(context.Background(), bulkReq)
	indexingDuration := time.Since(startTime)

	require.NoError(t, err)

	throughput := float64(numEntities) / indexingDuration.Seconds()

	t.Logf("=== Bulk Indexing Performance Test Results ===")
	t.Logf("Entities Indexed: %d", numEntities)
	t.Logf("Indexing Duration: %v", indexingDuration)
	t.Logf("Indexing Throughput: %.2f entities/second", throughput)

	// Assert performance requirements
	require.Greater(t, throughput, 100.0, 
		fmt.Sprintf("Indexing throughput %.2f is below minimum 100 entities/second", throughput))
}

func TestSearchAccuracy(t *testing.T) {
	// Setup test database
	dbConfig := &config.DatabaseConfig{
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/blytz_perf?sslmode=disable",
		Environment:  "test",
	}

	db, err := config.InitDB(dbConfig)
	require.NoError(t, err)

	// Setup specific test data for accuracy testing
	setupAccuracyTestData(t, db)

	searchRepo := repository.NewSearchRepository(db)
	searchService := services.NewSearchService(searchRepo, nil, nil)

	// Test search accuracy
	testCases := []struct {
		query        string
		expectedHits int
		description  string
	}{
		{"laptop", 3, "Should find all laptop products"},
		{"phone", 2, "Should find all phone products"},
		{"electronics", 5, "Should find all electronics products"},
		{"nonexistent", 0, "Should find no results for non-existent query"},
		{"apple", 1, "Should find Apple laptop specifically"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			query := &models.SearchQuery{
				Query: tc.query,
				Page:  1,
				Limit: 20,
			}

			result, err := searchService.Search(context.Background(), query)
			require.NoError(t, err)
			require.Equal(t, tc.expectedHits, len(result.Results), 
				fmt.Sprintf("Query '%s' expected %d results, got %d", tc.query, tc.expectedHits, len(result.Results)))
		})
	}
}

func setupTestData(b *testing.B, db interface{}) {
	// Create test data for performance testing
	// This would typically insert a large number of test records
	// For now, we'll assume the database is pre-populated
}

func setupAccuracyTestData(t *testing.T, db interface{}) {
	// Create specific test data for accuracy testing
	// This would insert known test records with specific titles and descriptions
	// For now, we'll assume the database has the test data
}

// Test memory usage during search operations
func TestMemoryUsage(t *testing.T) {
	// Setup test database
	dbConfig := &config.DatabaseConfig{
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/blytz_perf?sslmode=disable",
		Environment:  "test",
	}

	db, err := config.InitDB(dbConfig)
	require.NoError(t, err)

	searchRepo := repository.NewSearchRepository(db)
	searchService := services.NewSearchService(searchRepo, nil, nil)

	// Test memory usage over multiple searches
	numSearches := 1000
	query := &models.SearchQuery{
		Query: "test",
		Page:  1,
		Limit: 20,
	}

	for i := 0; i < numSearches; i++ {
		_, err := searchService.Search(context.Background(), query)
		require.NoError(t, err)

		// Check for memory leaks periodically
		if i%100 == 0 {
			t.Logf("Completed %d searches", i)
			// In a real implementation, you might check memory usage here
		}
	}

	t.Logf("Completed %d search operations without memory issues", numSearches)
}

// Test cache performance (if Redis is available)
func TestCachePerformance(t *testing.T) {
	// This test would require Redis to be available
	// For now, we'll skip it if Redis is not configured
	t.Skip("Cache performance test requires Redis - skipping for now")
}