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

// Auction struct
type Auction struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	ProductName  string    `json:"product_name"`
	Description  string    `json:"description"`
	StartingBid  float64   `json:"starting_bid"`
	CurrentBid   float64   `json:"current_bid"`
	BidCount     int       `json:"bid_count"`
	BuyNowPrice  *float64  `json:"buy_now_price,omitempty"`
	SellerID     string    `json:"seller_id"`
	CurrentBidder *string   `json:"current_bidder,omitempty"`
	Status       string    `json:"status"` // "active", "ended", "cancelled"
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Bid struct
type Bid struct {
	ID        string    `json:"id"`
	AuctionID string    `json:"auction_id"`
	BidderID  string    `json:"bidder_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"` // "active", "withdrawn", "accepted"
	CreatedAt time.Time `json:"created_at"`
}

// Request structs
type CreateAuctionRequest struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Description string  `json:"description"`
	StartingBid float64 `json:"starting_bid"`
	BuyNowPrice *float64 `json:"buy_now_price,omitempty"`
	SellerID    string  `json:"seller_id"`
	Duration    int     `json:"duration"` // Duration in hours
}

type PlaceBidRequest struct {
	AuctionID string  `json:"auction_id"`
	BidderID  string  `json:"bidder_id"`
	Amount    float64 `json:"amount"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Auction service with in-memory storage
type AuctionService struct {
	auctions []Auction
	bids     []Bid
	mu       sync.RWMutex
}

// New auction service
func NewAuctionService() *AuctionService {
	return &AuctionService{
		auctions: []Auction{
			{
				ID:          "auction-1",
				ProductID:   "product-1",
				ProductName: "Vintage Camera",
				Description: "Beautiful vintage camera from 1970s in excellent condition",
				StartingBid: 299.99,
				CurrentBid:  450.00,
				BidCount:    3,
				BuyNowPrice: &[]float64{899.99}[0],
				SellerID:    "seller-1",
				CurrentBidder: &[]string{"bidder-2"}[0],
				Status:      "active",
				StartTime:   time.Now().Add(-2 * time.Hour),
				EndTime:     time.Now().Add(6 * time.Hour),
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().Add(-10 * time.Minute),
			},
			{
				ID:          "auction-2",
				ProductID:   "product-2",
				ProductName: "Designer Handbag",
				Description: "Authentic designer handbag, barely used, comes with dust bag",
				StartingBid: 450.00,
				CurrentBid:  450.00,
				BidCount:    1,
				BuyNowPrice: &[]float64{1200.00}[0],
				SellerID:    "seller-2",
				CurrentBidder: &[]string{"bidder-1"}[0],
				Status:      "active",
				StartTime:   time.Now().Add(-1 * time.Hour),
				EndTime:     time.Now().Add(7 * time.Hour),
				CreatedAt:   time.Now().Add(-12 * time.Hour),
				UpdatedAt:   time.Now().Add(-5 * time.Minute),
			},
			{
				ID:          "auction-3",
				ProductID:   "product-3",
				ProductName: "Gaming Laptop",
				Description: "High-performance gaming laptop, RTX 3080, 32GB RAM, 1TB SSD",
				StartingBid: 1899.99,
				CurrentBid:  2100.00,
				BidCount:    5,
				BuyNowPrice: &[]float64{3500.00}[0],
				SellerID:    "seller-1",
				CurrentBidder: &[]string{"bidder-3"}[0],
				Status:      "active",
				StartTime:   time.Now().Add(-30 * time.Minute),
				EndTime:     time.Now().Add(7 * time.Hour).Add(30 * time.Minute),
				CreatedAt:   time.Now().Add(-6 * time.Hour),
				UpdatedAt:   time.Now().Add(-2 * time.Minute),
			},
		},
		bids: []Bid{
			{
				ID:        "bid-1",
				AuctionID: "auction-1",
				BidderID:  "bidder-1",
				Amount:    350.00,
				Status:    "withdrawn",
				CreatedAt: time.Now().Add(-45 * time.Minute),
			},
			{
				ID:        "bid-2",
				AuctionID: "auction-1",
				BidderID:  "bidder-2",
				Amount:    450.00,
				Status:    "active",
				CreatedAt: time.Now().Add(-10 * time.Minute),
			},
			{
				ID:        "bid-3",
				AuctionID: "auction-2",
				BidderID:  "bidder-1",
				Amount:    450.00,
				Status:    "active",
				CreatedAt: time.Now().Add(-5 * time.Minute),
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
func paginateAuctions(items []Auction, page, perPage int) ([]Auction, map[string]interface{}) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Auction{}, map[string]interface{}{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
		}
	}
	if end > total {
		end = total
	}

	return items[start:end], map[string]interface{}{
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	}
}

// Health check
func (s *AuctionService) health(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeAuctions := 0
	for _, auction := range s.auctions {
		if auction.Status == "active" {
			activeAuctions++
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Auction service is healthy and working!",
		Data: map[string]interface{}{
			"service":         "auction-service",
			"version":         "v2.0-working",
			"status":          "healthy",
			"timestamp":       time.Now(),
			"total_auctions":  len(s.auctions),
			"active_auctions": activeAuctions,
			"total_bids":      len(s.bids),
		},
	})
}

// Get all auctions
func (s *AuctionService) getAllAuctions(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter active auctions only
	var activeAuctions []Auction
	for _, auction := range s.auctions {
		if auction.Status == "active" {
			// Check if auction has ended
			if time.Now().After(auction.EndTime) {
				continue // Skip expired auctions
			}
			activeAuctions = append(activeAuctions, auction)
		}
	}

	// Apply search filter if provided
	search := strings.ToLower(r.URL.Query().Get("search"))
	if search != "" {
		var filteredAuctions []Auction
		for _, auction := range activeAuctions {
			if strings.Contains(strings.ToLower(auction.ProductName), search) ||
				strings.Contains(strings.ToLower(auction.Description), search) {
				filteredAuctions = append(filteredAuctions, auction)
			}
		}
		activeAuctions = filteredAuctions
	}

	// Apply status filter if provided
	if status := strings.ToLower(r.URL.Query().Get("status")); status != "" {
		var filteredAuctions []Auction
		for _, auction := range activeAuctions {
			if strings.Contains(strings.ToLower(auction.Status), status) {
				filteredAuctions = append(filteredAuctions, auction)
			}
		}
		activeAuctions = filteredAuctions
	}

	// Apply seller filter if provided
	if sellerID := r.URL.Query().Get("seller_id"); sellerID != "" {
		var filteredAuctions []Auction
		for _, auction := range activeAuctions {
			if auction.SellerID == sellerID {
				filteredAuctions = append(filteredAuctions, auction)
			}
		}
		activeAuctions = filteredAuctions
	}

	// Paginate results
	paginatedAuctions, pagination := paginateAuctions(activeAuctions, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Auctions retrieved successfully",
		Data: map[string]interface{}{
			"auctions":   paginatedAuctions,
			"pagination": pagination,
		},
	})
}

// Get auction by ID
func (s *AuctionService) getAuction(w http.ResponseWriter, r *http.Request) {
	auctionID := strings.TrimPrefix(r.URL.Path, "/api/v1/auctions/")
	if auctionID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Auction ID is required",
			Error:   "No auction ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, auction := range s.auctions {
		if auction.ID == auctionID {
			// Get bids for this auction
			var auctionBids []Bid
			for _, bid := range s.bids {
				if bid.AuctionID == auctionID && bid.Status == "active" {
					auctionBids = append(auctionBids, bid)
				}
			}

			// Sort bids by amount (descending)
			for i := 0; i < len(auctionBids)-1; i++ {
				for j := i + 1; j < len(auctionBids); j++ {
					if auctionBids[i].Amount < auctionBids[j].Amount {
						auctionBids[i], auctionBids[j] = auctionBids[j], auctionBids[i]
					}
				}
			}

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Auction retrieved successfully",
				Data: map[string]interface{}{
					"auction": auction,
					"bids":    auctionBids,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Auction not found",
		Error:   "Auction with ID " + auctionID + " does not exist",
	})
}

// Create auction
func (s *AuctionService) createAuction(w http.ResponseWriter, r *http.Request) {
	var req CreateAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid auction data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.ProductID == "" || req.ProductName == "" || req.SellerID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Product ID, product name, and seller ID are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.StartingBid <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Starting bid must be greater than 0",
			Error:   "Invalid starting bid",
		})
		return
	}

	if req.Duration <= 0 || req.Duration > 168 { // Max 7 days
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Duration must be between 1 and 168 hours",
			Error:   "Invalid duration",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	endTime := now.Add(time.Duration(req.Duration) * time.Hour)

	// Create new auction
	newAuction := Auction{
		ID:          fmt.Sprintf("auction-%d", time.Now().UnixNano()),
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
		Description: req.Description,
		StartingBid: req.StartingBid,
		CurrentBid:  req.StartingBid,
		BidCount:    0,
		BuyNowPrice: req.BuyNowPrice,
		SellerID:    req.SellerID,
		Status:      "active",
		StartTime:   now,
		EndTime:     endTime,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.auctions = append(s.auctions, newAuction)

	log.Printf("Auction created: %s (%s) by seller %s", newAuction.ID, newAuction.ProductName, newAuction.SellerID)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Auction created successfully",
		Data: map[string]interface{}{
			"auction": newAuction,
		},
	})
}

// Place bid
func (s *AuctionService) placeBid(w http.ResponseWriter, r *http.Request) {
	var req PlaceBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid bid data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.AuctionID == "" || req.BidderID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Auction ID and bidder ID are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.Amount <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Bid amount must be greater than 0",
			Error:   "Invalid bid amount",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find auction
	var auction *Auction
	for i := range s.auctions {
		if s.auctions[i].ID == req.AuctionID && s.auctions[i].Status == "active" {
			auction = &s.auctions[i]
			break
		}
	}

	if auction == nil {
		writeJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Message: "Auction not found or not active",
			Error:   "Invalid auction ID or auction is not active",
		})
		return
	}

	// Check if auction has ended
	if time.Now().After(auction.EndTime) {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Auction has ended",
			Error:   "Cannot place bid on ended auction",
		})
		return
	}

	// Check if bidder is the seller
	if auction.SellerID == req.BidderID {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot place bid on your own auction",
			Error:   "Seller cannot bid on their own auction",
		})
		return
	}

	// Check if bid amount is valid
	if req.Amount <= auction.CurrentBid {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Bid amount must be higher than current bid",
			Error:   "Bid too low",
		})
		return
	}

	// Create new bid
	newBid := Bid{
		ID:        fmt.Sprintf("bid-%d", time.Now().UnixNano()),
		AuctionID: req.AuctionID,
		BidderID:  req.BidderID,
		Amount:    req.Amount,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	s.bids = append(s.bids, newBid)

	// Update auction
	for i := range s.auctions {
		if s.auctions[i].ID == req.AuctionID {
			s.auctions[i].CurrentBid = req.Amount
			s.auctions[i].BidCount++
			s.auctions[i].CurrentBidder = &req.BidderID
			s.auctions[i].UpdatedAt = time.Now()
			break
		}
	}

	log.Printf("Bid placed: %s on auction %s by %s", newBid.ID, req.AuctionID, req.BidderID)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Bid placed successfully",
		Data: map[string]interface{}{
			"bid":     newBid,
			"current": map[string]interface{}{
				"current_bid":    req.Amount,
				"bid_count":     auction.BidCount + 1,
				"current_bidder": req.BidderID,
			},
		},
	})
}

// Get auction bids
func (s *AuctionService) getAuctionBids(w http.ResponseWriter, r *http.Request) {
	auctionID := r.URL.Query().Get("auction_id")
	if auctionID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Auction ID is required",
			Error:   "No auction ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get bids for this auction
	var auctionBids []Bid
	for _, bid := range s.bids {
		if bid.AuctionID == auctionID && bid.Status == "active" {
			auctionBids = append(auctionBids, bid)
		}
	}

	// Sort bids by amount (descending)
	for i := 0; i < len(auctionBids)-1; i++ {
		for j := i + 1; j < len(auctionBids); j++ {
			if auctionBids[i].Amount < auctionBids[j].Amount {
				auctionBids[i], auctionBids[j] = auctionBids[j], auctionBids[i]
			}
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Auction bids retrieved successfully",
		Data: map[string]interface{}{
			"bids":    auctionBids,
			"count":   len(auctionBids),
			"auction": auctionID,
		},
	})
}

// Get seller auctions
func (s *AuctionService) getSellerAuctions(w http.ResponseWriter, r *http.Request) {
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

	// Filter auctions by seller
	var sellerAuctions []Auction
	for _, auction := range s.auctions {
		if auction.SellerID == sellerID {
			sellerAuctions = append(sellerAuctions, auction)
		}
	}

	// Paginate results
	paginatedAuctions, pagination := paginateAuctions(sellerAuctions, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Seller auctions retrieved successfully",
		Data: map[string]interface{}{
			"auctions":   paginatedAuctions,
			"pagination": pagination,
			"seller_id":  sellerID,
		},
	})
}

// End auction
func (s *AuctionService) endAuction(w http.ResponseWriter, r *http.Request) {
	auctionID := strings.TrimPrefix(r.URL.Path, "/api/v1/auctions/")
	if auctionID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Auction ID is required",
			Error:   "No auction ID provided",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and end auction
	for i := range s.auctions {
		if s.auctions[i].ID == auctionID {
			if s.auctions[i].Status != "active" {
				writeJSONResponse(w, http.StatusBadRequest, Response{
					Success: false,
					Message: "Auction is not active",
					Error:   "Cannot end auction that is not active",
				})
				return
			}

			s.auctions[i].Status = "ended"
			s.auctions[i].UpdatedAt = time.Now()

			log.Printf("Auction ended: %s", auctionID)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Auction ended successfully",
				Data: map[string]interface{}{
					"auction": s.auctions[i],
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Auction not found",
		Error:   "Auction with ID " + auctionID + " does not exist",
	})
}

// Start auction service
func main() {
	service := NewAuctionService()

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/auctions", corsMiddleware(service.getAllAuctions))
	http.Handle("/api/v1/auctions/", corsMiddleware(service.getAuction)) // For GET by ID and end
	http.Handle("/api/v1/auctions/create", corsMiddleware(service.createAuction))
	http.Handle("/api/v1/auctions/bid", corsMiddleware(service.placeBid))
	http.Handle("/api/v1/auctions/bids", corsMiddleware(service.getAuctionBids))
	http.Handle("/api/v1/auctions/seller", corsMiddleware(service.getSellerAuctions))

	port := ":8087"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	fmt.Printf("🚀 AUCTION SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)
	fmt.Printf("🏦 Auctions list: http://localhost%s/api/v1/auctions\n", port)
	fmt.Printf("📝 Auction details: http://localhost%s/api/v1/auctions/{id}\n", port)
	fmt.Printf("➕ Create auction: http://localhost%s/api/v1/auctions/create\n", port)
	fmt.Printf("💰 Place bid: http://localhost%s/api/v1/auctions/bid\n", port)
	fmt.Printf("📋 Auction bids: http://localhost%s/api/v1/auctions/bids?auction_id={id}\n", port)
	fmt.Printf("👨‍💼 Seller auctions: http://localhost%s/api/v1/auctions/seller?seller_id={id}\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("📊 Total auctions: %d\n", len(service.auctions))
	fmt.Printf("💰 Total bids: %d\n", len(service.bids))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(port, nil))
}