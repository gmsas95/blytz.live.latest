// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gmsas95/blytz.live.latest/services/auction-service/internal/models"
)

const (
	auctionBaseURL = "http://localhost:8087"
	timeout        = 10 * time.Second
)

func waitForAuctionService(t *testing.T) {
	client := &http.Client{Timeout: timeout}
	maxRetries := 30

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(auctionBaseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	t.Fatal("Auction service did not become healthy within timeout")
}

func setupTestAuction() *models.Auction {
	return &models.Auction{
		Title:         "Integration Test Auction",
		Description:   "Test auction for integration testing",
		StartingPrice: 100.00,
		CurrentPrice:  100.00,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        "active",
		SellerID:      "seller-integration-test",
		Category:      "electronics",
		Condition:     "new",
		Images:        []string{"image1.jpg", "image2.jpg"},
	}
}

func TestAuctionIntegration(t *testing.T) {
	waitForAuctionService(t)
	client := &http.Client{Timeout: timeout}

	t.Run("Complete auction lifecycle", func(t *testing.T) {
		// Step 1: Create a new auction
		auction := setupTestAuction()
		
		jsonBody, err := json.Marshal(auction)
		if err != nil {
			t.Fatalf("Failed to marshal auction: %v", err)
		}

		resp, err := client.Post(auctionBaseURL+"/api/v1/auctions", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create auction: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			t.Fatalf("Expected status 201, got %d. Response: %s", resp.StatusCode, string(body[:n]))
		}

		var createdAuction models.Auction
		if err := json.NewDecoder(resp.Body).Decode(&createdAuction); err != nil {
			t.Fatalf("Failed to decode auction response: %v", err)
		}

		if createdAuction.ID == "" {
			t.Fatal("Expected auction ID to be set")
		}

		// Step 2: Get the created auction
		resp, err = client.Get(auctionBaseURL + "/api/v1/auctions/" + createdAuction.ID)
		if err != nil {
			t.Fatalf("Failed to get auction: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var retrievedAuction models.Auction
		if err := json.NewDecoder(resp.Body).Decode(&retrievedAuction); err != nil {
			t.Fatalf("Failed to decode auction response: %v", err)
		}

		if retrievedAuction.Title != auction.Title {
			t.Fatalf("Expected title %s, got %s", auction.Title, retrievedAuction.Title)
		}

		// Step 3: Get all active auctions
		resp, err = client.Get(auctionBaseURL + "/api/v1/auctions/active")
		if err != nil {
			t.Fatalf("Failed to get active auctions: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var activeAuctions []models.Auction
		if err := json.NewDecoder(resp.Body).Decode(&activeAuctions); err != nil {
			t.Fatalf("Failed to decode active auctions response: %v", err)
		}

		if len(activeAuctions) == 0 {
			t.Fatal("Expected at least one active auction")
		}

		// Step 4: Place a bid
		bid := &models.Bid{
			AuctionID: createdAuction.ID,
			BidderID:  "bidder-integration-test",
			Amount:    150.00,
		}

		jsonBody, err = json.Marshal(bid)
		if err != nil {
			t.Fatalf("Failed to marshal bid: %v", err)
		}

		resp, err = client.Post(auctionBaseURL+"/api/v1/auctions/"+createdAuction.ID+"/bids", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to place bid: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			t.Fatalf("Expected status 201, got %d. Response: %s", resp.StatusCode, string(body[:n]))
		}

		// Step 5: Get bids for the auction
		resp, err = client.Get(auctionBaseURL + "/api/v1/auctions/" + createdAuction.ID + "/bids")
		if err != nil {
			t.Fatalf("Failed to get bids: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var bids []models.Bid
		if err := json.NewDecoder(resp.Body).Decode(&bids); err != nil {
			t.Fatalf("Failed to decode bids response: %v", err)
		}

		if len(bids) == 0 {
			t.Fatal("Expected at least one bid")
		}

		// Step 6: Update auction status
		updateReq := map[string]string{"status": "ended"}
		jsonBody, err = json.Marshal(updateReq)
		if err != nil {
			t.Fatalf("Failed to marshal update request: %v", err)
		}

		req, err := http.NewRequest("PUT", auctionBaseURL+"/api/v1/auctions/"+createdAuction.ID+"/status", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create update request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("Failed to update auction status: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid bid amount", func(t *testing.T) {
		// Create an auction first
		auction := setupTestAuction()
		jsonBody, _ := json.Marshal(auction)
		
		resp, err := client.Post(auctionBaseURL+"/api/v1/auctions", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create auction: %v", err)
		}
		defer resp.Body.Close()

		var createdAuction models.Auction
		json.NewDecoder(resp.Body).Decode(&createdAuction)

		// Try to place a bid lower than starting price
		bid := &models.Bid{
			AuctionID: createdAuction.ID,
			BidderID:  "bidder-integration-test",
			Amount:    50.00, // Lower than starting price
		}

		jsonBody, err = json.Marshal(bid)
		if err != nil {
			t.Fatalf("Failed to marshal bid: %v", err)
		}

		resp, err = client.Post(auctionBaseURL+"/api/v1/auctions/"+createdAuction.ID+"/bids", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to place bid: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Get non-existent auction", func(t *testing.T) {
		resp, err := client.Get(auctionBaseURL + "/api/v1/auctions/non-existent-id")
		if err != nil {
			t.Fatalf("Failed to get auction: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("Search auctions", func(t *testing.T) {
		// Create multiple auctions for search testing
		auctions := []*models.Auction{
			{
				Title:         "iPhone 13 Auction",
				Description:   "Latest iPhone model",
				StartingPrice: 500.00,
				CurrentPrice:  500.00,
				StartTime:     time.Now(),
				EndTime:       time.Now().Add(24 * time.Hour),
				Status:        "active",
				SellerID:      "seller-1",
				Category:      "electronics",
			},
			{
				Title:         "Samsung TV Auction",
				Description:   "Smart TV 55 inch",
				StartingPrice: 300.00,
				CurrentPrice:  300.00,
				StartTime:     time.Now(),
				EndTime:       time.Now().Add(24 * time.Hour),
				Status:        "active",
				SellerID:      "seller-2",
				Category:      "electronics",
			},
		}

		for _, auction := range auctions {
			jsonBody, _ := json.Marshal(auction)
			client.Post(auctionBaseURL+"/api/v1/auctions", "application/json", bytes.NewBuffer(jsonBody))
		}

		// Search for "iPhone"
		resp, err := client.Get(auctionBaseURL + "/api/v1/auctions/search?q=iPhone")
		if err != nil {
			t.Fatalf("Failed to search auctions: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var searchResults []models.Auction
		if err := json.NewDecoder(resp.Body).Decode(&searchResults); err != nil {
			t.Fatalf("Failed to decode search results: %v", err)
		}

		if len(searchResults) == 0 {
			t.Fatal("Expected search results for 'iPhone'")
		}
	})
}

func TestAuctionHealthCheck(t *testing.T) {
	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(auctionBaseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to check health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var healthResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if healthResp["status"] != "healthy" {
		t.Fatalf("Expected status 'healthy', got %v", healthResp["status"])
	}

	if healthResp["service"] != "auction-service" {
		t.Fatalf("Expected service 'auction-service', got %v", healthResp["service"])
	}
}

func TestConcurrentBids(t *testing.T) {
	waitForAuctionService(t)
	client := &http.Client{Timeout: timeout}

	// Create an auction
	auction := setupTestAuction()
	jsonBody, _ := json.Marshal(auction)
	
	resp, err := client.Post(auctionBaseURL+"/api/v1/auctions", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create auction: %v", err)
	}
	defer resp.Body.Close()

	var createdAuction models.Auction
	json.NewDecoder(resp.Body).Decode(&createdAuction)

	// Place multiple concurrent bids
	numBids := 10
	done := make(chan bool, numBids)

	for i := 0; i < numBids; i++ {
		go func(index int) {
			defer func() { done <- true }()

			bid := &models.Bid{
				AuctionID: createdAuction.ID,
				BidderID:  fmt.Sprintf("bidder-%d", index),
				Amount:    float64(100 + (index+1)*10), // Increasing amounts
			}

			jsonBody, err := json.Marshal(bid)
			if err != nil {
				t.Errorf("Failed to marshal bid %d: %v", index, err)
				return
			}

			resp, err := client.Post(auctionBaseURL+"/api/v1/auctions/"+createdAuction.ID+"/bids", "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Errorf("Failed to place bid %d: %v", index, err)
				return
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Unexpected status for bid %d: %d", index, resp.StatusCode)
			}
		}(i)
	}

	// Wait for all bids to complete
	for i := 0; i < numBids; i++ {
		<-done
	}
}

func BenchmarkAuctionEndpoints(b *testing.B) {
	waitForAuctionService(b)
	client := &http.Client{Timeout: timeout}

	// Create a test auction
	auction := setupTestAuction()
	jsonBody, _ := json.Marshal(auction)
	resp, _ := client.Post(auctionBaseURL+"/api/v1/auctions", "application/json", bytes.NewBuffer(jsonBody))
	if resp != nil {
		defer resp.Body.Close()
		var createdAuction models.Auction
		json.NewDecoder(resp.Body).Decode(&createdAuction)

		b.Run("GetAuction", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resp, _ := client.Get(auctionBaseURL + "/api/v1/auctions/" + createdAuction.ID)
				if resp != nil {
					resp.Body.Close()
				}
			}
		})

		b.Run("GetActiveAuctions", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resp, _ := client.Get(auctionBaseURL + "/api/v1/auctions/active")
				if resp != nil {
					resp.Body.Close()
				}
			}
		})

		b.Run("HealthCheck", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resp, _ := client.Get(auctionBaseURL + "/health")
				if resp != nil {
					resp.Body.Close()
				}
			}
		})
	}
}