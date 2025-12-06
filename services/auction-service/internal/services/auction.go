package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/models"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/repository"
)

type AuctionService struct {
	repo   repository.Repository
	logger *zap.Logger
	db     *gorm.DB
}

func NewAuctionService(repo repository.Repository, logger *zap.Logger, db *gorm.DB) *AuctionService {
	return &AuctionService{
		repo:   repo,
		logger: logger,
		db:     db,
	}
}

// === AUCTION MANAGEMENT ===

// CreateAuction creates new auction
func (s *AuctionService) CreateAuction(ctx context.Context, sellerID string, req *models.CreateAuctionRequest) (*models.Auction, error) {
	s.logger.Info("🎵 Auction Service: Creating auction", 
		zap.String("seller_id", sellerID),
		zap.String("product_id", req.ProductID))

	// Validate request
	if err := s.validateCreateAuctionRequest(req); err != nil {
		return nil, err
	}

	// Create auction
	auction := &models.Auction{
		ProductID:       req.ProductID,
		SellerID:        sellerID,
		Title:           req.Title,
		Description:     req.Description,
		StartingPrice:   req.StartingPrice,
		CurrentPrice:    req.StartingPrice,
		ReservePrice:    req.ReservePrice,
		BuyItNowPrice:   req.BuyItNowPrice,
		MinBidIncrement: req.MinBidIncrement,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		Type:            req.Type,
		AutoExtend:      req.AutoExtend,
		ExtendMinutes:   req.ExtendMinutes,
		Featured:        req.Featured,
		ShippingInfo:    req.ShippingInfo,
		Condition:       req.Condition,
		Category:        req.Category,
		Subcategory:     req.Subcategory,
		Status:          models.AuctionStatusScheduled,
		IsActive:        false,
	}

	// Set tags array
	auction.SetTagsArray(req.Tags)

	// Create images
	var images []models.AuctionImage
	for i, imageURL := range req.Images {
		image := models.AuctionImage{
			AuctionID: auction.AuctionID,
			ImageURL:  imageURL,
			Order:     i + 1,
			IsMain:    i == 0, // First image is main
		}
		images = append(images, image)
	}

	// Save auction with transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create auction
		if err := tx.Create(auction).Error; err != nil {
			return fmt.Errorf("failed to create auction: %w", err)
		}

		// Create images
		for i := range images {
			images[i].AuctionID = auction.AuctionID
		}
		if len(images) > 0 {
			if err := tx.Create(&images).Error; err != nil {
				return fmt.Errorf("failed to create auction images: %w", err)
			}
		}

		// Log activity
		activity := &models.AuctionActivity{
			AuctionID:    auction.AuctionID,
			UserID:       sellerID,
			ActivityType: "create",
			ActivityData: fmt.Sprintf(`{"title": "%s", "starting_price": %.2f}`, auction.Title, auction.StartingPrice),
		}
		if err := tx.Create(activity).Error; err != nil {
			s.logger.Error("Failed to log auction activity", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("🎵 Auction Service: Failed to create auction", zap.Error(err))
		return nil, err
	}

	// Load relationships
	auction.Images = images

	s.logger.Info("🎵 Auction Service: Auction created successfully", 
		zap.String("auction_id", auction.AuctionID),
		zap.String("product_id", auction.ProductID))

	return auction, nil
}

// GetAuction retrieves auction by ID
func (s *AuctionService) GetAuction(ctx context.Context, auctionID string, userID *string) (*models.Auction, error) {
	s.logger.Info("🎵 Auction Service: Getting auction", zap.String("auction_id", auctionID))

	var auction models.Auction
	query := s.db.Preload("Images").Preload("Bids", func(db *gorm.DB) *gorm.DB {
		return db.Order("amount DESC, bid_time DESC")
	}).Preload("Bids.Bidder")

	err := query.Where("auction_id = ?", auctionID).First(&auction).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("auction not found")
		}
		return nil, err
	}

	// Increment views
	s.db.Model(&auction).Update("total_views", gorm.Expr("total_views + 1"))

	// Check if user is watching this auction
	if userID != nil {
		var watcher models.AuctionWatcher
		err := s.db.Where("auction_id = ? AND user_id = ? AND is_active = ?", auctionID, *userID, true).First(&watcher).Error
		if err == nil {
			// Set watched flag in response
			// This would be handled in response layer
		}
	}

	s.logger.Info("🎵 Auction Service: Auction retrieved successfully", zap.String("auction_id", auctionID))
	return &auction, nil
}

// UpdateAuction updates existing auction
func (s *AuctionService) UpdateAuction(ctx context.Context, auctionID, sellerID string, req *models.UpdateAuctionRequest) (*models.Auction, error) {
	s.logger.Info("🎵 Auction Service: Updating auction", 
		zap.String("auction_id", auctionID),
		zap.String("seller_id", sellerID))

	// Get existing auction
	var auction models.Auction
	err := s.db.Where("auction_id = ? AND seller_id = ?", auctionID, sellerID).First(&auction).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("auction not found or unauthorized")
		}
		return nil, err
	}

	// Check if auction can be updated (only scheduled auctions)
	if auction.Status != models.AuctionStatusScheduled {
		return nil, fmt.Errorf("auction cannot be updated - already started")
	}

	// Update fields
	if req.Title != "" {
		auction.Title = req.Title
	}
	if req.Description != "" {
		auction.Description = req.Description
	}
	if req.ReservePrice > 0 {
		auction.ReservePrice = req.ReservePrice
	}
	if req.BuyItNowPrice > 0 {
		auction.BuyItNowPrice = req.BuyItNowPrice
	}
	if req.MinBidIncrement > 0 {
		auction.MinBidIncrement = req.MinBidIncrement
	}
	if !req.StartTime.IsZero() {
		auction.StartTime = req.StartTime
	}
	if !req.EndTime.IsZero() {
		auction.EndTime = req.EndTime
	}

	// Update in transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&auction).Error; err != nil {
			return fmt.Errorf("failed to update auction: %w", err)
		}

		// Update images if provided
		if req.Images != nil {
			// Delete existing images
			tx.Where("auction_id = ?", auctionID).Delete(&models.AuctionImage{})

			// Create new images
			var images []models.AuctionImage
			for i, imageURL := range req.Images {
				image := models.AuctionImage{
					AuctionID: auctionID,
					ImageURL:  imageURL,
					Order:     i + 1,
					IsMain:    i == 0,
				}
				images = append(images, image)
			}
			if len(images) > 0 {
				if err := tx.Create(&images).Error; err != nil {
					return fmt.Errorf("failed to update auction images: %w", err)
				}
			}
		}

		// Log activity
		activity := &models.AuctionActivity{
			AuctionID:    auctionID,
			UserID:       sellerID,
			ActivityType: "update",
			ActivityData: fmt.Sprintf(`{"title": "%s"}`, auction.Title),
		}
		if err := tx.Create(activity).Error; err != nil {
			s.logger.Error("Failed to log auction activity", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("🎵 Auction Service: Failed to update auction", zap.Error(err))
		return nil, err
	}

	s.logger.Info("🎵 Auction Service: Auction updated successfully", zap.String("auction_id", auctionID))
	return &auction, nil
}

// DeleteAuction deletes auction
func (s *AuctionService) DeleteAuction(ctx context.Context, auctionID, sellerID string) error {
	s.logger.Info("🎵 Auction Service: Deleting auction", 
		zap.String("auction_id", auctionID),
		zap.String("seller_id", sellerID))

	// Get existing auction
	var auction models.Auction
	err := s.db.Where("auction_id = ? AND seller_id = ?", auctionID, sellerID).First(&auction).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("auction not found or unauthorized")
		}
		return err
	}

	// Check if auction can be deleted (only scheduled auctions)
	if auction.Status != models.AuctionStatusScheduled {
		return fmt.Errorf("auction cannot be deleted - already started")
	}

	// Delete in transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Delete images
		if err := tx.Where("auction_id = ?", auctionID).Delete(&models.AuctionImage{}).Error; err != nil {
			return fmt.Errorf("failed to delete auction images: %w", err)
		}

		// Delete bids
		if err := tx.Where("auction_id = ?", auctionID).Delete(&models.Bid{}).Error; err != nil {
			return fmt.Errorf("failed to delete auction bids: %w", err)
		}

		// Delete watchers
		if err := tx.Where("auction_id = ?", auctionID).Delete(&models.AuctionWatcher{}).Error; err != nil {
			return fmt.Errorf("failed to delete auction watchers: %w", err)
		}

		// Delete activities
		if err := tx.Where("auction_id = ?", auctionID).Delete(&models.AuctionActivity{}).Error; err != nil {
			return fmt.Errorf("failed to delete auction activities: %w", err)
		}

		// Delete auction
		if err := tx.Where("auction_id = ?", auctionID).Delete(&models.Auction{}).Error; err != nil {
			return fmt.Errorf("failed to delete auction: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("🎵 Auction Service: Failed to delete auction", zap.Error(err))
		return err
	}

	s.logger.Info("🎵 Auction Service: Auction deleted successfully", zap.String("auction_id", auctionID))
	return nil
}

// === AUCTION SEARCH & LISTING ===

// SearchAuctions searches auctions with filters
func (s *AuctionService) SearchAuctions(ctx context.Context, req *models.SearchAuctionsRequest) (*models.AuctionsResponse, error) {
	s.logger.Info("🎵 Auction Service: Searching auctions", zap.String("query", req.Query))

	// Build query
	query := s.db.Model(&models.Auction{}).Preload("Images")

	// Apply filters
	if req.Query != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", 
			"%"+req.Query+"%", "%"+req.Query+"%")
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Subcategory != "" {
		query = query.Where("subcategory = ?", req.Subcategory)
	}
	if req.MinPrice > 0 {
		query = query.Where("starting_price >= ?", req.MinPrice)
	}
	if req.MaxPrice > 0 {
		query = query.Where("starting_price <= ?", req.MaxPrice)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.SellerID != "" {
		query = query.Where("seller_id = ?", req.SellerID)
	}
	if req.Condition != "" {
		query = query.Where("condition = ?", req.Condition)
	}
	if req.Featured {
		query = query.Where("featured = ?", req.Featured)
	}
	if req.BuyItNow {
		query = query.Where("buy_it_now_price > 0")
	}
	if req.EndingSoon {
		// Ending in next 24 hours
		query = query.Where("status = ? AND end_time <= NOW() + INTERVAL '24 hours' AND end_time > NOW()", 
			models.AuctionStatusActive)
	}

	// Apply sorting
	sortBy := "created_at"
	if req.SortBy == "price" {
		sortBy = "starting_price"
	} else if req.SortBy == "end_time" {
		sortBy = "end_time"
	} else if req.SortBy == "views" {
		sortBy = "total_views"
	} else if req.SortBy == "bids" {
		sortBy = "total_bids"
	}

	sortOrder := "DESC"
	if req.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(sortBy + " " + sortOrder)

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var auctions []models.Auction
	err := query.Limit(limit).Offset(offset).Find(&auctions).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages

	response := &models.AuctionsResponse{
		Auctions:   auctions,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    hasNext,
	}

	s.logger.Info("🎵 Auction Service: Auctions searched successfully", 
		zap.Int64("total", total),
		zap.Int("count", len(auctions)))

	return response, nil
}

// GetSellerAuctions gets auctions for a specific seller
func (s *AuctionService) GetSellerAuctions(ctx context.Context, sellerID string, page, limit int) (*models.AuctionsResponse, error) {
	s.logger.Info("🎵 Auction Service: Getting seller auctions", zap.String("seller_id", sellerID))

	query := s.db.Model(&models.Auction{}).Preload("Images").Where("seller_id = ?", sellerID)

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var auctions []models.Auction
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&auctions).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages

	response := &models.AuctionsResponse{
		Auctions:   auctions,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    hasNext,
	}

	s.logger.Info("🎵 Auction Service: Seller auctions retrieved successfully", 
		zap.String("seller_id", sellerID),
		zap.Int64("total", total))

	return response, nil
}

// === BID MANAGEMENT ===

// PlaceBid places a bid on an auction
func (s *AuctionService) PlaceBid(ctx context.Context, auctionID, bidderID string, req *models.PlaceBidRequest) (*models.BidResponse, error) {
	s.logger.Info("🎵 Auction Service: Placing bid", 
		zap.String("auction_id", auctionID),
		zap.String("bidder_id", bidderID),
		zap.Float64("amount", req.Amount))

	// Get auction with lock
	var auction models.Auction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where("auction_id = ?", auctionID).First(&auction).Error
		if err != nil {
			return err
		}

		// Validate auction
		if auction.IsEnded() {
			return fmt.Errorf("auction has ended")
		}
		if auction.Status != models.AuctionStatusActive {
			return fmt.Errorf("auction is not active")
		}
		if time.Now().Before(auction.StartTime) {
			return fmt.Errorf("auction has not started")
		}
		if time.Now().After(auction.EndTime) {
			return fmt.Errorf("auction has ended")
		}
		if auction.SellerID == bidderID {
			return fmt.Errorf("cannot bid on own auction")
		}

		// Validate bid amount
		if !auction.IsValidBid(req.Amount) {
			return fmt.Errorf("bid amount must be at least $%.2f", auction.GetMinNextBid())
		}

		// Check if this is an auto-bid
		if req.IsAutoBid && req.MaxAutoBid > 0 {
			return s.handleAutoBid(tx, &auction, bidderID, req)
		}

		// Create bid
		bid := &models.Bid{
			AuctionID: auctionID,
			BidderID:  bidderID,
			BidderName: "Bidder Name", // Would get from user service
			Amount:    req.Amount,
			IsWinning: true,
			IsAutoBid: false,
			BidTime:   time.Now(),
			IPAddress: "127.0.0.1", // Would get from request context
			UserAgent:  "Auction Service",
		}

		// Update previous winning bid
		if auction.WinningBidID != nil {
			if err := tx.Model(&models.Bid{}).Where("bid_id = ?", *auction.WinningBidID).Update("is_winning", false).Error; err != nil {
				return fmt.Errorf("failed to update previous winning bid: %w", err)
			}
		}

		// Save new bid
		if err := tx.Create(bid).Error; err != nil {
			return fmt.Errorf("failed to create bid: %w", err)
		}

		// Update auction
		auction.CurrentPrice = req.Amount
		auction.TotalBids += 1
		auction.WinningBidID = &bid.BidID
		auction.WinningUserID = &bidderID

		// Handle auto-extension
		if auction.AutoExtend && time.Until(auction.EndTime) < 5*time.Minute {
			auction.EndTime = auction.EndTime.Add(time.Duration(auction.ExtendMinutes) * time.Minute)
		}

		if err := tx.Save(&auction).Error; err != nil {
			return fmt.Errorf("failed to update auction: %w", err)
		}

		// Log activity
		activity := &models.AuctionActivity{
			AuctionID:    auctionID,
			UserID:       bidderID,
			ActivityType: "bid",
			ActivityData: fmt.Sprintf(`{"amount": %.2f, "bid_id": "%s"}`, req.Amount, bid.BidID),
		}
		if err := tx.Create(activity).Error; err != nil {
			s.logger.Error("Failed to log auction activity", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("🎵 Auction Service: Failed to place bid", zap.Error(err))
		return nil, err
	}

	// Get updated auction with bid
	bid, err := s.getBidDetails(auctionID, req.Amount)
	if err != nil {
		return nil, err
	}

	response := &models.BidResponse{
		Bid:         *bid,
		IsOutbid:    false,
		NewHighBid:  auction.CurrentPrice,
		MinNextBid:  auction.GetMinNextBid(),
		TimeRemaining: auction.GetTimeRemaining(),
	}

	s.logger.Info("🎵 Auction Service: Bid placed successfully", 
		zap.String("auction_id", auctionID),
		zap.String("bid_id", bid.BidID),
		zap.Float64("amount", req.Amount))

	return response, nil
}

// GetAuctionBids gets bids for an auction
func (s *AuctionService) GetAuctionBids(ctx context.Context, auctionID string, page, limit int) (*models.BidsResponse, error) {
	s.logger.Info("🎵 Auction Service: Getting auction bids", zap.String("auction_id", auctionID))

	query := s.db.Model(&models.Bid{}).Preload("Bidder").Where("auction_id = ?", auctionID)

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	var bids []models.Bid
	err := query.Order("amount DESC, bid_time DESC").Limit(limit).Offset(offset).Find(&bids).Error
	if err != nil {
		return nil, err
	}

	hasNext := int64(page*limit) < total

	response := &models.BidsResponse{
		Bids:    bids,
		Total:   total,
		Page:    page,
		Limit:   limit,
		HasNext: hasNext,
	}

	s.logger.Info("🎵 Auction Service: Auction bids retrieved successfully", 
		zap.String("auction_id", auctionID),
		zap.Int64("total", total))

	return response, nil
}

// === AUCTION MANAGEMENT ===

// StartAuction starts an auction
func (s *AuctionService) StartAuction(ctx context.Context, auctionID string) error {
	s.logger.Info("🎵 Auction Service: Starting auction", zap.String("auction_id", auctionID))

	var auction models.Auction
	err := s.db.Where("auction_id = ?", auctionID).First(&auction).Error
	if err != nil {
		return fmt.Errorf("auction not found")
	}

	if auction.Status != models.AuctionStatusScheduled {
		return fmt.Errorf("auction cannot be started - not scheduled")
	}

	now := time.Now()
	if now.Before(auction.StartTime) {
		return fmt.Errorf("auction cannot be started - start time not reached")
	}

	// Update auction
	auction.Status = models.AuctionStatusActive
	auction.IsActive = true

	if err := s.db.Save(&auction).Error; err != nil {
		return fmt.Errorf("failed to start auction: %w", err)
	}

	// Log activity
	activity := &models.AuctionActivity{
		AuctionID:    auctionID,
		ActivityType: "start",
		ActivityData: `{"status": "active"}`,
	}
	if err := s.db.Create(activity).Error; err != nil {
		s.logger.Error("Failed to log auction activity", zap.Error(err))
	}

	s.logger.Info("🎵 Auction Service: Auction started successfully", zap.String("auction_id", auctionID))
	return nil
}

// EndAuction ends an auction
func (s *AuctionService) EndAuction(ctx context.Context, auctionID string) error {
	s.logger.Info("🎵 Auction Service: Ending auction", zap.String("auction_id", auctionID))

	var auction models.Auction
	err := s.db.Where("auction_id = ?", auctionID).First(&auction).Error
	if err != nil {
		return fmt.Errorf("auction not found")
	}

	if auction.IsEnded() {
		return fmt.Errorf("auction already ended")
	}

	// Process auction end
	endReason := "ended"
	winnerUserID := ""
	finalPrice := auction.CurrentPrice

	// Check if reserve price was met
	if auction.ReservePrice > 0 && auction.CurrentPrice < auction.ReservePrice {
		endReason = "reserve_not_met"
		winnerUserID = ""
		finalPrice = 0
	} else if auction.WinningUserID != nil {
		winnerUserID = *auction.WinningUserID
		endReason = "sold"
	}

	// Update auction
	auction.Status = models.AuctionStatusEnded
	auction.IsActive = false
	auction.FinalPrice = finalPrice
	auction.EndReason = endReason
	auction.EndTime = time.Now()

	// Update winning bid
	if auction.WinningBidID != nil && endReason == "sold" {
		if err := s.db.Model(&models.Bid{}).Where("bid_id = ?", *auction.WinningBidID).Update("is_winning", true).Error; err != nil {
			return fmt.Errorf("failed to update winning bid: %w", err)
		}
	}

	if err := s.db.Save(&auction).Error; err != nil {
		return fmt.Errorf("failed to end auction: %w", err)
	}

	// Log activity
	activityData := fmt.Sprintf(`{"status": "ended", "end_reason": "%s", "final_price": %.2f}`, endReason, finalPrice)
	activity := &models.AuctionActivity{
		AuctionID:    auctionID,
		ActivityType: "end",
		ActivityData: activityData,
	}
	if err := s.db.Create(activity).Error; err != nil {
		s.logger.Error("Failed to log auction activity", zap.Error(err))
	}

	// TODO: Create order for winner if reserve price was met
	if endReason == "sold" {
		// This would integrate with Order Service
		s.logger.Info("🎵 Auction Service: Auction ended with winner", 
			zap.String("auction_id", auctionID),
			zap.String("winner_user_id", winnerUserID),
			zap.Float64("final_price", finalPrice))
	}

	s.logger.Info("🎵 Auction Service: Auction ended successfully", 
		zap.String("auction_id", auctionID),
		zap.String("end_reason", endReason))

	return nil
}

// === HELPER METHODS ===

// validateCreateAuctionRequest validates auction creation request
func (s *AuctionService) validateCreateAuctionRequest(req *models.CreateAuctionRequest) error {
	if req.StartingPrice < 0 {
		return fmt.Errorf("starting price cannot be negative")
	}
	if req.MinBidIncrement <= 0 {
		return fmt.Errorf("minimum bid increment must be positive")
	}
	if req.StartTime.After(req.EndTime) {
		return fmt.Errorf("start time must be before end time")
	}
	if req.StartTime.Before(time.Now()) {
		return fmt.Errorf("start time cannot be in the past")
	}
	if req.BuyItNowPrice > 0 && req.BuyItNowPrice <= req.StartingPrice {
		return fmt.Errorf("buy it now price must be greater than starting price")
	}
	if req.ReservePrice > 0 && req.ReservePrice <= req.StartingPrice {
		return fmt.Errorf("reserve price must be greater than starting price")
	}
	return nil
}

// handleAutoBid handles automatic bidding
func (s *AuctionService) handleAutoBid(tx *gorm.DB, auction *models.Auction, bidderID string, req *models.PlaceBidRequest) error {
	// Get user's current highest bid
	var currentBid models.Bid
	err := tx.Where("auction_id = ? AND bidder_id = ?", auction.AuctionID, bidderID).Order("amount DESC").First(&currentBid).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Determine bid amount
	bidAmount := req.Amount
	if bidAmount <= auction.CurrentPrice {
		bidAmount = auction.CurrentPrice + auction.MinBidIncrement
	}

	// Check against max auto bid
	if bidAmount > req.MaxAutoBid {
		bidAmount = req.MaxAutoBid
	}

	// If bid is still not higher than current price, can't bid
	if bidAmount <= auction.CurrentPrice {
		return fmt.Errorf("auto-bid insufficient to outbid current bid")
	}

	// Create bid
	bid := &models.Bid{
		AuctionID: auction.AuctionID,
		BidderID:  bidderID,
		BidderName: "Auto Bidder",
		Amount:    bidAmount,
		IsWinning: true,
		IsAutoBid: true,
		MaxAutoBid: req.MaxAutoBid,
		BidTime:   time.Now(),
		IPAddress: "127.0.0.1",
		UserAgent:  "Auto Bid Service",
	}

	// Update previous winning bid
	if auction.WinningBidID != nil {
		if err := tx.Model(&models.Bid{}).Where("bid_id = ?", *auction.WinningBidID).Update("is_winning", false).Error; err != nil {
			return fmt.Errorf("failed to update previous winning bid: %w", err)
		}
	}

	// Save new bid
	if err := tx.Create(bid).Error; err != nil {
		return fmt.Errorf("failed to create auto bid: %w", err)
	}

	// Update auction
	auction.CurrentPrice = bidAmount
	auction.TotalBids += 1
	auction.WinningBidID = &bid.BidID
	auction.WinningUserID = &bidderID

	if err := tx.Save(auction).Error; err != nil {
		return fmt.Errorf("failed to update auction: %w", err)
	}

	return nil
}

// getBidDetails gets bid details for response
func (s *AuctionService) getBidDetails(auctionID string, amount float64) (*models.Bid, error) {
	var bid models.Bid
	err := s.db.Where("auction_id = ? AND amount = ?", auctionID, amount).First(&bid).Error
	if err != nil {
		return nil, err
	}
	return &bid, nil
}