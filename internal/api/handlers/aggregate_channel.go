package handlers

import (
	"time"

	"github.com/Akshat-Kumar-work/concurrent-api-gateway/pkg/service"
	"github.com/gin-gonic/gin"
)

// AggregateChannelHandler aggregates data from multiple services concurrently using channels.
// This version uses channel blocking for synchronization instead of WaitGroup.
// Key concept: Each <-resultChan blocks until data arrives, naturally waiting for all goroutines.
func AggregateChannelHandler(c *gin.Context) {
	// Extract user_id from query parameters, default to "123" if not provided
	userId := c.Query("user_id")
	if userId == "" {
		userId = "123"
	}

	start := time.Now()

	// result struct holds the response from each service call
	type result struct {
		service string // Name of the service (e.g., "user", "orders")
		data    any    // The actual data returned
		err     error  // Any error that occurred
	}

	// Map of service names to their fetch functions
	servicesToCall := map[string]func(string) (any, error){
		"user":          service.FetchUser,
		"orders":        service.FetchOrders,
		"notifications": service.FetchNotifications,
	}

	// Create a buffered channel that can hold len(servicesToCall) results (3 in this case)
	// Buffered channel allows goroutines to send without blocking (until buffer is full)
	// This means all 3 goroutines can start sending immediately
	resultChan := make(chan result, len(servicesToCall))

	// Launch a goroutine for each service to fetch data concurrently
	for name, fetcher := range servicesToCall {
		go func(svcName string, fn func(string) (any, error)) {
			// Fetch data from the service
			data, err := fn(userId)
			// Send result to the channel (non-blocking if buffer has space)
			resultChan <- result{service: svcName, data: data, err: err}
		}(name, fetcher)
	}

	// Initialize maps to collect results and errors
	results := make(map[string]any)
	errors := make([]string, 0)

	// Collect results from all goroutines
	// IMPORTANT: This loop runs exactly len(servicesToCall) times (3 times)
	// Each iteration blocks on <-resultChan until a goroutine sends its result
	// This blocking behavior acts as implicit synchronization - no WaitGroup needed!
	//
	// How it works:
	// 1. First iteration: blocks until first goroutine completes and sends result
	// 2. Second iteration: blocks until second goroutine completes and sends result
	// 3. Third iteration: blocks until third goroutine completes and sends result
	// 4. Loop ends: All 3 goroutines have finished!
	//
	// The blocking receive (<-resultChan) is doing the same job as wg.Wait(),
	// but it's implicit rather than explicit.
	for range servicesToCall {
		// Block here until a goroutine sends a result
		// Results can arrive in any order (fastest service first)
		res := <-resultChan

		if res.err != nil {
			// Store error with service name for debugging
			errors = append(errors, res.service)
		} else {
			// Store successful result using the service name from the result struct
			// (not from the loop variable, since results may arrive out of order)
			results[res.service] = res.data
		}
	}

	// Return aggregated results as JSON
	c.JSON(200, gin.H{
		"data":        results,
		"errors":      errors,
		"duration_ms": time.Since(start).Milliseconds(),
		"concurrency": "channels",
	})

}
