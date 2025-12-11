package threshold

import (
	"time"

	"github.com/google/uuid"
)

// ThresholdSettings định nghĩa các ngưỡng cảnh báo
type ThresholdSettings struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SurveyPointID uuid.UUID `json:"survey_point_id" gorm:"type:uuid;not null;uniqueIndex"`

	// Temperature thresholds
	TempMin         *float64 `json:"temp_min,omitempty" gorm:"type:float"`
	TempMax         *float64 `json:"temp_max,omitempty" gorm:"type:float"`
	TempCriticalMin *float64 `json:"temp_critical_min,omitempty" gorm:"type:float"`
	TempCriticalMax *float64 `json:"temp_critical_max,omitempty" gorm:"type:float"`

	// Humidity thresholds
	HumidityMin         *float64 `json:"humidity_min,omitempty" gorm:"type:float"`
	HumidityMax         *float64 `json:"humidity_max,omitempty" gorm:"type:float"`
	HumidityCriticalMin *float64 `json:"humidity_critical_min,omitempty" gorm:"type:float"`
	HumidityCriticalMax *float64 `json:"humidity_critical_max,omitempty" gorm:"type:float"`

	// Soil moisture thresholds
	SoilMoistureMin         *float64 `json:"soil_moisture_min,omitempty" gorm:"type:float"`
	SoilMoistureMax         *float64 `json:"soil_moisture_max,omitempty" gorm:"type:float"`
	SoilMoistureCriticalMin *float64 `json:"soil_moisture_critical_min,omitempty" gorm:"type:float"`
	SoilMoistureCriticalMax *float64 `json:"soil_moisture_critical_max,omitempty" gorm:"type:float"`

	// Light thresholds
	LightMin         *float64 `json:"light_min,omitempty" gorm:"type:float"`
	LightMax         *float64 `json:"light_max,omitempty" gorm:"type:float"`
	LightCriticalMin *float64 `json:"light_critical_min,omitempty" gorm:"type:float"`
	LightCriticalMax *float64 `json:"light_critical_max,omitempty" gorm:"type:float"`

	// Auto pump settings
	AutoPumpEnabled         bool     `json:"auto_pump_enabled" gorm:"default:false"`
	PumpTriggerSoilMoisture *float64 `json:"pump_trigger_soil_moisture,omitempty" gorm:"type:float"`
	PumpStopSoilMoisture    *float64 `json:"pump_stop_soil_moisture,omitempty" gorm:"type:float"`
	PumpDurationSeconds     int      `json:"pump_duration_seconds" gorm:"default:30"`
	PumpCooldownMinutes     int      `json:"pump_cooldown_minutes" gorm:"default:60"`

	// Alert settings
	AlertEnabled         bool `json:"alert_enabled" gorm:"default:true"`
	AlertCooldownMinutes int  `json:"alert_cooldown_minutes" gorm:"default:10"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (ThresholdSettings) TableName() string {
	return "tbl_threshold_settings"
}

// AlertHistory lịch sử cảnh báo
type AlertHistory struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SurveyPointID  uuid.UUID  `json:"survey_point_id" gorm:"type:uuid;not null;index"`
	AlertType      string     `json:"alert_type" gorm:"type:varchar(50);not null"`
	Severity       string     `json:"severity" gorm:"type:varchar(20);not null;index"`
	SensorValue    float64    `json:"sensor_value" gorm:"type:float;not null"`
	ThresholdValue float64    `json:"threshold_value" gorm:"type:float;not null"`
	Message        string     `json:"message" gorm:"type:text;not null"`
	Acknowledged   bool       `json:"acknowledged" gorm:"default:false;index"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *uuid.UUID `json:"acknowledged_by,omitempty" gorm:"type:uuid"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
}

// TableName specifies the table name
func (AlertHistory) TableName() string {
	return "tbl_alert_history"
}

// AutoPumpHistory lịch sử tự động bơm
type AutoPumpHistory struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SurveyPointID       uuid.UUID  `json:"survey_point_id" gorm:"type:uuid;not null;index"`
	CommandID           *uuid.UUID `json:"command_id,omitempty" gorm:"type:uuid"`
	TriggerSoilMoisture float64    `json:"trigger_soil_moisture" gorm:"type:float;not null"`
	TargetSoilMoisture  float64    `json:"target_soil_moisture" gorm:"type:float;not null"`
	PumpDurationSeconds int        `json:"pump_duration_seconds" gorm:"not null"`
	Status              string     `json:"status" gorm:"type:varchar(50);default:'triggered';index"`
	StartedAt           time.Time  `json:"started_at" gorm:"autoCreateTime;index"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	Notes               *string    `json:"notes,omitempty" gorm:"type:text"`
}

// TableName specifies the table name
func (AutoPumpHistory) TableName() string {
	return "tbl_auto_pump_history"
}

// UpdateThresholdRequest request để cập nhật threshold
type UpdateThresholdRequest struct {
	TempMin         *float64 `json:"temp_min"`
	TempMax         *float64 `json:"temp_max"`
	TempCriticalMin *float64 `json:"temp_critical_min"`
	TempCriticalMax *float64 `json:"temp_critical_max"`

	HumidityMin         *float64 `json:"humidity_min"`
	HumidityMax         *float64 `json:"humidity_max"`
	HumidityCriticalMin *float64 `json:"humidity_critical_min"`
	HumidityCriticalMax *float64 `json:"humidity_critical_max"`

	SoilMoistureMin         *float64 `json:"soil_moisture_min"`
	SoilMoistureMax         *float64 `json:"soil_moisture_max"`
	SoilMoistureCriticalMin *float64 `json:"soil_moisture_critical_min"`
	SoilMoistureCriticalMax *float64 `json:"soil_moisture_critical_max"`

	LightMin         *float64 `json:"light_min"`
	LightMax         *float64 `json:"light_max"`
	LightCriticalMin *float64 `json:"light_critical_min"`
	LightCriticalMax *float64 `json:"light_critical_max"`

	AutoPumpEnabled         *bool    `json:"auto_pump_enabled"`
	PumpTriggerSoilMoisture *float64 `json:"pump_trigger_soil_moisture"`
	PumpStopSoilMoisture    *float64 `json:"pump_stop_soil_moisture"`
	PumpDurationSeconds     *int     `json:"pump_duration_seconds"`
	PumpCooldownMinutes     *int     `json:"pump_cooldown_minutes"`

	AlertEnabled         *bool `json:"alert_enabled"`
	AlertCooldownMinutes *int  `json:"alert_cooldown_minutes"`
}

// AlertCheck kết quả kiểm tra ngưỡng
type AlertCheck struct {
	ShouldAlert    bool
	AlertType      string
	Severity       string
	SensorValue    float64
	ThresholdValue float64
	Message        string
}
