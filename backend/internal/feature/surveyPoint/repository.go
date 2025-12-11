package surveyPoint

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, surveyPoint *SurveyPoint) error
	GetByID(ctx context.Context, id uuid.UUID) (*SurveyPoint, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetMCUSurveyPoints(ctx context.Context, mcuID uuid.UUID) ([]*SurveyPointInfo, error)
	UpdateStatus(ctx context.Context, surveyPointID uuid.UUID, status string) (*SurveyPointOperationResult, error)
	ListByMCU(ctx context.Context, mcuID uuid.UUID) ([]*SurveyPoint, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*SurveyPoint, error)
	GetOwnerUserID(ctx context.Context, surveyPointID uuid.UUID) (uuid.UUID, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, surveyPoint *SurveyPoint) error {
	return r.db.WithContext(ctx).Create(surveyPoint).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*SurveyPoint, error) {
	var surveyPoint SurveyPoint
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&surveyPoint).Error
	if err != nil {
		return nil, err
	}
	return &surveyPoint, nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&SurveyPoint{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&SurveyPoint{}).Error
}

func (r *repository) GetMCUSurveyPoints(ctx context.Context, mcuID uuid.UUID) ([]*SurveyPointInfo, error) {
	var surveyPoints []*SurveyPointInfo
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_mcu_survey_points(?)", mcuID).
		Scan(&surveyPoints).Error
	if err != nil {
		return nil, err
	}
	return surveyPoints, nil
}

func (r *repository) UpdateStatus(ctx context.Context, surveyPointID uuid.UUID, status string) (*SurveyPointOperationResult, error) {
	var result SurveyPointOperationResult
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM update_survey_point_status(?, ?)", surveyPointID, status).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) ListByMCU(ctx context.Context, mcuID uuid.UUID) ([]*SurveyPoint, error) {
	var surveyPoints []*SurveyPoint
	err := r.db.WithContext(ctx).
		Where("mcu_id = ?", mcuID).
		Order("name ASC").
		Find(&surveyPoints).Error
	if err != nil {
		return nil, err
	}
	return surveyPoints, nil
}

func (r *repository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*SurveyPoint, error) {
	var surveyPoints []*SurveyPoint
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("updated_at DESC").
		Find(&surveyPoints).Error
	if err != nil {
		return nil, err
	}
	return surveyPoints, nil
}

func (r *repository) GetOwnerUserID(ctx context.Context, surveyPointID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.WithContext(ctx).
		Raw("SELECT get_survey_point_owner_user_id(?) AS user_id", surveyPointID).
		Scan(&userID).Error
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}
