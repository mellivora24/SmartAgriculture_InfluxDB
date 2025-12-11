package mcu

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, req *CreateMCURequest) (*MCU, error)
	GetByID(ctx context.Context, id uuid.UUID) (*MCU, error)
	GetByCode(ctx context.Context, mcuCode string) (*MCUInfo, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*MCU, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetFarmMCUs(ctx context.Context, farmID uuid.UUID) ([]*MCUWithDetails, error)
	UpdateStatus(ctx context.Context, mcuCode, status string) (*MCUOperationResult, error)
	ListByFarm(ctx context.Context, farmID uuid.UUID) ([]*MCU, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*MCU, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateMCURequest) (*MCU, error) {
	mcu := &MCU{
		FarmID:  req.FarmID,
		MCUCode: req.MCUCode,
		Status:  "online",
	}

	if err := s.repo.Create(ctx, mcu); err != nil {
		return nil, fmt.Errorf("failed to create mcu: %w", err)
	}

	return mcu, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*MCU, error) {
	mcu, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mcu not found")
		}
		return nil, fmt.Errorf("failed to get mcu: %w", err)
	}

	return mcu, nil
}

func (s *service) GetByCode(ctx context.Context, mcuCode string) (*MCUInfo, error) {
	info, err := s.repo.GetByCode(ctx, mcuCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get mcu by code: %w", err)
	}

	return info, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*MCU, error) {
	mcu, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mcu not found")
		}
		return nil, fmt.Errorf("failed to get mcu: %w", err)
	}

	if len(updates) > 0 {
		if err := s.repo.Update(ctx, id, updates); err != nil {
			return nil, fmt.Errorf("failed to update mcu: %w", err)
		}
	}

	mcu, err = s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated mcu: %w", err)
	}

	return mcu, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("mcu not found")
		}
		return fmt.Errorf("failed to get mcu: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete mcu: %w", err)
	}

	return nil
}

func (s *service) GetFarmMCUs(ctx context.Context, farmID uuid.UUID) ([]*MCUWithDetails, error) {
	mcus, err := s.repo.GetFarmMCUs(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get farm mcus: %w", err)
	}

	return mcus, nil
}

func (s *service) UpdateStatus(ctx context.Context, mcuCode, status string) (*MCUOperationResult, error) {
	result, err := s.repo.UpdateStatus(ctx, mcuCode, status)
	if err != nil {
		return nil, fmt.Errorf("failed to update mcu status: %w", err)
	}

	if !result.Success {
		return nil, errors.New(result.Message)
	}

	return result, nil
}

func (s *service) ListByFarm(ctx context.Context, farmID uuid.UUID) ([]*MCU, error) {
	mcus, err := s.repo.ListByFarm(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to list mcus by farm: %w", err)
	}

	return mcus, nil
}

func (s *service) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*MCU, error) {
	mcus, err := s.repo.ListByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list mcus by status: %w", err)
	}

	return mcus, nil
}
