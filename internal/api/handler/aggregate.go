package handlers

import (
	"sync"
	"time"

	"github.com/Akshat-Kumar-work/concurrent-api-gateway/pkg/service"

	"github.com/gin-gonic/gin"
)

// Version 1: Basic WaitGroup
func AggregateHandler(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = "123"
	}

	start := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex // To safely write to the map
	results := make(map[string]interface{})
	errors := make([]string, 0)

	// Define service to fetch
	servicesToCall := []struct {
		name string
		call func(string) (interface{}, error)
	}{
		{"user", service.FetchUser},
		{"orders", service.FetchOrders},
		{"notifications", service.FetchNotifications},
	}

	// Launch goroutines for each service
	for _, svc := range servicesToCall {
		wg.Add(1)
		go func(name string, fetcher func(string) (interface{}, error)) {
			defer wg.Done()

			data, err := fetcher(userID)
			mu.Lock()
			if err != nil {
				errors = append(errors, name+": "+err.Error())
			} else {
				results[name] = data
			}
			mu.Unlock()
		}(svc.name, svc.call)
	}

	wg.Wait() // Wait for all goroutines

	c.JSON(200, gin.H{
		"success":     len(errors) == 0,
		"data":        results,
		"errors":      errors,
		"duration_ms": time.Since(start).Milliseconds(),
		"concurrency": "waitgroup",
	})
}
