package mqtt

import (
	"backend/infra/network"
	"backend/internal/realtime/shared"
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type mqttService struct {
	client *network.MQTTClient
}

func NewService(client *network.MQTTClient) shared.MQTTService {
	return &mqttService{
		client: client,
	}
}

func (s *mqttService) Subscribe(topic string, handler mqtt.MessageHandler) error {
	if err := s.client.Subscribe(topic, handler); err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}
	log.Printf("[MQTT Service] Subscribed to topic: %s", topic)
	return nil
}

func (s *mqttService) Publish(topic string, payload interface{}) error {
	if err := s.client.Publish(topic, payload, false); err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}
	return nil
}

func (s *mqttService) PublishJSON(topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := s.client.Publish(topic, payload, false); err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}

	log.Printf("[MQTT Service] Published to topic: %s", topic)
	return nil
}

func (s *mqttService) IsConnected() bool {
	return s.client.IsConnected()
}
