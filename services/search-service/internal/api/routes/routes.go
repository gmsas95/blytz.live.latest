package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/search-service/internal/api/handlers"
	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
)

// SetupRoutes configures all API routes for the search service
func SetupRoutes(router *gin.Engine, searchHandler *handlers.SearchHandler) {
	// Apply global middleware
	router.Use(shared_utils.RequestIDMiddleware())
	router.Use(shared_utils.LoggingMiddleware(nil))
	router.Use(gin.Recovery())

	// API versioning
	v1 := router.Group("/api/v1")
	{
		// Public search endpoints
		search := v1.Group("/search")
		{
			search.GET("", searchHandler.Search)                           // General search
			search.GET("/type/:type", searchHandler.SearchByType)          // Search by entity type
			search.GET("/suggestions", searchHandler.GetSearchSuggestions)  // Search suggestions
			search.GET("/facets", searchHandler.GetSearchFacets)          // Search facets
			search.GET("/popular", searchHandler.GetPopularSearches)       // Popular searches
		}

		// Index management endpoints (protected)
		index := v1.Group("/index")
		{
			index.POST("", searchHandler.IndexEntity)           // Index single entity
			index.POST("/bulk", searchHandler.BulkIndexEntities)  // Bulk index entities
			index.DELETE("", searchHandler.DeleteFromIndex)       // Delete from index
			index.DELETE("/type/:type/id/:id", searchHandler.DeleteByType) // Delete by type and ID
		}

		// Admin endpoints
		admin := v1.Group("/admin")
		{
			admin.POST("/reindex", searchHandler.ReindexAll)      // Full reindex
			admin.GET("/stats", searchHandler.GetSearchStats)       // Search statistics
			admin.GET("/status", searchHandler.GetIndexStatus)     // Index status
		}
	}

	// Health and status endpoints
	router.GET("/health", searchHandler.Health)      // Health check
	router.GET("/status", searchHandler.GetIndexStatus) // Service status

	// Metrics endpoint (for Prometheus)
	router.GET("/metrics", func(c *gin.Context) {
		// Return basic metrics for now
		shared_utils.SendJSON(c, http.StatusOK, gin.H{
			"service": "search-service",
			"status":  "running",
			"metrics": gin.H{
				"uptime": "unknown",
				"requests": "unknown",
				"errors": "unknown",
			},
		})
	})
}