package ws

import (
	"log"
	"net/http"
	"sync"

	"govmonitor-it/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow CORS for WebSocket connection
	},
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan domain.WSMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan domain.WSMessage, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
			log.Println("[WebSocket] Client connected")

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Println("[WebSocket] Client disconnected")

		case msg := <-h.broadcast:
			h.mu.Lock()
			for conn := range h.clients {
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Printf("[WebSocket] Write error: %v", err)
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) BroadcastMessage(msg domain.WSMessage) {
	h.broadcast <- msg
}

func (h *Hub) Broadcast(v interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		err := conn.WriteJSON(v)
		if err != nil {
			log.Printf("[WebSocket] Direct write error: %v", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

func (h *Hub) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade error: %v", err)
		return
	}
	h.register <- conn
	log.Printf("[WebSocket] Client %s upgraded to WebSocket successfully", conn.RemoteAddr().String()) // New log line

	// Read pump goroutine: keeps socket alive and detects client disconnection
	go func() {
		defer func() {
			h.unregister <- conn
			conn.Close() // Ensure connection is closed on exit
		}()
		for {
			// ReadMessage reads a message from the WebSocket connection.
			// It blocks until a message is received or an error occurs.
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[WebSocket] Read error for client %s: %v", conn.RemoteAddr().String(), err)
				}
				break // Exit the loop on error
			}

			// Optionally, handle incoming messages here if needed in the future.
			// For now, we just read them to keep the connection alive.
			log.Printf("[WebSocket] Received message type %d from client %s: %s", messageType, conn.RemoteAddr().String(), message)
		}
	}()
}
