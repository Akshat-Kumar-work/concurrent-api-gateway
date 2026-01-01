package main

import (
	handlers "github.com/Akshat-Kumar-work/concurrent-api-gateway/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/health", func(ctx *gin.Context) {
		m := map[string]string{
			"status": "ok",
		}
		//ctx.JSON(200, gin.H{"status": "ok"})
		ctx.JSON(200, m)

	})

	router.GET("/api/aggregate", handlers.AggregateHandler)

	router.Run(":8080")
}
