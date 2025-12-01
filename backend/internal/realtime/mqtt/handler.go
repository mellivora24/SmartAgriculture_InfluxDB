package mqtt

import (
	"backend/internal/feature/sensorData"
	"backend/internal/realtime/shared"
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

type Handler struct {
	mqttService   shared.MQTTService
	wsService     shared.WebSocketService
	sensorService sensorData.Service
}

func NewHandler(mqttService shared.MQTTService, wsService shared.WebSocketService, sensorService sensorData.Service) *Handler {
	return &Handler{
		mqttService:   mqttService,
		wsService:     wsService,
		sensorService: sensorService,
	}
}

// Init subscribes to all required MQTT topics
func (h *Handler) Init() error {
	topics := map[string]mqtt.MessageHandler{
		shared.MQTTTopicHealthRequest:   h.onHealthCheck,
		shared.MQTTTopicSensorData:      h.onSensorData,
		shared.MQTTTopicControlResponse: h.onControlResponse,
		shared.MQTTTopicAlert:           h.onAlert,
	}

	for topic, handler := range topics {
		if err := h.mqttService.Subscribe(topic, handler); err != nil {
			return err
		}
	}

	log.Println("[MQTT Handler] Subscribed to all topics successfully")
	return nil
}

// onHealthCheck handles health check requests
func (h *Handler) onHealthCheck(client mqtt.Client, msg mqtt.Message) {
	if err := h.mqttService.Publish(shared.MQTTTopicHealthResponse, "ok"); err != nil {
		log.Printf("[MQTT Handler] Error publishing health response: %v", err)
	}
}

// onSensorData handles incoming sensor data from MCU
func (h *Handler) onSensorData(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg shared.MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling sensor data: %v", err)
		return
	}

	// Parse sensor data payload
	payloadBytes, err := json.Marshal(mqttMsg.Payload)
	if err != nil {
		log.Printf("[MQTT Handler] Error marshaling payload: %v", err)
		return
	}

	var sensorPayload shared.SensorDataPayload
	if err := json.Unmarshal(payloadBytes, &sensorPayload); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling sensor payload: %v", err)
		return
	}

	// Save sensor data to InfluxDB
	ctx := context.Background()
	sensorRecord := &sensorData.SensorData{
		SurveyPointID: sensorPayload.SurveyPointID,
		Temperature:   *sensorPayload.Temperature,
		Humidity:      *sensorPayload.Humidity,
		SoilMoisture:  *sensorPayload.SoilMoisture,
		Light:         *sensorPayload.Light,
		Timestamp:     time.Now(),
	}

	if err := h.sensorService.WriteSensorData(ctx, sensorRecord); err != nil {
		log.Printf("[MQTT Handler] Error saving sensor data: %v", err)
		// Continue to broadcast even if save fails
	}

	// Broadcast to WebSocket clients with the same MCU
	wsMsg := shared.WSMessage{
		Topic:   shared.WSTopicSensorData,
		Payload: payloadBytes,
	}

	if err := h.wsService.BroadcastToMCU(sensorPayload.MCUCode, wsMsg); err != nil {
		log.Printf("[MQTT Handler] Error broadcasting sensor data: %v", err)
	}

	log.Printf("[MQTT Handler] Processed sensor data for MCU: %s", sensorPayload.MCUCode)
}

// onControlResponse handles control response from MCU
func (h *Handler) onControlResponse(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg shared.MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling control response: %v", err)
		return
	}

	// Parse control response payload
	payloadBytes, err := json.Marshal(mqttMsg.Payload)
	if err != nil {
		log.Printf("[MQTT Handler] Error marshaling payload: %v", err)
		return
	}

	var controlPayload shared.ControlResponsePayload
	if err := json.Unmarshal(payloadBytes, &controlPayload); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling control payload: %v", err)
		return
	}

	// Get pending commands for this device
	ctx := context.Background()
	commands, err := h.sensorService.GetCommandHistory(ctx, nil, &controlPayload.DeviceName, 10)
	if err != nil {
		log.Printf("[MQTT Handler] Error getting command history: %v", err)
		return
	}

	// Find the most recent pending command
	var commandID uuid.UUID
	for _, cmd := range commands {
		if cmd.Status == "pending" {
			commandID = cmd.CommandID
			break
		}
	}

	if commandID != uuid.Nil {
		// Update command status
		status := "success"
		if controlPayload.Status == "failed" {
			status = "failed"
		}

		if _, err := h.sensorService.UpdateCommandStatus(ctx, commandID, status); err != nil {
			log.Printf("[MQTT Handler] Error updating command status: %v", err)
		} else {
			log.Printf("[MQTT Handler] Updated command %s to status: %s", commandID, status)
		}
	}

	// Broadcast response to WebSocket clients
	wsMsg := shared.WSMessage{
		Topic:   shared.WSTopicControlResponse,
		Payload: payloadBytes,
	}

	if err := h.wsService.BroadcastToMCU(controlPayload.MCUCode, wsMsg); err != nil {
		log.Printf("[MQTT Handler] Error broadcasting control response: %v", err)
	}

	log.Printf("[MQTT Handler] Processed control response for MCU: %s, Device: %s, Status: %s",
		controlPayload.MCUCode, controlPayload.DeviceName, controlPayload.Status)
}

// onAlert handles alert notifications from MCU
func (h *Handler) onAlert(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg shared.MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling alert: %v", err)
		return
	}

	// Parse alert payload
	payloadBytes, err := json.Marshal(mqttMsg.Payload)
	if err != nil {
		log.Printf("[MQTT Handler] Error marshaling payload: %v", err)
		return
	}

	var alertPayload shared.MQTTAlert
	if err := json.Unmarshal(payloadBytes, &alertPayload); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling alert payload: %v", err)
		return
	}

	// Broadcast alert to WebSocket clients
	wsMsg := shared.WSMessage{
		Topic:   shared.WSTopicAlert,
		Payload: payloadBytes,
	}

	if err := h.wsService.BroadcastToMCU(alertPayload.MCUCode, wsMsg); err != nil {
		log.Printf("[MQTT Handler] Error broadcasting alert: %v", err)
	}

	log.Printf("[MQTT Handler] Processed alert for MCU: %s, Severity: %s",
		alertPayload.MCUCode, alertPayload.Severity)
}

// Helper functions
func extractFromTopic(topic string, position int) string {
	parts := strings.Split(topic, "/")
	if len(parts) > position {
		return parts[position]
	}
	return ""
}

func extractUIDFromTopic(topic string) string {
	return extractFromTopic(topic, 1)
}

func extractMCUCodeFromTopic(topic string) string {
	return extractFromTopic(topic, 3)
}
