package shared

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MQTT Topics
const (
	MQTTTopicSensorData      = "user/+/mcu/+/data"
	MQTTTopicControlResponse = "user/+/mcu/+/control/response"
	MQTTTopicAlert           = "user/+/mcu/+/alert"
	MQTTTopicHealthRequest   = "/health/request"
	MQTTTopicHealthResponse  = "/health/response"
)

// WebSocket Topics
const (
	WSTopicConnect         = "connect"
	WSTopicSensorData      = "sensor_data"
	WSTopicControlRequest  = "control_request"
	WSTopicControlResponse = "control_response"
	WSTopicAlert           = "alert"
	WSTopicError           = "error"
)

// MQTTMessage represents an MQTT message
type MQTTMessage struct {
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
}

// SensorDataPayload represents sensor data from MCU
type SensorDataPayload struct {
	MCUCode       string                 `json:"mcu_code"`
	SurveyPointID uuid.UUID              `json:"survey_point_id"`
	Temperature   *float64               `json:"temperature,omitempty"`
	Humidity      *float64               `json:"humidity,omitempty"`
	SoilMoisture  *float64               `json:"soil_moisture,omitempty"`
	Light         *float64               `json:"light,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// ControlRequestPayload represents a control request from client
type ControlRequestPayload struct {
	SurveyPointID uuid.UUID              `json:"survey_point_id"`
	MCUCode       string                 `json:"mcu_code"`
	DeviceName    string                 `json:"device_name"`
	Command       string                 `json:"command"` // "on" or "off"
	Value         interface{}            `json:"value,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// ControlResponsePayload represents a control response from MCU
type ControlResponsePayload struct {
	SurveyPointID uuid.UUID   `json:"survey_point_id"`
	MCUCode       string      `json:"mcu_code"`
	DeviceName    string      `json:"device_name"`
	Command       string      `json:"command"`
	Status        string      `json:"status"` // success, failed, pending
	Message       string      `json:"message,omitempty"`
	Value         interface{} `json:"value,omitempty"`
}

// DeviceConfig represents device configuration
type DeviceConfig struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	DeviceType string    `json:"device_type"`
	MCUCode    string    `json:"mcu_code"`
	IsActive   bool      `json:"is_active"`
}

// MQTTAlert represents alert notification from MCU
type MQTTAlert struct {
	MCUCode  string    `json:"mcu_code"`
	Title    string    `json:"title"`
	Message  string    `json:"message"`
	Severity string    `json:"severity"` // info, warning, error, critical
	Time     time.Time `json:"time"`
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Topic     string          `json:"topic"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp,omitempty"`
}

// WSErrorPayload represents an error message
type WSErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ClientInfo represents WebSocket client information
type ClientInfo struct {
	UID       string
	MCUCode   string
	ConnectAt time.Time
}

type DiseaseDetectionPayload struct {
	MCUCode     string    `json:"mcu_code"`
	DiseaseName string    `json:"disease_name"`
	Confidence  float64   `json:"confidence"`
	DetectedAt  time.Time `json:"detected_at"`
}
