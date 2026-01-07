package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// Hub maintains the set of active clients for a specific session.
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
		case message := <-h.broadcast:
			for conn := range h.clients {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("WebSocket write error: %v", err)
				}
			}
		}
	}
}

// Manager manages all the hubs.
type Manager struct {
	hubs   map[string]*Hub
	hubsMu sync.Mutex
}

// NewManager creates a new WebSocket manager.
func NewManager() *Manager {
	return &Manager{
		hubs: make(map[string]*Hub),
	}
}

// getOrCreateHub finds an existing hub for a session or creates a new one.
func (m *Manager) getOrCreateHub(sessionID string) *Hub {
	m.hubsMu.Lock()
	defer m.hubsMu.Unlock()

	if hub, exists := m.hubs[sessionID]; exists {
		return hub
	}

	hub := newHub()
	m.hubs[sessionID] = hub
	return hub
}

// Run starts the goroutine for each hub.
func (m *Manager) Run() {
	// This is a placeholder for more complex logic if needed.
	// In this design, hubs are started on-demand in getOrCreateHub.
}

// HandleConnection upgrades an HTTP request to a WebSocket connection.
func (m *Manager) HandleConnection(w http.ResponseWriter, r *http.Request, sessionID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	hub := m.getOrCreateHub(sessionID)
	// Start the hub's processing goroutine if it's the first connection
	if len(hub.clients) == 0 {
		go hub.run()
	}
	hub.register <- conn

	defer func() {
		hub.unregister <- conn
	}()

	// Keep the connection alive by reading messages.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			// Client has disconnected.
			break
		}
	}
}

// Broadcast sends a message to all clients in a specific session hub.
func (m *Manager) Broadcast(sessionID string, message []byte) {
	m.hubsMu.Lock()
	hub, exists := m.hubs[sessionID]
	m.hubsMu.Unlock()

	if exists {
		hub.broadcast <- message
	}
}
