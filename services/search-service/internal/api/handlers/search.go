package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/models"
	"github.com/gmsas95/blytz.live.latest/services/search-service/internal/services"
	shared_errors "github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
	shared_utils "github.com/gmsas95/blytz.live.latest/shared/pkg/utils"
)

type SearchHandler struct {
	searchService *services.SearchService
}

func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// Search handles search requests
func (h *SearchHandler) Search(c *gin.Context) {
	var query models.SearchQuery

	// Bind query parameters
	if err := c.ShouldBindQuery(&query); err != nil {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_QUERY_PARAMS", "Invalid query parameters"))
		return
	}

	// Set defaults
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.EntityType == "" {
		query.EntityType = "all"
	}

	// Perform search
	result, err := h.searchService.Search(c.Request.Context(), &query)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, result)
}

// IndexEntity handles indexing a single entity
func (h *SearchHandler) IndexEntity(c *gin.Context) {
	var req models.IndexRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_REQUEST_BODY", "Invalid request body"))
		return
	}

	err := h.searchService.IndexEntity(c.Request.Context(), &req)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusCreated, gin.H{
		"message": "Entity indexed successfully",
		"entity_type": req.EntityType,
		"entity_id": req.EntityID,
	})
}

// BulkIndexEntities handles bulk indexing of entities
func (h *SearchHandler) BulkIndexEntities(c *gin.Context) {
	var req models.BulkIndexRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_REQUEST_BODY", "Invalid request body"))
		return
	}

	// Validate entity count
	if len(req.Entities) == 0 {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("EMPTY_ENTITIES", "No entities to index"))
		return
	}
	if len(req.Entities) > 1000 {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("TOO_MANY_ENTITIES", "Maximum 1000 entities allowed per request"))
		return
	}

	err := h.searchService.BulkIndexEntities(c.Request.Context(), &req)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusCreated, gin.H{
		"message": "Entities indexed successfully",
		"count": len(req.Entities),
	})
}

// DeleteFromIndex handles deletion of entities from search index
func (h *SearchHandler) DeleteFromIndex(c *gin.Context) {
	var req models.DeleteIndexRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_REQUEST_BODY", "Invalid request body"))
		return
	}

	err := h.searchService.DeleteFromIndex(c.Request.Context(), &req)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Entity deleted from index successfully",
		"entity_type": req.EntityType,
		"entity_id": req.EntityID,
	})
}

// GetSearchSuggestions handles search suggestions requests
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("MISSING_QUERY", "Query parameter 'q' is required"))
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	suggestions, err := h.searchService.GetSearchSuggestions(c.Request.Context(), query, limit)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"query": query,
		"suggestions": suggestions,
	})
}

// GetSearchFacets handles faceted search requests
func (h *SearchHandler) GetSearchFacets(c *gin.Context) {
	var query models.SearchQuery

	// Bind query parameters (excluding pagination for facets)
	if err := c.ShouldBindQuery(&query); err != nil {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_QUERY_PARAMS", "Invalid query parameters"))
		return
	}

	// Set defaults
	if query.EntityType == "" {
		query.EntityType = "all"
	}

	facets, err := h.searchService.GetSearchFacets(c.Request.Context(), &query)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, facets)
}

// GetPopularSearches handles popular searches requests
func (h *SearchHandler) GetPopularSearches(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	popularSearches, err := h.searchService.GetPopularSearches(c.Request.Context(), limit)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"popular_searches": popularSearches,
	})
}

// ReindexAll handles full reindexing
func (h *SearchHandler) ReindexAll(c *gin.Context) {
	// This is a potentially long-running operation, so we'll run it asynchronously
	go func() {
		err := h.searchService.ReindexAll(c.Request.Context())
		if err != nil {
			// Log error - in production, you might want to use a proper logging system
			// or send notifications
			return
		}
	}()

	shared_utils.SendSuccessResponse(c, http.StatusAccepted, gin.H{
		"message": "Reindexing started. This may take several minutes.",
	})
}

// GetSearchStats handles search statistics requests
func (h *SearchHandler) GetSearchStats(c *gin.Context) {
	stats, err := h.searchService.GetSearchStats(c.Request.Context())
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Health handles health check requests
func (h *SearchHandler) Health(c *gin.Context) {
	// Perform basic health checks
	stats, err := h.searchService.GetSearchStats(c.Request.Context())
	if err != nil {
		shared_utils.SendErrorResponse(c, shared_errors.NewInternalError("HEALTH_CHECK_FAILED", "Health check failed"))
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"service":   "search-service",
		"status":    "healthy",
		"cache":     stats["cache"],
		"timestamp":  shared_utils.GetCurrentTimestamp(),
	})
}

// SearchByType handles search requests for specific entity types
func (h *SearchHandler) SearchByType(c *gin.Context) {
	entityType := c.Param("type")
	if entityType == "" {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("MISSING_ENTITY_TYPE", "Entity type is required"))
		return
	}

	// Validate entity type
	validTypes := []string{"product", "auction", "user", "all"}
	isValid := false
	for _, t := range validTypes {
		if t == entityType {
			isValid = true
			break
		}
	}
	if !isValid {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_ENTITY_TYPE", "Invalid entity type"))
		return
	}

	var query models.SearchQuery

	// Bind query parameters
	if err := c.ShouldBindQuery(&query); err != nil {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_QUERY_PARAMS", "Invalid query parameters"))
		return
	}

	// Set entity type from URL parameter
	query.EntityType = entityType

	// Set defaults
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	// Perform search
	result, err := h.searchService.Search(c.Request.Context(), &query)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, result)
}

// DeleteByType handles deletion of entities from search index by type and ID
func (h *SearchHandler) DeleteByType(c *gin.Context) {
	entityType := c.Param("type")
	entityID := c.Param("id")

	if entityType == "" || entityID == "" {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("MISSING_PARAMS", "Entity type and ID are required"))
		return
	}

	// Validate entity type
	validTypes := []string{"product", "auction", "user"}
	isValid := false
	for _, t := range validTypes {
		if t == entityType {
			isValid = true
			break
		}
	}
	if !isValid {
		shared_utils.SendErrorResponse(c, shared_errors.NewValidationError("INVALID_ENTITY_TYPE", "Invalid entity type"))
		return
	}

	req := &models.DeleteIndexRequest{
		EntityType: entityType,
		EntityID:   entityID,
	}

	err := h.searchService.DeleteFromIndex(c.Request.Context(), req)
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Entity deleted from index successfully",
		"entity_type": entityType,
		"entity_id": entityID,
	})
}

// GetIndexStatus handles index status requests
func (h *SearchHandler) GetIndexStatus(c *gin.Context) {
	// This would typically return information about the index status
	// For now, we'll return basic status information
	stats, err := h.searchService.GetSearchStats(c.Request.Context())
	if err != nil {
		shared_utils.SendErrorResponse(c, err)
		return
	}

	status := gin.H{
		"service": "search-service",
		"status":  "running",
		"cache":   stats["cache"],
		"features": []string{
			"full_text_search",
			"fuzzy_search",
			"faceted_search",
			"suggestions",
			"analytics",
			"caching",
		},
		"performance": gin.H{
			"max_connections": 200,
			"cache_ttl":       "5-15 minutes",
			"query_timeout":   "30 seconds",
		},
	}

	shared_utils.SendSuccessResponse(c, http.StatusOK, status)
}