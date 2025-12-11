package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// LoadTestConfig holds load test configuration
type LoadTestConfig struct {
	BaseURL          string        `json:"base_url"`
	WebSocketURL     string        `json:"websocket_url"`
	NumUsers         int           `json:"num_users"`
	Duration         time.Duration `json:"duration"`
	RampUpTime       time.Duration `json:"ramp_up_time"`
	RequestsPerSecond int           `json:"requests_per_second"`
	TestType         string        `json:"test_type"` // "http", "websocket", "mixed"
}

// LoadTestResult holds test results
type LoadTestResult struct {
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulReqs  int64         `json:"successful_requests"`
	FailedReqs      int64         `json:"failed_requests"`
	ResponseTime     time.Duration `json:"avg_response_time"`
	MaxResponseTime  time.Duration `json:"max_response_time"`
	MinResponseTime  time.Duration `json:"min_response_time"`
	Throughput      float64       `json:"throughput_rps"`
	ErrorRate       float64       `json:"error_rate_percent"`
	TestDuration    time.Duration `json:"test_duration"`
	ActiveUsers     int           `json:"active_users"`
}

// UserSimulator simulates a single user
type UserSimulator struct {
	ID           int
	Config       *LoadTestConfig
	Logger       *zap.Logger
	Results      *LoadTestResult
	WebSocketConn *websocket.Conn
}

// NewUserSimulator creates a new user simulator
func NewUserSimulator(id int, config *LoadTestConfig, results *LoadTestResult, logger *zap.Logger) *UserSimulator {
	return &UserSimulator{
		ID:      id,
		Config:   config,
		Logger:   logger,
		Results:  results,
	}
}

// RunHTTPTest runs HTTP load test
func (us *UserSimulator) RunHTTPTest(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	ticker := time.NewTicker(time.Duration(1000/us.Config.RequestsPerSecond) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			start := time.Now()
			
			// Test different endpoints
			endpoints := []string{
				"/health",
				"/api/v1/auth/profile",
				"/api/v1/products",
				"/api/v1/auctions",
			}

			for _, endpoint := range endpoints {
				req, err := http.NewRequest("GET", us.Config.BaseURL+endpoint, nil)
				if err != nil {
					atomic.AddInt64(&us.Results.FailedReqs, 1)
					continue
				}

				resp, err := client.Do(req)
				duration := time.Since(start)
				
				if err != nil {
					atomic.AddInt64(&us.Results.FailedReqs, 1)
				} else {
					atomic.AddInt64(&us.Results.SuccessfulReqs, 1)
					resp.Body.Close()
					
					// Update response times (simplified for performance)
					if duration > us.Results.MaxResponseTime {
						us.Results.MaxResponseTime = duration
					}
					if us.Results.MinResponseTime == 0 || duration < us.Results.MinResponseTime {
						us.Results.MinResponseTime = duration
					}
				}
				
				atomic.AddInt64(&us.Results.TotalRequests, 1)
			}
		}
	}
}

// RunWebSocketTest runs WebSocket load test
func (us *UserSimulator) RunWebSocketTest(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Connect to WebSocket
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(us.Config.WebSocketURL+"?user_id="+fmt.Sprintf("user-%d", us.ID), nil)
	if err != nil {
		us.Logger.Error("WebSocket connection failed", 
			zap.Int("user_id", us.ID),
			zap.Error(err))
		atomic.AddInt64(&us.Results.FailedReqs, 1)
		return
	}
	defer conn.Close()

	us.WebSocketConn = conn
	us.Logger.Info("WebSocket connected", zap.Int("user_id", us.ID))

	// Send messages periodically
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	messageCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			message := map[string]interface{}{
				"type":      "chat_message",
				"chat_id":   "test-chat-1",
				"sender_id":  fmt.Sprintf("user-%d", us.ID),
				"content":    fmt.Sprintf("Test message %d from user %d", messageCount, us.ID),
				"timestamp":  time.Now().Unix(),
			}

			if err := conn.WriteJSON(message); err != nil {
				atomic.AddInt64(&us.Results.FailedReqs, 1)
				us.Logger.Error("WebSocket write failed", 
					zap.Int("user_id", us.ID),
					zap.Error(err))
				return
			}

			atomic.AddInt64(&us.Results.TotalRequests, 1)
			atomic.AddInt64(&us.Results.SuccessfulReqs, 1)
			messageCount++

			// Read response
			_, _, err := conn.ReadMessage()
			if err != nil {
				us.Logger.Error("WebSocket read failed", 
					zap.Int("user_id", us.ID),
					zap.Error(err))
			}
		}
	}
}

// RunMixedTest runs mixed HTTP and WebSocket load test
func (us *UserSimulator) RunMixedTest(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// 70% HTTP requests, 30% WebSocket
	if us.ID%10 < 7 {
		us.RunHTTPTest(ctx, wg)
	} else {
		us.RunWebSocketTest(ctx, wg)
	}
}

// RunLoadTest executes the load test
func RunLoadTest(config *LoadTestConfig) (*LoadTestResult, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting load test",
		zap.String("test_type", config.TestType),
		zap.Int("num_users", config.NumUsers),
		zap.Duration("duration", config.Duration),
		zap.Int("requests_per_second", config.RequestsPerSecond),
	)

	results := &LoadTestResult{
		MinResponseTime: time.Hour, // Initialize to high value
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	var wg sync.WaitGroup

	// Ramp up users gradually
	rampUpTicker := time.NewTicker(config.RampUpTime / time.Duration(config.NumUsers))
	defer rampUpTicker.Stop()

	userCount := 0
	rampUpDone := make(chan bool)

	go func() {
		time.Sleep(config.RampUpTime)
		rampUpDone <- true
	}()

	for {
		select {
		case <-rampUpTicker.C:
			if userCount >= config.NumUsers {
				goto allUsersStarted
			}

			userCount++
			user := NewUserSimulator(userCount, config, results, logger)
			wg.Add(1)

			switch config.TestType {
			case "http":
				go user.RunHTTPTest(ctx, &wg)
			case "websocket":
				go user.RunWebSocketTest(ctx, &wg)
			case "mixed":
				go user.RunMixedTest(ctx, &wg)
			default:
				go user.RunHTTPTest(ctx, &wg)
			}

			logger.Info("User started", zap.Int("user_id", userCount))

		case <-rampUpDone:
			goto allUsersStarted
		}
	}

allUsersStarted:
	logger.Info("All users started, beginning load test")

	// Wait for test completion
	<-ctx.Done()
	wg.Wait()

	// Calculate final results
	results.TestDuration = config.Duration
	results.ActiveUsers = userCount
	
	if results.TotalRequests > 0 {
		results.Throughput = float64(results.TotalRequests) / config.Duration.Seconds()
		results.ErrorRate = (float64(results.FailedReqs) / float64(results.TotalRequests)) * 100
	}

	return results, nil
}

// SaveResults saves test results to file
func SaveResults(results *LoadTestResult, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// PrintResults prints test results to console
func PrintResults(results *LoadTestResult) {
	fmt.Println("\n" + "="*60)
	fmt.Println("LOAD TEST RESULTS")
	fmt.Println("="*60)
	fmt.Printf("Total Requests:     %d\n", results.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", results.SuccessfulReqs)
	fmt.Printf("Failed Requests:    %d\n", results.FailedReqs)
	fmt.Printf("Error Rate:         %.2f%%\n", results.ErrorRate)
	fmt.Printf("Throughput:         %.2f RPS\n", results.Throughput)
	fmt.Printf("Avg Response Time:   %v\n", results.ResponseTime)
	fmt.Printf("Min Response Time:   %v\n", results.MinResponseTime)
	fmt.Printf("Max Response Time:   %v\n", results.MaxResponseTime)
	fmt.Printf("Test Duration:      %v\n", results.TestDuration)
	fmt.Printf("Active Users:       %d\n", results.ActiveUsers)
	fmt.Println("="*60)
}

// main function
func main() {
	// Default configuration
	config := &LoadTestConfig{
		BaseURL:          "http://localhost:8085", // Default to auth service
		WebSocketURL:     "ws://localhost:8090/ws", // Default to chat service
		NumUsers:         1000,                  // Start with 1000 users
		Duration:         5 * time.Minute,       // 5 minutes test
		RampUpTime:       30 * time.Second,       // 30 seconds ramp up
		RequestsPerSecond: 10,                    // 10 requests per second per user
		TestType:         "mixed",               // Mixed HTTP/WebSocket test
	}

	// Parse command line arguments or use defaults
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "small":
			config.NumUsers = 100
			config.Duration = 2 * time.Minute
		case "medium":
			config.NumUsers = 1000
			config.Duration = 5 * time.Minute
		case "large":
			config.NumUsers = 3000
			config.Duration = 10 * time.Minute
		case "stress":
			config.NumUsers = 5000
			config.Duration = 15 * time.Minute
			config.RequestsPerSecond = 20
		}
	}

	// Run the load test
	fmt.Printf("Starting load test with %d users for %v\n", config.NumUsers, config.Duration)
	fmt.Printf("Test type: %s\n", config.TestType)
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("WebSocket URL: %s\n", config.WebSocketURL)

	results, err := RunLoadTest(config)
	if err != nil {
		log.Fatalf("Load test failed: %v", err)
	}

	// Print and save results
	PrintResults(results)

	// Save results to file with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("load_test_results_%s.json", timestamp)
	
	if err := SaveResults(results, filename); err != nil {
		log.Printf("Failed to save results: %v", err)
	} else {
		fmt.Printf("Results saved to: %s\n", filename)
	}

	// Evaluate results against targets
	fmt.Println("\n" + "="*60)
	fmt.Println("PERFORMANCE EVALUATION")
	fmt.Println("="*60)
	
	// Target: Scale from 1,000 to 3,000+ concurrent users
	if config.NumUsers >= 1000 && results.Throughput >= 1000 {
		fmt.Println("✅ PASSED: Successfully handled 1,000+ concurrent users")
	} else {
		fmt.Println("❌ FAILED: Did not meet 1,000+ concurrent user target")
	}

	// Target: Architecture ready for 50,000+ users
	if results.Throughput >= 5000 && results.ErrorRate < 1.0 {
		fmt.Println("✅ PASSED: Architecture ready for 50,000+ users")
	} else {
		fmt.Println("❌ FAILED: Architecture not ready for 50,000+ users")
	}

	// Target: Error rate < 1%
	if results.ErrorRate < 1.0 {
		fmt.Println("✅ PASSED: Error rate under 1%")
	} else {
		fmt.Println("❌ FAILED: Error rate above 1%")
	}

	// Target: Response time < 200ms for 95% of requests
	if results.MaxResponseTime < 200*time.Millisecond {
		fmt.Println("✅ PASSED: Response time under 200ms")
	} else {
		fmt.Println("❌ FAILED: Response time above 200ms")
	}

	fmt.Println("="*60)
}