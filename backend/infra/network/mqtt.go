package network

import (
	"backend/internal/shared"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	Client mqtt.Client
	Config *shared.MQTTConfig
}

func NewMQTTClient(config *shared.MQTTConfig) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(config.GetBrokerURL())
	opts.SetClientID(config.ClientID)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetKeepAlive(time.Duration(config.KeepAlive) * time.Second)
	opts.SetCleanSession(config.CleanSession)
	opts.SetAutoReconnect(config.AutoReconnect)
	opts.SetMaxReconnectInterval(config.MaxReconnectInterval)
	opts.SetConnectTimeout(config.ConnectTimeout)
	opts.SetDefaultPublishHandler(defaultMessageHandler)
	opts.SetOnConnectHandler(onConnectHandler)
	opts.SetConnectionLostHandler(connectionLostHandler)
	opts.SetReconnectingHandler(reconnectingHandler)

	client := mqtt.NewClient(opts)

	token := client.Connect()
	if !token.WaitTimeout(config.ConnectTimeout) {
		return nil, fmt.Errorf("MQTT connection timeout")
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	log.Printf("MQTT connected successfully to %s", config.GetBrokerURL())

	return &MQTTClient{
		Client: client,
		Config: config,
	}, nil
}

func defaultMessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("[MQTT] Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
}

func onConnectHandler(client mqtt.Client) {
	log.Println("MQTT client connected")
}

func connectionLostHandler(client mqtt.Client, err error) {
	log.Printf("MQTT connection lost: %v", err)
}

func reconnectingHandler(client mqtt.Client, opts *mqtt.ClientOptions) {
	log.Println("MQTT client reconnecting...")
}

func (m *MQTTClient) Subscribe(topic string, handler mqtt.MessageHandler) error {
	if handler == nil {
		handler = defaultMessageHandler
	}

	token := m.Client.Subscribe(topic, m.Config.QoS, handler)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("subscription timeout for topic: %s", topic)
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

func (m *MQTTClient) SubscribeMultiple(filters map[string]byte, handler mqtt.MessageHandler) error {
	if handler == nil {
		handler = defaultMessageHandler
	}

	token := m.Client.SubscribeMultiple(filters, handler)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("subscription timeout")
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	log.Printf("Subscribed to %d topics", len(filters))
	return nil
}

func (m *MQTTClient) Unsubscribe(topics ...string) error {
	token := m.Client.Unsubscribe(topics...)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("unsubscribe timeout")
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	log.Printf("Unsubscribed from topics: %v", topics)
	return nil
}

func (m *MQTTClient) Publish(topic string, payload interface{}, retained bool) error {
	token := m.Client.Publish(topic, m.Config.QoS, retained, payload)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("publish timeout for topic: %s", topic)
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}

	return nil
}

func (m *MQTTClient) PublishAsync(topic string, payload interface{}, retained bool, callback func(mqtt.Token)) {
	token := m.Client.Publish(topic, m.Config.QoS, retained, payload)
	if callback != nil {
		go func() {
			token.Wait()
			callback(token)
		}()
	}
}

func (m *MQTTClient) Disconnect(quiesce uint) {
	m.Client.Disconnect(quiesce)
	log.Println("MQTT client disconnected")
}

func (m *MQTTClient) IsConnected() bool {
	return m.Client.IsConnected()
}

func (m *MQTTClient) HealthCheck() error {
	if !m.Client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}
	return nil
}

func NewMQTTClientWithTLS(config *shared.MQTTConfig, tlsConfig *tls.Config) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.GetBrokerURL())
	opts.SetClientID(config.ClientID)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetKeepAlive(time.Duration(config.KeepAlive) * time.Second)
	opts.SetCleanSession(config.CleanSession)
	opts.SetAutoReconnect(config.AutoReconnect)
	opts.SetMaxReconnectInterval(config.MaxReconnectInterval)
	opts.SetConnectTimeout(config.ConnectTimeout)
	opts.SetTLSConfig(tlsConfig)

	opts.SetDefaultPublishHandler(defaultMessageHandler)
	opts.SetOnConnectHandler(onConnectHandler)
	opts.SetConnectionLostHandler(connectionLostHandler)
	opts.SetReconnectingHandler(reconnectingHandler)

	client := mqtt.NewClient(opts)
	token := client.Connect()

	if !token.WaitTimeout(config.ConnectTimeout) {
		return nil, fmt.Errorf("MQTT TLS connection timeout")
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker with TLS: %w", err)
	}

	log.Printf("MQTT (TLS) connected successfully to %s", config.GetBrokerURL())

	return &MQTTClient{
		Client: client,
		Config: config,
	}, nil
}
