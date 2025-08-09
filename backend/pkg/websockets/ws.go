package websockets

import (
	"backend/pkg/utils"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var hub = NewWsHub()
var mutex = &sync.Mutex{}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	client := hub.NewClient(conn)
	go client.WriteMessages() // Start the write goroutine
	hub.register <- client    // Register with hub instead of direct map access
	defer func() {
		hub.unregister <- client // Unregister when function exits
	}()
	mutex.Unlock()

	// Handle WebSocket connection

	for {

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			hub.unregister <- client
			break
		}
		log.Println("Received message:", string(msg))

		hub.handleClientMessage(client, msg)

		//hub.broadcast <- msg
	}
}

func StartWebsocketServer() *WsHub {
	port := utils.GetEnv("PORT", "8080")

	go hub.Run() // Start the hub to handle broadcasting messages

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler)

	go func() {
		log.Println("WebSocket server running on port", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Println("Error starting WebSocket server:", err)
			return
		}
	}()

	time.Sleep(100 * time.Millisecond) // Give server time to start

	return hub
}
