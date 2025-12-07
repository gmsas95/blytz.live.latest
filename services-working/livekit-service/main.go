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

// Room struct
type Room struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Type           string    `json:"type"` // "auction", "product_demo", "chat", "meeting"
	HostID         string    `json:"host_id"`
	Participants   []Participant `json:"participants"`
	Status         string    `json:"status"` // "active", "ended", "waiting"
	Config         RoomConfig `json:"config"`
	Recording      bool      `json:"recording"`
	Recordings     []Recording `json:"recordings,omitempty"`
	StartedAt      time.Time `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Participant struct
type Participant struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Identity    string    `json:"identity"`
	Name        string    `json:"name"`
	Permission  string    `json:"permission"` // "host", "moderator", "speaker", "viewer"
	Video       bool      `json:"video"`
	Audio       bool      `json:"audio"`
	ScreenShare bool      `json:"screen_share"`
	JoinedAt    time.Time `json:"joined_at"`
	LeftAt      *time.Time `json:"left_at,omitempty"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
}

// RoomConfig struct
type RoomConfig struct {
	MaxParticipants   int           `json:"max_participants"`
	AllowScreenShare  bool          `json:"allow_screen_share"`
	EnableRecording   bool          `json:"enable_recording"`
	AutoStartRecording bool         `json:"auto_start_recording"`
	ChatEnabled       bool          `json:"chat_enabled"`
	Quality           string        `json:"quality"` // "low", "medium", "high", "ultra"
	Layout            string        `json:"layout"`  // "grid", "speaker", "gallery"
	Password          string        `json:"password,omitempty"`
	WaitingRoom       bool          `json:"waiting_room"`
}

// Recording struct
type Recording struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"room_id"`
	Type        string    `json:"type"` // "audio_video", "audio_only", "screen_only"
	URL         string    `json:"url"`
	Size        int64     `json:"size"` // bytes
	Duration    int64     `json:"duration"` // seconds
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at"`
	Status      string    `json:"status"` // "processing", "completed", "failed"
	StoragePath string    `json:"storage_path"`
}

// Request structs
type CreateRoomRequest struct {
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Type             string     `json:"type"`
	HostID           string     `json:"host_id"`
	Config           RoomConfig `json:"config"`
	MaxParticipants  int        `json:"max_participants"`
	EnableRecording  bool       `json:"enable_recording"`
}

type JoinRoomRequest struct {
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
	Password   string `json:"password,omitempty"`
}

type UpdateParticipantRequest struct {
	Video       bool `json:"video"`
	Audio       bool `json:"audio"`
	ScreenShare bool `json:"screen_share"`
	Permission  string `json:"permission,omitempty"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LiveKit service with in-memory storage
type LiveKitService struct {
	rooms        []Room
	participants  []Participant
	mu            sync.RWMutex
}

// New LiveKit service
func NewLiveKitService() *LiveKitService {
	return &LiveKitService{
		rooms: []Room{
			{
				ID:          "room-1",
				Name:        "Vintage Camera Auction",
				Description: "Live auction for vintage camera",
				Type:        "auction",
				HostID:      "seller-1",
				Participants: []Participant{
					{
						ID:         "participant-1",
						UserID:     "seller-1",
						Identity:   "seller-vintage-camera",
						Name:       "John Seller",
						Permission: "host",
						Video:      true,
						Audio:      true,
						ScreenShare: false,
						JoinedAt:   time.Now().Add(-15 * time.Minute),
						IP:         "192.168.1.100",
						UserAgent:  "Mozilla/5.0 (Chrome)",
					},
					{
						ID:         "participant-2",
						UserID:     "user-1",
						Identity:   "buyer-interested",
						Name:       "Jane Buyer",
						Permission: "speaker",
						Video:      true,
						Audio:      false,
						ScreenShare: false,
						JoinedAt:   time.Now().Add(-10 * time.Minute),
						IP:         "192.168.1.101",
						UserAgent:  "Mozilla/5.0 (Firefox)",
					},
				},
				Status: "active",
				Config: RoomConfig{
					MaxParticipants:    50,
					AllowScreenShare:   true,
					EnableRecording:    true,
					AutoStartRecording: true,
					ChatEnabled:        true,
					Quality:            "high",
					Layout:             "speaker",
					WaitingRoom:        false,
				},
				Recording: true,
				Recordings: []Recording{
					{
						ID:          "recording-1",
						RoomID:      "room-1",
						Type:        "audio_video",
						URL:         "https://storage.example.com/recordings/rec1.mp4",
						Size:        52428800, // 50MB
						Duration:    900, // 15 minutes
						StartedAt:   time.Now().Add(-15 * time.Minute),
						EndedAt:     time.Now(),
						Status:      "completed",
						StoragePath: "/recordings/2024/12/07/recording-1.mp4",
					},
				},
				StartedAt: time.Now().Add(-15 * time.Minute),
				CreatedAt: time.Now().Add(-20 * time.Minute),
				UpdatedAt: time.Now(),
			},
			{
				ID:          "room-2",
				Name:        "Designer Handbag Demo",
				Description: "Product demonstration for designer handbag",
				Type:        "product_demo",
				HostID:      "seller-2",
				Participants: []Participant{
					{
						ID:         "participant-3",
						UserID:     "seller-2",
						Identity:   "seller-handbag",
						Name:       "Maria Designer",
						Permission: "host",
						Video:      true,
						Audio:      true,
						ScreenShare: false,
						JoinedAt:   time.Now().Add(-5 * time.Minute),
						IP:         "192.168.1.102",
						UserAgent:  "Mozilla/5.0 (Safari)",
					},
				},
				Status: "active",
				Config: RoomConfig{
					MaxParticipants:    25,
					AllowScreenShare:   false,
					EnableRecording:    true,
					AutoStartRecording: false,
					ChatEnabled:        true,
					Quality:            "medium",
					Layout:             "gallery",
					WaitingRoom:        true,
				},
				Recording: false,
				StartedAt: time.Now().Add(-5 * time.Minute),
				CreatedAt: time.Now().Add(-10 * time.Minute),
				UpdatedAt: time.Now(),
			},
		},
	}
}

// Generate room token (simplified)
func generateRoomToken(roomID, userID, permission string) string {
	return fmt.Sprintf("livekit_token_%s_%s_%s_%d", roomID, userID, permission, time.Now().UnixNano())
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
func paginateRooms(items []Room, page, perPage int) ([]Room, map[string]interface{}) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Room{}, map[string]interface{}{
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
func (s *LiveKitService) health(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeRooms := 0
	totalParticipants := 0
	recordingRooms := 0

	for _, room := range s.rooms {
		if room.Status == "active" {
			activeRooms++
		}
		totalParticipants += len(room.Participants)
		if room.Recording {
			recordingRooms++
		}
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "LiveKit service is healthy and working!",
		Data: map[string]interface{}{
			"service":           "livekit-service",
			"version":           "v2.0-working",
			"status":            "healthy",
			"timestamp":         time.Now(),
			"total_rooms":       len(s.rooms),
			"active_rooms":      activeRooms,
			"total_participants": totalParticipants,
			"recording_rooms":   recordingRooms,
		},
	})
}

// Get all rooms
func (s *LiveKitService) getAllRooms(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply host filter if provided
	var filteredRooms []Room
	if hostID := r.URL.Query().Get("host_id"); hostID != "" {
		for _, room := range s.rooms {
			if room.HostID == hostID {
				filteredRooms = append(filteredRooms, room)
			}
		}
	} else {
		filteredRooms = s.rooms
	}

	// Apply status filter if provided
	if status := strings.ToLower(r.URL.Query().Get("status")); status != "" {
		var statusFilteredRooms []Room
		for _, room := range filteredRooms {
			if strings.Contains(strings.ToLower(room.Status), status) {
				statusFilteredRooms = append(statusFilteredRooms, room)
			}
		}
		filteredRooms = statusFilteredRooms
	}

	// Apply type filter if provided
	if roomType := strings.ToLower(r.URL.Query().Get("type")); roomType != "" {
		var typeFilteredRooms []Room
		for _, room := range filteredRooms {
			if strings.Contains(strings.ToLower(room.Type), roomType) {
				typeFilteredRooms = append(typeFilteredRooms, room)
			}
		}
		filteredRooms = typeFilteredRooms
	}

	// Sort rooms by creation date (newest first)
	for i := 0; i < len(filteredRooms)-1; i++ {
		for j := i + 1; j < len(filteredRooms); j++ {
			if filteredRooms[i].CreatedAt.Before(filteredRooms[j].CreatedAt) {
				filteredRooms[i], filteredRooms[j] = filteredRooms[j], filteredRooms[i]
			}
		}
	}

	// Paginate results
	paginatedRooms, pagination := paginateRooms(filteredRooms, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Rooms retrieved successfully",
		Data: map[string]interface{}{
			"rooms":     paginatedRooms,
			"pagination": pagination,
		},
	})
}

// Get room by ID
func (s *LiveKitService) getRoom(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/api/v1/rooms/")
	if roomID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Room ID is required",
			Error:   "No room ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, room := range s.rooms {
		if room.ID == roomID {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Room retrieved successfully",
				Data: map[string]interface{}{
					"room": room,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Room not found",
		Error:   "Room with ID " + roomID + " does not exist",
	})
}

// Create room
func (s *LiveKitService) createRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid room data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Name == "" || req.HostID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Room name and host ID are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.Type == "" {
		req.Type = "meeting"
	}

	// Validate room type
	validTypes := map[string]bool{
		"auction":      true,
		"product_demo": true,
		"chat":         true,
		"meeting":      true,
	}
	if !validTypes[req.Type] {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid room type",
			Error:   "Room type must be one of: auction, product_demo, chat, meeting",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create new room
	newRoom := Room{
		ID:          fmt.Sprintf("room-%d", time.Now().UnixNano()),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		HostID:      req.HostID,
		Participants: []Participant{},
		Status:      "waiting",
		Config: RoomConfig{
			MaxParticipants:     10,
			AllowScreenShare:    true,
			EnableRecording:     req.EnableRecording,
			AutoStartRecording:  false,
			ChatEnabled:         true,
			Quality:             "medium",
			Layout:              "grid",
			WaitingRoom:         false,
		},
		Recording:  false,
		Recordings: []Recording{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Apply config overrides if provided
	if req.Config.MaxParticipants > 0 {
		newRoom.Config.MaxParticipants = req.Config.MaxParticipants
	}
	if req.Config.Quality != "" {
		newRoom.Config.Quality = req.Config.Quality
	}
	if req.Config.Layout != "" {
		newRoom.Config.Layout = req.Config.Layout
	}

	s.rooms = append(s.rooms, newRoom)

	log.Printf("Room created: %s (%s) by host %s", newRoom.ID, newRoom.Name, newRoom.HostID)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Room created successfully",
		Data: map[string]interface{}{
			"room": newRoom,
		},
	})
}

// Join room
func (s *LiveKitService) joinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/api/v1/rooms/")
	if roomID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Room ID is required",
			Error:   "No room ID provided",
		})
		return
	}

	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid join request",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Name == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID and name are required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.Permission == "" {
		req.Permission = "viewer"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find room
	var roomIndex = -1
	for i := range s.rooms {
		if s.rooms[i].ID == roomID {
			roomIndex = i
			break
		}
	}

	if roomIndex == -1 {
		writeJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Message: "Room not found",
			Error:   "Room with ID " + roomID + " does not exist",
		})
		return
	}

	room := &s.rooms[roomIndex]

	// Check room status
	if room.Status == "ended" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot join ended room",
			Error:   "Room has ended",
		})
		return
	}

	// Check participant limit
	if len(room.Participants) >= room.Config.MaxParticipants {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Room is full",
			Error:   "Maximum participants reached",
		})
		return
	}

	// Create new participant
	newParticipant := Participant{
		ID:          fmt.Sprintf("participant-%d", time.Now().UnixNano()),
		UserID:      req.UserID,
		Identity:    fmt.Sprintf("%s_%s", req.Name, req.UserID),
		Name:        req.Name,
		Permission:  req.Permission,
		Video:       false,
		Audio:       false,
		ScreenShare: false,
		JoinedAt:    time.Now(),
		IP:          "192.168.1.xxx", // In real app, get from request
		UserAgent:   "Mozilla/5.0",     // In real app, get from request
	}

	// Add participant to room
	room.Participants = append(room.Participants, newParticipant)
	room.UpdatedAt = time.Now()

	// Update room status
	if room.Status == "waiting" {
		room.Status = "active"
		room.StartedAt = time.Now()
	}

	// Start recording if enabled and auto-start is true
	if room.Config.AutoStartRecording && room.Config.EnableRecording && !room.Recording {
		room.Recording = true
	}

	// Generate token
	token := generateRoomToken(roomID, req.UserID, req.Permission)

	log.Printf("User %s joined room %s", req.Name, roomID)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Joined room successfully",
		Data: map[string]interface{}{
			"room":        room,
			"participant": newParticipant,
			"token":       token,
		},
	})
}

// Leave room
func (s *LiveKitService) leaveRoom(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/api/v1/rooms/")
	if roomID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Room ID is required",
			Error:   "No room ID provided",
		})
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid leave request",
			Error:   err.Error(),
		})
		return
	}

	if req.UserID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User ID is required",
			Error:   "Missing user ID",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find room
	var roomIndex = -1
	for i := range s.rooms {
		if s.rooms[i].ID == roomID {
			roomIndex = i
			break
		}
	}

	if roomIndex == -1 {
		writeJSONResponse(w, http.StatusNotFound, Response{
			Success: false,
			Message: "Room not found",
			Error:   "Room with ID " + roomID + " does not exist",
		})
		return
	}

	room := &s.rooms[roomIndex]

	// Find and remove participant
	var removedParticipant *Participant
	var newParticipants []Participant
	for i := range room.Participants {
		if room.Participants[i].UserID == req.UserID {
			removedParticipant = &room.Participants[i]
			now := time.Now()
			room.Participants[i].LeftAt = &now
		}
		if room.Participants[i].UserID != req.UserID {
			newParticipants = append(newParticipants, room.Participants[i])
		}
	}

	if removedParticipant == nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "User not in room",
			Error:   "User with ID " + req.UserID + " is not in room " + roomID,
		})
		return
	}

	room.Participants = newParticipants
	room.UpdatedAt = time.Now()

	// Update room status if empty
	if len(room.Participants) == 0 && room.Status == "active" {
		room.Status = "ended"
		now := time.Now()
		room.EndedAt = &now
		room.Recording = false
	}

	log.Printf("User %s left room %s", removedParticipant.Name, roomID)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Left room successfully",
		Data: map[string]interface{}{
			"participant": removedParticipant,
			"room":        room,
		},
	})
}

// End room
func (s *LiveKitService) endRoom(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/api/v1/rooms/")
	if roomID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Room ID is required",
			Error:   "No room ID provided",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.rooms {
		if s.rooms[i].ID == roomID {
			if s.rooms[i].Status == "ended" {
				writeJSONResponse(w, http.StatusBadRequest, Response{
					Success: false,
					Message: "Room is already ended",
					Error:   "Cannot end an already ended room",
				})
				return
			}

			s.rooms[i].Status = "ended"
			now := time.Now()
			s.rooms[i].EndedAt = &now
			s.rooms[i].Recording = false
			s.rooms[i].UpdatedAt = now

			// Mark all participants as left
			for j := range s.rooms[i].Participants {
				s.rooms[i].Participants[j].LeftAt = &now
			}

			log.Printf("Room ended: %s", roomID)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Room ended successfully",
				Data: map[string]interface{}{
					"room": s.rooms[i],
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Room not found",
		Error:   "Room with ID " + roomID + " does not exist",
	})
}

// Get room statistics
func (s *LiveKitService) getRoomStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_rooms":        len(s.rooms),
		"active_rooms":       0,
		"ended_rooms":        0,
		"waiting_rooms":      0,
		"total_participants": 0,
		"active_participants": 0,
		"recording_rooms":    0,
		"room_types":         make(map[string]int),
	}

	totalParticipants := 0

	for _, room := range s.rooms {
		switch room.Status {
		case "active":
			stats["active_rooms"] = stats["active_rooms"].(int) + 1
		case "ended":
			stats["ended_rooms"] = stats["ended_rooms"].(int) + 1
		case "waiting":
			stats["waiting_rooms"] = stats["waiting_rooms"].(int) + 1
		}

		participantCount := len(room.Participants)
		totalParticipants += participantCount

		if room.Status == "active" {
			stats["active_participants"] = stats["active_participants"].(int) + participantCount
		}

		if room.Recording {
			stats["recording_rooms"] = stats["recording_rooms"].(int) + 1
		}

		// Count room types
		if count, exists := stats["room_types"].(map[string]int)[room.Type]; exists {
			stats["room_types"].(map[string]int)[room.Type] = count + 1
		} else {
			stats["room_types"].(map[string]int)[room.Type] = 1
		}
	}

	stats["total_participants"] = totalParticipants

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Room statistics retrieved successfully",
		Data: map[string]interface{}{
			"stats": stats,
		},
	})
}

// Start LiveKit service
func main() {
	service := NewLiveKitService()

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/rooms", corsMiddleware(service.getAllRooms))
	http.Handle("/api/v1/rooms/", corsMiddleware(service.getRoom)) // For GET by ID
	http.Handle("/api/v1/rooms/create", corsMiddleware(service.createRoom))
	http.Handle("/api/v1/rooms/join", corsMiddleware(service.joinRoom))
	http.Handle("/api/v1/rooms/leave", corsMiddleware(service.leaveRoom))
	http.Handle("/api/v1/rooms/stats", corsMiddleware(service.getRoomStats))

	port := ":8093"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	fmt.Printf("🚀 LIVEKIT SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)
	fmt.Printf("🏠 Rooms list: http://localhost%s/api/v1/rooms\n", port)
	fmt.Printf("📝 Room details: http://localhost%s/api/v1/rooms/{id}\n", port)
	fmt.Printf("➕ Create room: http://localhost%s/api/v1/rooms/create\n", port)
	fmt.Printf("👋 Join room: http://localhost%s/api/v1/rooms/join\n", port)
	fmt.Printf("👋 Leave room: http://localhost%s/api/v1/rooms/leave\n", port)
	fmt.Printf("📈 Room stats: http://localhost%s/api/v1/rooms/stats\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("🏠 Total rooms: %d\n", len(service.rooms))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(port, nil))
}