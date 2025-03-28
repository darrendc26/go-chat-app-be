package main

import (
	"context"

	"github.com/gin-gonic/gin"
)

func main() {
	InitDB()
	defer db.Close(context.Background())
	defer redisClient.Close()

	r := gin.Default()
	r.POST("/send", sendMessage)
	r.GET("/history", getChatHistory)
	r.GET("/ws", handleWebSocket)

	r.Run(":8080")
}
