package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gmsas95/blytz.live.latest/shared/pkg/errors"
)

// Response represents a standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Response
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// HealthStatus represents health check response
type HealthStatus struct {
	Status  string             `json:"status"`
	Checks  map[string]CheckResult `json:"checks,omitempty"`
	Uptime  string             `json:"uptime,omitempty"`
	Version string             `json:"version,omitempty"`
}

// CheckResult represents individual health check result
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// NewHealthStatus creates a new health status
func NewHealthStatus(status string) *HealthStatus {
	return &HealthStatus{
		Status: status,
		Checks: make(map[string]CheckResult),
	}
}

// AddService adds a service check result
func (h *HealthStatus) AddService(serviceName, status, message string) {
	h.Checks[serviceName] = CheckResult{
		Status:  status,
		Message: message,
	}
}

// SendSuccessResponse sends a successful response
func SendSuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	response := Response{
		Success:   true,
		Message:   getSuccessMessage(statusCode),
		Data:      data,
		Timestamp: time.Now().UTC(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(statusCode, response)
}

// SendSuccessResponseWithMessage sends a successful response with custom message
func SendSuccessResponseWithMessage(c *gin.Context, statusCode int, message string, data interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(statusCode, response)
}

// SendErrorResponse sends an error response
func SendErrorResponse(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	message := "Internal server error"
	var errorData interface{}
	
	// Check if it's an AppError
	if appErr, ok := errors.IsAppError(err); ok {
		statusCode = appErr.HTTPStatus
		message = appErr.Message
		errorData = gin.H{
			"type":    appErr.Type,
			"code":    appErr.Code,
			"details": appErr.Details,
		}
	} else {
		errorData = gin.H{
			"type": "INTERNAL_ERROR",
			"code": "INTERNAL_SERVER_ERROR",
		}
	}
	
	response := Response{
		Success:   false,
		Message:   message,
		Error:     errorData,
		Timestamp: time.Now().UTC(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(statusCode, response)
}

// SendPaginatedResponse sends a paginated response
func SendPaginatedResponse(c *gin.Context, statusCode int, data interface{}, pagination PaginationInfo) {
	response := PaginatedResponse{
		Response: Response{
			Success:   true,
			Message:   getSuccessMessage(statusCode),
			Data:      data,
			Timestamp: time.Now().UTC(),
			RequestID: getRequestID(c),
		},
		Pagination: pagination,
	}
	
	c.JSON(statusCode, response)
}

// SendValidationErrorResponse sends a validation error response
func SendValidationErrorResponse(c *gin.Context, validationErrors map[string]string) {
	response := Response{
		Success:   false,
		Message:   "Validation failed",
		Error: gin.H{
			"type":    "VALIDATION_ERROR",
			"code":    "VALIDATION_FAILED",
			"details": validationErrors,
		},
		Timestamp: time.Now().UTC(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(http.StatusBadRequest, response)
}

// SendJSON sends a raw JSON response
func SendJSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// SendEmptyResponse sends an empty response with status code
func SendEmptyResponse(c *gin.Context, statusCode int) {
	c.Status(statusCode)
}

// getRequestID extracts request ID from context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// getSuccessMessage returns a default success message based on status code
func getSuccessMessage(statusCode int) string {
	switch statusCode {
	case http.StatusOK:
		return "Request successful"
	case http.StatusCreated:
		return "Resource created successfully"
	case http.StatusAccepted:
		return "Request accepted for processing"
	case http.StatusNoContent:
		return "Request processed successfully"
	default:
		return "Success"
	}
}

// CalculatePagination calculates pagination information
func CalculatePagination(page, perPage int, total int64) PaginationInfo {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}
	
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	
	return PaginationInfo{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// GetPaginationParams extracts pagination parameters from query
func GetPaginationParams(c *gin.Context) (page, perPage int) {
	page = 1
	perPage = 10
	
	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	if pp := c.Query("per_page"); pp != "" {
		if parsed, err := parseInt(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}
	
	return page, perPage
}

// parseInt is a helper to parse integer strings
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}