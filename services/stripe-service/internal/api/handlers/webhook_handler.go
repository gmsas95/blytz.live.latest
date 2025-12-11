package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz-mvp/services/stripe-service/internal/services"
	"go.uber.org/zap"
)

// WebhookHandler handles webhook requests
type WebhookHandler struct {
	webhookService *services.WebhookService
	logger         *zap.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService *services.WebhookService, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		logger:         logger,
	}
}

// HandleStripeWebhook handles incoming Stripe webhook events
func (h *WebhookHandler) HandleStripeWebhook(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to read request body",
		})
		return
	}

	// Get the Stripe signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		h.logger.Error("Missing Stripe signature header")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Missing Stripe signature header",
		})
		return
	}

	// Process the webhook event
	ctx := context.Background()
	err = h.webhookService.ProcessWebhookEvent(ctx, body, signature)
	if err != nil {
		h.logger.Error("Failed to process webhook event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to process webhook event",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook processed successfully",
	})
}

// GetWebhookEvents retrieves webhook events with pagination
func (h *WebhookHandler) GetWebhookEvents(c *gin.Context) {
	// Parse query parameters
	limit := 10 // default limit
	offset := 0 // default offset
	eventType := c.Query("event_type")

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := parseInt(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get events from service
	events, err := h.webhookService.ListEvents(limit, offset, eventType)
	if err != nil {
		h.logger.Error("Failed to retrieve webhook events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to retrieve webhook events",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook events retrieved successfully",
		"data": gin.H{
			"events": events,
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
				"count":  len(events),
			},
		},
	})
}

// GetWebhookEvent retrieves a specific webhook event by ID
func (h *WebhookHandler) GetWebhookEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Event ID is required",
		})
		return
	}

	// Get event from service
	event, err := h.webhookService.GetEventByID(eventID)
	if err != nil {
		h.logger.Error("Failed to retrieve webhook event", zap.String("event_id", eventID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to retrieve webhook event",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook event retrieved successfully",
		"data":    event,
	})
}

// ReplayWebhookEvent replays a previously processed webhook event
func (h *WebhookHandler) ReplayWebhookEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Event ID is required",
		})
		return
	}

	// Replay event using service
	ctx := context.Background()
	err := h.webhookService.ReplayEvent(ctx, eventID)
	if err != nil {
		h.logger.Error("Failed to replay webhook event", zap.String("event_id", eventID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to replay webhook event",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook event replayed successfully",
	})
}

// WebhookMiddleware is a middleware to verify webhook signatures before processing
func (h *WebhookHandler) WebhookMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to webhook endpoints
		if !isWebhookEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get the Stripe signature header
		signature := c.GetHeader("Stripe-Signature")
		if signature == "" {
			h.logger.Error("Missing Stripe signature header")
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Missing Stripe signature header",
			})
			c.Abort()
			return
		}

		// Read the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			h.logger.Error("Failed to read webhook request body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Failed to read request body",
			})
			c.Abort()
			return
		}

		// Restore the request body for subsequent handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Verify the signature
		_, err = h.webhookService.VerifyWebhookSignature(body, signature)
		if err != nil {
			h.logger.Error("Webhook signature verification failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Webhook signature verification failed",
			})
			c.Abort()
			return
		}

		// Store the verified body and signature in the context for later use
		c.Set("webhook_body", body)
		c.Set("webhook_signature", signature)

		c.Next()
	}
}

// Helper function to check if the endpoint is a webhook endpoint
func isWebhookEndpoint(path string) bool {
	return path == "/api/v1/webhooks/stripe"
}

// Helper function to parse integer from string with error handling
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}