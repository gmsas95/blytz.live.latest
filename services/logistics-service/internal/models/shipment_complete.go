package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Shipment Status Constants
const (
	ShipmentStatusCreated         = "order_created"
	ShipmentStatusPendingPickup   = "pending_pickup"
	ShipmentStatusPickedUp        = "picked_up"
	ShipmentStatusInTransit       = "in_transit"
	ShipmentStatusOutForDelivery  = "out_for_delivery"
	ShipmentStatusDelivered       = "delivered"
	ShipmentStatusFailedDelivery  = "failed_delivery"
	ShipmentStatusReturned        = "returned_to_sender"
	ShipmentStatusCancelled       = "cancelled"
)

// Service Type Constants
const (
	ServiceTypeStandard = "standard"
	ServiceTypeExpress  = "express"
	ServiceTypeSameDay  = "same_day"
	ServiceTypeEconomy  = "economy"
	ServiceTypeInternational = "international"
)

// Country Code Constants
const (
	CountryMalaysia = "MY"
	CountrySingapore = "SG"
	CountryThailand = "TH"
	CountryIndonesia = "ID"
	CountryPhilippines = "PH"
)

// Enhanced Shipment Model
type Shipment struct {
	ID               string          `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID          string          `json:"order_id" gorm:"not null;index"`
	NinjaOrderID     string          `json:"ninja_order_id" gorm:"index"` // Ninjavan order ID
	TrackingNumber   string          `json:"tracking_number" gorm:"not null;uniqueIndex"`
	ServiceType      string          `json:"service_type" gorm:"not null"`
	ServiceLevel     string          `json:"service_level" gorm:"not null"`
	Status           string          `json:"status" gorm:"not null;default:order_created"`
	EstimatedDelivery time.Time       `json:"estimated_delivery"`
	ActualDelivery   *time.Time      `json:"actual_delivery"`
	PickupDate       time.Time       `json:"pickup_date"`
	PickupTimeFrom   string          `json:"pickup_time_from" gorm:""`
	PickupTimeTo     string          `json:"pickup_time_to" gorm:""`
	Weight           float64         `json:"weight" gorm:"not null"`
	DeclaredValue    float64         `json:"declared_value" gorm:"not null"`
	TotalPrice       float64         `json:"total_price" gorm:"not null"`
	Currency         string          `json:"currency" gorm:"not null;default:MYR"`
	Insurance        bool            `json:"insurance" gorm:"default:false"`
	InsuranceAmount  float64         `json:"insurance_amount" gorm:"default:0"`
	IsFragile       bool            `json:"is_fragile" gorm:"default:false"`
	RequiresSignature bool          `json:"requires_signature" gorm:"default:false"`
	PickupInstructions string        `json:"pickup_instructions" gorm:"type:text"`
	DeliveryInstructions string     `json:"delivery_instructions" gorm:"type:text"`
	PackageDimensions string        `json:"package_dimensions" gorm:"type:text"` // JSON
	TrackingEvents   string        `json:"tracking_events" gorm:"type:text"` // JSON array
	WebhookEvents    string        `json:"webhook_events" gorm:"type:text"` // JSON array
	Metadata         string        `json:"metadata" gorm:"type:text"` // Additional data
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	
	// Addresses
	ShipperAddress   string         `json:"shipper_address" gorm:"type:text"` // JSON
	RecipientAddress string         `json:"recipient_address" gorm:"type:text"` // JSON
	
	// Relationships
	Order            *Order         `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Packages         []Package      `json:"packages,omitempty" gorm:"foreignKey:ShipmentID"`
	WebhookLogs      []WebhookLog   `json:"webhook_logs,omitempty" gorm:"foreignKey:ShipmentID"`
}

// Package Model
type Package struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ShipmentID     string    `json:"shipment_id" gorm:"not null;index"`
	PackageNumber  int       `json:"package_number" gorm:"not null"`
	Weight         float64   `json:"weight" gorm:"not null"`
	Length        float64   `json:"length" gorm:"not null"`
	Width         float64   `json:"width" gorm:"not null"`
	Height        float64   `json:"height" gorm:"not null"`
	Description    string    `json:"description" gorm:"type:text"`
	ContentValue  float64   `json:"content_value" gorm:"default:0"`
	IsFragile     bool      `json:"is_fragile" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	
	// Relationships
	Shipment       *Shipment `json:"shipment,omitempty" gorm:"foreignKey:ShipmentID"`
}

// Address Model
type Address struct {
	Name         string `json:"name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Email        string `json:"email" binding:"omitempty,email"`
	Address1     string `json:"address1" binding:"required"`
	Address2     string `json:"address2" binding:"omitempty"`
	Area         string `json:"area" binding:"omitempty"`
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	Country      string `json:"country" binding:"required"`
	PostalCode   string `json:"postal_code" binding:"required"`
	Instructions string `json:"instructions" binding:"omitempty"`
}

// Tracking Event Model
type TrackingEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
}

// Shipping Rate Model
type ShippingRate struct {
	ServiceType      string  `json:"service_type"`
	ServiceLevel     string  `json:"service_level"`
	Price            float64 `json:"price"`
	EstimatedDays    int     `json:"estimated_days"`
	MaxWeight        float64 `json:"max_weight"`
	MaxDimensions    string  `json:"max_dimensions"`
	IsAvailable      bool    `json:"is_available"`
	Description      string  `json:"description"`
}

// Coverage Area Model
type CoverageArea struct {
	PostalCode      string   `json:"postal_code"`
	Area            string   `json:"area"`
	City            string   `json:"city"`
	State           string   `json:"state"`
	Country         string   `json:"country"`
	IsCovered       bool     `json:"is_covered"`
	AvailableServices []string `json:"available_services"`
	DeliveryTime    string   `json:"delivery_time"`
	PriceFrom       float64  `json:"price_from"`
}

// Webhook Log Model
type WebhookLog struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ShipmentID string    `json:"shipment_id" gorm:"not null;index"`
	Event      string    `json:"event" gorm:"not null"` // tracking.updated, order.created, etc.
	Payload    string    `json:"payload" gorm:"type:text"` // JSON payload
	Signature  string    `json:"signature" gorm:"not null"` // Webhook signature
	IsVerified bool      `json:"is_verified" gorm:"default:false"`
	Processed  bool      `json:"processed" gorm:"default:false"`
	Response   string    `json:"response" gorm:"type:text"` // Response data
	StatusCode int       `json:"status_code" gorm:"default:200"`
	IPAddress  string    `json:"ip_address" gorm:""`
	UserAgent   string    `json:"user_agent" gorm:""`
	CreatedAt  time.Time `json:"created_at"`
	
	// Relationships
	Shipment    *Shipment `json:"shipment,omitempty" gorm:"foreignKey:ShipmentID"`
}

// Ninjavan API Request/Response Models

// NinjavanCreateOrderRequest for Ninjavan API
type NinjavanCreateOrderRequest struct {
	ServiceType       string             `json:"service_type"`
	TrackingNumber    string             `json:"tracking_number"`
	Reference         string             `json:"reference"`
	Shipper           Address            `json:"shipper"`
	Recipient         Address            `json:"recipient"`
	Pickup            NinjavanPickup    `json:"pickup"`
	Parcels           []NinjavanParcel   `json:"parcels"`
}

// NinjavanPickup Model
type NinjavanPickup struct {
	PickupDate     string `json:"pickup_date"`
	PickupTimeFrom string `json:"pickup_time_from"`
	PickupTimeTo   string `json:"pickup_time_to"`
	Instructions   string `json:"instructions"`
}

// NinjavanParcel Model
type NinjavanParcel struct {
	Weight         float64 `json:"weight"`
	Dimension      NinjavanDimension `json:"dimension"`
	Description    string  `json:"description"`
	DeclaredValue  float64 `json:"declared_value"`
}

// NinjavanDimension Model
type NinjavanDimension struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// NinjavanCreateOrderResponse from Ninjavan API
type NinjavanCreateOrderResponse struct {
	Status               string    `json:"status"`
	OrderID              string    `json:"order_id"`
	TrackingNumber       string    `json:"tracking_number"`
	ServiceType          string    `json:"service_type"`
	EstimatedDelivery    string    `json:"estimated_delivery"`
	TotalPrice           float64   `json:"total_price"`
	CreatedAt            string    `json:"created_at"`
}

// NinjavanOrderDetailsResponse from Ninjavan API
type NinjavanOrderDetailsResponse struct {
	OrderID              string              `json:"order_id"`
	TrackingNumber       string              `json:"tracking_number"`
	Status               string              `json:"status"`
	ServiceType          string              `json:"service_type"`
	CreatedAt            string              `json:"created_at"`
	PickedUpAt           string              `json:"picked_up_at"`
	EstimatedDelivery    string              `json:"estimated_delivery"`
	Shipper              Address             `json:"shipper"`
	Recipient            Address             `json:"recipient"`
	Parcels              []NinjavanParcel    `json:"parcels"`
	TrackingEvents       []TrackingEvent     `json:"tracking_events"`
}

// NinjavanCancelOrderRequest for Ninjavan API
type NinjavanCancelOrderRequest struct {
	Reason    string `json:"reason"`
	CancelBy  string `json:"cancel_by"`
}

// NinjavanCancelOrderResponse from Ninjavan API
type NinjavanCancelOrderResponse struct {
	Status             string  `json:"status"`
	Message            string  `json:"message"`
	RefundAmount       float64 `json:"refund_amount"`
	RefundProcessing   string  `json:"refund_processing_time"`
}

// NinjavanTrackingResponse from Ninjavan API
type NinjavanTrackingResponse struct {
	TrackingNumber      string           `json:"tracking_number"`
	Status              string           `json:"status"`
	EstimatedDelivery    string           `json:"estimated_delivery"`
	CurrentLocation      string           `json:"current_location"`
	ProgressPercentage  int              `json:"progress_percentage"`
	Events              []TrackingEvent  `json:"events"`
}

// NinjavanRatesRequest for Ninjavan API
type NinjavanRatesRequest struct {
	Origin      NinjavanAddress     `json:"origin"`
	Destination NinjavanAddress     `json:"destination"`
	Parcels     []NinjavanParcel    `json:"parcels"`
	ServiceType string              `json:"service_type"`
}

// NinjavanAddress Model
type NinjavanAddress struct {
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// NinjavanRatesResponse from Ninjavan API
type NinjavanRatesResponse struct {
	Rates    []ShippingRate `json:"rates"`
	Currency string         `json:"currency"`
}

// NinjavanServicesResponse from Ninjavan API
type NinjavanServicesResponse struct {
	Services []ShippingRate `json:"services"`
	Currency string         `json:"currency"`
}

// NinjavanAddressValidationRequest for Ninjavan API
type NinjavanAddressValidationRequest struct {
	Address Address `json:"address"`
}

// NinjavanAddressValidationResponse from Ninjavan API
type NinjavanAddressValidationResponse struct {
	IsValid                bool     `json:"is_valid"`
	StandardizedAddress     Address  `json:"standardized_address"`
	Suggestions            []Address `json:"suggestions"`
	PostalCodeInfo         CoverageArea `json:"postal_code_info"`
}

// Request Models

// CreateShipmentRequest represents shipment creation request
type CreateShipmentRequest struct {
	OrderID                string             `json:"order_id" binding:"required"`
	ServiceType            string             `json:"service_type" binding:"required"`
	ServiceLevel           string             `json:"service_level" binding:"required"`
	Shipper                Address            `json:"shipper" binding:"required"`
	Recipient              Address            `json:"recipient" binding:"required"`
	PickupDate             string             `json:"pickup_date" binding:"required"`
	PickupTimeFrom         string             `json:"pickup_time_from" binding:"required"`
	PickupTimeTo           string             `json:"pickup_time_to" binding:"required"`
	PickupInstructions      string             `json:"pickup_instructions" binding:"max=500"`
	DeliveryInstructions    string             `json:"delivery_instructions" binding:"max=500"`
	Packages               []NinjavanParcel   `json:"packages" binding:"required,min=1"`
	Insurance              bool               `json:"insurance"`
	IsFragile              bool               `json:"is_fragile"`
	RequiresSignature       bool               `json:"requires_signature"`
	Metadata               string             `json:"metadata"`
}

// UpdateShipmentRequest represents shipment update request
type UpdateShipmentRequest struct {
	PickupDate          string `json:"pickup_date" binding:"omitempty"`
	PickupTimeFrom      string `json:"pickup_time_from" binding:"omitempty"`
	PickupTimeTo        string `json:"pickup_time_to" binding:"omitempty"`
	PickupInstructions  string `json:"pickup_instructions" binding:"omitempty,max=500"`
	DeliveryInstructions string `json:"delivery_instructions" binding:"omitempty,max=500"`
	Insurance           bool   `json:"insurance"`
	IsFragile          bool   `json:"is_fragile"`
	RequiresSignature   bool   `json:"requires_signature"`
}

// CancelShipmentRequest represents shipment cancellation request
type CancelShipmentRequest struct {
	Reason   string `json:"reason" binding:"required,max=500"`
	CancelBy string `json:"cancel_by" binding:"required,oneof=merchant customer ninjavan"`
}

// GetRatesRequest represents shipping rates request
type GetRatesRequest struct {
	OriginPostalCode    string             `json:"origin_postal_code" binding:"required"`
	OriginCountry      string             `json:"origin_country" binding:"required"`
	DestinationPostalCode string          `json:"destination_postal_code" binding:"required"`
	DestinationCountry string          `json:"destination_country" binding:"required"`
	Packages           []NinjavanParcel  `json:"packages" binding:"required,min=1"`
	ServiceType       string             `json:"service_type"`
}

// ValidateAddressRequest represents address validation request
type ValidateAddressRequest struct {
	Address Address `json:"address" binding:"required"`
}

// SearchShipmentsRequest represents search shipments request
type SearchShipmentsRequest struct {
	Query           string    `json:"query" form:"query"`
	OrderID         string    `json:"order_id" form:"order_id"`
	TrackingNumber  string    `json:"tracking_number" form:"tracking_number"`
	Status          string    `json:"status" form:"status"`
	ServiceType     string    `json:"service_type" form:"service_type"`
	ShipperName     string    `json:"shipper_name" form:"shipper_name"`
	RecipientName   string    `json:"recipient_name" form:"recipient_name"`
	MinWeight       float64   `json:"min_weight" form:"min_weight"`
	MaxWeight       float64   `json:"max_weight" form:"max_weight"`
	MinPrice        float64   `json:"min_price" form:"min_price"`
	MaxPrice        float64   `json:"max_price" form:"max_price"`
	StartDate       string    `json:"start_date" form:"start_date"`
	EndDate         string    `json:"end_date" form:"end_date"`
	SortBy          string    `json:"sort_by" form:"sort_by"` // created_at, estimated_delivery, price
	SortOrder       string    `json:"sort_order" form:"sort_order"` // asc, desc
	Page            int       `json:"page" form:"page"`
	Limit           int       `json:"limit" form:"limit"`
}

// Response Models

// ShipmentResponse represents shipment response
type ShipmentResponse struct {
	Shipment           Shipment          `json:"shipment"`
	Packages           []Package         `json:"packages"`
	TrackingEvents     []TrackingEvent   `json:"tracking_events"`
	TimeRemaining      string            `json:"time_remaining"`
	ProgressPercentage int               `json:"progress_percentage"`
	CanTrack          bool              `json:"can_track"`
	CanCancel         bool              `json:"can_cancel"`
	EstimatedDaysLeft int               `json:"estimated_days_left"`
}

// ShipmentsResponse represents shipments list response
type ShipmentsResponse struct {
	Shipments    []Shipment `json:"shipments"`
	Total        int64      `json:"total"`
	Page         int        `json:"page"`
	Limit        int        `json:"limit"`
	TotalPages   int        `json:"total_pages"`
	HasNext      bool       `json:"has_next"`
	HasPrevious  bool       `json:"has_previous"`
}

// TrackingResponse represents tracking response
type TrackingResponse struct {
	TrackingNumber      string           `json:"tracking_number"`
	Status              string           `json:"status"`
	EstimatedDelivery    string           `json:"estimated_delivery"`
	CurrentLocation      string           `json:"current_location"`
	ProgressPercentage  int              `json:"progress_percentage"`
	Events              []TrackingEvent  `json:"events"`
	StatusText           string           `json:"status_text"`
}

// RatesResponse represents rates response
type RatesResponse struct {
	Rates    []ShippingRate `json:"rates"`
	Currency string         `json:"currency"`
	Total    int            `json:"total"`
	Origin   string         `json:"origin"`
	Destination string      `json:"destination"`
}

// ServicesResponse represents services response
type ServicesResponse struct {
	Services []ShippingRate `json:"services"`
	Currency string         `json:"currency"`
	Total    int            `json:"total"`
}

// AddressValidationResponse represents address validation response
type AddressValidationResponse struct {
	IsValid            bool            `json:"is_valid"`
	StandardizedAddress Address         `json:"standardized_address"`
	Suggestions       []Address       `json:"suggestions"`
	PostalCodeInfo     CoverageArea    `json:"postal_code_info"`
	Message            string          `json:"message"`
}

// Webhook Messages (from Ninjavan)

// NinjavanWebhookOrderCreated from Ninjavan
type NinjavanWebhookOrderCreated struct {
	Event         string    `json:"event"`
	OrderID       string    `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	Reference     string    `json:"reference"`
	ServiceType   string    `json:"service_type"`
	Timestamp     string    `json:"timestamp"`
	Signature     string    `json:"signature"`
}

// NinjavanWebhookTrackingUpdated from Ninjavan
type NinjavanWebhookTrackingUpdated struct {
	Event          string          `json:"event"`
	TrackingNumber string          `json:"tracking_number"`
	Status         string          `json:"status"`
	Location       string          `json:"location"`
	Description    string          `json:"description"`
	Timestamp      string          `json:"timestamp"`
	Signature      string          `json:"signature"`
}

// NinjavanWebhookOrderDelivered from Ninjavan
type NinjavanWebhookOrderDelivered struct {
	Event          string    `json:"event"`
	OrderID        string    `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	DeliveredAt    string    `json:"delivered_at"`
	SignedBy       string    `json:"signed_by"`
	Timestamp      string    `json:"timestamp"`
	Signature      string    `json:"signature"`
}

// Order Model (simplified for logistics service)
type Order struct {
	OrderID        string  `json:"order_id"`
	UserID         string  `json:"user_id"`
	TotalAmount    float64 `json:"total_amount"`
	Status         string  `json:"status"`
	ShippingStatus string  `json:"shipping_status"`
	CreatedAt      string  `json:"created_at"`
}

// Helper Methods

// BeforeCreate hook for Shipment
func (s *Shipment) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = generateUUID()
	}
	if s.Currency == "" {
		s.Currency = "MYR"
	}
	if s.Status == "" {
		s.Status = ShipmentStatusCreated
	}
	return nil
}

// BeforeCreate hook for Package
func (p *Package) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = generateUUID()
	}
	return nil
}

// BeforeCreate hook for WebhookLog
func (w *WebhookLog) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = generateUUID()
	}
	return nil
}

// GetShipperAddress returns shipper address as Address struct
func (s *Shipment) GetShipperAddress() Address {
	if s.ShipperAddress == "" {
		return Address{}
	}
	var address Address
	json.Unmarshal([]byte(s.ShipperAddress), &address)
	return address
}

// SetShipperAddress sets shipper address from Address struct
func (s *Shipment) SetShipperAddress(address Address) {
	data, _ := json.Marshal(address)
	s.ShipperAddress = string(data)
}

// GetRecipientAddress returns recipient address as Address struct
func (s *Shipment) GetRecipientAddress() Address {
	if s.RecipientAddress == "" {
		return Address{}
	}
	var address Address
	json.Unmarshal([]byte(s.RecipientAddress), &address)
	return address
}

// SetRecipientAddress sets recipient address from Address struct
func (s *Shipment) SetRecipientAddress(address Address) {
	data, _ := json.Marshal(address)
	s.RecipientAddress = string(data)
}

// GetPackageDimensions returns package dimensions as array
func (s *Shipment) GetPackageDimensions() []NinjavanDimension {
	if s.PackageDimensions == "" {
		return []NinjavanDimension{}
	}
	var dimensions []NinjavanDimension
	json.Unmarshal([]byte(s.PackageDimensions), &dimensions)
	return dimensions
}

// SetPackageDimensions sets package dimensions from array
func (s *Shipment) SetPackageDimensions(dimensions []NinjavanDimension) {
	if len(dimensions) == 0 {
		s.PackageDimensions = ""
		return
	}
	data, _ := json.Marshal(dimensions)
	s.PackageDimensions = string(data)
}

// GetTrackingEvents returns tracking events as array
func (s *Shipment) GetTrackingEvents() []TrackingEvent {
	if s.TrackingEvents == "" {
		return []TrackingEvent{}
	}
	var events []TrackingEvent
	json.Unmarshal([]byte(s.TrackingEvents), &events)
	return events
}

// SetTrackingEvents sets tracking events from array
func (s *Shipment) SetTrackingEvents(events []TrackingEvent) {
	if len(events) == 0 {
		s.TrackingEvents = ""
		return
	}
	data, _ := json.Marshal(events)
	s.TrackingEvents = string(data)
}

// SetWebhookEvents sets webhook events from string array
func (s *Shipment) SetWebhookEvents(events []string) {
	if len(events) == 0 {
		s.WebhookEvents = ""
		return
	}
	data, _ := json.Marshal(events)
	s.WebhookEvents = string(data)
}

// GetWebhookEvents returns webhook events as string array
func (s *Shipment) GetWebhookEvents() []string {
	if s.WebhookEvents == "" {
		return []string{}
	}
	var events []string
	json.Unmarshal([]byte(s.WebhookEvents), &events)
	return events
}

// GetStatusText returns human readable status text
func (s *Shipment) GetStatusText() string {
	switch s.Status {
	case ShipmentStatusCreated:
		return "Order Created"
	case ShipmentStatusPendingPickup:
		return "Pending Pickup"
	case ShipmentStatusPickedUp:
		return "Picked Up"
	case ShipmentStatusInTransit:
		return "In Transit"
	case ShipmentStatusOutForDelivery:
		return "Out for Delivery"
	case ShipmentStatusDelivered:
		return "Delivered"
	case ShipmentStatusFailedDelivery:
		return "Delivery Failed"
	case ShipmentStatusReturned:
		return "Returned to Sender"
	case ShipmentStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown Status"
	}
}

// GetServiceTypeText returns human readable service type text
func (s *Shipment) GetServiceTypeText() string {
	switch s.ServiceType {
	case ServiceTypeStandard:
		return "Standard Delivery"
	case ServiceTypeExpress:
		return "Express Delivery"
	case ServiceTypeSameDay:
		return "Same Day Delivery"
	case ServiceTypeEconomy:
		return "Economy Delivery"
	case ServiceTypeInternational:
		return "International Delivery"
	default:
		return s.ServiceType
	}
}

// IsDelivered checks if shipment is delivered
func (s *Shipment) IsDelivered() bool {
	return s.Status == ShipmentStatusDelivered
}

// CanCancel checks if shipment can be cancelled
func (s *Shipment) CanCancel() bool {
	return s.Status == ShipmentStatusCreated || s.Status == ShipmentStatusPendingPickup
}

// GetProgressPercentage calculates progress percentage based on status
func (s *Shipment) GetProgressPercentage() int {
	switch s.Status {
	case ShipmentStatusCreated:
		return 5
	case ShipmentStatusPendingPickup:
		return 10
	case ShipmentStatusPickedUp:
		return 25
	case ShipmentStatusInTransit:
		return 60
	case ShipmentStatusOutForDelivery:
		return 80
	case ShipmentStatusDelivered:
		return 100
	case ShipmentStatusFailedDelivery:
		return 70
	case ShipmentStatusReturned:
		return 100
	case ShipmentStatusCancelled:
		return 100
	default:
		return 0
	}
}

// GetDaysRemaining calculates days remaining until delivery
func (s *Shipment) GetDaysRemaining() int {
	if s.IsDelivered() {
		return 0
	}
	
	now := time.Now()
	if now.After(s.EstimatedDelivery) {
		return 0
	}
	
	days := int(s.EstimatedDelivery.Sub(now).Hours() / 24)
	if days < 0 {
		return 0
	}
	
	return days
}

// GetTimeRemaining returns time remaining as string
func (s *Shipment) GetTimeRemaining() string {
	if s.IsDelivered() {
		return "Delivered"
	}
	
	now := time.Now()
	if now.After(s.EstimatedDelivery) {
		return "Delayed"
	}
	
	duration := s.EstimatedDelivery.Sub(now)
	if duration <= 0 {
		return "Today"
	}
	
	if duration < 24*time.Hour {
		return fmt.Sprintf("%.0f hours", duration.Hours())
	} else {
		return fmt.Sprintf("%.0f days", duration.Hours()/24)
	}
}

// Helper function to generate UUID (simplified)
func generateUUID() string {
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
}