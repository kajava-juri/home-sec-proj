package websockets

import (
	"log"
)

type WsHub struct {
	// Register is a channel for clients to register with the hub
	register chan *Client
	// Unregister is a channel for clients to unregister from the hub
	unregister chan *Client
	// Clients is a set of all connected clients
	clients map[*Client]bool
	// Broadcast is a channel for broadcasting messages to all clients
	broadcast chan []byte
}

func NewWsHub() *WsHub {
	return &WsHub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256), // Buffered channel for broadcasting messages
	}
}

func (h *WsHub) BroadcastMessage(message []byte) {
	h.broadcast <- message
}

func (h *WsHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send) // Close the send channel to stop writing
			}
		// Grab the next message from the broadcast channel
		case message := <-h.broadcast:
			log.Println("Broadcasting message:", string(message))

			// Send the message to all connected clients
			for client := range h.clients {
				client.SendMessage(message)
			}
		}
	}
}
