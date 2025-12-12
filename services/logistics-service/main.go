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

// Shipment struct
type Shipment struct {
	ID                 string               `json:"id"`
	OrderID            string               `json:"order_id"`
	SellerID           string               `json:"seller_id"`
	BuyerID            string               `json:"buyer_id"`
	PickupAddress      Address              `json:"pickup_address"`
	DeliveryAddress    Address              `json:"delivery_address"`
	Package            Package              `json:"package"`
	Status             string               `json:"status"` // "pending", "picked_up", "in_transit", "out_for_delivery", "delivered", "cancelled", "returned"
	TrackingNumber     string               `json:"tracking_number"`
	Carrier            CarrierInfo          `json:"carrier"`
	TrackingHistory    []TrackingEvent      `json:"tracking_history"`
	EstimatedDelivery  *time.Time           `json:"estimated_delivery,omitempty"`
	ActualDelivery     *time.Time           `json:"actual_delivery,omitempty"`
	PickupTime        *time.Time           `json:"pickup_time,omitempty"`
	Cost              float64              `json:"cost"`
	InsuranceAmount    float64              `json:"insurance_amount"`
	SignatureRequired  bool                 `json:"signature_required"`
	SpecialInstructions string              `json:"special_instructions,omitempty"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}

// Package struct
type Package struct {
	Weight     float64 `json:"weight"`     // kg
	Dimensions PackageDimensions `json:"dimensions"`
	Value      float64 `json:"value"`       // package value for insurance
	Fragile    bool    `json:"fragile"`
	Description string  `json:"description"`
}

// PackageDimensions struct
type PackageDimensions struct {
	Length float64 `json:"length"` // cm
	Width  float64 `json:"width"`  // cm
	Height float64 `json:"height"` // cm
}

// Address struct
type Address struct {
	Name      string `json:"name"`
	Street    string `json:"street"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`
	Country   string `json:"country"`
	Phone     string `json:"phone"`
	Email     string `json:"email,omitempty"`
}

// CarrierInfo struct
type CarrierInfo struct {
	Name    string `json:"name"`
	Service string `json:"service"` // "standard", "express", "overnight", "international"
	Type    string `json:"type"`    // "ground", "air", "sea"
}

// TrackingEvent struct
type TrackingEvent struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Timestamp   time.Time `json:"timestamp"`
}

// Request structs
type CreateShipmentRequest struct {
	OrderID            string        `json:"order_id"`
	PickupAddress      Address       `json:"pickup_address"`
	DeliveryAddress    Address       `json:"delivery_address"`
	Package            Package       `json:"package"`
	Carrier            CarrierInfo   `json:"carrier"`
	Cost               float64       `json:"cost"`
	InsuranceAmount    float64       `json:"insurance_amount"`
	SignatureRequired  bool          `json:"signature_required"`
	SpecialInstructions string        `json:"special_instructions,omitempty"`
}

type UpdateShipmentStatusRequest struct {
	Status    string  `json:"status"`
	Location  string  `json:"location,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Logistics service with in-memory storage
type LogisticsService struct {
	shipments []Shipment
	mu        sync.RWMutex
}

// New logistics service
func NewLogisticsService() *LogisticsService {
	return &LogisticsService{
		shipments: []Shipment{
			{
				ID:               "shipment-1",
				OrderID:          "order-1",
				SellerID:         "seller-1",
				BuyerID:          "user-1",
				PickupAddress: Address{
					Name:    "Seller Store",
					Street:  "123 Seller St",
					City:    "New York",
					State:   "NY",
					ZipCode: "10001",
					Country: "USA",
					Phone:   "+1-555-0123",
					Email:   "seller@example.com",
				},
				DeliveryAddress: Address{
					Name:    "Buyer Home",
					Street:  "456 Buyer Ave",
					City:    "Los Angeles",
					State:   "CA",
					ZipCode: "90210",
					Country: "USA",
					Phone:   "+1-555-0456",
					Email:   "buyer@example.com",
				},
				Package: Package{
					Weight: 2.5,
					Dimensions: PackageDimensions{
						Length: 30.0,
						Width:  20.0,
						Height: 15.0,
					},
					Value:      450.00,
					Fragile:    true,
					Description: "Vintage Camera",
				},
				Status:           "shipped",
				TrackingNumber:  "TRACK123456",
				Carrier: CarrierInfo{
					Name:    "FedEx",
					Service: "Express",
					Type:    "air",
				},
				TrackingHistory: []TrackingEvent{
					{
						ID:          "event-1",
						Status:      "created",
						Description: "Shipment created",
						Location:    "New York, NY",
						Timestamp:   time.Now().Add(-24 * time.Hour),
					},
					{
						ID:          "event-2",
						Status:      "picked_up",
						Description: "Package picked up from seller",
						Location:    "New York, NY",
						Timestamp:   time.Now().Add(-12 * time.Hour),
					},
					{
						ID:          "event-3",
						Status:      "in_transit",
						Description: "Package in transit",
						Location:    "Memphis, TN",
						Timestamp:   time.Now().Add(-6 * time.Hour),
					},
				},
				EstimatedDelivery:  &[]time.Time{time.Now().Add(2 * 24 * time.Hour)}[0],
				PickupTime:        &[]time.Time{time.Now().Add(-12 * time.Hour)}[0],
				Cost:              15.99,
				InsuranceAmount:    4.50,
				SignatureRequired:  true,
				SpecialInstructions: "Handle with care - fragile item",
				CreatedAt:          time.Now().Add(-24 * time.Hour),
				UpdatedAt:          time.Now().Add(-6 * time.Hour),
			},
			{
				ID:               "shipment-2",
				OrderID:          "order-2",
				SellerID:         "seller-2",
				BuyerID:          "user-2",
				PickupAddress: Address{
					Name:    "Seller Store",
					Street:  "789 Seller Blvd",
					City:    "Miami",
					State:   "FL",
					ZipCode: "33101",
					Country: "USA",
					Phone:   "+1-555-0789",
					Email:   "seller2@example.com",
				},
				DeliveryAddress: Address{
					Name:    "Buyer Home",
					Street:  "321 Buyer St",
					City:    "Chicago",
					State:   "IL",
					ZipCode: "60601",
					Country: "USA",
					Phone:   "+1-555-0321",
					Email:   "buyer2@example.com",
				},
				Package: Package{
					Weight: 1.8,
					Dimensions: PackageDimensions{
						Length: 35.0,
						Width:  25.0,
						Height: 10.0,
					},
					Value:      450.00,
					Fragile:    false,
					Description: "Designer Handbag",
				},
				Status:           "delivered",
				TrackingNumber:  "TRACK789012",
				Carrier: CarrierInfo{
					Name:    "UPS",
					Service: "Ground",
					Type:    "ground",
				},
				TrackingHistory: []TrackingEvent{
					{
						ID:          "event-4",
						Status:      "created",
						Description: "Shipment created",
						Location:    "Miami, FL",
						Timestamp:   time.Now().Add(-48 * time.Hour),
					},
					{
						ID:          "event-5",
						Status:      "delivered",
						Description: "Package delivered",
						Location:    "Chicago, IL",
						Timestamp:   time.Now().Add(-1 * time.Hour),
					},
				},
				EstimatedDelivery:  &[]time.Time{time.Now().Add(-12 * time.Hour)}[0],
				ActualDelivery:     &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
				PickupTime:        &[]time.Time{time.Now().Add(-46 * time.Hour)}[0],
				Cost:              12.50,
				InsuranceAmount:    9.00,
				SignatureRequired:  false,
				CreatedAt:          time.Now().Add(-48 * time.Hour),
				UpdatedAt:          time.Now().Add(-1 * time.Hour),
			},
		},
	}
}

// Generate tracking number
func generateTrackingNumber() string {
	return fmt.Sprintf("TRACK%d", time.Now().UnixNano())
}

// Calculate shipping cost
func calculateShippingCost(pkg Package, carrier CarrierInfo, distance float64) float64 {
	baseCost := 5.00
	weightCost := pkg.Weight * 2.50
	dimensionCost := (pkg.Dimensions.Length + pkg.Dimensions.Width + pkg.Dimensions.Height) * 0.10
	
	cost := baseCost + weightCost + dimensionCost + (distance * 0.05)
	
	// Apply carrier multiplier
	switch carrier.Name {
	case "FedEx":
		cost *= 1.2
	case "UPS":
		cost *= 1.1
	case "DHL":
		cost *= 1.3
	}
	
	// Apply service multiplier
	switch carrier.Service {
	case "Express":
		cost *= 1.5
	case "Overnight":
		cost *= 2.0
	case "International":
		cost *= 3.0
	}
	
	return cost
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
func paginateShipments(items []Shipment, page, perPage int) ([]Shipment, map[string]interface{}) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		return []Shipment{}, map[string]interface{}{
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
func (s *LogisticsService) health(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeShipments := 0
	deliveredShipments := 0
	totalCost := 0.0

	for _, shipment := range s.shipments {
		switch shipment.Status {
		case "pending", "picked_up", "in_transit", "out_for_delivery":
			activeShipments++
		case "delivered":
			deliveredShipments++
		}
		totalCost += shipment.Cost
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Logistics service is healthy and working!",
		Data: map[string]interface{}{
			"service":            "logistics-service",
			"version":            "v2.0-working",
			"status":             "healthy",
			"timestamp":          time.Now(),
			"total_shipments":    len(s.shipments),
			"active_shipments":   activeShipments,
			"delivered_shipments": deliveredShipments,
			"total_cost":         totalCost,
		},
	})
}

// Get all shipments
func (s *LogisticsService) getAllShipments(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePagination(r)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply user filter if provided
	var filteredShipments []Shipment
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		for _, shipment := range s.shipments {
			if shipment.BuyerID == userID {
				filteredShipments = append(filteredShipments, shipment)
			}
		}
	} else if sellerID := r.URL.Query().Get("seller_id"); sellerID != "" {
		for _, shipment := range s.shipments {
			if shipment.SellerID == sellerID {
				filteredShipments = append(filteredShipments, shipment)
			}
		}
	} else if orderID := r.URL.Query().Get("order_id"); orderID != "" {
		for _, shipment := range s.shipments {
			if shipment.OrderID == orderID {
				filteredShipments = append(filteredShipments, shipment)
			}
		}
	} else {
		filteredShipments = s.shipments
	}

	// Apply status filter if provided
	if status := strings.ToLower(r.URL.Query().Get("status")); status != "" {
		var statusFilteredShipments []Shipment
		for _, shipment := range filteredShipments {
			if strings.Contains(strings.ToLower(shipment.Status), status) {
				statusFilteredShipments = append(statusFilteredShipments, shipment)
			}
		}
		filteredShipments = statusFilteredShipments
	}

	// Apply carrier filter if provided
	if carrier := strings.ToLower(r.URL.Query().Get("carrier")); carrier != "" {
		var carrierFilteredShipments []Shipment
		for _, shipment := range filteredShipments {
			if strings.Contains(strings.ToLower(shipment.Carrier.Name), carrier) {
				carrierFilteredShipments = append(carrierFilteredShipments, shipment)
			}
		}
		filteredShipments = carrierFilteredShipments
	}

	// Sort shipments by creation date (newest first)
	for i := 0; i < len(filteredShipments)-1; i++ {
		for j := i + 1; j < len(filteredShipments); j++ {
			if filteredShipments[i].CreatedAt.Before(filteredShipments[j].CreatedAt) {
				filteredShipments[i], filteredShipments[j] = filteredShipments[j], filteredShipments[i]
			}
		}
	}

	// Paginate results
	paginatedShipments, pagination := paginateShipments(filteredShipments, page, perPage)

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Shipments retrieved successfully",
		Data: map[string]interface{}{
			"shipments":   paginatedShipments,
			"pagination": pagination,
		},
	})
}

// Get shipment by ID
func (s *LogisticsService) getShipment(w http.ResponseWriter, r *http.Request) {
	shipmentID := strings.TrimPrefix(r.URL.Path, "/api/v1/shipments/")
	if shipmentID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Shipment ID is required",
			Error:   "No shipment ID provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, shipment := range s.shipments {
		if shipment.ID == shipmentID {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Shipment retrieved successfully",
				Data: map[string]interface{}{
					"shipment": shipment,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Shipment not found",
		Error:   "Shipment with ID " + shipmentID + " does not exist",
	})
}

// Create shipment
func (s *LogisticsService) createShipment(w http.ResponseWriter, r *http.Request) {
	var req CreateShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid shipment data",
			Error:   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.OrderID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Order ID is required",
			Error:   "Missing required fields",
		})
		return
	}

	if req.PickupAddress.Street == "" || req.DeliveryAddress.Street == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Pickup and delivery addresses are required",
			Error:   "Incomplete address information",
		})
		return
	}

	if req.Package.Weight <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Package weight must be greater than 0",
			Error:   "Invalid package weight",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create new shipment
	newShipment := Shipment{
		ID:                 fmt.Sprintf("shipment-%d", time.Now().UnixNano()),
		OrderID:            req.OrderID,
		SellerID:           "seller-1", // In real app, get from order
		BuyerID:            "user-1",    // In real app, get from order
		PickupAddress:      req.PickupAddress,
		DeliveryAddress:    req.DeliveryAddress,
		Package:            req.Package,
		Status:             "pending",
		TrackingNumber:     generateTrackingNumber(),
		Carrier:            req.Carrier,
		TrackingHistory: []TrackingEvent{
			{
				ID:          fmt.Sprintf("event-%d", time.Now().UnixNano()),
				Status:      "created",
				Description: "Shipment created",
				Location:    req.PickupAddress.City + ", " + req.PickupAddress.State,
				Timestamp:   time.Now(),
			},
		},
		EstimatedDelivery:  &[]time.Time{time.Now().Add(5 * 24 * time.Hour)}[0],
		Cost:              req.Cost,
		InsuranceAmount:    req.InsuranceAmount,
		SignatureRequired:  req.SignatureRequired,
		SpecialInstructions: req.SpecialInstructions,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	s.shipments = append(s.shipments, newShipment)

	log.Printf("Shipment created: %s for order %s", newShipment.ID, newShipment.OrderID)

	writeJSONResponse(w, http.StatusCreated, Response{
		Success: true,
		Message: "Shipment created successfully",
		Data: map[string]interface{}{
			"shipment": newShipment,
		},
	})
}

// Update shipment status
func (s *LogisticsService) updateShipmentStatus(w http.ResponseWriter, r *http.Request) {
	shipmentID := strings.TrimPrefix(r.URL.Path, "/api/v1/shipments/")
	if shipmentID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Shipment ID is required",
			Error:   "No shipment ID provided",
		})
		return
	}

	var req UpdateShipmentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid status update data",
			Error:   err.Error(),
		})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":           true,
		"picked_up":         true,
		"in_transit":        true,
		"out_for_delivery":  true,
		"delivered":         true,
		"cancelled":         true,
		"returned":          true,
	}

	if !validStatuses[req.Status] {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid status",
			Error:   "Status must be one of: pending, picked_up, in_transit, out_for_delivery, delivered, cancelled, returned",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and update shipment
	for i := range s.shipments {
		if s.shipments[i].ID == shipmentID {
			oldStatus := s.shipments[i].Status
			s.shipments[i].Status = req.Status
			s.shipments[i].UpdatedAt = time.Now()

			// Add tracking event
			description := fmt.Sprintf("Status changed from %s to %s", oldStatus, req.Status)
			if req.Notes != "" {
				description = req.Notes
			}

			newEvent := TrackingEvent{
				ID:          fmt.Sprintf("event-%d", time.Now().UnixNano()),
				Status:      req.Status,
				Description: description,
				Location:    req.Location,
				Timestamp:   time.Now(),
			}

			s.shipments[i].TrackingHistory = append(s.shipments[i].TrackingHistory, newEvent)

			// Update delivery times
			if req.Status == "picked_up" {
				now := time.Now()
				s.shipments[i].PickupTime = &now
			} else if req.Status == "delivered" {
				now := time.Now()
				s.shipments[i].ActualDelivery = &now
			}

			log.Printf("Shipment status updated: %s -> %s", shipmentID, req.Status)

			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Shipment status updated successfully",
				Data: map[string]interface{}{
					"shipment": s.shipments[i],
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Shipment not found",
		Error:   "Shipment with ID " + shipmentID + " does not exist",
	})
}

// Track shipment
func (s *LogisticsService) trackShipment(w http.ResponseWriter, r *http.Request) {
	shipmentID := r.URL.Query().Get("shipment_id")
	if shipmentID == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Shipment ID is required",
			Error:   "No shipment ID provided",
		})
		return
	}

	trackingNumber := r.URL.Query().Get("tracking_number")
	if trackingNumber == "" {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Tracking number is required",
			Error:   "No tracking number provided",
		})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find shipment
	for _, shipment := range s.shipments {
		if shipment.ID == shipmentID || shipment.TrackingNumber == trackingNumber {
			writeJSONResponse(w, http.StatusOK, Response{
				Success: true,
				Message: "Shipment tracking retrieved successfully",
				Data: map[string]interface{}{
					"shipment":          shipment,
					"tracking_history":  shipment.TrackingHistory,
					"current_status":    shipment.Status,
					"estimated_delivery": shipment.EstimatedDelivery,
				},
			})
			return
		}
	}

	writeJSONResponse(w, http.StatusNotFound, Response{
		Success: false,
		Message: "Shipment not found",
		Error:   "Shipment with ID " + shipmentID + " or tracking number " + trackingNumber + " does not exist",
	})
}

// Calculate shipping cost
func (s *LogisticsService) calculateShippingCost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Package   Package     `json:"package"`
		Carrier   CarrierInfo `json:"carrier"`
		Distance  float64     `json:"distance"` // km
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid cost calculation data",
			Error:   err.Error(),
		})
		return
	}

	// Validate package
	if req.Package.Weight <= 0 {
		writeJSONResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "Package weight must be greater than 0",
			Error:   "Invalid package weight",
		})
		return
	}

	if req.Distance <= 0 {
		req.Distance = 1000 // Default distance
	}

	// Calculate cost
	cost := calculateShippingCost(req.Package, req.Carrier, req.Distance)
	insuranceCost := req.Package.Value * 0.01 // 1% of package value
	totalCost := cost + insuranceCost

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Shipping cost calculated successfully",
		Data: map[string]interface{}{
			"base_cost":       cost,
			"insurance_cost":  insuranceCost,
			"total_cost":      totalCost,
			"package":        req.Package,
			"carrier":        req.Carrier,
			"distance":       req.Distance,
		},
	})
}

// Get shipment statistics
func (s *LogisticsService) getShipmentStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_shipments":     len(s.shipments),
		"pending_shipments":   0,
		"picked_up_shipments": 0,
		"in_transit_shipments": 0,
		"out_for_delivery":    0,
		"delivered_shipments": 0,
		"cancelled_shipments": 0,
		"returned_shipments":  0,
		"total_cost":          0.0,
		"average_cost":        0.0,
		"total_weight":        0.0,
		"average_weight":      0.0,
	}

	totalCost := 0.0
	totalWeight := 0.0

	for _, shipment := range s.shipments {
		switch shipment.Status {
		case "pending":
			stats["pending_shipments"] = stats["pending_shipments"].(int) + 1
		case "picked_up":
			stats["picked_up_shipments"] = stats["picked_up_shipments"].(int) + 1
		case "in_transit":
			stats["in_transit_shipments"] = stats["in_transit_shipments"].(int) + 1
		case "out_for_delivery":
			stats["out_for_delivery"] = stats["out_for_delivery"].(int) + 1
		case "delivered":
			stats["delivered_shipments"] = stats["delivered_shipments"].(int) + 1
		case "cancelled":
			stats["cancelled_shipments"] = stats["cancelled_shipments"].(int) + 1
		case "returned":
			stats["returned_shipments"] = stats["returned_shipments"].(int) + 1
		}

		totalCost += shipment.Cost
		totalWeight += shipment.Package.Weight
	}

	stats["total_cost"] = totalCost
	stats["total_weight"] = totalWeight

	if len(s.shipments) > 0 {
		stats["average_cost"] = totalCost / float64(len(s.shipments))
		stats["average_weight"] = totalWeight / float64(len(s.shipments))
	}

	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Message: "Shipment statistics retrieved successfully",
		Data: map[string]interface{}{
			"stats": stats,
		},
	})
}

// Start logistics service
func main() {
	service := NewLogisticsService()

	// Setup routes with CORS
	http.Handle("/health", corsMiddleware(service.health))
	http.Handle("/api/v1/shipments", corsMiddleware(service.getAllShipments))
	http.Handle("/api/v1/shipments/", corsMiddleware(service.getShipment)) // For GET by ID and update status
	http.Handle("/api/v1/shipments/create", corsMiddleware(service.createShipment))
	http.Handle("/api/v1/shipments/track", corsMiddleware(service.trackShipment))
	http.Handle("/api/v1/shipments/calculate-cost", corsMiddleware(service.calculateShippingCost))
	http.Handle("/api/v1/shipments/stats", corsMiddleware(service.getShipmentStats))

	port := ":8091"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	fmt.Printf("🚀 LOGISTICS SERVICE - WORKING VERSION\n")
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)
	fmt.Printf("📦 Shipments list: http://localhost%s/api/v1/shipments\n", port)
	fmt.Printf("📝 Shipment details: http://localhost%s/api/v1/shipments/{id}\n", port)
	fmt.Printf("➕ Create shipment: http://localhost%s/api/v1/shipments/create\n", port)
	fmt.Printf("📍 Track shipment: http://localhost%s/api/v1/shipments/track?shipment_id={id}&tracking_number={track}\n", port)
	fmt.Printf("💰 Calculate cost: http://localhost%s/api/v1/shipments/calculate-cost\n", port)
	fmt.Printf("📈 Shipment stats: http://localhost%s/api/v1/shipments/stats\n", port)
	fmt.Printf("⏰ Started at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("📦 Total shipments: %d\n", len(service.shipments))
	fmt.Printf("🎯 Status: Ready to serve!\n")

	log.Fatal(http.ListenAndServe(port, nil))
}