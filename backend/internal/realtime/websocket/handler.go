package websocket

import (
	"backend/internal/feature/sensorData"
	realtimeShared "backend/internal/realtime/shared"
	"backend/internal/shared"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure properly in production
	},
}

type Handler struct {
	wsService     realtimeShared.WebSocketService
	mqttService   realtimeShared.MQTTService
	sensorService sensorData.Service
}

func NewHandler(wsService realtimeShared.WebSocketService, mqttService realtimeShared.MQTTService, sensorService sensorData.Service) *Handler {
	return &Handler{
		wsService:     wsService,
		mqttService:   mqttService,
		sensorService: sensorService,
	}
}

// HandleConnection handles new WebSocket connections
func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Get token from query params
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter", http.StatusUnauthorized)
		return
	}

	// Validate token and get claims
	claims, err := shared.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	mcuCode := r.URL.Query().Get("mcu_code")
	if mcuCode == "" {
		http.Error(w, "Missing mcu_code parameter", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS Handler] Upgrade error: %v", err)
		http.Error(w, "Could not upgrade to WebSocket", http.StatusBadRequest)
		return
	}

	// Use username or user_id as UID
	uid := claims.Username
	if uid == "" {
		uid = claims.UserID.String()
	}

	client := h.wsService.AddClient(uid, mcuCode, conn)

	// Store user_id in client for later use
	client.Mu.Lock()
	client.UID = claims.UserID.String()
	client.Mu.Unlock()

	// Send welcome message
	welcomeMsg := realtimeShared.WSMessage{
		Topic: realtimeShared.WSTopicConnect,
		Payload: json.RawMessage(fmt.Sprintf(`{"status":"connected","mcu_code":"%s","user_id":"%s"}`,
			mcuCode, claims.UserID.String())),
	}

	welcomeData, _ := json.Marshal(welcomeMsg)
	select {
	case client.Send <- welcomeData:
	default:
		log.Printf("[WS Handler] Failed to send welcome message to client %s", client.ID)
	}

	log.Printf("[WS Handler] New client connected: UID=%s, MCU=%s, UserID=%s", uid, mcuCode, claims.UserID.String())

	// Start goroutines for read/write
	go h.wsService.HandleClientWrite(client)
	h.wsService.HandleClientRead(client, func(clientID string, msg realtimeShared.WSMessage) {
		h.onClientMessage(clientID, msg, claims.UserID.String())
	})
}

// onClientMessage handles messages from WebSocket clients
func (h *Handler) onClientMessage(clientID string, msg realtimeShared.WSMessage, userIDStr string) {
	log.Printf("[WS Handler] Message from client %s, topic: %s", clientID, msg.Topic)

	switch msg.Topic {
	case realtimeShared.WSTopicControlRequest:
		h.handleControlRequest(clientID, msg, userIDStr)
	default:
		log.Printf("[WS Handler] Unknown message topic: %s", msg.Topic)
	}
}

// handleControlRequest processes control requests from clients
func (h *Handler) handleControlRequest(clientID string, msg realtimeShared.WSMessage, userIDStr string) {
	var controlReq realtimeShared.ControlRequestPayload
	if err := json.Unmarshal(msg.Payload, &controlReq); err != nil {
		log.Printf("[WS Handler] Error unmarshaling control request: %v", err)
		h.sendError(clientID, "invalid_payload", "Invalid control request format")
		return
	}

	// Validate required fields
	if controlReq.DeviceName == "" || controlReq.Command == "" {
		h.sendError(clientID, "missing_fields", "Device name and command are required")
		return
	}

	if controlReq.MCUCode == "" {
		h.sendError(clientID, "missing_fields", "MCU code is required")
		return
	}

	// IMPORTANT: Validate survey_point_id
	if controlReq.SurveyPointID == uuid.Nil {
		h.sendError(clientID, "missing_fields", "Survey point ID is required")
		return
	}

	ctx := context.Background()

	// Parse user_id for permission check (optional)
	userID, err := shared.ParseUUID(userIDStr)
	if err != nil {
		log.Printf("[WS Handler] Error parsing user ID: %v", err)
		h.sendError(clientID, "invalid_user", "Invalid user ID")
		return
	}

	// TODO: Add permission check here
	// Check if user has permission to control this survey point
	// hasPermission := h.checkUserPermission(ctx, userID, controlReq.SurveyPointID)
	// if !hasPermission {
	//     h.sendError(clientID, "permission_denied", "You don't have permission to control this survey point")
	//     return
	// }

	// Create command in database with pending status
	cmdReq := &sensorData.CreateCommandRequest{
		SurveyPointID: controlReq.SurveyPointID,
		DeviceName:    controlReq.DeviceName,
		Command:       controlReq.Command,
	}

	result, err := h.sensorService.CreateCommand(ctx, cmdReq)
	if err != nil {
		log.Printf("[WS Handler] Error creating command: %v", err)
		h.sendError(clientID, "database_error", fmt.Sprintf("Failed to create command: %v", err))
		return
	}

	log.Printf("[WS Handler] Command created with ID: %s, Status: pending, SurveyPointID: %s",
		result.CommandID, controlReq.SurveyPointID)

	// Publish command to MQTT
	mqttTopic := fmt.Sprintf("user/%s/mcu/%s/control/request", userIDStr, controlReq.MCUCode)

	mqttMsg := realtimeShared.MQTTMessage{
		Topic:   "control_request",
		Payload: controlReq,
	}

	if err := h.mqttService.PublishJSON(mqttTopic, mqttMsg); err != nil {
		log.Printf("[WS Handler] Error publishing to MQTT: %v", err)

		// Update command status to failed
		_, err := h.sensorService.UpdateCommandStatus(ctx, *result.CommandID, "failed")
		if err != nil {
			log.Printf("[WS Handler] Error updating command status to failed: %v", err)
		}

		h.sendError(clientID, "mqtt_error", "Failed to send command to device")
		return
	}

	// Send acknowledgment to client
	ackPayload := map[string]interface{}{
		"command_id":      result.CommandID,
		"survey_point_id": controlReq.SurveyPointID.String(),
		"status":          "pending",
		"message":         "Command sent to device",
	}
	ackBytes, _ := json.Marshal(ackPayload)

	ackMsg := realtimeShared.WSMessage{
		Topic:   realtimeShared.WSTopicControlResponse,
		Payload: ackBytes,
	}

	if err := h.wsService.SendToClient(clientID, ackMsg); err != nil {
		log.Printf("[WS Handler] Error sending acknowledgment: %v", err)
	}

	log.Printf("[WS Handler] Control request processed: SurveyPointID=%s, Device=%s, Command=%s, UserID=%s",
		controlReq.SurveyPointID, controlReq.DeviceName, controlReq.Command, userID)
}

// sendError sends error message to client
func (h *Handler) sendError(clientID, code, message string) {
	errorPayload := realtimeShared.WSErrorPayload{
		Code:    code,
		Message: message,
	}

	payloadBytes, _ := json.Marshal(errorPayload)

	errorMsg := realtimeShared.WSMessage{
		Topic:   realtimeShared.WSTopicError,
		Payload: payloadBytes,
	}

	if err := h.wsService.SendToClient(clientID, errorMsg); err != nil {
		log.Printf("[WS Handler] Error sending error message: %v", err)
	}
}
