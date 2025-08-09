package websockets

import (
	"encoding/json"
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
	// Topic subscriptions
	topicSubscriptions map[string]map[*Client]bool
}

func NewWsHub() *WsHub {
	return &WsHub{
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		clients:            make(map[*Client]bool),
		broadcast:          make(chan []byte, 256), // Buffered channel for broadcasting messages
		topicSubscriptions: make(map[string]map[*Client]bool),
	}
}

func (h *WsHub) BroadcastMessage(message []byte) {
	h.broadcast <- message
}

func (h *WsHub) BroadcastToTopic(message []byte, topic string) {
	if clients, ok := h.topicSubscriptions[topic]; ok {
		for client := range clients {
			client.SendMessage(message)
		}
	}
}

func (h *WsHub) SubscribeClientToTopics(client *Client, topics []string) {
	for _, topic := range topics {
		if _, exists := h.topicSubscriptions[topic]; !exists {
			h.topicSubscriptions[topic] = make(map[*Client]bool)
		}
		h.topicSubscriptions[topic][client] = true
		client.subscribedTopics[topic] = true
	}
}

func (h *WsHub) UnsubscribeClientFromTopics(client *Client, topics []string) {
	for _, topic := range topics {
		if clients, exists := h.topicSubscriptions[topic]; exists {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.topicSubscriptions, topic)
			}
		}
		delete(h.topicSubscriptions[topic], client)
		delete(client.subscribedTopics, topic)
	}
}

func (h *WsHub) handleClientMessage(client *Client, message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("Error unmarshalling message:", err)
		return
	}

	switch msg.Action {
	case "subscribe":
		h.SubscribeClientToTopics(client, msg.Topics)
	case "unsubscribe":
		h.UnsubscribeClientFromTopics(client, msg.Topics)
	}
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
