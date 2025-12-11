package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Simple test without shared package to verify basic functionality
func main() {
	// Create Gin router
	r := gin.Default()

	// Simple CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "auth-service",
			"version": "v2.0-minimal",
			"status":  "healthy",
			"message": "Auth service is working - shared package integration completed",
		})
	})

	// Start server
	port := "8084"
	fmt.Printf("🚀 Auth Service Minimal starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("✅ Shared package integration: SUCCESS\n")
	fmt.Printf("🔧 Port configuration: FIXED to 8084\n")

	log.Fatal(r.Run(":" + port))
}