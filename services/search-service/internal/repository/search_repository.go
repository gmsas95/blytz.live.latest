package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/models"
	"gorm.io/gorm"
)

type SearchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// Search performs full-text search with advanced filtering
func (r *SearchRepository) Search(ctx context.Context, query *models.SearchQuery) (*models.SearchResponse, error) {
	startTime := time.Now()

	// Build the base query
	baseQuery := r.db.WithContext(ctx).Table("search_indices")
	
	// Build where conditions
	whereConditions := []string{}
	whereArgs := []interface{}{}

	// Always filter by active status
	whereConditions = append(whereConditions, "is_active = ?")
	whereArgs = append(whereArgs, true)

	// Entity type filter
	if query.EntityType != "" && query.EntityType != "all" {
		whereConditions = append(whereConditions, "entity_type = ?")
		whereArgs = append(whereArgs, query.EntityType)
	}

	// Full-text search
	if query.Query != "" {
		if query.FuzzySearch {
			// Use trigram similarity for fuzzy search
			whereConditions = append(whereConditions, "(title % ? OR description % ? OR category % ? OR tags % ?)")
			whereArgs = append(whereArgs, query.Query, query.Query, query.Query, query.Query)
		} else {
			// Use tsvector for exact full-text search
			whereConditions = append(whereConditions, "search_vector @@ plainto_tsquery('english', ?)")
			whereArgs = append(whereArgs, query.Query)
		}
	}

	// Category filter
	if query.Category != "" {
		whereConditions = append(whereConditions, "category = ?")
		whereArgs = append(whereArgs, query.Category)
	}

	// Sub-category filter
	if query.SubCategory != "" {
		whereConditions = append(whereConditions, "sub_category = ?")
		whereArgs = append(whereArgs, query.SubCategory)
	}

	// Price range filter
	if query.MinPrice > 0 {
		whereConditions = append(whereConditions, "price >= ?")
		whereArgs = append(whereArgs, query.MinPrice)
	}
	if query.MaxPrice > 0 {
		whereConditions = append(whereConditions, "price <= ?")
		whereArgs = append(whereArgs, query.MaxPrice)
	}

	// Location filter
	if query.Location != "" {
		whereConditions = append(whereConditions, "location ILIKE ?")
		whereArgs = append(whereArgs, "%"+query.Location+"%")
	}

	// Seller filter
	if query.SellerID != "" {
		whereConditions = append(whereConditions, "seller_id = ?")
		whereArgs = append(whereArgs, query.SellerID)
	}

	// Status filter
	if query.Status != "" {
		whereConditions = append(whereConditions, "status = ?")
		whereArgs = append(whereArgs, query.Status)
	}

	// Tags filter
	if len(query.Tags) > 0 {
		tagConditions := make([]string, len(query.Tags))
		for i, tag := range query.Tags {
			tagConditions[i] = "tags ILIKE ?"
			whereArgs = append(whereArgs, "%"+tag+"%")
		}
		whereConditions = append(whereConditions, "("+strings.Join(tagConditions, " OR ")+")")
	}

	// Combine all where conditions
	whereClause := strings.Join(whereConditions, " AND ")
	baseQuery = baseQuery.Where(whereClause, whereArgs...)

	// Get total count
	var total int64
	countQuery := baseQuery.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Build order by clause
	orderBy := r.buildOrderBy(query.SortBy, query.SortOrder, query.Query)

	// Apply pagination
	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Execute the main query
	var results []models.SearchIndex
	queryErr := baseQuery.
		Select("*").
		Order(orderBy).
		Limit(limit).
		Offset(offset).
		Find(&results).Error

	if queryErr != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", queryErr)
	}

	// Convert to search results with relevance scores
	searchResults := make([]models.SearchResult, len(results))
	for i, result := range results {
		relevanceScore := r.calculateRelevanceScore(result, query.Query)
		searchResults[i] = result.ToSearchResult(relevanceScore)
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	queryTime := time.Since(startTime).Milliseconds()

	response := &models.SearchResponse{
		Results:    searchResults,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		QueryTime:  queryTime,
	}

	return response, nil
}

// buildOrderBy constructs the ORDER BY clause based on sort parameters
func (r *SearchRepository) buildOrderBy(sortBy, sortOrder, query string) string {
	if sortBy == "" {
		if query != "" {
			// Default to relevance for text searches
			return "ts_rank(search_vector, plainto_tsquery('english', ?)) DESC"
		}
		// Default to created_at for non-text searches
		return "created_at DESC"
	}

	switch sortBy {
	case "relevance":
		if query != "" {
			return "ts_rank(search_vector, plainto_tsquery('english', ?)) DESC"
		}
		return "created_at DESC"
	case "price_asc":
		return "price ASC"
	case "price_desc":
		return "price DESC"
	case "rating_desc":
		return "rating DESC, review_count DESC"
	case "created_at_desc":
		return "created_at DESC"
	case "view_count_desc":
		return "view_count DESC"
	default:
		return "created_at DESC"
	}
}

// calculateRelevanceScore calculates relevance score for search results
func (r *SearchRepository) calculateRelevanceScore(result models.SearchIndex, query string) float64 {
	if query == "" {
		return 1.0
	}

	// Simple relevance calculation based on text matching
	// In production, you might want to use more sophisticated algorithms
	score := 0.0
	
	// Title matches are most important
	if strings.Contains(strings.ToLower(result.Title), strings.ToLower(query)) {
		score += 0.8
	}
	
	// Description matches
	if strings.Contains(strings.ToLower(result.Description), strings.ToLower(query)) {
		score += 0.5
	}
	
	// Category matches
	if strings.Contains(strings.ToLower(result.Category), strings.ToLower(query)) {
		score += 0.3
	}
	
	// Tags matches
	if strings.Contains(strings.ToLower(result.Tags), strings.ToLower(query)) {
		score += 0.2
	}
	
	// Boost popular items
	if result.ViewCount > 100 {
		score += 0.1
	}
	
	if result.Rating > 4.0 {
		score += 0.1
	}
	
	if score == 0 {
		score = 0.1 // Minimum score for any result
	}
	
	return score
}

// IndexEntity adds or updates an entity in the search index
func (r *SearchRepository) IndexEntity(ctx context.Context, req *models.IndexRequest) error {
	// Convert tags to JSON string
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	// Create search index record
	searchIndex := &models.SearchIndex{
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		Title:       req.Title,
		Description: req.Description,
		Tags:        string(tagsJSON),
		Category:    req.Category,
		SubCategory: req.SubCategory,
		Price:       req.Price,
		Status:      req.Status,
		Location:    req.Location,
		SellerID:    req.SellerID,
		SellerName:  req.SellerName,
		Rating:      req.Rating,
		ReviewCount: req.ReviewCount,
		ViewCount:   req.ViewCount,
		LikesCount:  req.LikesCount,
		IsActive:    req.IsActive,
	}

	// Use UPSERT to handle both insert and update
	err = r.db.WithContext(ctx).
		Exec(`
			INSERT INTO search_indices (
				entity_type, entity_id, title, description, tags, category, sub_category,
				price, status, location, seller_id, seller_name, rating, review_count,
				view_count, likes_count, is_active, created_at, updated_at, indexed_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), NOW())
			ON CONFLICT (entity_type, entity_id) DO UPDATE SET
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				tags = EXCLUDED.tags,
				category = EXCLUDED.category,
				sub_category = EXCLUDED.sub_category,
				price = EXCLUDED.price,
				status = EXCLUDED.status,
				location = EXCLUDED.location,
				seller_id = EXCLUDED.seller_id,
				seller_name = EXCLUDED.seller_name,
				rating = EXCLUDED.rating,
				review_count = EXCLUDED.review_count,
				view_count = EXCLUDED.view_count,
				likes_count = EXCLUDED.likes_count,
				is_active = EXCLUDED.is_active,
				updated_at = NOW(),
				indexed_at = NOW()
		`,
			searchIndex.EntityType, searchIndex.EntityID, searchIndex.Title, searchIndex.Description,
			searchIndex.Tags, searchIndex.Category, searchIndex.SubCategory, searchIndex.Price,
			searchIndex.Status, searchIndex.Location, searchIndex.SellerID, searchIndex.SellerName,
			searchIndex.Rating, searchIndex.ReviewCount, searchIndex.ViewCount, searchIndex.LikesCount,
			searchIndex.IsActive,
		).Error

	if err != nil {
		return fmt.Errorf("failed to index entity: %w", err)
	}

	return nil
}

// BulkIndexEntities adds or updates multiple entities in the search index
func (r *SearchRepository) BulkIndexEntities(ctx context.Context, req *models.BulkIndexRequest) error {
	// Start transaction
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, entity := range req.Entities {
		tagsJSON, err := json.Marshal(entity.Tags)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to marshal tags: %w", err)
		}

		err = tx.Exec(`
			INSERT INTO search_indices (
				entity_type, entity_id, title, description, tags, category, sub_category,
				price, status, location, seller_id, seller_name, rating, review_count,
				view_count, likes_count, is_active, created_at, updated_at, indexed_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), NOW())
			ON CONFLICT (entity_type, entity_id) DO UPDATE SET
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				tags = EXCLUDED.tags,
				category = EXCLUDED.category,
				sub_category = EXCLUDED.sub_category,
				price = EXCLUDED.price,
				status = EXCLUDED.status,
				location = EXCLUDED.location,
				seller_id = EXCLUDED.seller_id,
				seller_name = EXCLUDED.seller_name,
				rating = EXCLUDED.rating,
				review_count = EXCLUDED.review_count,
				view_count = EXCLUDED.view_count,
				likes_count = EXCLUDED.likes_count,
				is_active = EXCLUDED.is_active,
				updated_at = NOW(),
				indexed_at = NOW()
		`,
			entity.EntityType, entity.EntityID, entity.Title, entity.Description,
			string(tagsJSON), entity.Category, entity.SubCategory, entity.Price,
			entity.Status, entity.Location, entity.SellerID, entity.SellerName,
			entity.Rating, entity.ReviewCount, entity.ViewCount, entity.LikesCount,
			entity.IsActive,
		).Error

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to bulk index entity %s: %w", entity.EntityID, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit bulk index transaction: %w", err)
	}

	return nil
}

// DeleteFromIndex removes an entity from the search index
func (r *SearchRepository) DeleteFromIndex(ctx context.Context, req *models.DeleteIndexRequest) error {
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", req.EntityType, req.EntityID).
		Delete(&models.SearchIndex{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete from index: %w", err)
	}

	return nil
}

// GetSearchSuggestions returns search suggestions based on query
func (r *SearchRepository) GetSearchSuggestions(ctx context.Context, query string, limit int) ([]models.SearchSuggestion, error) {
	if query == "" {
		return []models.SearchSuggestion{}, nil
	}

	var suggestions []models.SearchSuggestion
	
	// Get popular queries that start with the given query
	err := r.db.WithContext(ctx).
		Table("search_analytics").
		Select("query, COUNT(*) as frequency, MAX(created_at) as last_seen").
		Where("query ILIKE ? AND query != ?", query+"%", query).
		Group("query").
		Order("frequency DESC, last_seen DESC").
		Limit(limit).
		Find(&suggestions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get search suggestions: %w", err)
	}

	// Add confidence scores
	for i := range suggestions {
		suggestions[i].Type = "query"
		suggestions[i].Confidence = float64(suggestions[i].Frequency) / 100.0 // Normalize
	}

	return suggestions, nil
}

// GetSearchFacets returns faceted search results
func (r *SearchRepository) GetSearchFacets(ctx context.Context, query *models.SearchQuery) (*models.SearchFacets, error) {
	// Build the base query (same as in Search method)
	baseQuery := r.db.WithContext(ctx).Table("search_indices")
	
	whereConditions := []string{}
	whereArgs := []interface{}{}

	// Always filter by active status
	whereConditions = append(whereConditions, "is_active = ?")
	whereArgs = append(whereArgs, true)

	// Apply the same filters as in Search method
	if query.EntityType != "" && query.EntityType != "all" {
		whereConditions = append(whereConditions, "entity_type = ?")
		whereArgs = append(whereArgs, query.EntityType)
	}

	if query.Query != "" {
		if query.FuzzySearch {
			whereConditions = append(whereConditions, "(title % ? OR description % ? OR category % ? OR tags % ?)")
			whereArgs = append(whereArgs, query.Query, query.Query, query.Query, query.Query)
		} else {
			whereConditions = append(whereConditions, "search_vector @@ plainto_tsquery('english', ?)")
			whereArgs = append(whereArgs, query.Query)
		}
	}

	// Apply other filters...
	whereClause := strings.Join(whereConditions, " AND ")
	baseQuery = baseQuery.Where(whereClause, whereArgs...)

	facets := &models.SearchFacets{
		Categories:    make(map[string]int64),
		SubCategories: make(map[string]int64),
		Locations:     make(map[string]int64),
		Ratings:       make(map[string]int64),
	}

	// Get category facets
	var categoryFacets []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	
	err := baseQuery.Select("category, COUNT(*) as count").
		Where("category != ''").
		Group("category").
		Order("count DESC").
		Limit(10).
		Find(&categoryFacets).Error

	if err == nil {
		for _, facet := range categoryFacets {
			facets.Categories[facet.Category] = facet.Count
		}
	}

	// Get location facets
	var locationFacets []struct {
		Location string `json:"location"`
		Count    int64  `json:"count"`
	}
	
	err = baseQuery.Select("location, COUNT(*) as count").
		Where("location != ''").
		Group("location").
		Order("count DESC").
		Limit(10).
		Find(&locationFacets).Error

	if err == nil {
		for _, facet := range locationFacets {
			facets.Locations[facet.Location] = facet.Count
		}
	}

	// Get rating facets
	var ratingFacets []struct {
		Rating string `json:"rating"`
		Count  int64  `json:"count"`
	}
	
	err = baseQuery.Select("CASE WHEN rating >= 4.5 THEN '4.5+' WHEN rating >= 4.0 THEN '4.0+' WHEN rating >= 3.5 THEN '3.5+' WHEN rating >= 3.0 THEN '3.0+' ELSE '2.5+' END as rating, COUNT(*) as count").
		Where("rating > 0").
		Group("rating").
		Order("rating DESC").
		Find(&ratingFacets).Error

	if err == nil {
		for _, facet := range ratingFacets {
			facets.Ratings[facet.Rating] = facet.Count
		}
	}

	// Get price range facets
	var priceRanges []struct {
		Min  float64 `json:"min"`
		Max  float64 `json:"max"`
		Count int64  `json:"count"`
	}
	
	err = baseQuery.Select(`
		CASE 
			WHEN price < 50 THEN 0
			WHEN price < 100 THEN 50
			WHEN price < 200 THEN 100
			WHEN price < 500 THEN 200
			WHEN price < 1000 THEN 500
			ELSE 1000
		END as min,
		CASE 
			WHEN price < 50 THEN 50
			WHEN price < 100 THEN 100
			WHEN price < 200 THEN 200
			WHEN price < 500 THEN 500
			WHEN price < 1000 THEN 1000
			ELSE 999999
		END as max,
		COUNT(*) as count
	`).
		Where("price > 0").
		Group("min, max").
		Order("min").
		Find(&priceRanges).Error

	if err == nil {
		for _, pr := range priceRanges {
			facets.PriceRanges = append(facets.PriceRanges, models.PriceRange{
				Min:  pr.Min,
				Max:  pr.Max,
				Count: pr.Count,
			})
		}
	}

	return facets, nil
}

// RecordSearchAnalytics records search analytics data
func (r *SearchRepository) RecordSearchAnalytics(ctx context.Context, analytics *models.SearchAnalytics) error {
	err := r.db.WithContext(ctx).Create(analytics).Error
	if err != nil {
		return fmt.Errorf("failed to record search analytics: %w", err)
	}
	return nil
}

// GetPopularSearches returns popular search queries
func (r *SearchRepository) GetPopularSearches(ctx context.Context, limit int) ([]models.PopularSearch, error) {
	var popularSearches []models.PopularSearch
	
	err := r.db.WithContext(ctx).
		Table("search_analytics").
		Select("query, COUNT(*) as frequency, MAX(created_at) as last_seen").
		Where("created_at > ?", time.Now().AddDate(0, 0, -30)). // Last 30 days
		Group("query").
		Order("frequency DESC, last_seen DESC").
		Limit(limit).
		Find(&popularSearches).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get popular searches: %w", err)
	}

	return popularSearches, nil
}