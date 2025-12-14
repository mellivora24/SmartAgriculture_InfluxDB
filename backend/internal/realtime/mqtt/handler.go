package mqtt

import (
	"backend/internal/feature/sensorData"
	"backend/internal/feature/surveyPoint"
	"backend/internal/feature/threshold"
	"backend/internal/realtime/shared"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

type Handler struct {
	mqttService        shared.MQTTService
	wsService          shared.WebSocketService
	sensorService      sensorData.Service
	thresholdService   threshold.Service
	surveyPointService surveyPoint.Service
}

func NewHandler(
	mqttService shared.MQTTService,
	wsService shared.WebSocketService,
	sensorService sensorData.Service,
	thresholdService threshold.Service,
	surveyPointService surveyPoint.Service,
) *Handler {
	return &Handler{
		mqttService:        mqttService,
		wsService:          wsService,
		sensorService:      sensorService,
		thresholdService:   thresholdService,
		surveyPointService: surveyPointService,
	}
}

func (h *Handler) Init() error {
	topics := map[string]mqtt.MessageHandler{
		shared.MQTTTopicHealthRequest:    h.onHealthCheck,
		shared.MQTTTopicSensorData:       h.onSensorData,
		shared.MQTTTopicControlResponse:  h.onControlResponse,
		shared.MQTTTopicAlert:            h.onAlert,
		"user/+/mcu/+/disease_detection": h.onDiseaseDetection,
	}

	for topic, handler := range topics {
		if err := h.mqttService.Subscribe(topic, handler); err != nil {
			return err
		}
	}

	log.Println("[MQTT Handler] Subscribed to all topics successfully")
	return nil
}

func (h *Handler) onHealthCheck(client mqtt.Client, msg mqtt.Message) {
	if err := h.mqttService.Publish(shared.MQTTTopicHealthResponse, "ok"); err != nil {
		log.Printf("[MQTT Handler] Error publishing health response: %v", err)
	}
}

func (h *Handler) onSensorData(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg shared.MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling sensor data: %v", err)
		return
	}

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

	if sensorPayload.SurveyPointID == uuid.Nil {
		log.Printf("[MQTT Handler] Invalid survey_point_id in sensor data")
		return
	}

	ctx := context.Background()

	// Save sensor data to InfluxDB
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
	}

	// Check thresholds and send alerts
	alerts, err := h.thresholdService.CheckSensorThresholds(
		ctx,
		sensorPayload.SurveyPointID,
		sensorPayload.Temperature,
		sensorPayload.Humidity,
		sensorPayload.SoilMoisture,
		sensorPayload.Light,
	)

	if err != nil {
		log.Printf("[MQTT Handler] Error checking thresholds: %v", err)
	}

	// Send alerts via WebSocket and record them
	for _, alert := range alerts {
		if err := h.thresholdService.RecordAlert(ctx, sensorPayload.SurveyPointID, alert); err != nil {
			log.Printf("[MQTT Handler] Error recording alert: %v", err)
		}

		alertPayload := shared.MQTTAlert{
			MCUCode:  sensorPayload.MCUCode,
			Title:    fmt.Sprintf("Cảnh báo: %s", alert.AlertType),
			Message:  alert.Message,
			Severity: alert.Severity,
			Time:     time.Now(),
		}

		alertBytes, _ := json.Marshal(alertPayload)
		wsMsg := shared.WSMessage{
			Topic:     shared.WSTopicAlert,
			Payload:   alertBytes,
			Timestamp: time.Now(),
		}

		if err := h.wsService.BroadcastToMCU(sensorPayload.MCUCode, wsMsg); err != nil {
			log.Printf("[MQTT Handler] Error broadcasting alert: %v", err)
		}

		log.Printf("[MQTT Handler] Alert sent: Type=%s, Severity=%s, MCU=%s",
			alert.AlertType, alert.Severity, sensorPayload.MCUCode)
	}

	// Check if auto pump should be triggered
	if sensorPayload.SoilMoisture != nil {
		shouldPump, err := h.thresholdService.ShouldTriggerAutoPump(ctx, sensorPayload.SurveyPointID, *sensorPayload.SoilMoisture)
		if err != nil {
			log.Printf("[MQTT Handler] Error checking auto pump: %v", err)
		} else if shouldPump {
			h.triggerAutoPump(ctx, sensorPayload.SurveyPointID, sensorPayload.MCUCode, *sensorPayload.SoilMoisture)
		}
	}

	// Broadcast sensor data to WebSocket clients
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

func (h *Handler) triggerAutoPump(ctx context.Context, surveyPointID uuid.UUID, mcuCode string, soilMoisture float64) {
	log.Printf("[MQTT Handler] Auto pump triggered for SurveyPoint: %s, Soil Moisture: %.2f%%", surveyPointID, soilMoisture)

	// Create command to turn on pump
	cmdReq := &sensorData.CreateCommandRequest{
		SurveyPointID: surveyPointID,
		DeviceName:    "pump",
		Command:       "on",
	}

	result, err := h.sensorService.CreateCommand(ctx, cmdReq)
	if err != nil {
		log.Printf("[MQTT Handler] Error creating auto pump command: %v", err)
		return
	}

	// Record auto pump history
	pumpHistory, err := h.thresholdService.RecordAutoPump(ctx, surveyPointID, result.CommandID, soilMoisture)
	if err != nil {
		log.Printf("[MQTT Handler] Error recording auto pump history: %v", err)
	}

	userID, err := h.surveyPointService.GetOwnerUserID(ctx, surveyPointID)
	if err != nil {
		log.Printf("[MQTT Handler] Error getting owner user ID: %v", err)
		userID = uuid.Nil
	}

	// Construct correct MQTT topic based on user ID
	var mqttTopic string
	if userID != uuid.Nil {
		// Send to user-specific topic (matching ESP8266 subscription)
		mqttTopic = fmt.Sprintf("user/%s/mcu/%s/control/request", userID, mcuCode)
		log.Printf("[MQTT Handler] Using user-specific topic for auto pump: %s", mqttTopic)
	} else {
		// Fallback to system topic
		mqttTopic = fmt.Sprintf("system/mcu/%s/control/request", mcuCode)
		log.Printf("[MQTT Handler] WARNING: Using fallback system topic: %s", mqttTopic)
	}

	mqttMsg := shared.MQTTMessage{
		Topic: "control_request",
		Payload: shared.ControlRequestPayload{
			SurveyPointID: surveyPointID,
			MCUCode:       mcuCode,
			DeviceName:    "pump",
			Command:       "on",
		},
	}

	if err := h.mqttService.PublishJSON(mqttTopic, mqttMsg); err != nil {
		log.Printf("[MQTT Handler] Error publishing auto pump command: %v", err)

		// Update command and pump history status to failed
		if result.CommandID != nil {
			h.sensorService.UpdateCommandStatus(ctx, *result.CommandID, "failed")
		}
		if pumpHistory != nil {
			notes := fmt.Sprintf("Failed to publish MQTT command to topic: %s", mqttTopic)
			h.thresholdService.UpdateAutoPumpStatus(ctx, pumpHistory.ID, "failed", &notes)
		}
		return
	}

	// Update auto pump history to running
	if pumpHistory != nil {
		notes := fmt.Sprintf("Auto pump started. Command ID: %s, Topic: %s, User ID: %s",
			result.CommandID, mqttTopic, userID)
		h.thresholdService.UpdateAutoPumpStatus(ctx, pumpHistory.ID, "running", &notes)
	}

	// Send notification via WebSocket
	alertPayload := shared.MQTTAlert{
		MCUCode:  mcuCode,
		Title:    "Tự động bơm nước",
		Message:  fmt.Sprintf("Độ ẩm đất thấp (%.2f%%). Hệ thống đã tự động bật máy bơm.", soilMoisture),
		Severity: "info",
		Time:     time.Now(),
	}

	alertBytes, _ := json.Marshal(alertPayload)
	wsMsg := shared.WSMessage{
		Topic:     shared.WSTopicAlert,
		Payload:   alertBytes,
		Timestamp: time.Now(),
	}

	if err := h.wsService.BroadcastToMCU(mcuCode, wsMsg); err != nil {
		log.Printf("[MQTT Handler] Error broadcasting auto pump notification: %v", err)
	}

	log.Printf("[MQTT Handler] ✅ Auto pump command sent successfully: Topic=%s, SurveyPoint=%s, Command=%s, UserID=%s",
		mqttTopic, surveyPointID, result.CommandID, userID)
}

func (h *Handler) onControlResponse(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg shared.MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling control response: %v", err)
		return
	}

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

	if controlPayload.SurveyPointID == uuid.Nil {
		log.Printf("[MQTT Handler] Invalid survey_point_id in control response")
		return
	}

	ctx := context.Background()

	commands, err := h.sensorService.GetCommandHistory(ctx, &controlPayload.SurveyPointID, &controlPayload.DeviceName, 10)
	if err != nil {
		log.Printf("[MQTT Handler] Error getting command history: %v", err)
		return
	}

	var commandID uuid.UUID
	for _, cmd := range commands {
		if cmd.Status == "pending" {
			commandID = cmd.CommandID
			break
		}
	}

	if commandID != uuid.Nil {
		status := "success"
		if controlPayload.Status == "failed" || controlPayload.Status == "error" {
			status = "failed"
		}

		if _, err := h.sensorService.UpdateCommandStatus(ctx, commandID, status); err != nil {
			log.Printf("[MQTT Handler] Error updating command status: %v", err)
		}

		// Update auto pump history if this was an auto pump command
		autoPumpHistory, err := h.thresholdService.GetAutoPumpHistory(ctx, controlPayload.SurveyPointID, 1)
		if err == nil && len(autoPumpHistory) > 0 && autoPumpHistory[0].CommandID != nil && *autoPumpHistory[0].CommandID == commandID {
			notes := fmt.Sprintf("Pump %s: %s", controlPayload.Status, controlPayload.Message)
			finalStatus := "completed"
			if status == "failed" {
				finalStatus = "failed"
			}
			h.thresholdService.UpdateAutoPumpStatus(ctx, autoPumpHistory[0].ID, finalStatus, &notes)
			log.Printf("[MQTT Handler] Auto pump history updated: %s", finalStatus)
		}
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

	log.Printf("[MQTT Handler] Processed control response for MCU: %s, Device: %s, Status: %s",
		controlPayload.MCUCode, controlPayload.DeviceName, controlPayload.Status)
}

func (h *Handler) onAlert(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg shared.MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling alert: %v", err)
		return
	}

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

func (h *Handler) onDiseaseDetection(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg shared.MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling disease detection: %v", err)
		return
	}

	payloadBytes, err := json.Marshal(mqttMsg.Payload)
	if err != nil {
		log.Printf("[MQTT Handler] Error marshaling payload: %v", err)
		return
	}

	var diseasePayload shared.DiseaseDetectionPayload
	if err := json.Unmarshal(payloadBytes, &diseasePayload); err != nil {
		log.Printf("[MQTT Handler] Error unmarshaling disease payload: %v", err)
		return
	}

	wsMsg := shared.WSMessage{
		Topic:     "disease_detection",
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}

	if err := h.wsService.BroadcastToMCU(diseasePayload.MCUCode, wsMsg); err != nil {
		log.Printf("[MQTT Handler] Error broadcasting disease detection: %v", err)
	}

	log.Printf("[MQTT Handler] Processed disease detection for MCU: %s, Disease: %s", diseasePayload.MCUCode, diseasePayload.DiseaseName)

}

func getFloatValue(val *float64) float64 {
	if val == nil {
		return 0.0
	}
	return *val
}
