package network

import (
	"backend/internal/shared"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *WebSocketHub
	UserID   string
	Metadata map[string]interface{}
}

type WebSocketHub struct {
	Clients    map[string]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
	Config     *shared.WebSocketConfig
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebSocketHub(config *shared.WebSocketConfig) *WebSocketHub {
	upgrader.ReadBufferSize = config.ReadBufferSize
	upgrader.WriteBufferSize = config.WriteBufferSize

	hub := &WebSocketHub{
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Config:     config,
	}

	go hub.run()
	log.Println("WebSocket hub started")

	return hub
}

func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s (Total: %d)", client.ID, len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
				log.Printf("Client unregistered: %s (Total: %d)", client.ID, len(h.Clients))
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					go func(c *Client) {
						h.Unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *WebSocketHub) HandleWebSocket(c *gin.Context, clientID string) error {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(h.Config.PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(h.Config.PongWait))
		return nil
	})

	client := &Client{
		ID:       clientID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h,
		Metadata: make(map[string]interface{}),
	}

	h.Register <- client

	go client.readPump()
	go client.writePump()

	return nil
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.Hub.Config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.Hub.Config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.Hub.Config.PongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.Hub.Config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Hub.Config.WriteTimeout))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Hub.Config.WriteTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		log.Println("Message type not found")
		return
	}

	switch msgType {
	case "ping":
		c.SendJSON(map[string]interface{}{
			"type": "pong",
			"time": time.Now().Unix(),
		})
	case "subscribe":
		log.Printf("Client %s subscribed", c.ID)
	case "unsubscribe":
		log.Printf("Client %s unsubscribed", c.ID)
	default:
		log.Printf("Unknown message type: %s", msgType)
	}
}

func (c *Client) SendJSON(data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return err
	}

	select {
	case c.Send <- message:
		return nil
	default:
		return websocket.ErrCloseSent
	}
}

func (h *WebSocketHub) BroadcastJSON(data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return err
	}

	h.Broadcast <- message
	return nil
}

func (h *WebSocketHub) SendToClient(clientID string, data interface{}) error {
	h.mu.RLock()
	client, ok := h.Clients[clientID]
	h.mu.RUnlock()

	if !ok {
		return websocket.ErrCloseSent
	}

	return client.SendJSON(data)
}

func (h *WebSocketHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

func (h *WebSocketHub) GetClients() map[string]*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make(map[string]*Client)
	for k, v := range h.Clients {
		clients[k] = v
	}
	return clients
}

func (h *WebSocketHub) DisconnectClient(clientID string) error {
	h.mu.RLock()
	client, ok := h.Clients[clientID]
	h.mu.RUnlock()

	if !ok {
		return websocket.ErrCloseSent
	}

	h.Unregister <- client
	return nil
}

func (h *WebSocketHub) HealthCheck() error {
	if h.Clients == nil {
		return fmt.Errorf("WebSocket hub not initialized")
	}
	return nil
}
