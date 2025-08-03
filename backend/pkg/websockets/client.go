package websockets

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	// Hub reference to the WebSocket hub
	hub *WsHub

	// The actual WebSocket connection
	conn *websocket.Conn

	// Channel to send messages to this client
	send chan []byte

	// Client type (e.g., "sensor", "dashboard")
	clientType string
}

// NewClient creates a new client with the given connection
func (h *WsHub) NewClient(conn *websocket.Conn, clientType string) *Client {
	// TODO: implement
	return &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan []byte, 256), // Buffered channel for sending messages
		clientType: clientType,
	}
}

func (c *Client) handleMessage(message []byte) {
	switch c.clientType {
	case "sensor":
		// Handle sensor data messages
		log.Printf("Sensor message: %s", message)
	case "dashboard":
		// Handle dashboard messages
		log.Printf("Dashboard message: %s", message)
	default:
		// Echo message back for testing
		c.hub.broadcast <- message
	}
}

// WriteMessages continuously writes messages from the send channel to the WebSocket
func (c *Client) WriteMessages() {
	// TODO: implement
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Error writing message:", err)
			break
		}
		log.Println("Sent message:", string(message))
	}
}

// SendMessage sends a message to this client (puts it in the send channel)
func (c *Client) SendMessage(message []byte) {
	select {
	case c.send <- message:
		// Message sent successfully
	default:
		log.Println("Failed to send message, channel is full")
	}
}

// Close closes the client connection and cleans up
func (c *Client) Close() {
	// TODO: implement
	if err := c.conn.Close(); err != nil {
		log.Println("Error closing connection:", err)
	}
	close(c.send) // Close the send channel to stop writing
}
