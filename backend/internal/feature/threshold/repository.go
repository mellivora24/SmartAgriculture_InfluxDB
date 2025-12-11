package threshold

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	GetBySurveyPoint(ctx context.Context, surveyPointID uuid.UUID) (*ThresholdSettings, error)
	CreateOrUpdate(ctx context.Context, settings *ThresholdSettings) error
	Update(ctx context.Context, surveyPointID uuid.UUID, req *UpdateThresholdRequest) error

	CreateAlertHistory(ctx context.Context, alert *AlertHistory) error
	GetAlertHistory(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]AlertHistory, error)
	GetLastAlert(ctx context.Context, surveyPointID uuid.UUID, alertType string) (*AlertHistory, error)
	AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID) error

	CreateAutoPumpHistory(ctx context.Context, history *AutoPumpHistory) error
	GetAutoPumpHistory(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]AutoPumpHistory, error)
	GetLastAutoPump(ctx context.Context, surveyPointID uuid.UUID) (*AutoPumpHistory, error)
	UpdateAutoPumpStatus(ctx context.Context, id uuid.UUID, status string, notes *string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetBySurveyPoint(ctx context.Context, surveyPointID uuid.UUID) (*ThresholdSettings, error) {
	var settings ThresholdSettings
	err := r.db.WithContext(ctx).
		Where("survey_point_id = ?", surveyPointID).
		First(&settings).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &settings, err
}

func (r *repository) CreateOrUpdate(ctx context.Context, settings *ThresholdSettings) error {
	// Check if exists
	var existing ThresholdSettings
	err := r.db.WithContext(ctx).
		Where("survey_point_id = ?", settings.SurveyPointID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		return r.db.WithContext(ctx).Create(settings).Error
	} else if err != nil {
		return err
	}

	// Update existing
	settings.ID = existing.ID
	settings.CreatedAt = existing.CreatedAt
	settings.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&existing).
		Updates(settings).Error
}

func (r *repository) Update(ctx context.Context, surveyPointID uuid.UUID, req *UpdateThresholdRequest) error {
	updates := make(map[string]interface{})

	// Temperature
	if req.TempMin != nil {
		updates["temp_min"] = *req.TempMin
	}
	if req.TempMax != nil {
		updates["temp_max"] = *req.TempMax
	}
	if req.TempCriticalMin != nil {
		updates["temp_critical_min"] = *req.TempCriticalMin
	}
	if req.TempCriticalMax != nil {
		updates["temp_critical_max"] = *req.TempCriticalMax
	}

	// Humidity
	if req.HumidityMin != nil {
		updates["humidity_min"] = *req.HumidityMin
	}
	if req.HumidityMax != nil {
		updates["humidity_max"] = *req.HumidityMax
	}
	if req.HumidityCriticalMin != nil {
		updates["humidity_critical_min"] = *req.HumidityCriticalMin
	}
	if req.HumidityCriticalMax != nil {
		updates["humidity_critical_max"] = *req.HumidityCriticalMax
	}

	// Soil Moisture
	if req.SoilMoistureMin != nil {
		updates["soil_moisture_min"] = *req.SoilMoistureMin
	}
	if req.SoilMoistureMax != nil {
		updates["soil_moisture_max"] = *req.SoilMoistureMax
	}
	if req.SoilMoistureCriticalMin != nil {
		updates["soil_moisture_critical_min"] = *req.SoilMoistureCriticalMin
	}
	if req.SoilMoistureCriticalMax != nil {
		updates["soil_moisture_critical_max"] = *req.SoilMoistureCriticalMax
	}

	// Light
	if req.LightMin != nil {
		updates["light_min"] = *req.LightMin
	}
	if req.LightMax != nil {
		updates["light_max"] = *req.LightMax
	}
	if req.LightCriticalMin != nil {
		updates["light_critical_min"] = *req.LightCriticalMin
	}
	if req.LightCriticalMax != nil {
		updates["light_critical_max"] = *req.LightCriticalMax
	}

	// Auto Pump
	if req.AutoPumpEnabled != nil {
		updates["auto_pump_enabled"] = *req.AutoPumpEnabled
	}
	if req.PumpTriggerSoilMoisture != nil {
		updates["pump_trigger_soil_moisture"] = *req.PumpTriggerSoilMoisture
	}
	if req.PumpStopSoilMoisture != nil {
		updates["pump_stop_soil_moisture"] = *req.PumpStopSoilMoisture
	}
	if req.PumpDurationSeconds != nil {
		updates["pump_duration_seconds"] = *req.PumpDurationSeconds
	}
	if req.PumpCooldownMinutes != nil {
		updates["pump_cooldown_minutes"] = *req.PumpCooldownMinutes
	}

	// Alerts
	if req.AlertEnabled != nil {
		updates["alert_enabled"] = *req.AlertEnabled
	}
	if req.AlertCooldownMinutes != nil {
		updates["alert_cooldown_minutes"] = *req.AlertCooldownMinutes
	}

	updates["updated_at"] = time.Now()

	return r.db.WithContext(ctx).
		Model(&ThresholdSettings{}).
		Where("survey_point_id = ?", surveyPointID).
		Updates(updates).Error
}

func (r *repository) CreateAlertHistory(ctx context.Context, alert *AlertHistory) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *repository) GetAlertHistory(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]AlertHistory, error) {
	var alerts []AlertHistory
	err := r.db.WithContext(ctx).
		Where("survey_point_id = ?", surveyPointID).
		Order("created_at DESC").
		Limit(limit).
		Find(&alerts).Error

	return alerts, err
}

func (r *repository) GetLastAlert(ctx context.Context, surveyPointID uuid.UUID, alertType string) (*AlertHistory, error) {
	var alert AlertHistory
	err := r.db.WithContext(ctx).
		Where("survey_point_id = ? AND alert_type = ?", surveyPointID, alertType).
		Order("created_at DESC").
		First(&alert).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &alert, err
}

func (r *repository) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&AlertHistory{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"acknowledged":    true,
			"acknowledged_at": now,
			"acknowledged_by": userID,
		}).Error
}

func (r *repository) CreateAutoPumpHistory(ctx context.Context, history *AutoPumpHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *repository) GetAutoPumpHistory(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]AutoPumpHistory, error) {
	var history []AutoPumpHistory
	err := r.db.WithContext(ctx).
		Where("survey_point_id = ?", surveyPointID).
		Order("started_at DESC").
		Limit(limit).
		Find(&history).Error

	return history, err
}

func (r *repository) GetLastAutoPump(ctx context.Context, surveyPointID uuid.UUID) (*AutoPumpHistory, error) {
	var history AutoPumpHistory
	err := r.db.WithContext(ctx).
		Where("survey_point_id = ?", surveyPointID).
		Order("started_at DESC").
		First(&history).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &history, err
}

func (r *repository) UpdateAutoPumpStatus(ctx context.Context, id uuid.UUID, status string, notes *string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"completed_at": now,
	}

	if notes != nil {
		updates["notes"] = *notes
	}

	return r.db.WithContext(ctx).
		Model(&AutoPumpHistory{}).
		Where("id = ?", id).
		Updates(updates).Error
}
