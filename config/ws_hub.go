package config

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients map[uint]map[*websocket.Conn]bool // userID → connections
	mu      sync.Mutex
}

var WSHub = &Hub{
	clients: make(map[uint]map[*websocket.Conn]bool),
}

// Register a new client under a userID
func (h *Hub) AddClient(userID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]bool)
	}
	h.clients[userID][conn] = true
	fmt.Printf("Client connected for user %d, total: %d\n", userID, len(h.clients[userID]))
}

// Remove disconnected client
func (h *Hub) RemoveClient(userID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[userID], conn)
	fmt.Printf("Client removed for user %d, total: %d\n", userID, len(h.clients[userID]))
}

// Send message only to one user
func (h *Hub) SendToUser(userID uint, msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients[userID] {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			conn.Close()
			delete(h.clients[userID], conn)
		}
	}
}
