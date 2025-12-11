package surveyPoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, req *CreateSurveyPointRequest) (*SurveyPoint, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SurveyPoint, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateSurveyPointRequest) (*SurveyPoint, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetMCUSurveyPoints(ctx context.Context, mcuID uuid.UUID) ([]*SurveyPointInfo, error)
	UpdateStatus(ctx context.Context, surveyPointID uuid.UUID, status string) (*SurveyPointOperationResult, error)
	ListByMCU(ctx context.Context, mcuID uuid.UUID) ([]*SurveyPoint, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*SurveyPoint, error)
	GetOwnerUserID(ctx context.Context, surveyPointID uuid.UUID) (uuid.UUID, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateSurveyPointRequest) (*SurveyPoint, error) {
	surveyPoint := &SurveyPoint{
		MCUID:       req.MCUID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "connecting",
	}

	if err := s.repo.Create(ctx, surveyPoint); err != nil {
		return nil, fmt.Errorf("failed to create survey point: %w", err)
	}

	return surveyPoint, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*SurveyPoint, error) {
	surveyPoint, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("survey point not found")
		}
		return nil, fmt.Errorf("failed to get survey point: %w", err)
	}

	return surveyPoint, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *UpdateSurveyPointRequest) (*SurveyPoint, error) {
	surveyPoint, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("survey point not found")
		}
		return nil, fmt.Errorf("failed to get survey point: %w", err)
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := s.repo.Update(ctx, id, updates); err != nil {
			return nil, fmt.Errorf("failed to update survey point: %w", err)
		}
	}

	surveyPoint, err = s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated survey point: %w", err)
	}

	return surveyPoint, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("survey point not found")
		}
		return fmt.Errorf("failed to get survey point: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete survey point: %w", err)
	}

	return nil
}

func (s *service) GetMCUSurveyPoints(ctx context.Context, mcuID uuid.UUID) ([]*SurveyPointInfo, error) {
	surveyPoints, err := s.repo.GetMCUSurveyPoints(ctx, mcuID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mcu survey points: %w", err)
	}

	return surveyPoints, nil
}

func (s *service) UpdateStatus(ctx context.Context, surveyPointID uuid.UUID, status string) (*SurveyPointOperationResult, error) {
	result, err := s.repo.UpdateStatus(ctx, surveyPointID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to update survey point status: %w", err)
	}

	if !result.Success {
		return nil, errors.New(result.Message)
	}

	return result, nil
}

func (s *service) ListByMCU(ctx context.Context, mcuID uuid.UUID) ([]*SurveyPoint, error) {
	surveyPoints, err := s.repo.ListByMCU(ctx, mcuID)
	if err != nil {
		return nil, fmt.Errorf("failed to list survey points by mcu: %w", err)
	}

	return surveyPoints, nil
}

func (s *service) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*SurveyPoint, error) {
	surveyPoints, err := s.repo.ListByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list survey points by status: %w", err)
	}

	return surveyPoints, nil
}

func (s *service) GetOwnerUserID(ctx context.Context, surveyPointID uuid.UUID) (uuid.UUID, error) {
	userID, err := s.repo.GetOwnerUserID(ctx, surveyPointID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get owner user id: %w", err)
	}

	return userID, nil
}
