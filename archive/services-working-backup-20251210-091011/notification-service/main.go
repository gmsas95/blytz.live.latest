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

// Notification struct
type Notification struct {
	ID             string               `json:"id"`
	UserID         string               `json:"user_id"`
	Type           string               `json:"type"` // "message", "auction", "order", "payment", "shipment", "system"
	Title          string               `json:"title"`
	Message        string               `json:"message"`
	Content        string               `json:"content,omitempty"`
	Image          string               `json:"image,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Priority       string               `json:"priority"` // "low", "normal", "high", "urgent"
	Status         string               `json:"status"` // "unread", "read", "archived"
	ReadAt         *time.Time           `json:"read_at,omitempty"`
	ArchivedAt     *time.Time           `json:"archived_at,omitempty"`
	ExpiresAt      *time.Time           `json:"expires_at,omitempty"`
	ActionType     string               `json:"action_type,omitempty"` // "view", "accept", "reject", "custom"
	ActionURL      string               `json:"action_url,omitempty"`
	ActionText     string               `json:"action_text,omitempty"`
	Settings       NotificationSettings  `json:"settings"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// NotificationSettings struct
type NotificationSettings struct {
	EmailEnabled    bool `json:"email_enabled"`
	PushEnabled     bool `json:"push_enabled"`
	SMSEnabled      bool `json:"sms_enabled"`
	InAppEnabled    bool `json:"in_app_enabled"`
	EmailSent       bool `json:"email_sent,omitempty"`
	PushSent        bool `json:"push_sent,omitempty"`
	SMSSent         bool `json:"sms_sent,omitempty"`
	InAppDelivered  bool `json:"in_app_delivered,omitempty"`
}

// UserNotificationPreferences struct
type UserNotificationPreferences struct {
	UserID                string `json:"user_id"`
	MessageNotifications   bool   `json:"message_notifications"`
	AuctionNotifications   bool   `json:"auction_notifications"`
	OrderNotifications     bool   `json:"order_notifications"`
	PaymentNotifications   bool   `json:"payment_notifications"`
	ShipmentNotifications  bool   `json:"shipment_notifications"`
	SystemNotifications    bool   `json:"system_notifications"`
	EmailNotifications     bool   `json:"email_notifications"`
	PushNotifications      bool   `json:"push_notifications"`
	SMSNotifications       bool   `json:"sms_notifications"`
	QuietHours            bool   `json:"quiet_hours"`
	QuietStartTime        string `json:"quiet_start_time"` // HH:MM format
	QuietEndTime          string `json:"quiet_end_time"`   // HH:MM format
	Timezone              string `json:"timezone"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// Request structs
type CreateNotificationRequest struct {
	UserID     string                 `json:"user_id"`
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Content    string                 `json:"content,omitempty"`
	Image      string                 `json:"image,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Priority   string                 `json:"priority"`
	ActionType string                 `json:"action_type,omitempty"`
	ActionURL  string                 `json:"action_url,omitempty"`
	ActionText string                 `json:"action_text,omitempty"`
	ExpiresIn  int                    `json:"expires_in,omitempty"` // minutes
}

type UpdateNotificationRequest struct {
	Status   string    `json:"status,omitempty"`
	ReadAt   *time.Time `json:"read_at,omitempty"`
	Archived *bool      `json:"archived,omitempty"`
}

type UpdatePreferencesRequest struct {
	UserID string `json:"user_id"`
	MessageNotifications  bool `json:"message_notifications,omitempty"`
	AuctionNotifications  bool `json:"auction_notifications,omitempty"`
	OrderNotifications    bool `json:"order_notifications,omitempty"`
	PaymentNotifications  bool `json:"payment_notifications,omitempty"`
	ShipmentNotifications bool `json:"shipment_notifications,omitempty"`
	SystemNotifications   bool `json:"system_notifications,omitempty"`
	EmailNotifications    bool `json:"email_notifications,omitempty"`
	PushNotifications     bool `json:"push_notifications,omitempty"`
	SMSNotifications      bool `json:"sms_notifications,omitempty"`
	QuietHours           bool `json:"quiet_hours,omitempty"`
	QuietStartTime       string `json:"quiet_start_time,omitempty"`
	QuietEndTime         string `json:"quiet_end_time,omitempty"`
	Timezone             string `json:"timezone,omitempty"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Notification service with in-memory storage
type NotificationService struct {
	notifications []Notification
	preferences   []UserNotificationPreferences
	mu            sync.RWMutex
}

// New notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		notifications: []Notification{
			{
				ID:          "notification-1",
				UserID:      "user-1",
				Type:        "auction",
				Title:       "Auction Ending Soon",
				Message:     "Your bid on 'Vintage Camera' is currently winning. Auction ends in 10 minutes!",
				Image:       "https://picsum.photos/100/100?random=1",
				Priority:    "high",
				Status:      "unread",
				ActionType:  "view",
				ActionURL:   "/auctions/auction-1",
				ActionText:  "View Auction",
				Data: map[string]interface{}{
					"auction_id": "auction-1",
					"product_id": "product-1",
					"current_bid": 450.00,
				},
				Settings: NotificationSettings{
					EmailEnabled:   true,
					PushEnabled:    true,
					SMSEnabled:     false,
					InAppEnabled:   true,
					EmailSent:      true,
					PushSent:       true,
					InAppDelivered: true,
				},
				CreatedAt: time.Now().Add(-5 * time.Minute),
				UpdatedAt: time.Now().Add(-5 * time.Minute),
			},
			{
				ID:          "notification-2",
				UserID:      "user-1",
				Type:        "message",
				Title:       "New Message",
				Message:     "You received a new message from John Seller about your inquiry on the vintage camera.",
				Content:     "Hi! I'm interested in your vintage camera. Is it still available?",
				Priority:    "normal",
				Status:      "read",
				ReadAt:      &[]time.Time{time.Now().Add(-30 * time.Minute)}[0],
				ActionType:  "view",
				ActionURL:   "/chats/chat-1",
				ActionText:  "View Chat",
				Data: map[string]interface{}{
					"chat_id": "chat-1",
					"sender_id": "user-2",
					"sender_name": "John Seller",
				},
				Settings: NotificationSettings{
					EmailEnabled:   false,
					PushEnabled:    true,
					SMSEnabled:     false,
					InAppEnabled:   true,
					PushSent:       true,
					InAppDelivered: true,
				},
				CreatedAt: time.Now().Add(-45 * time.Minute),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			},
			{
				ID:          "notification-3",
				UserID:      "user-2",
				Type:        "order",
				Title:       "Order Shipped",
				Message:     "Your order #order-1 has been shipped! Track your package with tracking number: TRACK123456",
				Priority:    "normal",
				Status:      "unread",
				ActionType:  "view",
				ActionURL:   "/logistics/shipment-1",
				ActionText:  "Track Package",
				Data: map[string]interface{}{
					"order_id": "order-1",
					"shipment_id": "shipment-1",
					"tracking_number": "TRACK123456",
					"carrier": "FedEx",
				},
				Settings: NotificationSettings{
					EmailEnabled:   true,
					PushEnabled:    true,
					SMSEnabled:     true,
					InAppEnabled:   true,
					EmailSent:      true,
					PushSent:       true,
					SMSSent:        true,
					InAppDelivered: true,
				},
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
			},
			{
				ID:          "notification-4",
				UserID:      "user-3",
				Type:        "payment",
				Title:       "Payment Successful",
				Message:     "Your payment of $450.00 for order #order-2 has been successfully processed.",
				Priority:    "high",
				Status:      "read",
				ReadAt:      &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
				ActionType:  "view",
				ActionURL:   "/payments/payment-2",
				ActionText:  "View Receipt",
				Data: map[string]interface{}{
					"payment_id": "payment-2",
					"order_id": "order-2",
					"amount": 450.00,
					"payment_method": "paypal",
				},
				Settings: NotificationSettings{
					EmailEnabled:   true,
					PushEnabled:    true,
					SMSEnabled:     false,
					InAppEnabled:   true,
					EmailSent:      true,
					PushSent:       true,
					InAppDelivered: true,
				},
				CreatedAt: time.Now().Add(-3 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
		},
		preferences: []UserNotificationPreferences{
			{
				UserID:                "user-1",
				MessageNotifications:   true,
				AuctionNotifications:   true,
				OrderNotifications:     true,
				PaymentNotifications:   true,
				ShipmentNotifications:  true,
				SystemNotifications:    true,
				EmailNotifications:     true,
				PushNotifications:      true,
				SMSNotifications:       false,
				QuietHours:            false,
				QuietStartTime:        "22:00",
				QuietEndTime:          "08:00",
				Timezone:              "America/New_York",
				UpdatedAt:             time.Now().Add(-24 * time.Hour),
			},
			{
				UserID:                "user-2",
				MessageNotifications:   true,
				AuctionNotifications:   true,
				OrderNotifications:     true,
				PaymentNotifications:   true,
				ShipmentNotifications:  true,
				SystemNotifications:    false,
				EmailNotifications:     true,
				PushNotifications:      true,
				SMSNotifications:       true,
				QuietHours:            true,
				QuietStartTime:        "22:00",
				QuietEndTime:          "08:00",
				Timezone:              "America/Los_Angeles",
				UpdatedAt:             time.Now().Add(-12 * time.Hour),
			},
			{
				UserID:                "user-3",
				MessageNotifications:   false,
				AuctionNotifications:   true,
				OrderNotifications:     true,
				PaymentNotifications:   true,
				ShipmentNotifications:  true,
				SystemNotifications:    true,
				EmailNotifications:     true,
				PushNotifications:      false,
				SMSNotifications:       false,
				QuietHours:            false,
				QuietStartTime:        "23:00",
				QuietEndTime:          "09:00",
				Timezone:              "America/Chicago",
				UpdatedAt:             time.Now().Add(-6 * time.Hour),
			},
		},
	}
}

// Check if user is in quiet hours
func (s *NotificationService) isInQuietHours(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, pref := range s.preferences {
		if pref.UserID == userID && pref.QuietHours {
			// Simplified check - in real app, would use timezone-aware time
			now := time.Now()
			hour := now.Hour()
			
			// Parse quiet hours (simplified - assumes same day)
			startHour := 22 // Default 22:00
			endHour := 8    // Default 08:00
			
			if pref.QuietStartTime != "" {
				if parsed, err := strconv.Atoi(pref.QuietStartTime[:2]); err == nil {
					startHour = parsed
				}
			}
			
			if pref.QuietEndTime != "" {
				if parsed, err := strconv.Atoi(pref.QuietEndTime[:2]); err == nil {
					endHour = parsed
				}
			}
			
			// Check if current time is in quiet hours
			if startHour > endHour {
				// Overnight quiet hours (e.g., 22:00 to 08:00)
				return hour >= startHour || hour < endHour
			} else {
				// Same day quiet hours (e.g., 02:00 to 06:00)
				return hour >= startHour && hour < endHour
			}
		}
	}
	return false
}

// Check notification preferences
func (s *NotificationService) shouldSendNotification(userID, notificationType string) (bool, bool, bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, pref := range s.preferences {
		if pref.UserID == userID {
			// Check specific notification type preferences
			enabled := false
			switch notificationType {
			case "message":
				enabled = pref.MessageNotifications
			case "auction":
				enabled = pref.AuctionNotifications
			case "order":
				enabled = pref.OrderNotifications
			case "payment":
				enabled = pref.PaymentNotifications
			case "shipment":
				enabled = pref.ShipmentNotifications
			case "system":
				enabled = pref.SystemNotifications
			}

			// Check channel preferences
			if enabled {
				emailEnabled := pref.EmailNotifications
				pushEnabled := pref.PushNotifications
				smsEnabled := pref.SMSNotifications
				inAppEnabled := true // Always enabled for in-app
				
				return emailEnabled, pushEnabled, smsEnabled, inAppEnabled
			}
			break
		}
	}
	
	// Default preferences
	return false, false, false, true
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
	perPage := 20 // More notifications per page

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
func paginateNotifications(items []Notification, page, perPage int) ([]Notification, map[string]interface{}) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Notification{}, map[string]interface{}{
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
func (s *NotificationService) health(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	unreadNotifications := 0
	totalNotifications := len(s.notifications)
	
	for _, notification := range s.notifications {
		if notification.Status == "unread" {
			unreadNotifications++
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Notification service is healthy and working!",
		Data: map[string]interface{}{
			"service":             "notification-service",
			"version":             "v2.0-working",
			"status":              "healthy",
			"timestamp":           time.Now(),
			"total_notifications": totalNotifications,
			"unread_notifications": unreadNotifications,
			"user_preferences":    len(s.preferences),
		},
	})
}

// Get all notifications
func (s *NotificationService) getAllNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID is required",
			Error:   "No user ID provided",
		})
		return
	}

	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter notifications by user
	var userNotifications []Notification
	for _, notification := range s.notifications {
		if notification.UserID == userID {
			userNotifications = append(userNotifications, notification)
		}
	}

	// Apply status filter if provided
	if status := strings.ToLower(r.URL.Query().Get("status")); status != "" {
		var statusFilteredNotifications []Notification
		for _, notification := range userNotifications {
			if strings.Contains(strings.ToLower(notification.Status), status) {
				statusFilteredNotifications = append(statusFilteredNotifications, notification)
			}
		}
		userNotifications = statusFilteredNotifications
	}

	// Apply type filter if provided
	if notificationType := strings.ToLower(r.URL.Query().Get("type")); notificationType != "" {
		var typeFilteredNotifications []Notification
		for _, notification := range userNotifications {
			if strings.Contains(strings.ToLower(notification.Type), notificationType) {
				typeFilteredNotifications = append(typeFilteredNotifications, notification)
			}
		}
		userNotifications = typeFilteredNotifications
	}

	// Sort notifications by creation date (newest first)
	for i := 0; i < len(userNotifications)-1; i++ {
		for j := i + 1; j < len(userNotifications); j++ {
			if userNotifications[i].CreatedAt.Before(userNotifications[j].CreatedAt) {
				userNotifications[i], userNotifications[j] = userNotifications[j], userNotifications[i]
			}
		}
	}

	// Paginate results
	paginatedNotifications, pagination := paginateNotifications(userNotifications, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Notifications retrieved successfully",
		Data: map[string]interface{}{
			"notifications": paginatedNotifications,
			"pagination":    pagination,
			"user_id":       userID,
		},
	})
}

// Get notification by ID
func (s *NotificationService) getNotification(w http.ResponseWriter, r *http.Request) {
	notificationID := strings.TrimPrefix(r.URL.Path, "/api/v1/notifications/")
	if notificationID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Notification ID is required",
			Error:   "No notification ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, notification := range s.notifications {
		if notification.ID == notificationID {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Notification retrieved successfully",
				Data: map[string]interface{}{
					"notification": notification,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Notification not found",
		Error:   "Notification with ID " + notificationID + " does not exist",
	})
}

// Create notification
func (s *NotificationService) createNotification(w http.ResponseWriter, r *http.Request) {
	var req CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid notification data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Type == "" || req.Title == "" || req.Message == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID, type, title, and message are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.Priority == "" {
		req.Priority = "normal"
	}

	// Validate notification type
	validTypes := map[string]bool{
		"message":   true,
		"auction":   true,
		"order":     true,
		"payment":   true,
		"shipment":  true,
		"system":    true,
	}
	if !validTypes[req.Type] {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid notification type",
			Error:   "Notification type must be one of: message, auction, order, payment, shipment, system",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check notification preferences
	emailEnabled, pushEnabled, smsEnabled, inAppEnabled := s.shouldSendNotification(req.UserID, req.Type)
	
	// Check quiet hours (only affects email/push/SMS, not in-app)
	inQuietHours := s.isInQuietHours(req.UserID)
	if inQuietHours {
		emailEnabled = false
		pushEnabled = false
		smsEnabled = false
	}

	// Create new notification
	newNotification := Notification{
		ID:          fmt.Sprintf("notification-%d", time.Now().UnixNano()),
		UserID:      req.UserID,
		Type:        req.Type,
		Title:       req.Title,
		Message:     req.Message,
		Content:     req.Content,
		Image:       req.Image,
		Data:        req.Data,
		Priority:    req.Priority,
		Status:      "unread",
		ActionType:  req.ActionType,
		ActionURL:   req.ActionURL,
		ActionText:  req.ActionText,
		Settings: NotificationSettings{
			EmailEnabled:   emailEnabled,
			PushEnabled:    pushEnabled,
			SMSEnabled:     smsEnabled,
			InAppEnabled:   inAppEnabled,
			EmailSent:      emailEnabled,
			PushSent:       pushEnabled,
			SMSSent:        smsEnabled,
			InAppDelivered: inAppEnabled,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set expiration if provided
	if req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Minute)
		newNotification.ExpiresAt = &expiresAt
	}

	s.notifications = append(s.notifications, newNotification)

	log.Printf("Notification created: %s for user %s (%s)", newNotification.ID, newNotification.UserID, newNotification.Type)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Notification created successfully",
		Data: map[string]interface{}{
			"notification": newNotification,
		},
	})
}

// Update notification
func (s *NotificationService) updateNotification(w http.ResponseWriter, r *http.Request) {
	notificationID := strings.TrimPrefix(r.URL.Path, "/api/v1/notifications/")
	if notificationID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Notification ID is required",
			Error:   "No notification ID provided",
		})
		return
	}

	var req UpdateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid update data",
			Error:   err.Error(),
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and update notification
	for i := range s.notifications {
		if s.notifications[i].ID == notificationID {
			if req.Status != "" {
				s.notifications[i].Status = req.Status
				if req.Status == "read" && req.ReadAt == nil {
					now := time.Now()
					s.notifications[i].ReadAt = &now
				}
			}
			
			if req.Archived != nil && *req.Archived {
				s.notifications[i].Status = "archived"
				now := time.Now()
				s.notifications[i].ArchivedAt = &now
			}

			s.notifications[i].UpdatedAt = time.Now()

			log.Printf("Notification updated: %s", notificationID)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Notification updated successfully",
				Data: map[string]interface{}{
					"notification": s.notifications[i],
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Notification not found",
		Error:   "Notification with ID " + notificationID + " does not exist",
	})
}

// Mark all notifications as read
func (s *NotificationService) markAllAsRead(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if req.UserID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID is required",
			Error:   "No user ID provided",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	readCount := 0

	// Mark all user notifications as read
	for i := range s.notifications {
		if s.notifications[i].UserID == req.UserID && s.notifications[i].Status == "unread" {
			s.notifications[i].Status = "read"
			s.notifications[i].ReadAt = &now
			s.notifications[i].UpdatedAt = now
			readCount++
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "All notifications marked as read",
		Data: map[string]interface{}{
			"user_id":    req.UserID,
			"read_count": readCount,
		},
	})
}

// Get user preferences
func (s *NotificationService) getUserPreferences(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID is required",
			Error:   "No user ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, pref := range s.preferences {
		if pref.UserID == userID {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "User preferences retrieved successfully",
				Data: map[string]interface{}{
					"preferences": pref,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "User preferences not found",
		Error:   "Preferences for user " + userID + " do not exist",
	})
}

// Update user preferences
func (s *NotificationService) updateUserPreferences(w http.ResponseWriter, r *http.Request) {
	var req UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid preference data",
			Error:   err.Error(),
		})
		return
	}

	if req.UserID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID is required",
			Error:   "No user ID provided",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and update preferences
	for i := range s.preferences {
		if s.preferences[i].UserID == req.UserID {
			// Update fields if provided
			if req.MessageNotifications {
				s.preferences[i].MessageNotifications = req.MessageNotifications
			}
			if req.AuctionNotifications {
				s.preferences[i].AuctionNotifications = req.AuctionNotifications
			}
			if req.OrderNotifications {
				s.preferences[i].OrderNotifications = req.OrderNotifications
			}
			if req.PaymentNotifications {
				s.preferences[i].PaymentNotifications = req.PaymentNotifications
			}
			if req.ShipmentNotifications {
				s.preferences[i].ShipmentNotifications = req.ShipmentNotifications
			}
			if req.SystemNotifications {
				s.preferences[i].SystemNotifications = req.SystemNotifications
			}
			if req.EmailNotifications {
				s.preferences[i].EmailNotifications = req.EmailNotifications
			}
			if req.PushNotifications {
				s.preferences[i].PushNotifications = req.PushNotifications
			}
			if req.SMSNotifications {
				s.preferences[i].SMSNotifications = req.SMSNotifications
			}
			if req.QuietHours {
				s.preferences[i].QuietHours = req.QuietHours
			}
			if req.QuietStartTime != "" {
				s.preferences[i].QuietStartTime = req.QuietStartTime
			}
			if req.QuietEndTime != "" {
				s.preferences[i].QuietEndTime = req.QuietEndTime
			}
			if req.Timezone != "" {
				s.preferences[i].Timezone = req.Timezone
			}

			s.preferences[i].UpdatedAt = time.Now()

			log.Printf("User preferences updated: %s", req.UserID)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "User preferences updated successfully",
				Data: map[string]interface{}{
					"preferences": s.preferences[i],
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "User preferences not found",
		Error:   "Preferences for user " + req.UserID + " do not exist",
	})
}

// Get notification statistics
func (s *NotificationService) getNotificationStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_notifications":  len(s.notifications),
		"unread_notifications": 0,
		"read_notifications":   0,
		"archived_notifications": 0,
		"notification_types": make(map[string]int),
		"notification_priorities": make(map[string]int),
		"delivery_stats": map[string]int{
			"email_sent":    0,
			"push_sent":     0,
			"sms_sent":      0,
			"in_app_delivered": 0,
		},
	}

	for _, notification := range s.notifications {
		// Count by status
		switch notification.Status {
		case "unread":
			stats["unread_notifications"] = stats["unread_notifications"].(int) + 1
		case "read":
			stats["read_notifications"] = stats["read_notifications"].(int) + 1
		case "archived":
			stats["archived_notifications"] = stats["archived_notifications"].(int) + 1
		}

		// Count by type
		if count, exists := stats["notification_types"].(map[string]int)[notification.Type]; exists {
			stats["notification_types"].(map[string]int)[notification.Type] = count + 1
		} else {
			stats["notification_types"].(map[string]int)[notification.Type] = 1
		}

		// Count by priority
		if count, exists := stats["notification_priorities"].(map[string]int)[notification.Priority]; exists {
			stats["notification_priorities"].(map[string]int)[notification.Priority] = count + 1
		} else {
			stats["notification_priorities"].(map[string]int)[notification.Priority] = 1
		}

		// Count delivery stats
		if notification.Settings.EmailSent {
			stats["delivery_stats"].(map[string]int)["email_sent"]++
		}
		if notification.Settings.PushSent {
			stats["delivery_stats"].(map[string]int)["push_sent"]++
		}
		if notification.Settings.SMSSent {
			stats["delivery_stats"].(map[string]int)["sms_sent"]++
		}
		if notification.Settings.InAppDelivered {
			stats["delivery_stats"].(map[string]int)["in_app_delivered"]++
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Notification statistics retrieved successfully",
		Data: map[string]interface{}{
			"stats": stats,
		},
	})
}

// Start notification service
func main() {
	service := NewNotificationService()

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/notifications", corsMiddleware(service.getAllNotifications))
	http.Handle("/api/v1/notifications/", corsMiddleware(service.getNotification)) // For GET by ID and update
	http.Handle("/api/v1/notifications/create", corsMiddleware(service.createNotification))
	http.Handle("/api/v1/notifications/mark-all-read", corsMiddleware(service.markAllAsRead))
	http.Handle("/api/v1/notifications/preferences", corsMiddleware(service.getUserPreferences))
	http.Handle("/api/v1/notifications/update-preferences", corsMiddleware(service.updateUserPreferences))
	http.Handle("/api/v1/notifications/stats", corsMiddleware(service.getNotificationStats))

	port := ":8094"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	fmt.Printf("🚀 NOTIFICATION SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)
	fmt.Printf("🔔 Notifications list: http://localhost%s/api/v1/notifications?user_id={id}\n", port)
	fmt.Printf("📝 Notification details: http://localhost%s/api/v1/notifications/{id}\n", port)
	fmt.Printf("➕ Create notification: http://localhost%s/api/v1/notifications/create\n", port)
	fmt.Printf("✅ Mark all read: http://localhost%s/api/v1/notifications/mark-all-read\n", port)
	fmt.Printf("⚙️  User preferences: http://localhost%s/api/v1/notifications/preferences?user_id={id}\n", port)
	fmt.Printf("📈 Notification stats: http://localhost%s/api/v1/notifications/stats\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("🔔 Total notifications: %d\n", len(service.notifications))
	fmt.Printf("👥 User preferences: %d\n", len(service.preferences))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(port, nil))
}