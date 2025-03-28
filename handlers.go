package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func sendMessage(c *gin.Context) {
	var msg struct {
		Sender   string `json:"sender"`
		Receiver string `json:"receiver"`
		Content  string `json:"content"`
	}

	if err := c.BindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	query := `INSERT into messages (sender, receiver, content) VALUES ($1, $2, $3)`
	_, err := db.Exec(context.Background(), query, msg.Sender, msg.Receiver, msg.Content)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	err = redisClient.Publish(context.Background(), "chat", msg.Content).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

func getChatHistory(c *gin.Context) {
	user1 := c.Query("user1")
	user2 := c.Query("user2")

	query := `SELECT sender, receiver, content, timestamp FROM messages WHERE (sender=$1 AND receiver=$2) OR (sender=$2 AND receiver=$1) ORDER BY timestamp ASC`
	rows, err := db.Query(context.Background(), query, user1, user2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}
	defer rows.Close()

	var messages []map[string]string
	for rows.Next() {
		var sender, receiver, content string
		var timestamp time.Time
		err := rows.Scan(&sender, &receiver, &content, &timestamp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row", "details": err.Error()})
			return
		}

		timestampStr := timestamp.Format("2006-01-02 15:04:05") // Format as YYYY-MM-DD HH:MM:SS

		messages = append(messages, map[string]string{
			"sender":    sender,
			"receiver":  receiver,
			"content":   content,
			"timestamp": timestampStr,
		})
	}

	c.JSON(http.StatusOK, messages)

}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to websocket"})
		return
	}
	defer conn.Close()

	pubsub := redisClient.Subscribe(context.Background(), "chat")
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write message"})
			break
		}
	}

}
