package shared

import (
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

type MQTTService interface {
	Subscribe(topic string, handler mqtt.MessageHandler) error
	Publish(topic string, payload interface{}) error
	PublishJSON(topic string, data interface{}) error
	IsConnected() bool
}

type Client struct {
	ID      string
	UID     string
	MCUCode string
	Conn    *websocket.Conn
	Send    chan []byte
	Mu      sync.RWMutex
}

type WebSocketService interface {
	AddClient(uid string, mcuCode string, conn *websocket.Conn) *Client
	HandleClientRead(client *Client, onMessage func(clientID string, msg WSMessage))
	HandleClientWrite(client *Client)
	BroadcastToMCU(mcuCode string, message WSMessage) error
	SendToClient(clientID string, message WSMessage) error
}
