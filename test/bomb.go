package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Configuration parameters
const (
	apiURL         = "http://localhost:8888/greeting/world" // Replace with your API endpoint
	numRequests    = 1000                                   // Total number of requests to send
	numGoroutines  = 50                                     // Number of concurrent goroutines
	timeoutSeconds = 10                                     // Timeout for each request
)

func main() {
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	// Channel to signal when all requests are done
	var wg sync.WaitGroup

	// Create a channel to distribute requests
	reqChan := make(chan int, numRequests)

	// Populate the request channel
	for i := 0; i < numRequests; i++ {
		reqChan <- i
	}
	close(reqChan)

	successCount := 0

	// Start time for measuring duration
	start := time.Now()

	// Launch goroutines to process requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range reqChan {
				// Send a GET request
				resp, err := client.Get(apiURL)
				if err != nil {
					fmt.Printf("Request failed: %v\n", err)
					continue
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					fmt.Printf("Non-OK HTTP status: %d\n", resp.StatusCode)
				} else {
					successCount++
				}
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Calculate and print the duration
	elapsed := time.Since(start)
	fmt.Printf("Sent %d requests in %s\n", numRequests, elapsed)
	fmt.Printf("Successes: %d\n", successCount)
}
