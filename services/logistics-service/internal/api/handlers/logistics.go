package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gmsas95/blytz-mvp/services/logistics-service/internal/models"
	"github.com/gmsas95/blytz-mvp/services/logistics-service/internal/services"
	"github.com/gmsas95/blytz-mvp/shared/pkg/errors"
	"github.com/gmsas95/blytz-mvp/shared/pkg/utils"
)

type LogisticsHandler struct {
	logisticsService *services.LogisticsService
	logger           *zap.Logger
}

func NewLogisticsHandler(logisticsService *services.LogisticsService, logger *zap.Logger) *LogisticsHandler {
	return &LogisticsHandler{
		logisticsService: logisticsService,
		logger:           logger,
	}
}

func (h *LogisticsHandler) CreateShipment(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		utils.SendErrorResponse(c, errors.ErrUnauthorizedError)
		return
	}

	var req services.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, errors.ErrInvalidRequestBodyError)
		return
	}

	shipment, err := h.logisticsService.CreateShipment(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create shipment", zap.Error(err))
		utils.SendErrorResponse(c, err)
		return
	}

	response := h.mapShipmentToResponse(shipment)
	utils.SendSuccessResponse(c, http.StatusOK, response)
}

func (h *LogisticsHandler) GetShipment(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		utils.SendErrorResponse(c, errors.ErrUnauthorizedError)
		return
	}

	shipmentID := c.Param("id")
	if shipmentID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request"))
		return
	}

	shipment, err := h.logisticsService.GetShipment(c.Request.Context(), shipmentID, userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok && appErr.Type == "NOT_FOUND_ERROR" {
			utils.SendErrorResponse(c, err)
			return
		}
		h.logger.Error("Failed to get shipment", zap.Error(err))
		utils.SendErrorResponse(c, err)
		return
	}

	response := h.mapShipmentToResponse(shipment)
	utils.SendSuccessResponse(c, http.StatusOK, response)
}

func (h *LogisticsHandler) UpdateShipmentStatus(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		utils.SendErrorResponse(c, errors.ErrUnauthorizedError)
		return
	}

	shipmentID := c.Param("id")
	if shipmentID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request"))
		return
	}

	var req services.UpdateShipmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, errors.ErrInvalidRequestBodyError)
		return
	}

	shipment, err := h.logisticsService.UpdateShipmentStatus(c.Request.Context(), shipmentID, userID, models.ShipmentStatus(req.Status))
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok && appErr.Type == "NOT_FOUND_ERROR" {
			utils.SendErrorResponse(c, err)
			return
		}
		h.logger.Error("Failed to update shipment status", zap.Error(err))
		utils.SendErrorResponse(c, err)
		return
	}

	response := h.mapShipmentToResponse(shipment)
	utils.SendSuccessResponse(c, http.StatusOK, response)
}

func (h *LogisticsHandler) GetShipmentByOrder(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		utils.SendErrorResponse(c, errors.ErrUnauthorizedError)
		return
	}

	orderID := c.Param("orderId")
	if orderID == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request"))
		return
	}

	shipment, err := h.logisticsService.GetShipmentByOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok && appErr.Type == "NOT_FOUND_ERROR" {
			utils.SendErrorResponse(c, err)
			return
		}
		h.logger.Error("Failed to get shipment by order", zap.Error(err))
		utils.SendErrorResponse(c, err)
		return
	}

	response := h.mapShipmentToResponse(shipment)
	utils.SendSuccessResponse(c, http.StatusOK, response)
}

func (h *LogisticsHandler) TrackShipment(c *gin.Context) {
	trackingNumber := c.Param("trackingNumber")
	if trackingNumber == "" {
		utils.SendErrorResponse(c, errors.NewValidationError("INVALID_REQUEST", "Invalid request"))
		return
	}

	shipment, events, err := h.logisticsService.TrackShipment(c.Request.Context(), trackingNumber)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok && appErr.Type == "NOT_FOUND_ERROR" {
			utils.SendErrorResponse(c, err)
			return
		}
		h.logger.Error("Failed to track shipment", zap.Error(err))
		utils.SendErrorResponse(c, err)
		return
	}

	shipmentResponse := h.mapShipmentToResponse(shipment)
	eventResponses := make([]services.TrackingEventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = services.TrackingEventResponse{
			ID:          event.ID,
			Status:      event.Status,
			Description: event.Description,
			Location:    event.Location,
			Timestamp:   event.Timestamp.Format("2006-01-02T15:04:05Z"),
		}
	}

	response := services.TrackingResponse{
		Shipment: *shipmentResponse,
		Events:   eventResponses,
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}

func (h *LogisticsHandler) mapShipmentToResponse(shipment *models.Shipment) *services.ShipmentResponse {
	var estimatedDelivery, actualDelivery *string
	if shipment.EstimatedDelivery != nil {
		str := shipment.EstimatedDelivery.Format("2006-01-02T15:04:05Z")
		estimatedDelivery = &str
	}
	if shipment.ActualDelivery != nil {
		str := shipment.ActualDelivery.Format("2006-01-02T15:04:05Z")
		actualDelivery = &str
	}

	return &services.ShipmentResponse{
		ID:             shipment.ID,
		OrderID:        shipment.OrderID,
		UserID:         shipment.UserID,
		TrackingNumber: shipment.TrackingNumber,
		Carrier:        shipment.Carrier,
		Service:        shipment.Service,
		Status:         shipment.Status,
		OriginAddress: services.AddressResponse{
			Name:        shipment.OriginAddress.Name,
			Street:      shipment.OriginAddress.Street,
			City:        shipment.OriginAddress.City,
			State:       shipment.OriginAddress.State,
			PostalCode:  shipment.OriginAddress.PostalCode,
			Country:     shipment.OriginAddress.Country,
			PhoneNumber: shipment.OriginAddress.PhoneNumber,
		},
		DestinationAddress: services.AddressResponse{
			Name:        shipment.DestinationAddress.Name,
			Street:      shipment.DestinationAddress.Street,
			City:        shipment.DestinationAddress.City,
			State:       shipment.DestinationAddress.State,
			PostalCode:  shipment.DestinationAddress.PostalCode,
			Country:     shipment.DestinationAddress.Country,
			PhoneNumber: shipment.DestinationAddress.PhoneNumber,
		},
		Weight:            shipment.Weight,
		Dimensions: services.DimensionsResponse{
			Length: shipment.Dimensions.Length,
			Width:  shipment.Dimensions.Width,
			Height: shipment.Dimensions.Height,
		},
		EstimatedDelivery: estimatedDelivery,
		ActualDelivery:    actualDelivery,
		Cost:              shipment.Cost,
		Notes:             shipment.Notes,
		CreatedAt:         shipment.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         shipment.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *LogisticsHandler) ListShipments(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		utils.SendErrorResponse(c, errors.ErrUnauthorizedError)
		return
	}

	// Get pagination parameters
	page, perPage := utils.GetPaginationParams(c)

	// Get shipments with pagination
	var shipments []models.Shipment
	var total int64

	// Count total shipments for this user
	if err := h.logisticsService.GetDB().Model(&models.Shipment{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		h.logger.Error("Failed to count shipments", zap.Error(err))
		utils.SendErrorResponse(c, errors.ErrInternalServerError)
		return
	}

	// Get paginated shipments
	offset := (page - 1) * perPage
	if err := h.logisticsService.GetDB().Where("user_id = ?", userID).Offset(offset).Limit(perPage).Order("created_at DESC").Find(&shipments).Error; err != nil {
		h.logger.Error("Failed to get shipments", zap.Error(err))
		utils.SendErrorResponse(c, errors.ErrInternalServerError)
		return
	}

	// Convert to response format
	shipmentResponses := make([]*services.ShipmentResponse, len(shipments))
	for i, shipment := range shipments {
		shipmentResponses[i] = h.mapShipmentToResponse(&shipment)
	}

	// Calculate pagination info
	pagination := utils.CalculatePagination(page, perPage, total)

	// Send paginated response
	utils.SendPaginatedResponse(c, http.StatusOK, shipmentResponses, pagination)
}


func (h *LogisticsHandler) ListCarriers(c *gin.Context) {
	carriers := []gin.H{
		{
			"id":   "ninja_van",
			"name": "NinjaVan",
			"services": []string{
				"standard",
				"express",
				"economy",
			},
			"countries": []string{
				"MY", "SG", "ID", "TH", "PH", "VN",
			},
		},
		{
			"id":   "pos_laju",
			"name": "PosLaju",
			"services": []string{
				"standard",
				"express",
				"next_day",
			},
			"countries": []string{
				"MY",
			},
		},
		{
			"id":   "gd_ex",
			"name": "GDex",
			"services": []string{
				"standard",
				"express",
				"economy",
			},
			"countries": []string{
				"MY",
			},
		},
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"carriers": carriers,
	})
}

func (h *LogisticsHandler) CalculateShippingRates(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		utils.SendErrorResponse(c, errors.ErrUnauthorizedError)
		return
	}

	var req services.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, errors.ErrInvalidRequestBodyError)
		return
	}

	// Calculate rates for different carriers
	rates := []gin.H{
		{
			"carrier":      "ninja_van",
			"service":      "standard",
			"cost":         req.Cost,
			"currency":     "MYR",
			"delivery_days": 2,
		},
		{
			"carrier":      "ninja_van",
			"service":      "express",
			"cost":         req.Cost * 1.5,
			"currency":     "MYR",
			"delivery_days": 1,
		},
		{
			"carrier":      "pos_laju",
			"service":      "standard",
			"cost":         req.Cost * 0.9,
			"currency":     "MYR",
			"delivery_days": 3,
		},
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"origin":      req.OriginAddress,
		"destination": req.DestinationAddress,
		"weight":      req.Weight,
		"dimensions":  req.Dimensions,
		"rates":       rates,
	})
}
