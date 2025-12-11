package utils

import (
	"os"
	"regexp"
	"unicode"
)

// GetEnv gets an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

// IsValidPassword validates password strength
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// IsValidPhone validates phone number format
func IsValidPhone(phone string) bool {
	// Basic phone validation - can be enhanced based on requirements
	pattern := `^\+?[1-9]\d{1,14}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(phone)
}

// IsValidUUID validates UUID format
func IsValidUUID(uuid string) bool {
	pattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(uuid)
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove HTML tags and potentially dangerous characters
	regex := regexp.MustCompile(`[<>&'"()]`)
	return regex.ReplaceAllString(input, "")
}

// ValidateRequired checks if required fields are present
func ValidateRequired(data map[string]interface{}, requiredFields []string) map[string]string {
	errors := make(map[string]string)
	
	for _, field := range requiredFields {
		if value, exists := data[field]; !exists || value == nil || value == "" {
			errors[field] = "This field is required"
		}
	}
	
	return errors
}

// ValidateStringLength validates string length constraints
func ValidateStringLength(value string, fieldName string, minLength, maxLength int) map[string]string {
	errors := make(map[string]string)
	
	if len(value) < minLength {
		errors[fieldName] = "Field is too short"
	}
	
	if len(value) > maxLength {
		errors[fieldName] = "Field is too long"
	}
	
	return errors
}