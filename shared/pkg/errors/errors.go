package errors

import (
	"fmt"
)

// AppError represents a structured application error
type AppError struct {
	Type       string `json:"type"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	HTTPStatus int    `json:"http_status"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s - %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

// ToJSON returns the error as a JSON-serializable map
func (e *AppError) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"type":        e.Type,
		"code":        e.Code,
		"message":     e.Message,
		"http_status": e.HTTPStatus,
	}
	
	if e.Details != "" {
		result["details"] = e.Details
	}
	
	return result
}