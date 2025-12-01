package shared

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	MQTTTopicSensorData      = "user/+/mcu/+/data"
	MQTTTopicControlResponse = "user/+/mcu/+/control/response"
	MQTTTopicAlert           = "user/+/mcu/+/alert"
	MQTTTopicHealthRequest   = "/health/request"
	MQTTTopicHealthResponse  = "/health/response"
)

const (
	WSTopicConnect         = "connect"
	WSTopicSensorData      = "sensor_data"
	WSTopicControlRequest  = "control_request"
	WSTopicControlResponse = "control_response"
	WSTopicAlert           = "alert"
	WSTopicError           = "error"
)

type MQTTMessage struct {
	Topic     string      `json:"topic"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type SensorDataPayload struct {
	MCUCode       string                 `json:"mcu_code"`
	SurveyPointID uuid.UUID              `json:"survey_point_id"`
	Temperature   *float64               `json:"temperature,omitempty"`
	Humidity      *float64               `json:"humidity,omitempty"`
	SoilMoisture  *float64               `json:"soil_moisture,omitempty"`
	Light         *float64               `json:"light,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

type ControlRequestPayload struct {
	MCUCode    string                 `json:"mcu_code"`
	DeviceName string                 `json:"device_name"`
	Command    string                 `json:"command"`
	Value      interface{}            `json:"value,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

type ControlResponsePayload struct {
	MCUCode    string      `json:"mcu_code"`
	DeviceName string      `json:"device_name"`
	Command    string      `json:"command"`
	Status     string      `json:"status"` // success, failed, pending
	Message    string      `json:"message,omitempty"`
	Value      interface{} `json:"value,omitempty"`
	ExecutedAt time.Time   `json:"executed_at"`
}

type DeviceConfig struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	DeviceType string    `json:"device_type"`
	MCUCode    string    `json:"mcu_code"`
	IsActive   bool      `json:"is_active"`
}

type MQTTAlert struct {
	MCUCode  string    `json:"mcu_code"`
	Title    string    `json:"title"`
	Message  string    `json:"message"`
	Severity string    `json:"severity"` // info, warning, error, critical
	Time     time.Time `json:"time"`
}

type WSMessage struct {
	Topic     string          `json:"topic"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp,omitempty"`
}

type WSErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ClientInfo struct {
	UID       string
	MCUCode   string
	ConnectAt time.Time
}
