package sensorData

import (
	"time"

	"github.com/google/uuid"
)

type SensorData struct {
	SurveyPointID   uuid.UUID              `json:"survey_point_id"`
	SurveyPointName string                 `json:"survey_point_name"`
	MCUCode         string                 `json:"mcu_code"`
	FarmName        string                 `json:"farm_name"`
	Temperature     float64                `json:"temperature"`
	Humidity        float64                `json:"humidity"`
	SoilMoisture    float64                `json:"soil_moisture"`
	Light           float64                `json:"light"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type DeviceCommand struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SurveyPointID *uuid.UUID `json:"survey_point_id,omitempty" gorm:"type:uuid"`
	DeviceName    string     `json:"device_name" gorm:"type:varchar(255);not null"`
	Command       string     `json:"command" gorm:"type:varchar(50);not null"`
	Status        string     `json:"status" gorm:"type:varchar(50);default:'pending'"`
	ExecutedAt    *time.Time `json:"executed_at,omitempty" gorm:"type:timestamp"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (DeviceCommand) TableName() string {
	return "tbl_device_commands"
}

type CommandInfo struct {
	CommandID       uuid.UUID  `json:"command_id"`
	SurveyPointID   *uuid.UUID `json:"survey_point_id"`
	SurveyPointName *string    `json:"survey_point_name"`
	DeviceName      string     `json:"device_name"`
	Command         string     `json:"command"`
	Status          string     `json:"status"`
	ExecutedAt      *time.Time `json:"executed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CreateCommandRequest struct {
	SurveyPointID uuid.UUID `json:"survey_point_id" validate:"required"`
	DeviceName    string    `json:"device_name" validate:"required"`
	Command       string    `json:"command" validate:"required,oneof=on off"`
}

type CommandOperationResult struct {
	Success   bool       `json:"success"`
	CommandID *uuid.UUID `json:"command_id,omitempty"`
	Message   string     `json:"message"`
}

type QuerySensorDataRequest struct {
	SurveyPointID *uuid.UUID `json:"survey_point_id,omitempty"`
	MCUCode       *string    `json:"mcu_code,omitempty"`
	FarmID        *uuid.UUID `json:"farm_id,omitempty"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Limit         int        `json:"limit" validate:"min=1,max=1000"`
}

type AggregationRequest struct {
	SurveyPointID uuid.UUID `json:"survey_point_id" validate:"required"`
	Field         string    `json:"field" validate:"required,oneof=temperature humidity soil_moisture light"`
	Aggregation   string    `json:"aggregation" validate:"required,oneof=mean min max sum count"`
	Window        string    `json:"window" validate:"required"`
	StartTime     time.Time `json:"start_time" validate:"required"`
	EndTime       time.Time `json:"end_time" validate:"required"`
}
