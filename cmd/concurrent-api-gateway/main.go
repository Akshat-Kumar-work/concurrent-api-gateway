package main

import (
	handlers "github.com/Akshat-Kumar-work/concurrent-api-gateway/internal/api/handlers"
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

	router.GET("/api/aggregate/wg", handlers.AggregateHandler)

	router.GET("/api/aggregate/channel", handlers.AggregateChannelHandler)

	router.GET("/api/aggregate/channel-with-context-timeout", handlers.AggregateHandlerWithTimeout)

	router.Run(":8080")
}
