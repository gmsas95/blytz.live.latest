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

// Message struct
type Message struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	ReceiverID string   `json:"receiver_id,omitempty"`
	Content   string    `json:"content"`
	MessageType string   `json:"message_type"` // "text", "image", "file", "product_share", "order_share"
	Attachments []Attachment `json:"attachments,omitempty"`
	IsRead    bool      `json:"is_read"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	IsEdited  bool      `json:"is_edited"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Attachment struct
type Attachment struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "image", "file", "audio", "video"
	Name     string `json:"name"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// Chat struct
type Chat struct {
	ID         string    `json:"id"`
	Participants []string `json:"participants"` // User IDs
	Type       string    `json:"type"` // "direct", "group"
	Title      string    `json:"title,omitempty"`
	LastMessage *Message  `json:"last_message,omitempty"`
	UnreadCounts map[string]int `json:"unread_counts,omitempty"` // user_id -> count
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Request structs
type SendMessageRequest struct {
	SenderID    string        `json:"sender_id"`
	ChatID       string        `json:"chat_id"`
	ReceiverID   string        `json:"receiver_id,omitempty"`
	Content      string        `json:"content"`
	MessageType  string        `json:"message_type"`
	Attachments  []Attachment  `json:"attachments,omitempty"`
}

type CreateChatRequest struct {
	Participants []string `json:"participants"`
	Type         string    `json:"type"`
	Title        string    `json:"title,omitempty"`
}

type UpdateMessageRequest struct {
	Content     string `json:"content"`
	IsEdited    bool   `json:"is_edited"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Chat service with in-memory storage
type ChatService struct {
	messages []Message
	chats    []Chat
	mu       sync.RWMutex
}

// New chat service
func NewChatService() *ChatService {
	return &ChatService{
		messages: []Message{
			{
				ID:        "msg-1",
				ChatID:    "chat-1",
				SenderID:  "user-1",
				ReceiverID: "user-2",
				Content:   "Hi! I'm interested in your vintage camera. Is it still available?",
				MessageType: "text",
				IsRead:    true,
				ReadAt:    &[]time.Time{time.Now().Add(-50 * time.Minute)}[0],
				CreatedAt: time.Now().Add(-1 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			{
				ID:        "msg-2",
				ChatID:    "chat-1",
				SenderID:  "user-2",
				ReceiverID: "user-1",
				Content:   "Yes, it's still available! It's in excellent condition. Would you like to see more photos?",
				MessageType: "text",
				IsRead:    true,
				ReadAt:    &[]time.Time{time.Now().Add(-45 * time.Minute)}[0],
				CreatedAt: time.Now().Add(-55 * time.Minute),
				UpdatedAt: time.Now().Add(-55 * time.Minute),
			},
			{
				ID:        "msg-3",
				ChatID:    "chat-1",
				SenderID:  "user-1",
				ReceiverID: "user-2",
				Content:   "Yes, please! Also, what's your best price if I buy it now?",
				MessageType: "text",
				IsRead:    false,
				CreatedAt: time.Now().Add(-30 * time.Minute),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			},
			{
				ID:        "msg-4",
				ChatID:    "chat-2",
				SenderID:  "user-3",
				ReceiverID: "seller-2",
				Content:   "Hello, I won the auction for the designer handbag. How do we proceed with payment and shipping?",
				MessageType: "text",
				IsRead:    true,
				ReadAt:    &[]time.Time{time.Now().Add(-20 * time.Minute)}[0],
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
			},
			{
				ID:        "msg-5",
				ChatID:    "chat-2",
				SenderID:  "seller-2",
				ReceiverID: "user-3",
				Content:   "Congratulations on winning! I'll send you the payment link. Do you prefer PayPal or bank transfer?",
				MessageType: "text",
				IsRead:    false,
				CreatedAt: time.Now().Add(-25 * time.Minute),
				UpdatedAt: time.Now().Add(-25 * time.Minute),
			},
		},
		chats: []Chat{
			{
				ID:        "chat-1",
				Participants: []string{"user-1", "user-2"},
				Type:      "direct",
				UnreadCounts: map[string]int{
					"user-1": 0,
					"user-2": 0,
				},
				CreatedAt: time.Now().Add(-24 * time.Hour),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			},
			{
				ID:        "chat-2",
				Participants: []string{"user-3", "seller-2"},
				Type:      "direct",
				UnreadCounts: map[string]int{
					"user-3": 0,
					"seller-2": 1,
				},
				CreatedAt: time.Now().Add(-12 * time.Hour),
				UpdatedAt: time.Now().Add(-25 * time.Minute),
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
	perPage := 20 // More messages per page for chat

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
func paginateMessages(items []Message, page, perPage int) ([]Message, map[string]interface{}) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Message{}, map[string]interface{}{
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

func paginateChats(items []Chat, page, perPage int) ([]Chat, map[string]interface{}) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Chat{}, map[string]interface{}{
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

// Update unread counts
func (s *ChatService) updateUnreadCounts(chatID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate unread counts for each participant
	unreadCounts := make(map[string]int)
	participants := make([]string, 0)

	// Get chat participants
	for _, chat := range s.chats {
		if chat.ID == chatID {
			participants = chat.Participants
			break
		}
	}

	if len(participants) == 0 {
		return
	}

	// Count unread messages for each participant
	for _, message := range s.messages {
		if message.ChatID == chatID && !message.IsRead {
			// Mark as unread for all participants except sender
			for _, participant := range participants {
				if participant != message.SenderID {
					unreadCounts[participant]++
				}
			}
		}
	}

	// Update chat with new counts
	for i := range s.chats {
		if s.chats[i].ID == chatID {
			s.chats[i].UnreadCounts = unreadCounts
			break
		}
	}
}

// Update last message
func (s *ChatService) updateLastMessage(chatID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the most recent message
	var lastMessage *Message
	lastTime := time.Time{}

	for _, message := range s.messages {
		if message.ChatID == chatID && message.CreatedAt.After(lastTime) {
			lastMessage = &message
			lastTime = message.CreatedAt
		}
	}

	// Update chat with last message
	for i := range s.chats {
		if s.chats[i].ID == chatID {
			s.chats[i].LastMessage = lastMessage
			s.chats[i].UpdatedAt = time.Now()
			break
		}
	}
}

// Health check
func (s *ChatService) health(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalUnread := 0
	for _, chat := range s.chats {
		for _, count := range chat.UnreadCounts {
			totalUnread += count
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Chat service is healthy and working!",
		Data: map[string]interface{}{
			"service":        "chat-service",
			"version":        "v2.0-working",
			"status":         "healthy",
			"timestamp":      time.Now(),
			"total_chats":    len(s.chats),
			"total_messages": len(s.messages),
			"total_unread":   totalUnread,
		},
	})
}

// Get all chats for user
func (s *ChatService) getUserChats(w http.ResponseWriter, r *http.Request) {
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

	// Get chats where user is a participant
	var userChats []Chat
	for _, chat := range s.chats {
		for _, participant := range chat.Participants {
			if participant == userID {
				userChats = append(userChats, chat)
				break
			}
		}
	}

	// Sort chats by last message time (newest first)
	for i := 0; i < len(userChats)-1; i++ {
		for j := i + 1; j < len(userChats); j++ {
			timeI := userChats[i].UpdatedAt
			timeJ := userChats[j].UpdatedAt
			if timeI.Before(timeJ) {
				userChats[i], userChats[j] = userChats[j], userChats[i]
			}
		}
	}

	// Paginate results
	paginatedChats, pagination := paginateChats(userChats, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "User chats retrieved successfully",
		Data: map[string]interface{}{
			"chats":      paginatedChats,
			"pagination": pagination,
			"user_id":    userID,
		},
	})
}

// Get chat by ID
func (s *ChatService) getChat(w http.ResponseWriter, r *http.Request) {
	chatID := strings.TrimPrefix(r.URL.Path, "/api/v1/chats/")
	if chatID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Chat ID is required",
			Error:   "No chat ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, chat := range s.chats {
		if chat.ID == chatID {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Chat retrieved successfully",
				Data: map[string]interface{}{
					"chat": chat,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Chat not found",
		Error:   "Chat with ID " + chatID + " does not exist",
	})
}

// Create chat
func (s *ChatService) createChat(w http.ResponseWriter, r *http.Request) {
	var req CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid chat data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if len(req.Participants) < 2 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "At least 2 participants are required",
			Error:   "Invalid participants",
		})
		return
	}

	if req.Type == "" {
		req.Type = "direct"
	}

	// Validate chat type
	validTypes := map[string]bool{
		"direct": true,
		"group":   true,
	}
	if !validTypes[req.Type] {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid chat type",
			Error:   "Chat type must be 'direct' or 'group'",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if direct chat already exists
	if req.Type == "direct" && len(req.Participants) == 2 {
		for _, chat := range s.chats {
			if chat.Type == "direct" && len(chat.Participants) == 2 {
				// Check if participants match (order doesn't matter)
				match := true
				for _, p := range req.Participants {
					found := false
					for _, cp := range chat.Participants {
						if p == cp {
							found = true
							break
						}
					}
					if !found {
						match = false
						break
					}
				}
				if match {
					writeJSONResponse(w, http.StatusConflict, Response{
						Success: false,
						Message: "Direct chat already exists",
						Error:   "Chat with these participants already exists",
					})
					return
				}
			}
		}
	}

	// Create new chat
	newChat := Chat{
		ID:           fmt.Sprintf("chat-%d", time.Now().UnixNano()),
		Participants: req.Participants,
		Type:         req.Type,
		Title:        req.Title,
		UnreadCounts: make(map[string]int),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Initialize unread counts for all participants
	for _, participant := range req.Participants {
		newChat.UnreadCounts[participant] = 0
	}

	s.chats = append(s.chats, newChat)

	log.Printf("Chat created: %s with %d participants", newChat.ID, len(req.Participants))

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Chat created successfully",
		Data: map[string]interface{}{
			"chat": newChat,
		},
	})
}

// Get chat messages
func (s *ChatService) getChatMessages(w http.ResponseWriter, r *http.Request) {
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Chat ID is required",
			Error:   "No chat ID provided",
		})
		return
	}

	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get messages for this chat
	var chatMessages []Message
	for _, message := range s.messages {
		if message.ChatID == chatID {
			chatMessages = append(chatMessages, message)
		}
	}

	// Sort messages by creation date (oldest first for chat)
	for i := 0; i < len(chatMessages)-1; i++ {
		for j := i + 1; j < len(chatMessages); j++ {
			if chatMessages[i].CreatedAt.After(chatMessages[j].CreatedAt) {
				chatMessages[i], chatMessages[j] = chatMessages[j], chatMessages[i]
			}
		}
	}

	// Paginate results
	paginatedMessages, pagination := paginateMessages(chatMessages, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Chat messages retrieved successfully",
		Data: map[string]interface{}{
			"messages":   paginatedMessages,
			"pagination": pagination,
			"chat_id":    chatID,
		},
	})
}

// Send message
func (s *ChatService) sendMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid message data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.SenderID == "" || req.Content == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Sender ID and content are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Validate message type
	validTypes := map[string]bool{
		"text":          true,
		"image":         true,
		"file":          true,
		"product_share": true,
		"order_share":   true,
	}
	if !validTypes[req.MessageType] {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid message type",
			Error:   "Message type must be one of: text, image, file, product_share, order_share",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Determine chat ID
	chatID := req.ChatID
	if chatID == "" && req.ReceiverID != "" {
		// Find or create direct chat
		for _, chat := range s.chats {
			if chat.Type == "direct" && len(chat.Participants) == 2 {
				// Check if participants match
				match := true
				participants := []string{req.SenderID, req.ReceiverID}
				for _, p := range participants {
					found := false
					for _, cp := range chat.Participants {
						if p == cp {
							found = true
							break
						}
					}
					if !found {
						match = false
						break
					}
				}
				if match {
					chatID = chat.ID
					break
				}
			}
		}

		// Create new chat if not found
		if chatID == "" {
			newChat := Chat{
				ID:           fmt.Sprintf("chat-%d", time.Now().UnixNano()),
				Participants: []string{req.SenderID, req.ReceiverID},
				Type:         "direct",
				UnreadCounts: map[string]int{
					req.SenderID:  0,
					req.ReceiverID: 0,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			s.chats = append(s.chats, newChat)
			chatID = newChat.ID
		}
	}

	// Create new message
	newMessage := Message{
		ID:          fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		ChatID:      chatID,
		SenderID:    req.SenderID,
		ReceiverID:  req.ReceiverID,
		Content:     req.Content,
		MessageType: req.MessageType,
		Attachments: req.Attachments,
		IsRead:      false,
		IsEdited:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.messages = append(s.messages, newMessage)

	// Update chat metadata
	s.updateLastMessage(chatID)
	s.updateUnreadCounts(chatID)

	log.Printf("Message sent: %s in chat %s by %s", newMessage.ID, chatID, req.SenderID)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Message sent successfully",
		Data: map[string]interface{}{
			"message": newMessage,
		},
	})
}

// Mark message as read
func (s *ChatService) markMessageRead(w http.ResponseWriter, r *http.Request) {
	messageID := strings.TrimPrefix(r.URL.Path, "/api/v1/messages/")
	if messageID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Message ID is required",
			Error:   "No message ID provided",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and update message
	var foundMessage *Message
	for i := range s.messages {
		if s.messages[i].ID == messageID {
			s.messages[i].IsRead = true
			now := time.Now()
			s.messages[i].ReadAt = &now
			s.messages[i].UpdatedAt = now
			foundMessage = &s.messages[i]
			break
		}
	}

	if foundMessage == nil {
		writeJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Message: "Message not found",
			Error:   "Message with ID " + messageID + " does not exist",
		})
		return
	}

	// Update unread counts
	s.updateUnreadCounts(foundMessage.ChatID)

	log.Printf("Message marked as read: %s", messageID)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Message marked as read",
		Data: map[string]interface{}{
			"message": *foundMessage,
		},
	})
}

// Mark all messages in chat as read
func (s *ChatService) markChatRead(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChatID string `json:"chat_id"`
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

	if req.ChatID == "" || req.UserID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Chat ID and User ID are required",
			Error:   "Missing required fields",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	readCount := 0

	// Mark all messages as read for user
	for i := range s.messages {
		if s.messages[i].ChatID == req.ChatID && 
		   s.messages[i].SenderID != req.UserID && 
		   !s.messages[i].IsRead {
			s.messages[i].IsRead = true
			s.messages[i].ReadAt = &now
			s.messages[i].UpdatedAt = now
			readCount++
		}
	}

	// Update unread counts
	s.updateUnreadCounts(req.ChatID)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Chat messages marked as read",
		Data: map[string]interface{}{
			"chat_id":    req.ChatID,
			"user_id":    req.UserID,
			"read_count": readCount,
		},
	})
}

// Get unread message count
func (s *ChatService) getUnreadCount(w http.ResponseWriter, r *http.Request) {
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

	totalUnread := 0
	for _, chat := range s.chats {
		for _, participant := range chat.Participants {
			if participant == userID {
				if count, exists := chat.UnreadCounts[participant]; exists {
					totalUnread += count
				}
				break
			}
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Unread count retrieved successfully",
		Data: map[string]interface{}{
			"user_id":      userID,
			"unread_count": totalUnread,
		},
	})
}

// Start chat service
func main() {
	service := NewChatService()

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/chats", corsMiddleware(service.getUserChats))
	http.Handle("/api/v1/chats/", corsMiddleware(service.getChat)) // For GET by ID
	http.Handle("/api/v1/chats/create", corsMiddleware(service.createChat))
	http.Handle("/api/v1/messages", corsMiddleware(service.getChatMessages))
	http.Handle("/api/v1/messages/send", corsMiddleware(service.sendMessage))
	http.Handle("/api/v1/messages/", corsMiddleware(service.markMessageRead)) // For mark as read
	http.Handle("/api/v1/messages/read-all", corsMiddleware(service.markChatRead))
	http.Handle("/api/v1/messages/unread", corsMiddleware(service.getUnreadCount))

	port := ":8090"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	fmt.Printf("🚀 CHAT SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)
	fmt.Printf("💬 User chats: http://localhost%s/api/v1/chats?user_id={id}\n", port)
	fmt.Printf("📝 Chat details: http://localhost%s/api/v1/chats/{id}\n", port)
	fmt.Printf("➕ Create chat: http://localhost%s/api/v1/chats/create\n", port)
	fmt.Printf("📨 Chat messages: http://localhost%s/api/v1/messages?chat_id={id}\n", port)
	fmt.Printf("💭 Send message: http://localhost%s/api/v1/messages/send\n", port)
	fmt.Printf("✅ Mark read: http://localhost%s/api/v1/messages/{id}\n", port)
	fmt.Printf("🔢 Unread count: http://localhost%s/api/v1/messages/unread?user_id={id}\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("💬 Total chats: %d\n", len(service.chats))
	fmt.Printf("📨 Total messages: %d\n", len(service.messages))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(port, nil))
}