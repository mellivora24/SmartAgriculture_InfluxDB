package mcu

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, mcu *MCU) error
	GetByID(ctx context.Context, id uuid.UUID) (*MCU, error)
	GetByCode(ctx context.Context, mcuCode string) (*MCUInfo, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetFarmMCUs(ctx context.Context, farmID uuid.UUID) ([]*MCUWithDetails, error)
	UpdateStatus(ctx context.Context, mcuCode, status string) (*MCUOperationResult, error)
	ListByFarm(ctx context.Context, farmID uuid.UUID) ([]*MCU, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*MCU, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, mcu *MCU) error {
	return r.db.WithContext(ctx).Create(mcu).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*MCU, error) {
	var mcu MCU
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&mcu).Error
	if err != nil {
		return nil, err
	}
	return &mcu, nil
}

func (r *repository) GetByCode(ctx context.Context, mcuCode string) (*MCUInfo, error) {
	var info MCUInfo
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_mcu_by_code(?)", mcuCode).
		Scan(&info).Error
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&MCU{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&MCU{}).Error
}

func (r *repository) GetFarmMCUs(ctx context.Context, farmID uuid.UUID) ([]*MCUWithDetails, error) {
	var mcus []*MCUWithDetails
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_farm_mcus(?)", farmID).
		Scan(&mcus).Error
	if err != nil {
		return nil, err
	}
	return mcus, nil
}

func (r *repository) UpdateStatus(ctx context.Context, mcuCode, status string) (*MCUOperationResult, error) {
	var result MCUOperationResult
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM update_mcu_status(?, ?)", mcuCode, status).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) ListByFarm(ctx context.Context, farmID uuid.UUID) ([]*MCU, error) {
	var mcus []*MCU
	err := r.db.WithContext(ctx).
		Where("farm_id = ?", farmID).
		Order("mcu_code ASC").
		Find(&mcus).Error
	if err != nil {
		return nil, err
	}
	return mcus, nil
}

func (r *repository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*MCU, error) {
	var mcus []*MCU
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("updated_at DESC").
		Find(&mcus).Error
	if err != nil {
		return nil, err
	}
	return mcus, nil
}
