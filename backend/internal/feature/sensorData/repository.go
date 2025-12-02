package sensorData

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"gorm.io/gorm"
)

type InfluxClient interface {
	WritePoint(point *write.Point)
	Query(ctx context.Context, query string) (interface{}, error)
	CreatePoint(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) *write.Point
}

type Repository interface {
	WriteSensorData(ctx context.Context, data *SensorData) error
	QuerySensorData(ctx context.Context, req *QuerySensorDataRequest) ([]map[string]interface{}, error)
	QueryLatestData(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]map[string]interface{}, error)
	QueryAggregation(ctx context.Context, req *AggregationRequest) ([]map[string]interface{}, error)

	CreateCommand(ctx context.Context, surveyPointID uuid.UUID, deviceName, command string) (*CommandOperationResult, error)
	UpdateCommandStatus(ctx context.Context, commandID uuid.UUID, status string) (*CommandOperationResult, error)
	GetPendingCommands(ctx context.Context, limit int) ([]*CommandInfo, error)
	GetCommandHistory(ctx context.Context, surveyPointID *uuid.UUID, deviceName *string, limit int) ([]*CommandInfo, error)
	GetCommandByID(ctx context.Context, commandID uuid.UUID) (*DeviceCommand, error)
}

type repository struct {
	db     *gorm.DB
	influx InfluxClient
	bucket string
}

func NewRepository(db *gorm.DB, influx InfluxClient, bucket string) Repository {
	return &repository{
		db:     db,
		influx: influx,
		bucket: bucket,
	}
}

func (r *repository) WriteSensorData(ctx context.Context, data *SensorData) error {
	tags := map[string]string{
		"survey_point_id":   data.SurveyPointID.String(),
		"survey_point_name": data.SurveyPointName,
		"mcu_code":          data.MCUCode,
		"farm_name":         data.FarmName,
	}

	fields := map[string]interface{}{
		"temperature":   data.Temperature,
		"humidity":      data.Humidity,
		"soil_moisture": data.SoilMoisture,
		"light":         data.Light,
	}

	point := r.influx.CreatePoint("sensor_data", tags, fields, data.Timestamp)
	r.influx.WritePoint(point)

	return nil
}

func (r *repository) QuerySensorData(ctx context.Context, req *QuerySensorDataRequest) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "sensor_data")`,
		r.bucket,
		req.StartTime.Format(time.RFC3339),
		req.EndTime.Format(time.RFC3339),
	)

	if req.SurveyPointID != nil {
		query += fmt.Sprintf(`
			|> filter(fn: (r) => r.survey_point_id == "%s")`, req.SurveyPointID.String())
	}

	if req.MCUCode != nil {
		query += fmt.Sprintf(`
			|> filter(fn: (r) => r.mcu_code == "%s")`, *req.MCUCode)
	}

	query += fmt.Sprintf(`
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: %d)`, req.Limit)

	result, err := r.influx.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	records, ok := result.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected query result type")
	}

	return records, nil
}

func (r *repository) QueryLatestData(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -24h)
			|> filter(fn: (r) => r._measurement == "sensor_data")
			|> filter(fn: (r) => r.survey_point_id == "%s")
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: %d)`,
		r.bucket,
		surveyPointID.String(),
		limit,
	)

	result, err := r.influx.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	records, ok := result.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected query result type")
	}

	return records, nil
}

func (r *repository) QueryAggregation(ctx context.Context, req *AggregationRequest) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "sensor_data")
			|> filter(fn: (r) => r.survey_point_id == "%s")
			|> filter(fn: (r) => r._field == "%s")
			|> aggregateWindow(every: %s, fn: %s, createEmpty: false)`,
		r.bucket,
		req.StartTime.Format(time.RFC3339),
		req.EndTime.Format(time.RFC3339),
		req.SurveyPointID.String(),
		req.Field,
		req.Window,
		req.Aggregation,
	)

	result, err := r.influx.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	records, ok := result.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected query result type")
	}

	return records, nil
}

func (r *repository) CreateCommand(ctx context.Context, surveyPointID uuid.UUID, deviceName, command string) (*CommandOperationResult, error) {
	var result CommandOperationResult
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM create_device_command(?, ?, ?)", surveyPointID, deviceName, command).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) UpdateCommandStatus(ctx context.Context, commandID uuid.UUID, status string) (*CommandOperationResult, error) {
	var result CommandOperationResult
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM update_command_status(?, ?)", commandID, status).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) GetPendingCommands(ctx context.Context, limit int) ([]*CommandInfo, error) {
	var commands []*CommandInfo
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_pending_commands(?)", limit).
		Scan(&commands).Error
	if err != nil {
		return nil, err
	}
	return commands, nil
}

func (r *repository) GetCommandHistory(ctx context.Context, surveyPointID *uuid.UUID, deviceName *string, limit int) ([]*CommandInfo, error) {
	var commands []*CommandInfo
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_command_history(?, ?, ?)", surveyPointID, deviceName, limit).
		Scan(&commands).Error
	if err != nil {
		return nil, err
	}
	return commands, nil
}

func (r *repository) GetCommandByID(ctx context.Context, commandID uuid.UUID) (*DeviceCommand, error) {
	var command DeviceCommand
	err := r.db.WithContext(ctx).
		Where("id = ?", commandID).
		First(&command).Error
	if err != nil {
		return nil, err
	}
	return &command, nil
}
