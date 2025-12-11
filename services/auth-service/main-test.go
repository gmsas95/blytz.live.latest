package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	shared_utils "github.com/gmsas95/blytz-mvp/shared/pkg/utils"
	shared_errors "github.com/gmsas95/blytz-mvp/shared/pkg/errors"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// CORS middleware using shared package
	r.Use(shared_utils.CORSMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		shared_utils.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
			"service": "auth-service",
			"version": "v2.0-test",
			"status":  "healthy",
			"message": "Auth service is working with shared package",
		})
	})

	// Test error endpoint
	r.GET("/test-error", func(c *gin.Context) {
		shared_utils.SendErrorResponse(c, shared_errors.ValidationError("TEST_ERROR", "This is a test error"))
	})

	// Start server
	port := "8084"
	fmt.Printf("🚀 Auth Service Test starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🧪 Test error: http://localhost:%s/test-error\n", port)

	log.Fatal(r.Run(":" + port))
}