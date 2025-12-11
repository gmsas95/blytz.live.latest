package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SearchIndex represents the main search index table
type SearchIndex struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
 EntityType   string    `gorm:"type:varchar(50);not null;index" json:"entity_type"`     // product, auction, user
	EntityID     string    `gorm:"type:varchar(255);not null;index" json:"entity_id"`       // Reference to original entity
	Title        string    `gorm:"type:varchar(500);not null" json:"title"`
	Description  string    `gorm:"type:text" json:"description"`
	Tags         string    `gorm:"type:text" json:"tags"`                                   // JSON array of tags
	Category     string    `gorm:"type:varchar(100);index" json:"category"`
	SubCategory  string    `gorm:"type:varchar(100);index" json:"sub_category"`
	Price        float64   `gorm:"type:decimal(10,2);index" json:"price"`
	Status       string    `gorm:"type:varchar(50);index" json:"status"`                   // active, inactive, sold, etc
	Location     string    `gorm:"type:varchar(255);index" json:"location"`
	SellerID     string    `gorm:"type:varchar(255);index" json:"seller_id"`
	SellerName   string    `gorm:"type:varchar(255);index" json:"seller_name"`
	Rating       float32   `gorm:"type:decimal(3,2);index" json:"rating"`
	ReviewCount  int       `gorm:"index" json:"review_count"`
	ViewCount    int64     `gorm:"index" json:"view_count"`
	LikesCount   int64     `gorm:"index" json:"likes_count"`
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`
	SearchVector string    `gorm:"type:tsvector" json:"-"`                                 // PostgreSQL full-text search vector
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	IndexedAt    time.Time `gorm:"autoCreateTime" json:"indexed_at"`
}

// SearchQuery represents a search request
type SearchQuery struct {
	Query        string   `json:"query" binding:"required,min=1,max=500"`
	EntityType   string   `json:"entity_type" binding:"omitempty,oneof=product auction user all"`
	Category     string   `json:"category"`
	SubCategory  string   `json:"sub_category"`
	MinPrice     float64  `json:"min_price"`
	MaxPrice     float64  `json:"max_price"`
	Location     string   `json:"location"`
	SellerID     string   `json:"seller_id"`
	Status       string   `json:"status"`
	Tags         []string `json:"tags"`
	SortBy       string   `json:"sort_by" binding:"omitempty,oneof=relevance price_asc price_desc rating_desc created_at_desc view_count_desc"`
	SortOrder    string   `json:"sort_order" binding:"omitempty,oneof=asc desc"`
	Page         int      `json:"page" binding:"min=1"`
	Limit        int      `json:"limit" binding:"min=1,max=100"`
	FuzzySearch  bool     `json:"fuzzy_search"`
}

// SearchResponse represents the search response
type SearchResponse struct {
	Results      []SearchResult `json:"results"`
	Total        int64          `json:"total"`
	Page         int            `json:"page"`
	Limit        int            `json:"limit"`
	TotalPages   int            `json:"total_pages"`
	QueryTime    int64          `json:"query_time_ms"`
	Suggestions  []string       `json:"suggestions,omitempty"`
	Facets       *SearchFacets  `json:"facets,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID           string    `json:"id"`
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Tags         []string  `json:"tags"`
	Category     string    `json:"category"`
	SubCategory  string    `json:"sub_category"`
	Price        float64   `json:"price"`
	Status       string    `json:"status"`
	Location     string    `json:"location"`
	SellerID     string    `json:"seller_id"`
	SellerName   string    `json:"seller_name"`
	Rating       float32   `json:"rating"`
	ReviewCount  int       `json:"review_count"`
	ViewCount    int64     `json:"view_count"`
	LikesCount   int64     `json:"likes_count"`
	RelevanceScore float64 `json:"relevance_score"`
	CreatedAt    time.Time `json:"created_at"`
}

// SearchFacets represents search facets for filtering
type SearchFacets struct {
	Categories    map[string]int64 `json:"categories"`
	SubCategories map[string]int64 `json:"sub_categories"`
	PriceRanges   []PriceRange    `json:"price_ranges"`
	Locations     map[string]int64 `json:"locations"`
	Ratings       map[string]int64 `json:"ratings"`
}

// PriceRange represents a price range facet
type PriceRange struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Count int64  `json:"count"`
}

// SearchSuggestion represents a search suggestion
type SearchSuggestion struct {
	Text        string `json:"text"`
	Type        string `json:"type"`        // query, category, product
	Frequency   int64  `json:"frequency"`
	Confidence  float64 `json:"confidence"`
}

// SearchAnalytics represents search analytics data
type SearchAnalytics struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Query       string    `gorm:"type:varchar(500);not null;index" json:"query"`
	UserID      string    `gorm:"type:varchar(255);index" json:"user_id"`
	Results     int       `gorm:"index" json:"results"`
	Clicks      int       `gorm:"index" json:"clicks"`
	IPAddress   string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent   string    `gorm:"type:text" json:"user_agent"`
	SessionID   string    `gorm:"type:varchar(255);index" json:"session_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// PopularSearch represents popular search queries
type PopularSearch struct {
	Query     string `json:"query"`
	Frequency int64  `json:"frequency"`
	LastSeen  time.Time `json:"last_seen"`
}

// TableName returns the table name for SearchIndex
func (SearchIndex) TableName() string {
	return "search_indices"
}

// TableName returns the table name for SearchAnalytics
func (SearchAnalytics) TableName() string {
	return "search_analytics"
}

// BeforeCreate GORM hook to generate search vector
func (s *SearchIndex) BeforeCreate(tx *gorm.DB) error {
	// This will be handled by database triggers for better performance
	// But we can also do it here if needed
	return nil
}

// BeforeUpdate GORM hook to update search vector
func (s *SearchIndex) BeforeUpdate(tx *gorm.DB) error {
	// This will be handled by database triggers for better performance
	// But we can also do it here if needed
	return nil
}

// ToSearchResult converts SearchIndex to SearchResult
func (s *SearchIndex) ToSearchResult(relevanceScore float64) SearchResult {
	return SearchResult{
		ID:             s.ID,
		EntityType:     s.EntityType,
		EntityID:       s.EntityID,
		Title:          s.Title,
		Description:    s.Description,
		Tags:           parseTags(s.Tags),
		Category:       s.Category,
		SubCategory:    s.SubCategory,
		Price:          s.Price,
		Status:         s.Status,
		Location:       s.Location,
		SellerID:       s.SellerID,
		SellerName:     s.SellerName,
		Rating:         s.Rating,
		ReviewCount:    s.ReviewCount,
		ViewCount:      s.ViewCount,
		LikesCount:     s.LikesCount,
		RelevanceScore: relevanceScore,
		CreatedAt:      s.CreatedAt,
	}
}

// parseTags is a helper function to parse tags from JSON string
func parseTags(tagsJSON string) []string {
	// In a real implementation, you'd parse this from JSON
	// For now, return empty slice
	return []string{}
}

// IndexRequest represents a request to index an entity
type IndexRequest struct {
	EntityType  string  `json:"entity_type" binding:"required,oneof=product auction user"`
	EntityID    string  `json:"entity_id" binding:"required"`
	Title       string  `json:"title" binding:"required,min=1,max=500"`
	Description string  `json:"description"`
	Tags        []string `json:"tags"`
	Category    string  `json:"category"`
	SubCategory string  `json:"sub_category"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
	Location    string  `json:"location"`
	SellerID    string  `json:"seller_id"`
	SellerName  string  `json:"seller_name"`
	Rating      float32 `json:"rating"`
	ReviewCount int     `json:"review_count"`
	ViewCount   int64   `json:"view_count"`
	LikesCount  int64   `json:"likes_count"`
	IsActive    bool    `json:"is_active"`
}

// BulkIndexRequest represents a request to bulk index multiple entities
type BulkIndexRequest struct {
	Entities []IndexRequest `json:"entities" binding:"required,min=1,max=1000"`
}

// DeleteIndexRequest represents a request to delete an entity from search index
type DeleteIndexRequest struct {
	EntityType string `json:"entity_type" binding:"required,oneof=product auction user"`
	EntityID   string `json:"entity_id" binding:"required"`
}