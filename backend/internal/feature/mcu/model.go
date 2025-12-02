package mcu

import (
	"time"

	"github.com/google/uuid"
)

type MCU struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FarmID    uuid.UUID `json:"farm_id" gorm:"type:uuid;not null"`
	MCUCode   string    `json:"mcu_code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Status    string    `json:"status" gorm:"type:varchar(50);default:'offline'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (MCU) TableName() string {
	return "tbl_mcus"
}

type MCUWithDetails struct {
	MCUID            uuid.UUID `json:"mcu_id" gorm:"column:mcu_id"`
	MCUCode          string    `json:"mcu_code" gorm:"column:mcu_code"`
	Status           string    `json:"status" gorm:"column:status"`
	SurveyPointCount int64     `json:"survey_point_count" gorm:"column:survey_point_count"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at"`
}

type MCUInfo struct {
	MCUID            uuid.UUID `json:"mcu_id"`
	MCUCode          string    `json:"mcu_code"`
	FarmID           uuid.UUID `json:"farm_id"`
	FarmName         string    `json:"farm_name"`
	Status           string    `json:"status"`
	SurveyPointCount int64     `json:"survey_point_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateMCURequest struct {
	FarmID  uuid.UUID `json:"farm_id" validate:"required"`
	MCUCode string    `json:"mcu_code" validate:"required"`
}

type UpdateMCUStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=online offline"`
}

type MCUOperationResult struct {
	Success bool       `json:"success"`
	MCUID   *uuid.UUID `json:"mcu_id,omitempty"`
	Message string     `json:"message"`
}
