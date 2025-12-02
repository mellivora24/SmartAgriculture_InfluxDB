package mqtt

import (
	"backend/internal/feature/sensorData"
	"backend/internal/realtime/shared"
	"context"
	"encoding/json"
	"log"
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

	// Validate sensor data
	if sensorPayload.SurveyPointID == uuid.Nil {
		log.Printf("[MQTT Handler] Invalid survey_point_id in sensor data")
		return
	}

	// Save sensor data to InfluxDB
	ctx := context.Background()
	sensorRecord := &sensorData.SensorData{
		SurveyPointID: sensorPayload.SurveyPointID,
		MCUCode:       sensorPayload.MCUCode,
		Temperature:   getFloatValue(sensorPayload.Temperature),
		Humidity:      getFloatValue(sensorPayload.Humidity),
		SoilMoisture:  getFloatValue(sensorPayload.SoilMoisture),
		Light:         getFloatValue(sensorPayload.Light),
		Timestamp:     time.Now(),
	}

	if err := h.sensorService.WriteSensorData(ctx, sensorRecord); err != nil {
		log.Printf("[MQTT Handler] Error saving sensor data: %v", err)
		// Continue to broadcast even if save fails
	}

	// Broadcast to WebSocket clients with the same MCU
	wsMsg := shared.WSMessage{
		Topic:     shared.WSTopicSensorData,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	if err := h.wsService.BroadcastToMCU(sensorPayload.MCUCode, wsMsg); err != nil {
		log.Printf("[MQTT Handler] Error broadcasting sensor data: %v", err)
	}

	log.Printf("[MQTT Handler] Processed sensor data for MCU: %s, SurveyPoint: %s",
		sensorPayload.MCUCode, sensorPayload.SurveyPointID)
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

	// Validate survey_point_id
	if controlPayload.SurveyPointID == uuid.Nil {
		log.Printf("[MQTT Handler] Invalid survey_point_id in control response")
		return
	}

	ctx := context.Background()

	// Get pending commands for this survey point and device
	commands, err := h.sensorService.GetCommandHistory(ctx, &controlPayload.SurveyPointID, &controlPayload.DeviceName, 10)
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
		// Map response status to database status
		status := "success"
		if controlPayload.Status == "failed" || controlPayload.Status == "error" {
			status = "failed"
		}

		// Update command status
		if _, err := h.sensorService.UpdateCommandStatus(ctx, commandID, status); err != nil {
			log.Printf("[MQTT Handler] Error updating command status: %v", err)
		} else {
			log.Printf("[MQTT Handler] Updated command %s to status: %s", commandID, status)
		}
	} else {
		log.Printf("[MQTT Handler] No pending command found for SurveyPoint: %s, Device: %s",
			controlPayload.SurveyPointID, controlPayload.DeviceName)
	}

	// Broadcast response to WebSocket clients
	wsMsg := shared.WSMessage{
		Topic:     shared.WSTopicControlResponse,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	if err := h.wsService.BroadcastToMCU(controlPayload.MCUCode, wsMsg); err != nil {
		log.Printf("[MQTT Handler] Error broadcasting control response: %v", err)
	}

	log.Printf("[MQTT Handler] Processed control response for MCU: %s, SurveyPoint: %s, Device: %s, Status: %s",
		controlPayload.MCUCode, controlPayload.SurveyPointID, controlPayload.DeviceName, controlPayload.Status)
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
		Topic:     shared.WSTopicAlert,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	if err := h.wsService.BroadcastToMCU(alertPayload.MCUCode, wsMsg); err != nil {
		log.Printf("[MQTT Handler] Error broadcasting alert: %v", err)
	}

	log.Printf("[MQTT Handler] Processed alert for MCU: %s, Severity: %s",
		alertPayload.MCUCode, alertPayload.Severity)
}

// Helper function to safely get float value from pointer
func getFloatValue(val *float64) float64 {
	if val == nil {
		return 0.0
	}
	return *val
}
