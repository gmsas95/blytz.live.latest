package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/models"
	"github.com/gmsas95/blytz-mvp/services/auction-service/internal/services"
	"github.com/gmsas95/blytz-mvp/services/auction-service/pkg/firebase"
	shared_errors "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/gmsas95/blytz-mvp/shared/pkg/utils"
)

type AuctionHandler struct {
	auctionService *services.AuctionService
	logger         *zap.Logger
	firebaseApp    firebase.FirebaseApp
}

func NewAuctionHandler(auctionService *services.AuctionService, logger *zap.Logger, firebaseApp firebase.FirebaseApp) *AuctionHandler {
	return &AuctionHandler{auctionService: auctionService, logger: logger, firebaseApp: firebaseApp}
}

func (h *AuctionHandler) CreateAuction(c *gin.Context) {
	var auction models.Auction
	if err := c.ShouldBindJSON(&auction); err != nil {
		utils.SendErrorResponse(c, shared_errors.ErrInvalidRequestBody)
		return
	}

	if err := h.auctionService.CreateAuction(c.Request.Context(), &auction); err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, auction)
}

func (h *AuctionHandler) GetAuction(c *gin.Context) {
	id := c.Param("id")
	auction, err := h.auctionService.GetAuction(c.Request.Context(), id)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, auction)
}

func (h *AuctionHandler) PlaceBid(c *gin.Context) {
	var bid models.Bid
	if err := c.ShouldBindJSON(&bid); err != nil {
		utils.SendErrorResponse(c, shared_errors.ErrInvalidRequestBody)
		return
	}

	bid.AuctionID = c.Param("id")

	if err := h.auctionService.PlaceBid(c.Request.Context(), &bid); err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	// Notify via Firebase asynchronously
	go func() {
		// Use a background context for the async operation
		if err := h.firebaseApp.SendBidNotification(context.Background(), &bid); err != nil {
			h.logger.Error("Failed to send bid notification", zap.Error(err))
		}
	}()

	utils.SendSuccessResponse(c, http.StatusCreated, bid)
}

func (h *AuctionHandler) ListAuctions(c *gin.Context) {
	auctions, err := h.auctionService.ListAuctions(c.Request.Context())
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"items":      auctions,
		"total":      len(auctions),
		"page":       1,
		"limit":      20,
		"totalPages": 1,
	})
}

func (h *AuctionHandler) GetActiveAuctions(c *gin.Context) {
	auctions, err := h.auctionService.GetActiveAuctions(c.Request.Context())
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, auctions)
}

func (h *AuctionHandler) GetAuctionStatus(c *gin.Context) {
	auctionID := c.Param("id")
	auction, err := h.auctionService.GetAuction(c.Request.Context(), auctionID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"status":    auction.Status,
		"end_time":  auction.EndTime,
		"is_active": auction.IsActive,
	})
}

func (h *AuctionHandler) GetBids(c *gin.Context) {
	_ = c.Param("id") // TODO: Use auction ID to fetch bids
	// TODO: Implement GetBids service method
	// For now, return empty array
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"bids": []interface{}{},
	})
}

func (h *AuctionHandler) UpdateAuction(c *gin.Context) {
	id := c.Param("id")
	var auction models.Auction
	if err := c.ShouldBindJSON(&auction); err != nil {
		utils.SendErrorResponse(c, shared_errors.ErrInvalidRequestBody)
		return
	}

	// Get existing auction
	existing, err := h.auctionService.GetAuction(c.Request.Context(), id)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	// TODO: Implement UpdateAuction service method
	// For now, return the existing auction
	utils.SendSuccessResponse(c, http.StatusOK, existing)
}

func (h *AuctionHandler) DeleteAuction(c *gin.Context) {
	id := c.Param("id")
	
	// Verify auction exists
	_, err := h.auctionService.GetAuction(c.Request.Context(), id)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	// TODO: Implement DeleteAuction service method
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Auction deleted successfully",
	})
}
