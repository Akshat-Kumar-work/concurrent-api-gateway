package handlers

import (
	"time"

	"github.com/Akshat-Kumar-work/concurrent-api-gateway/pkg/service"
	"github.com/gin-gonic/gin"
)

func AggregateChannelHandler(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		userId = "123"
	}

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

	for name, fetcher := range servicesToCall {
		go func(svcName string, fn func(string) (any, error)) {
			data, err := fn(userId)
			resultChan <- result{service: svcName, data: data, err: err}
		}(name, fetcher)
	}

	// Collect results
	results := make(map[string]any)
	errors := make([]string, 0)

	for service := range servicesToCall {
		res := <-resultChan
		if res.err != nil {
			errors = append(errors, res.service)
		} else {
			results[service] = res.data
		}
	}

	c.JSON(200, gin.H{
		"data":        results,
		"errors":      errors,
		"duration_ms": time.Since(start).Milliseconds(),
		"concurrency": "channels",
	})

}
