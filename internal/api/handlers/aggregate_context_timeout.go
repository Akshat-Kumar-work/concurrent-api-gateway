package handlers

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Akshat-Kumar-work/concurrent-api-gateway/pkg/service"
	"github.com/gin-gonic/gin"
)

// Version 3: With Context and Timeout
func AggregateHandlerWithTimeout(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = "123"
	}

	// Set overall timeout for the aggregation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
	defer cancel()

	start := time.Now()
	type result struct {
		service string
		data    any
		err     error
	}

	servicesToCall := map[string]func(string) (any, error){
		"user":          service.FetchUser,
		"orders":        service.FetchOrders,
		"notifications": service.FetchNotifications,
	}

	resultChan := make(chan result, len(servicesToCall))
	var wg sync.WaitGroup

	// Launch goroutines with context
	for name, fetcher := range servicesToCall {
		wg.Add(1)
		go func(svcName string, fn func(string) (any, error)) {
			defer wg.Done()

			// Create a channel for the actual fetch operation
			innerChan := make(chan result, 1)
			go func() {
				data, err := fn(userID)
				innerChan <- result{service: svcName, data: data, err: err}
			}()

			// Wait for either result or context cancellation
			select {
			case res := <-innerChan:
				resultChan <- res
			case <-ctx.Done():
				resultChan <- result{
					service: svcName,
					err:     errors.New("service timeout: " + ctx.Err().Error()),
				}
			}
		}(name, fetcher)
	}

	// Close resultChan when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make(map[string]any)
	errors := make([]string, 0)

	for res := range resultChan {
		if res.err != nil {
			errors = append(errors, res.service+": "+res.err.Error())
		} else {
			results[res.service] = res.data
		}
	}

	c.JSON(200, gin.H{
		"data":        results,
		"errors":      errors,
		"duration_ms": time.Since(start).Milliseconds(),
		"concurrency": "context_with_timeout",
		"timed_out":   ctx.Err() != nil,
	})
}
