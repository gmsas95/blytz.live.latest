package models

import (
	"time"
)

// Auction Status Constants
const (
	AuctionStatusScheduled = "scheduled"
	AuctionStatusActive    = "active"
	AuctionStatusEnded     = "ended"
	AuctionStatusCancelled = "cancelled"
)

// Auction Type Constants
const (
	AuctionTypeLive       = "live"
	AuctionTypeScheduled  = "scheduled"
)

// Bid Status Constants
const (
	BidStatusActive     = "active"
	BidStatusWinning   = "winning"
	BidStatusOutbid    = "outbid"
	BidStatusWinning   = "winning"
	BidStatusCancelled  = "cancelled"
)

// Enhanced Auction Model with more fields
type Auction struct {
	AuctionID       string          `json:"auction_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProductID       string          `json:"product_id" gorm:"not null;index"`
	SellerID        string          `json:"seller_id" gorm:"not null;index"`
	Title           string          `json:"title" gorm:"not null"`
	Description     string          `json:"description" gorm:"type:text"`
	StartingPrice   float64         `json:"starting_price" gorm:"not null"`
	CurrentPrice    float64         `json:"current_price" gorm:"not null;default:0"`
	ReservePrice    float64         `json:"reserve_price" gorm:"default:0"`
	MinBidIncrement float64         `json:"min_bid_increment" gorm:"not null;default:1"`
	BuyItNowPrice  float64         `json:"buy_it_now_price" gorm:"default:0"`
	TotalBids       int             `json:"total_bids" gorm:"default:0"`
	TotalViews      int             `json:"total_views" gorm:"default:0"`
	TotalWatchers  int             `json:"total_watchers" gorm:"default:0"`
	StartTime       time.Time       `json:"start_time" gorm:"not null"`
	EndTime         time.Time       `json:"end_time" gorm:"not null"`
	Status          string          `json:"status" gorm:"not null;default:scheduled"`
	Type            string          `json:"type" gorm:"not null;default:scheduled"`
	IsActive        bool            `json:"is_active" gorm:"default:false"`
	AutoExtend      bool            `json:"auto_extend" gorm:"default:false"`
	ExtendMinutes   int             `json:"extend_minutes" gorm:"default:5"`
	Featured        bool            `json:"featured" gorm:"default:false"`
	ShippingInfo    string          `json:"shipping_info"`
	Condition       string          `json:"condition"`
	Category        string          `json:"category"`
	Subcategory     string          `json:"subcategory"`
	Tags            string          `json:"tags" gorm:"type:text"` // JSON array
	Metadata        string          `json:"metadata" gorm:"type:text"` // Additional data
	WinningBidID   *string         `json:"winning_bid_id" gorm:"index"`
	WinningUserID   *string         `json:"winning_user_id" gorm:"index"`
	FinalPrice      float64         `json:"final_price" gorm:"default:0"`
	EndReason       string          `json:"end_reason" gorm:""` // ended, cancelled, reserve_not_met, sold
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	
	// Relationships
	Bids            []Bid           `json:"bids,omitempty" gorm:"foreignKey:AuctionID"`
	Images          []AuctionImage  `json:"images,omitempty" gorm:"foreignKey:AuctionID"`
	Watchers        []AuctionWatcher `json:"watchers,omitempty" gorm:"foreignKey:AuctionID"`
}

// Enhanced Bid Model
type Bid struct {
	BidID       string    `json:"bid_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AuctionID   string    `json:"auction_id" gorm:"not null;index"`
	BidderID    string    `json:"bidder_id" gorm:"not null;index"`
	BidderName  string    `json:"bidder_name" gorm:"not null"`
	Amount      float64   `json:"amount" gorm:"not null"`
	IsWinning   bool      `json:"is_winning" gorm:"default:false"`
	IsAutoBid   bool      `json:"is_auto_bid" gorm:"default:false"`
	MaxAutoBid  float64   `json:"max_auto_bid" gorm:"default:0"`
	Status      string    `json:"status" gorm:"default:active"`
	BidTime     time.Time `json:"bid_time" gorm:"not null"`
	IPAddress   string    `json:"ip_address" gorm:"not null"`
	UserAgent   string    `json:"user_agent" gorm:""`
	Notes       string    `json:"notes" gorm:""`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Relationships
	Bidder      *User     `json:"bidder,omitempty" gorm:"foreignKey:BidderID"`
}

// Enhanced Auction Image Model
type AuctionImage struct {
	ImageID    string `json:"image_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AuctionID  string `json:"auction_id" gorm:"not null;index"`
	ImageURL   string `json:"image_url" gorm:"not null"`
	AltText    string `json:"alt_text" gorm:""`
	Order      int    `json:"order" gorm:"not null;default:0"`
	IsMain     bool   `json:"is_main" gorm:"default:false"`
	FileSize   int64  `json:"file_size" gorm:"default:0"`
	FileType   string `json:"file_type" gorm:""`
	Width      int    `json:"width" gorm:"default:0"`
	Height     int    `json:"height" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Auction Watcher Model for tracking users watching auctions
type AuctionWatcher struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AuctionID  string    `json:"auction_id" gorm:"not null;index"`
	UserID     string    `json:"user_id" gorm:"not null;index"`
	WatchStart time.Time `json:"watch_start" gorm:"not null"`
	WatchEnd   time.Time `json:"watch_end"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Auction Activity Log Model
type AuctionActivity struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AuctionID  string    `json:"auction_id" gorm:"not null;index"`
	UserID     string    `json:"user_id" gorm:"not null;index"`
	ActivityType string   `json:"activity_type" gorm:"not null"` // bid, view, watch, end, cancel
	ActivityData string   `json:"activity_data" gorm:"type:text"` // JSON data
	IPAddress   string    `json:"ip_address" gorm:""`
	UserAgent   string    `json:"user_agent" gorm:""`
	CreatedAt  time.Time `json:"created_at"`
	
	// Relationships
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// User Model for auction service (simplified)
type User struct {
	UserID      string    `json:"user_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username    string    `json:"username" gorm:"not null;uniqueIndex"`
	Email       string    `json:"email" gorm:"not null;uniqueIndex"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Avatar      string    `json:"avatar"`
	Rating      float64   `json:"rating" gorm:"default:5"`
	TotalBids   int       `json:"total_bids" gorm:"default:0"`
	TotalWins   int       `json:"total_wins" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	IsVerified  bool      `json:"is_verified" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Request Models

// CreateAuctionRequest represents auction creation request
type CreateAuctionRequest struct {
	ProductID       string    `json:"product_id" binding:"required"`
	Title           string    `json:"title" binding:"required,min=3,max=200"`
	Description     string    `json:"description" binding:"required,min=10,max=2000"`
	StartingPrice   float64   `json:"starting_price" binding:"required,min=0"`
	ReservePrice    float64   `json:"reserve_price" binding:"min=0"`
	BuyItNowPrice  float64   `json:"buy_it_now_price" binding:"min=0"`
	MinBidIncrement float64   `json:"min_bid_increment" binding:"required,min=0.01"`
	StartTime       time.Time `json:"start_time" binding:"required"`
	EndTime         time.Time `json:"end_time" binding:"required"`
	Type            string    `json:"type" binding:"required,oneof=live scheduled"`
	Images          []string  `json:"images" binding:"required,min=1,dive,url"`
	AutoExtend      bool      `json:"auto_extend"`
	ExtendMinutes   int       `json:"extend_minutes" binding:"min=1,max=30"`
	Featured        bool      `json:"featured"`
	ShippingInfo    string    `json:"shipping_info" binding:"max=500"`
	Condition       string    `json:"condition" binding:"max=50"`
	Category        string    `json:"category" binding:"required"`
	Subcategory     string    `json:"subcategory"`
	Tags            []string  `json:"tags"`
	Metadata        string    `json:"metadata"`
}

// UpdateAuctionRequest represents auction update request
type UpdateAuctionRequest struct {
	Title           string    `json:"title" binding:"omitempty,min=3,max=200"`
	Description     string    `json:"description" binding:"omitempty,min=10,max=2000"`
	ReservePrice    float64   `json:"reserve_price" binding:"omitempty,min=0"`
	BuyItNowPrice  float64   `json:"buy_it_now_price" binding:"omitempty,min=0"`
	MinBidIncrement float64   `json:"min_bid_increment" binding:"omitempty,min=0.01"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Images          []string  `json:"images" binding:"omitempty,dive,url"`
	AutoExtend      bool      `json:"auto_extend"`
	ExtendMinutes   int       `json:"extend_minutes" binding:"omitempty,min=1,max=30"`
	Featured        bool      `json:"featured"`
	ShippingInfo    string    `json:"shipping_info" binding:"omitempty,max=500"`
	Condition       string    `json:"condition" binding:"omitempty,max=50"`
	Category        string    `json:"category"`
	Subcategory     string    `json:"subcategory"`
	Tags            []string  `json:"tags"`
	Metadata        string    `json:"metadata"`
}

// PlaceBidRequest represents bid placement request
type PlaceBidRequest struct {
	Amount     float64 `json:"amount" binding:"required,min=0"`
	IsAutoBid  bool    `json:"is_auto_bid"`
	MaxAutoBid float64 `json:"max_auto_bid" binding:"omitempty,min=0"`
}

// SearchAuctionsRequest represents auction search request
type SearchAuctionsRequest struct {
	Query        string    `json:"query" form:"query"`
	Category     string    `json:"category" form:"category"`
	Subcategory  string    `json:"subcategory" form:"subcategory"`
	MinPrice     float64   `json:"min_price" form:"min_price"`
	MaxPrice     float64   `json:"max_price" form:"max_price"`
	Status       string    `json:"status" form:"status"`
	Type         string    `json:"type" form:"type"`
	SellerID     string    `json:"seller_id" form:"seller_id"`
	Condition    string    `json:"condition" form:"condition"`
	Featured     bool      `json:"featured" form:"featured"`
	BuyItNow     bool      `json:"buy_it_now" form:"buy_it_now"`
	EndingSoon   bool      `json:"ending_soon" form:"ending_soon"`
	SortBy       string    `json:"sort_by" form:"sort_by"` // price, end_time, views, bids
	SortOrder    string    `json:"sort_order" form:"sort_order"` // asc, desc
	Page         int       `json:"page" form:"page"`
	Limit        int       `json:"limit" form:"limit"`
}

// Auction Statistics
type AuctionStats struct {
	TotalAuctions  int64   `json:"total_auctions"`
	ActiveAuctions int64   `json:"active_auctions"`
	TotalBids      int64   `json:"total_bids"`
	TotalRevenue    float64 `json:"total_revenue"`
	AvgFinalPrice  float64 `json:"avg_final_price"`
	TopCategories  []string `json:"top_categories"`
}

// Response Models

// AuctionResponse represents auction response
type AuctionResponse struct {
	Auction     Auction `json:"auction"`
	TimeRemaining string  `json:"time_remaining"`
	BidCount     int     `json:"bid_count"`
	IsWatched    bool    `json:"is_watched"`
	CanBid       bool    `json:"can_bid"`
	MinNextBid   float64 `json:"min_next_bid"`
}

// BidResponse represents bid response
type BidResponse struct {
	Bid           Bid     `json:"bid"`
	IsOutbid      bool    `json:"is_outbid"`
	NewHighBid    float64 `json:"new_high_bid"`
	MinNextBid    float64 `json:"min_next_bid"`
	TimeRemaining string  `json:"time_remaining"`
}

// AuctionsResponse represents auctions list response
type AuctionsResponse struct {
	Auctions    []Auction `json:"auctions"`
	Total       int64     `json:"total"`
	Page        int       `json:"page"`
	Limit       int       `json:"limit"`
	TotalPages  int       `json:"total_pages"`
	HasNext     bool      `json:"has_next"`
}

// BidsResponse represents bids list response
type BidsResponse struct {
	Bids    []Bid `json:"bids"`
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	Limit   int   `json:"limit"`
	HasNext bool  `json:"has_next"`
}

// WebSocket Messages

// BidUpdateMessage represents bid update for WebSocket
type BidUpdateMessage struct {
	Type        string    `json:"type"` // new_bid, bid_outbid, auction_end
	AuctionID   string    `json:"auction_id"`
	Bid         Bid       `json:"bid,omitempty"`
	NewPrice    float64   `json:"new_price"`
	TimeRemaining string    `json:"time_remaining"`
	TotalBids   int       `json:"total_bids"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// AuctionUpdateMessage represents auction update for WebSocket
type AuctionUpdateMessage struct {
	Type          string  `json:"type"` // auction_start, auction_end, auction_cancel
	AuctionID     string  `json:"auction_id"`
	Status        string  `json:"status"`
	CurrentPrice  float64 `json:"current_price"`
	TotalBids     int     `json:"total_bids"`
	WinnerUserID  string  `json:"winner_user_id,omitempty"`
	FinalPrice    float64 `json:"final_price,omitempty"`
	EndReason     string  `json:"end_reason,omitempty"`
	Message       string  `json:"message"`
	Timestamp     time.Time `json:"timestamp"`
}

// Helper Methods

// BeforeCreate hook for Auction
func (a *Auction) BeforeCreate(tx *gorm.DB) error {
	if a.AuctionID == "" {
		a.AuctionID = generateUUID()
	}
	if a.CurrentPrice == 0 {
		a.CurrentPrice = a.StartingPrice
	}
	return nil
}

// BeforeCreate hook for Bid
func (b *Bid) BeforeCreate(tx *gorm.DB) error {
	if b.BidID == "" {
		b.BidID = generateUUID()
	}
	if b.Status == "" {
		b.Status = BidStatusActive
	}
	return nil
}

// IsValidBid checks if bid amount is valid
func (a *Auction) IsValidBid(amount float64) bool {
	minNextBid := a.CurrentPrice + a.MinBidIncrement
	return amount >= minNextBid
}

// GetMinNextBid returns minimum next bid amount
func (a *Auction) GetMinNextBid() float64 {
	return a.CurrentPrice + a.MinBidIncrement
}

// IsEnded checks if auction has ended
func (a *Auction) IsEnded() bool {
	return a.Status == AuctionStatusEnded || a.Status == AuctionStatusCancelled
}

// IsActive checks if auction is active
func (a *Auction) IsActive() bool {
	return a.Status == AuctionStatusActive && a.IsActive
}

// GetTimeRemaining returns time remaining as string
func (a *Auction) GetTimeRemaining() string {
	if a.IsEnded() {
		return "ended"
	}
	
	now := time.Now()
	if now.Before(a.StartTime) {
		return "scheduled"
	}
	
	duration := a.EndTime.Sub(now)
	if duration <= 0 {
		return "ending"
	}
	
	return formatDuration(duration)
}

// SetTagsArray sets tags from string array
func (a *Auction) SetTagsArray(tags []string) {
	if len(tags) == 0 {
		a.Tags = ""
		return
	}
	data, _ := json.Marshal(tags)
	a.Tags = string(data)
}

// GetTagsArray returns tags as string array
func (a *Auction) GetTagsArray() []string {
	if a.Tags == "" {
		return []string{}
	}
	var tags []string
	json.Unmarshal([]byte(a.Tags), &tags)
	return tags
}

// Helper function to generate UUID (simplified)
func generateUUID() string {
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	if d <= time.Hour {
		return fmt.Sprintf("%dm", d/time.Minute)
	}
	return fmt.Sprintf("%dh%dm", d/time.Hour, (d%time.Hour)/time.Minute)
}