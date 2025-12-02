package sensorData

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	WriteSensorData(ctx context.Context, data *SensorData) error
	QuerySensorData(ctx context.Context, req *QuerySensorDataRequest) ([]map[string]interface{}, error)
	QueryLatestData(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]map[string]interface{}, error)
	QueryAggregation(ctx context.Context, req *AggregationRequest) ([]map[string]interface{}, error)

	CreateCommand(ctx context.Context, req *CreateCommandRequest) (*CommandOperationResult, error)
	UpdateCommandStatus(ctx context.Context, commandID uuid.UUID, status string) (*CommandOperationResult, error)
	GetPendingCommands(ctx context.Context, limit int) ([]*CommandInfo, error)
	GetCommandHistory(ctx context.Context, surveyPointID *uuid.UUID, deviceName *string, limit int) ([]*CommandInfo, error)
	GetCommandByID(ctx context.Context, commandID uuid.UUID) (*DeviceCommand, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) WriteSensorData(ctx context.Context, data *SensorData) error {
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	if err := s.repo.WriteSensorData(ctx, data); err != nil {
		return fmt.Errorf("failed to write sensor data: %w", err)
	}

	return nil
}

func (s *service) QuerySensorData(ctx context.Context, req *QuerySensorDataRequest) ([]map[string]interface{}, error) {
	if req.StartTime == nil {
		t := time.Now().Add(-24 * time.Hour)
		req.StartTime = &t
	}
	if req.EndTime == nil {
		t := time.Now()
		req.EndTime = &t
	}
	if req.Limit == 0 {
		req.Limit = 100
	}

	records, err := s.repo.QuerySensorData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query sensor data: %w", err)
	}

	return records, nil
}

func (s *service) QueryLatestData(ctx context.Context, surveyPointID uuid.UUID, limit int) ([]map[string]interface{}, error) {
	if limit == 0 {
		limit = 10
	}

	records, err := s.repo.QueryLatestData(ctx, surveyPointID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest data: %w", err)
	}

	return records, nil
}

func (s *service) QueryAggregation(ctx context.Context, req *AggregationRequest) ([]map[string]interface{}, error) {
	records, err := s.repo.QueryAggregation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregation: %w", err)
	}

	return records, nil
}

func (s *service) CreateCommand(ctx context.Context, req *CreateCommandRequest) (*CommandOperationResult, error) {
	result, err := s.repo.CreateCommand(ctx, req.SurveyPointID, req.DeviceName, req.Command)
	if err != nil {
		return nil, fmt.Errorf("failed to create command: %w", err)
	}

	if !result.Success {
		return nil, errors.New(result.Message)
	}

	return result, nil
}

func (s *service) UpdateCommandStatus(ctx context.Context, commandID uuid.UUID, status string) (*CommandOperationResult, error) {
	result, err := s.repo.UpdateCommandStatus(ctx, commandID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to update command status: %w", err)
	}

	if !result.Success {
		return nil, errors.New(result.Message)
	}

	return result, nil
}

func (s *service) GetPendingCommands(ctx context.Context, limit int) ([]*CommandInfo, error) {
	if limit == 0 {
		limit = 100
	}

	commands, err := s.repo.GetPendingCommands(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending commands: %w", err)
	}

	return commands, nil
}

func (s *service) GetCommandHistory(ctx context.Context, surveyPointID *uuid.UUID, deviceName *string, limit int) ([]*CommandInfo, error) {
	if limit == 0 {
		limit = 50
	}

	commands, err := s.repo.GetCommandHistory(ctx, surveyPointID, deviceName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get command history: %w", err)
	}

	return commands, nil
}

func (s *service) GetCommandByID(ctx context.Context, commandID uuid.UUID) (*DeviceCommand, error) {
	command, err := s.repo.GetCommandByID(ctx, commandID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("command not found")
		}
		return nil, fmt.Errorf("failed to get command: %w", err)
	}

	return command, nil
}
