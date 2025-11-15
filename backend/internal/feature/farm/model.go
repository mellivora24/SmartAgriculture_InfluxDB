package farm

import (
	"time"

	"github.com/google/uuid"
)

type Farm struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description *string   `json:"description,omitempty" gorm:"type:text"`
	Location    *string   `json:"location,omitempty" gorm:"type:varchar(500)"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Farm) TableName() string {
	return "tbl_farms"
}

type FarmUser struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	FarmID    uuid.UUID `json:"farm_id" gorm:"type:uuid;not null"`
	Role      string    `json:"role" gorm:"type:varchar(50);default:'viewer'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (FarmUser) TableName() string {
	return "tbl_farm_users"
}

type UserFarm struct {
	FarmID           uuid.UUID `json:"farm_id"`
	FarmName         string    `json:"farm_name"`
	FarmDescription  *string   `json:"farm_description"`
	FarmLocation     *string   `json:"farm_location"`
	UserRole         string    `json:"user_role"`
	MCUCount         int64     `json:"mcu_count"`
	SurveyPointCount int64     `json:"survey_point_count"`
	OnlineMCUCount   int64     `json:"online_mcu_count"`
	CreatedAt        time.Time `json:"created_at"`
}

type FarmOverview struct {
	FarmID             uuid.UUID `json:"farm_id"`
	FarmName           string    `json:"farm_name"`
	FarmDescription    *string   `json:"farm_description"`
	FarmLocation       *string   `json:"farm_location"`
	TotalMCUs          int64     `json:"total_mcus"`
	OnlineMCUs         int64     `json:"online_mcus"`
	OfflineMCUs        int64     `json:"offline_mcus"`
	TotalSurveyPoints  int64     `json:"total_survey_points"`
	ConnectingPoints   int64     `json:"connecting_points"`
	ConnectedPoints    int64     `json:"connected_points"`
	DisconnectedPoints int64     `json:"disconnected_points"`
	CreatedAt          time.Time `json:"created_at"`
}

type FarmStructure struct {
	FarmID            uuid.UUID  `json:"farm_id"`
	FarmName          string     `json:"farm_name"`
	MCUID             *uuid.UUID `json:"mcu_id"`
	MCUCode           *string    `json:"mcu_code"`
	MCUStatus         *string    `json:"mcu_status"`
	SurveyPointID     *uuid.UUID `json:"survey_point_id"`
	SurveyPointName   *string    `json:"survey_point_name"`
	SurveyPointStatus *string    `json:"survey_point_status"`
}

type CreateFarmRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	Location    *string `json:"location,omitempty"`
}

type UpdateFarmRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Location    *string `json:"location,omitempty"`
}

type AddUserToFarmRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=owner manager viewer"`
}

type OperationResult struct {
	Success bool       `json:"success"`
	FarmID  *uuid.UUID `json:"farm_id,omitempty"`
	Message string     `json:"message"`
}
