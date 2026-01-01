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

	// create buffered channel to collect results
	// in buffered channel, send only blocks main goroutine if buffer is full
	// in buffered channel, receive only blocks main goroutine if buffer is empty
	resultChan := make(chan result, len(servicesToCall))
	var wg sync.WaitGroup
	// WaitGroup uses an internal atomic counter:
	// - wg.Add(1) increments the counter
	// - wg.Done() decrements the counter (equivalent to wg.Add(-1))
	// - wg.Wait() blocks until counter reaches 0

	// Launch goroutines with context
	for name, fetcher := range servicesToCall {
		wg.Add(1) // Increment counter: +1 (now counter = 1, 2, 3 as we loop)
		go func(svcName string, fn func(string) (any, error)) {
			defer wg.Done() // Decrement counter when goroutine exits: -1

			// Create a channel for the actual fetch operation
			// We need innerChan because select can only wait on receives, not sends
			// This allows us to race between the fetch completing and the timeout
			innerChan := make(chan result, 1)
			go func() {
				data, err := fn(userID)
				// Only send if channel is still open (non-blocking check)
				select {
				case innerChan <- result{service: svcName, data: data, err: err}:
				case <-ctx.Done():
					//“Wait for the context to be cancelled (channel closed). Once it’s closed, proceed with this case.”
					// Context cancelled, don't send (timeout already handled)
				}
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
	// This goroutine runs ONCE per request (not continuously):
	//
	// How wg.Wait() knows all are done:
	// - WaitGroup maintains an internal atomic counter
	// - We called wg.Add(1) 3 times (counter = 3)
	// - Each worker calls wg.Done() when finished (counter decrements: 3→2→1→0)
	// - wg.Wait() blocks until counter reaches 0
	// - When counter = 0, wg.Wait() unblocks (all workers have called Done())
	//
	// Execution flow:
	// 1. Blocks on wg.Wait() until counter reaches 0 (all 3 workers called wg.Done())
	// 2. Once counter = 0, wg.Wait() unblocks
	// 3. Closes the channel to signal completion
	// 4. Goroutine exits
	// The close() signals the main goroutine's range loop to exit
	go func() {
		wg.Wait()         // Blocks until internal counter reaches 0
		close(resultChan) // Close channel to signal completion
		// Goroutine exits here
	}()

	// Collect results
	// IMPORTANT: This runs in the MAIN goroutine (same thread as the HTTP handler)
	//
	// Execution Timeline:
	// 1. Main goroutine launches 3 worker goroutines (lines 45-75)
	// 2. Main goroutine launches cleanup goroutine (line 93) - waits for wg.Wait()
	// 3. Main goroutine IMMEDIATELY starts reading here (line 106) - does NOT wait for workers!
	// 4. Main goroutine blocks on first <-resultChan (waiting for first result)
	// 5. Workers send results as they complete (concurrently)
	// 6. Main goroutine reads results as they arrive (concurrently with workers)
	// 7. When all workers finish → cleanup goroutine closes channel
	// 8. Main goroutine's range loop exits when channel closes
	//
	// Key Points:
	// - resultChan is SHARED between: worker goroutines (send), main goroutine (receive), cleanup goroutine (close)
	// - Reading happens CONCURRENTLY with workers sending (not sequentially after)
	// - Channel is buffered (size=3), so workers can send without blocking
	// - Range loop blocks on each read until data arrives or channel closes
	// - When channel closes, range loop automatically exits (even if not all results read)
	results := make(map[string]any)
	errors := make([]string, 0)

	// This range loop runs in the MAIN goroutine
	// It blocks on each iteration until:
	// - A result arrives from a worker (reads it)
	// - OR channel is closed (loop exits)
	// - read one by one reading is blocking
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
