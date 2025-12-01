package surveyPoint

import (
	"time"

	"github.com/google/uuid"
)

type SurveyPoint struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MCUID       uuid.UUID `json:"mcu_id" gorm:"column:mcu_id;type:uuid;not null"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description *string   `json:"description,omitempty" gorm:"type:text"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'connecting'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SurveyPoint) TableName() string {
	return "tbl_survey_points"
}

type SurveyPointInfo struct {
	SurveyPointID   uuid.UUID `json:"survey_point_id"`
	SurveyPointName string    `json:"survey_point_name"`
	Description     *string   `json:"description"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateSurveyPointRequest struct {
	MCUID       uuid.UUID `json:"mcu_id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description *string   `json:"description,omitempty"`
}

type UpdateSurveyPointRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=connecting connected disconnected"`
}

type SurveyPointOperationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
