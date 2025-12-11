package threshold

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	GetBySurveyPoint(ctx context.Context, surveyPointID uuid.UUID) (*ThresholdSettings, error)
	UpdateThresholds(ctx context.Context, surveyPointID uuid.UUID, req *UpdateThresholdRequest) error

	CheckSensorThresholds(ctx context.Context, surveyPointID uuid.UUID, temperature, humidity, soilMoisture, light *float64) ([]AlertCheck, error)
	RecordAlert(ctx context.Context, surveyPointID uuid.UUID, check AlertCheck) error
	GetAlertHistory(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]AlertHistory, error)
	AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID) error

	ShouldTriggerAutoPump(ctx context.Context, surveyPointID uuid.UUID, soilMoisture float64) (bool, error)
	RecordAutoPump(ctx context.Context, surveyPointID uuid.UUID, commandID *uuid.UUID, soilMoisture float64) (*AutoPumpHistory, error)
	GetAutoPumpHistory(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]AutoPumpHistory, error)
	UpdateAutoPumpStatus(ctx context.Context, id uuid.UUID, status string, notes *string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetBySurveyPoint(ctx context.Context, surveyPointID uuid.UUID) (*ThresholdSettings, error) {
	return s.repo.GetBySurveyPoint(ctx, surveyPointID)
}

func (s *service) UpdateThresholds(ctx context.Context, surveyPointID uuid.UUID, req *UpdateThresholdRequest) error {
	return s.repo.Update(ctx, surveyPointID, req)
}

func (s *service) CheckSensorThresholds(ctx context.Context, surveyPointID uuid.UUID, temperature, humidity, soilMoisture, light *float64) ([]AlertCheck, error) {
	settings, err := s.repo.GetBySurveyPoint(ctx, surveyPointID)
	if err != nil || settings == nil || !settings.AlertEnabled {
		return nil, err
	}

	var alerts []AlertCheck

	// Check temperature
	if temperature != nil {
		if alert := s.checkThreshold("temperature", *temperature,
			settings.TempMin, settings.TempMax,
			settings.TempCriticalMin, settings.TempCriticalMax, "°C"); alert != nil {
			if s.shouldSendAlert(ctx, surveyPointID, "temperature", settings.AlertCooldownMinutes) {
				alerts = append(alerts, *alert)
			}
		}
	}

	// Check humidity
	if humidity != nil {
		if alert := s.checkThreshold("humidity", *humidity,
			settings.HumidityMin, settings.HumidityMax,
			settings.HumidityCriticalMin, settings.HumidityCriticalMax, "%"); alert != nil {
			if s.shouldSendAlert(ctx, surveyPointID, "humidity", settings.AlertCooldownMinutes) {
				alerts = append(alerts, *alert)
			}
		}
	}

	// Check soil moisture
	if soilMoisture != nil {
		if alert := s.checkThreshold("soil_moisture", *soilMoisture,
			settings.SoilMoistureMin, settings.SoilMoistureMax,
			settings.SoilMoistureCriticalMin, settings.SoilMoistureCriticalMax, "%"); alert != nil {
			if s.shouldSendAlert(ctx, surveyPointID, "soil_moisture", settings.AlertCooldownMinutes) {
				alerts = append(alerts, *alert)
			}
		}
	}

	// Check light
	if light != nil {
		if alert := s.checkThreshold("light", *light,
			settings.LightMin, settings.LightMax,
			settings.LightCriticalMin, settings.LightCriticalMax, "lux"); alert != nil {
			if s.shouldSendAlert(ctx, surveyPointID, "light", settings.AlertCooldownMinutes) {
				alerts = append(alerts, *alert)
			}
		}
	}

	return alerts, nil
}

func (s *service) checkThreshold(alertType string, value float64, min, max, criticalMin, criticalMax *float64, unit string) *AlertCheck {
	if criticalMin != nil && value < *criticalMin {
		return &AlertCheck{
			ShouldAlert:    true,
			AlertType:      alertType,
			Severity:       "critical",
			SensorValue:    value,
			ThresholdValue: *criticalMin,
			Message:        fmt.Sprintf("%s quá thấp: %.2f%s (ngưỡng tối thiểu: %.2f%s)", getVietnameseName(alertType), value, unit, *criticalMin, unit),
		}
	}

	if criticalMax != nil && value > *criticalMax {
		return &AlertCheck{
			ShouldAlert:    true,
			AlertType:      alertType,
			Severity:       "critical",
			SensorValue:    value,
			ThresholdValue: *criticalMax,
			Message:        fmt.Sprintf("%s quá cao: %.2f%s (ngưỡng tối đa: %.2f%s)", getVietnameseName(alertType), value, unit, *criticalMax, unit),
		}
	}

	if min != nil && value < *min {
		return &AlertCheck{
			ShouldAlert:    true,
			AlertType:      alertType,
			Severity:       "warning",
			SensorValue:    value,
			ThresholdValue: *min,
			Message:        fmt.Sprintf("%s thấp: %.2f%s (ngưỡng khuyến nghị: %.2f%s)", getVietnameseName(alertType), value, unit, *min, unit),
		}
	}

	if max != nil && value > *max {
		return &AlertCheck{
			ShouldAlert:    true,
			AlertType:      alertType,
			Severity:       "warning",
			SensorValue:    value,
			ThresholdValue: *max,
			Message:        fmt.Sprintf("%s cao: %.2f%s (ngưỡng khuyến nghị: %.2f%s)", getVietnameseName(alertType), value, unit, *max, unit),
		}
	}

	return nil
}

func getVietnameseName(alertType string) string {
	names := map[string]string{
		"temperature":   "Nhiệt độ",
		"humidity":      "Độ ẩm không khí",
		"soil_moisture": "Độ ẩm đất",
		"light":         "Ánh sáng",
	}
	return names[alertType]
}

func (s *service) shouldSendAlert(ctx context.Context, surveyPointID uuid.UUID, alertType string, cooldownMinutes int) bool {
	lastAlert, err := s.repo.GetLastAlert(ctx, surveyPointID, alertType)
	if err != nil || lastAlert == nil {
		return true
	}

	timeSinceLastAlert := time.Since(lastAlert.CreatedAt)
	cooldown := time.Duration(cooldownMinutes) * time.Minute

	return timeSinceLastAlert >= cooldown
}

func (s *service) RecordAlert(ctx context.Context, surveyPointID uuid.UUID, check AlertCheck) error {
	alert := &AlertHistory{
		SurveyPointID:  surveyPointID,
		AlertType:      check.AlertType,
		Severity:       check.Severity,
		SensorValue:    check.SensorValue,
		ThresholdValue: check.ThresholdValue,
		Message:        check.Message,
	}

	return s.repo.CreateAlertHistory(ctx, alert)
}

func (s *service) GetAlertHistory(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]AlertHistory, error) {
	return s.repo.GetAlertHistory(ctx, surveyPointID, limit)
}

func (s *service) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID) error {
	return s.repo.AcknowledgeAlert(ctx, alertID, userID)
}

func (s *service) ShouldTriggerAutoPump(ctx context.Context, surveyPointID uuid.UUID, soilMoisture float64) (bool, error) {
	settings, err := s.repo.GetBySurveyPoint(ctx, surveyPointID)
	if err != nil || settings == nil {
		return false, err
	}

	if !settings.AutoPumpEnabled || settings.PumpTriggerSoilMoisture == nil {
		return false, nil
	}

	if soilMoisture >= *settings.PumpTriggerSoilMoisture {
		return false, nil
	}

	lastPump, err := s.repo.GetLastAutoPump(ctx, surveyPointID)
	if err != nil {
		return false, err
	}

	if lastPump != nil {
		timeSinceLastPump := time.Since(lastPump.StartedAt)
		cooldown := time.Duration(settings.PumpCooldownMinutes) * time.Minute

		if timeSinceLastPump < cooldown {
			return false, nil
		}
	}

	return true, nil
}

func (s *service) RecordAutoPump(ctx context.Context, surveyPointID uuid.UUID, commandID *uuid.UUID, soilMoisture float64) (*AutoPumpHistory, error) {
	settings, err := s.repo.GetBySurveyPoint(ctx, surveyPointID)
	if err != nil || settings == nil {
		return nil, err
	}

	targetSoilMoisture := float64(60.0)
	if settings.PumpStopSoilMoisture != nil {
		targetSoilMoisture = *settings.PumpStopSoilMoisture
	}

	history := &AutoPumpHistory{
		SurveyPointID:       surveyPointID,
		CommandID:           commandID,
		TriggerSoilMoisture: soilMoisture,
		TargetSoilMoisture:  targetSoilMoisture,
		PumpDurationSeconds: settings.PumpDurationSeconds,
		Status:              "triggered",
	}

	if err := s.repo.CreateAutoPumpHistory(ctx, history); err != nil {
		return nil, err
	}

	return history, nil
}

func (s *service) GetAutoPumpHistory(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]AutoPumpHistory, error) {
	return s.repo.GetAutoPumpHistory(ctx, surveyPointID, limit)
}

func (s *service) UpdateAutoPumpStatus(ctx context.Context, id uuid.UUID, status string, notes *string) error {
	return s.repo.UpdateAutoPumpStatus(ctx, id, status, notes)
}
