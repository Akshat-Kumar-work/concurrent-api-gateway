// mock_services/main.go
package main

import (
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	rand.Seed(time.Now().UnixNano())

	// Mock service 1: User Service
	r.GET("/mock/user/:id", func(c *gin.Context) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // Random delay
		c.JSON(200, gin.H{
			"service":   "user",
			"id":        c.Param("id"),
			"name":      "John Doe",
			"email":     "john@example.com",
			"timestamp": time.Now().Unix(),
		})
	})

	// Mock service 2: Order Service
	r.GET("/mock/orders/:userId", func(c *gin.Context) {
		time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)
		c.JSON(200, gin.H{
			"service": "orders",
			"userId":  c.Param("userId"),
			"orders": []gin.H{
				{"id": "ORD001", "total": 99.99},
				{"id": "ORD002", "total": 149.99},
			},
			"timestamp": time.Now().Unix(),
		})
	})

	// Mock service 3: Notification Service
	r.GET("/mock/notifications/:userId", func(c *gin.Context) {
		time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
		c.JSON(200, gin.H{
			"service":   "notifications",
			"userId":    c.Param("userId"),
			"unread":    3,
			"messages":  []string{"Welcome back!", "Order shipped", "New feature available"},
			"timestamp": time.Now().Unix(),
		})
	})

	// Mock service 4: Inventory Service (for later)
	r.GET("/mock/inventory/:productId", func(c *gin.Context) {
		time.Sleep(time.Duration(rand.Intn(80)) * time.Millisecond)
		c.JSON(200, gin.H{
			"service":   "inventory",
			"productId": c.Param("productId"),
			"stock":     rand.Intn(100),
			"price":     49.99,
			"timestamp": time.Now().Unix(),
		})
	})

	println("Mock services running on :9090")
	r.Run(":9090")
}
