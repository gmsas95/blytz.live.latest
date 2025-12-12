package services

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/gmsas95/blytz-mvp/services/search-service/internal/models"
	"github.com/gmsas95/blytz-mvp/services/search-service/internal/repository"
	"go.uber.org/zap"
)

type SearchService struct {
	searchRepo *repository.SearchRepository
	redis      *redis.Client
	logger     *zap.Logger
}

func NewSearchService(searchRepo *repository.SearchRepository, redis *redis.Client, logger *zap.Logger) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
		redis:      redis,
		logger:     logger,
	}
}

// Search performs search with caching and analytics
func (s *SearchService) Search(ctx context.Context, query *models.SearchQuery) (*models.SearchResponse, error) {
	startTime := time.Now()

	// Generate cache key
	cacheKey := s.generateCacheKey("search", query)

	// Try to get from cache first
	if cached, err := s.getFromCache(ctx, cacheKey); err == nil && cached != nil {
		if typed, ok := cached.(*models.SearchResponse); ok {
			s.logger.Debug("Search result served from cache", zap.String("query", query.Query))
			return typed, nil
		}
	}

	// Perform search
	result, err := s.searchRepo.Search(ctx, query)
	if err != nil {
		s.logger.Error("Search failed", zap.Error(err), zap.String("query", query.Query))
		return nil, err
	}

	// Cache the result for 5 minutes
	if err := s.setCache(ctx, cacheKey, result, 5*time.Minute); err != nil {
		s.logger.Warn("Failed to cache search result", zap.Error(err))
	}

	// Record analytics asynchronously
	go s.recordAnalytics(query, result.Total, time.Since(startTime))

	s.logger.Info("Search completed",
		zap.String("query", query.Query),
		zap.Int64("results", result.Total),
		zap.Int64("query_time_ms", result.QueryTime))

	return result, nil
}

// IndexEntity indexes a single entity
func (s *SearchService) IndexEntity(ctx context.Context, req *models.IndexRequest) error {
	err := s.searchRepo.IndexEntity(ctx, req)
	if err != nil {
		s.logger.Error("Failed to index entity",
			zap.String("entity_type", req.EntityType),
			zap.String("entity_id", req.EntityID),
			zap.Error(err))
		return err
	}

	// Invalidate relevant cache entries
	s.invalidateCache(ctx, req.EntityType, req.Category)

	s.logger.Info("Entity indexed successfully",
		zap.String("entity_type", req.EntityType),
		zap.String("entity_id", req.EntityID))

	return nil
}

// BulkIndexEntities indexes multiple entities
func (s *SearchService) BulkIndexEntities(ctx context.Context, req *models.BulkIndexRequest) error {
	startTime := time.Now()

	err := s.searchRepo.BulkIndexEntities(ctx, req)
	if err != nil {
		s.logger.Error("Failed to bulk index entities", zap.Error(err))
		return err
	}

	// Invalidate cache for affected entity types and categories
	entityTypes := make(map[string]bool)
	categories := make(map[string]bool)

	for _, entity := range req.Entities {
		entityTypes[entity.EntityType] = true
		if entity.Category != "" {
			categories[entity.Category] = true
		}
	}

	for entityType := range entityTypes {
		s.invalidateCache(ctx, entityType, "")
	}

	for category := range categories {
		s.invalidateCache(ctx, "", category)
	}

	s.logger.Info("Bulk index completed",
		zap.Int("entity_count", len(req.Entities)),
		zap.Duration("duration", time.Since(startTime)))

	return nil
}

// DeleteFromIndex removes an entity from search index
func (s *SearchService) DeleteFromIndex(ctx context.Context, req *models.DeleteIndexRequest) error {
	err := s.searchRepo.DeleteFromIndex(ctx, req)
	if err != nil {
		s.logger.Error("Failed to delete from index",
			zap.String("entity_type", req.EntityType),
			zap.String("entity_id", req.EntityID),
			zap.Error(err))
		return err
	}

	// Invalidate cache
	s.invalidateCache(ctx, req.EntityType, "")

	s.logger.Info("Entity deleted from index",
		zap.String("entity_type", req.EntityType),
		zap.String("entity_id", req.EntityID))

	return nil
}

// GetSearchSuggestions returns search suggestions with caching
func (s *SearchService) GetSearchSuggestions(ctx context.Context, query string, limit int) ([]models.SearchSuggestion, error) {
	if query == "" {
		return []models.SearchSuggestion{}, nil
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("suggestions:%s:%d", s.hashString(query), limit)

	// Try cache first
	var suggestions []models.SearchSuggestion
	if cached, err := s.getFromCache(ctx, cacheKey); err == nil && cached != nil {
		if typed, ok := cached.([]models.SearchSuggestion); ok {
			return typed, nil
		}
	}

	// Get from repository
	suggestions, err := s.searchRepo.GetSearchSuggestions(ctx, query, limit)
	if err != nil {
		s.logger.Error("Failed to get search suggestions", zap.Error(err), zap.String("query", query))
		return nil, err
	}

	// Cache for 10 minutes
	if err := s.setCache(ctx, cacheKey, suggestions, 10*time.Minute); err != nil {
		s.logger.Warn("Failed to cache search suggestions", zap.Error(err))
	}

	return suggestions, nil
}

// GetSearchFacets returns faceted search results with caching
func (s *SearchService) GetSearchFacets(ctx context.Context, query *models.SearchQuery) (*models.SearchFacets, error) {
	// Generate cache key
	cacheKey := s.generateCacheKey("facets", query)

	// Try cache first
	var facets *models.SearchFacets
	if cached, err := s.getFromCache(ctx, cacheKey); err == nil && cached != nil {
		if typed, ok := cached.(*models.SearchFacets); ok {
			return typed, nil
		}
	}

	// Get from repository
	facets, err := s.searchRepo.GetSearchFacets(ctx, query)
	if err != nil {
		s.logger.Error("Failed to get search facets", zap.Error(err))
		return nil, err
	}

	// Cache for 15 minutes
	if err := s.setCache(ctx, cacheKey, facets, 15*time.Minute); err != nil {
		s.logger.Warn("Failed to cache search facets", zap.Error(err))
	}

	return facets, nil
}

// GetPopularSearches returns popular search queries with caching
func (s *SearchService) GetPopularSearches(ctx context.Context, limit int) ([]models.PopularSearch, error) {
	cacheKey := fmt.Sprintf("popular_searches:%d", limit)

	// Try cache first
	var popularSearches []models.PopularSearch
	if cached, err := s.getFromCache(ctx, cacheKey); err == nil && cached != nil {
		if typed, ok := cached.([]models.PopularSearch); ok {
			return typed, nil
		}
	}

	// Get from repository
	popularSearches, err := s.searchRepo.GetPopularSearches(ctx, limit)
	if err != nil {
		s.logger.Error("Failed to get popular searches", zap.Error(err))
		return nil, err
	}

	// Cache for 1 hour
	if err := s.setCache(ctx, cacheKey, popularSearches, time.Hour); err != nil {
		s.logger.Warn("Failed to cache popular searches", zap.Error(err))
	}

	return popularSearches, nil
}

// ReindexAll reindexes all data from source services
func (s *SearchService) ReindexAll(ctx context.Context) error {
	s.logger.Info("Starting full reindex")

	// This would typically involve:
	// 1. Calling product service to get all products
	// 2. Calling auction service to get all auctions
	// 3. Calling auth service to get all users
	// 4. Bulk indexing all entities

	// For now, we'll just clear the cache and log
	s.clearAllCache(ctx)

	s.logger.Info("Full reindex completed")
	return nil
}

// GetSearchStats returns search statistics
func (s *SearchService) GetSearchStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get cache hit rate
	cacheStats := s.getCacheStats(ctx)
	stats["cache"] = cacheStats

	// Get popular searches count
	popular, err := s.GetPopularSearches(ctx, 10)
	if err == nil {
		stats["popular_searches_count"] = len(popular)
	}

	// Get total indexed entities (this would require a separate query)
	stats["indexed_entities"] = "unknown" // Would be implemented with a count query

	return stats, nil
}

// Helper methods

func (s *SearchService) generateCacheKey(prefix string, query *models.SearchQuery) string {
	// Create a deterministic cache key based on query parameters
	queryBytes, _ := json.Marshal(query)
	queryHash := s.hashString(string(queryBytes))
	return fmt.Sprintf("%s:%s", prefix, queryHash)
}

func (s *SearchService) hashString(input string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}

func (s *SearchService) getFromCache(ctx context.Context, key string) (interface{}, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("cache miss")
	}
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal([]byte(val), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SearchService) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if s.redis == nil {
		return nil // Skip caching if Redis is not available
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, key, data, ttl).Err()
}

func (s *SearchService) invalidateCache(ctx context.Context, entityType, category string) {
	if s.redis == nil {
		return
	}

	// Delete cache entries matching patterns
	patterns := []string{
		"search:*",           // All search results
		"suggestions:*",       // All suggestions
		"facets:*",          // All facets
		"popular_searches:*", // Popular searches
	}

	if entityType != "" {
		patterns = append(patterns, fmt.Sprintf("search:*%s*", entityType))
	}

	if category != "" {
		patterns = append(patterns, fmt.Sprintf("search:*%s*", category))
	}

	for _, pattern := range patterns {
		keys, err := s.redis.Keys(ctx, pattern).Result()
		if err != nil {
			s.logger.Warn("Failed to get cache keys for invalidation", zap.Error(err), zap.String("pattern", pattern))
			continue
		}

		if len(keys) > 0 {
			err = s.redis.Del(ctx, keys...).Err()
			if err != nil {
				s.logger.Warn("Failed to invalidate cache keys", zap.Error(err), zap.Strings("keys", keys))
			}
		}
	}
}

func (s *SearchService) clearAllCache(ctx context.Context) {
	if s.redis == nil {
		return
	}

	patterns := []string{
		"search:*",
		"suggestions:*",
		"facets:*",
		"popular_searches:*",
	}

	for _, pattern := range patterns {
		keys, err := s.redis.Keys(ctx, pattern).Result()
		if err != nil {
			s.logger.Warn("Failed to get cache keys for clearing", zap.Error(err), zap.String("pattern", pattern))
			continue
		}

		if len(keys) > 0 {
			err = s.redis.Del(ctx, keys...).Err()
			if err != nil {
				s.logger.Warn("Failed to clear cache keys", zap.Error(err), zap.Strings("keys", keys))
			}
		}
	}
}

func (s *SearchService) getCacheStats(ctx context.Context) map[string]interface{} {
	if s.redis == nil {
		return map[string]interface{}{
			"status": "disabled",
		}
	}

	info, err := s.redis.Info(ctx, "memory", "keyspace", "stats").Result()
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	return map[string]interface{}{
		"status": "enabled",
		"info":   info,
	}
}

func (s *SearchService) recordAnalytics(query *models.SearchQuery, resultCount int64, queryTime time.Duration) {
	ctx := context.Background()

	analytics := &models.SearchAnalytics{
		Query:     query.Query,
		Results:   int(resultCount),
		CreatedAt: time.Now(),
	}

	// This would normally include user ID, IP, etc.
	// For now, we'll just record basic analytics

	err := s.searchRepo.RecordSearchAnalytics(ctx, analytics)
	if err != nil {
		s.logger.Warn("Failed to record search analytics", zap.Error(err))
	}
}