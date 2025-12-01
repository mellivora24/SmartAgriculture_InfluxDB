package websocket

import (
	"backend/infra/network"
	realtimeShared "backend/internal/realtime/shared"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsService struct {
	hub     *network.WebSocketHub
	clients map[string]*realtimeShared.Client
	mcuMap  map[string][]string
	mu      sync.RWMutex
}

func NewService(hub *network.WebSocketHub) realtimeShared.WebSocketService {
	return &wsService{
		hub:     hub,
		clients: make(map[string]*realtimeShared.Client),
		mcuMap:  make(map[string][]string),
	}
}

func (s *wsService) AddClient(uid, mcuCode string, conn *websocket.Conn) *realtimeShared.Client {
	clientID := fmt.Sprintf("%s_%s_%d", uid, mcuCode, time.Now().UnixNano())

	client := &realtimeShared.Client{
		ID:      clientID,
		UID:     uid,
		MCUCode: mcuCode,
		Conn:    conn,
		Send:    make(chan []byte, 256),
	}

	s.mu.Lock()
	s.clients[clientID] = client
	s.mcuMap[mcuCode] = append(s.mcuMap[mcuCode], clientID)
	s.mu.Unlock()

	log.Printf("[WS Service] Client added: %s (UID: %s, MCU: %s)", clientID, uid, mcuCode)
	return client
}

func (s *wsService) RemoveClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[clientID]
	if !ok {
		return
	}

	mcuClients := s.mcuMap[client.MCUCode]
	for i, id := range mcuClients {
		if id == clientID {
			s.mcuMap[client.MCUCode] = append(mcuClients[:i], mcuClients[i+1:]...)
			break
		}
	}

	if len(s.mcuMap[client.MCUCode]) == 0 {
		delete(s.mcuMap, client.MCUCode)
	}

	close(client.Send)
	delete(s.clients, clientID)

	log.Printf("[WS Service] Client removed: %s", clientID)
}

func (s *wsService) BroadcastToMCU(mcuCode string, message realtimeShared.WSMessage) error {
	message.Timestamp = time.Now()
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	s.mu.RLock()
	clientIDs := s.mcuMap[mcuCode]
	s.mu.RUnlock()

	if len(clientIDs) == 0 {
		log.Printf("[WS Service] No clients connected for MCU: %s", mcuCode)
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, clientID := range clientIDs {
		if client, ok := s.clients[clientID]; ok {
			select {
			case client.Send <- data:
			default:
				log.Printf("[WS Service] Client %s buffer full, skipping", clientID)
			}
		}
	}

	log.Printf("[WS Service] Broadcasted %d messages to MCU: %s", len(clientIDs), mcuCode)
	return nil
}

func (s *wsService) SendToClient(clientID string, message realtimeShared.WSMessage) error {
	message.Timestamp = time.Now()
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	s.mu.RLock()
	client, ok := s.clients[clientID]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("client not found: %s", clientID)
	}

	select {
	case client.Send <- data:
		return nil
	default:
		return fmt.Errorf("client buffer full")
	}
}

func (s *wsService) HandleClientRead(client *realtimeShared.Client, onMessage func(clientID string, msg realtimeShared.WSMessage)) {
	defer func() {
		s.RemoveClient(client.ID)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512 * 1024)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS Service] Read error client %s: %v", client.ID, err)
			}
			break
		}

		var wsMsg realtimeShared.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("[WS Service] Unmarshal error from client %s: %v", client.ID, err)
			continue
		}

		if onMessage != nil {
			onMessage(client.ID, wsMsg)
		}
	}
}

func (s *wsService) HandleClientWrite(client *realtimeShared.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(msg)

			// batch messages if available
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *wsService) GetClientsByMCU(mcuCode string) []*realtimeShared.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clientIDs := s.mcuMap[mcuCode]
	clients := make([]*realtimeShared.Client, 0, len(clientIDs))

	for _, id := range clientIDs {
		if client, ok := s.clients[id]; ok {
			clients = append(clients, client)
		}
	}

	return clients
}
