package farm

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, req *CreateFarmRequest) (*OperationResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Farm, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateFarmRequest) (*Farm, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Farm, error)

	GetUserFarms(ctx context.Context, userID uuid.UUID) ([]*UserFarm, error)
	GetFarmOverview(ctx context.Context, farmID uuid.UUID) (*FarmOverview, error)
	GetFarmStructure(ctx context.Context, farmID uuid.UUID) ([]*FarmStructure, error)
	AddUserToFarm(ctx context.Context, farmID uuid.UUID, req *AddUserToFarmRequest) (*OperationResult, error)
	RemoveUserFromFarm(ctx context.Context, farmID, userID uuid.UUID) (*OperationResult, error)
	CheckUserPermission(ctx context.Context, userID, farmID uuid.UUID, requiredRole string) (bool, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, req *CreateFarmRequest) (*OperationResult, error) {
	result, err := s.repo.CreateFarmWithOwner(ctx, userID, req.Name, req.Description, req.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to create farm: %w", err)
	}

	if !result.Success {
		return nil, errors.New(result.Message)
	}

	return result, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Farm, error) {
	farm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("farm not found")
		}
		return nil, fmt.Errorf("failed to get farm: %w", err)
	}

	return farm, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *UpdateFarmRequest) (*Farm, error) {
	farm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("farm not found")
		}
		return nil, fmt.Errorf("failed to get farm: %w", err)
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}

	if len(updates) > 0 {
		if err := s.repo.Update(ctx, id, updates); err != nil {
			return nil, fmt.Errorf("failed to update farm: %w", err)
		}
	}

	farm, err = s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated farm: %w", err)
	}

	return farm, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("farm not found")
		}
		return fmt.Errorf("failed to get farm: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete farm: %w", err)
	}

	return nil
}

func (s *service) List(ctx context.Context, limit, offset int) ([]*Farm, error) {
	farms, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list farms: %w", err)
	}

	return farms, nil
}

func (s *service) GetUserFarms(ctx context.Context, userID uuid.UUID) ([]*UserFarm, error) {
	farms, err := s.repo.GetUserFarms(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user farms: %w", err)
	}

	return farms, nil
}

func (s *service) GetFarmOverview(ctx context.Context, farmID uuid.UUID) (*FarmOverview, error) {
	overview, err := s.repo.GetFarmOverview(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get farm overview: %w", err)
	}

	return overview, nil
}

func (s *service) GetFarmStructure(ctx context.Context, farmID uuid.UUID) ([]*FarmStructure, error) {
	structure, err := s.repo.GetFarmStructure(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get farm structure: %w", err)
	}

	return structure, nil
}

func (s *service) AddUserToFarm(ctx context.Context, farmID uuid.UUID, req *AddUserToFarmRequest) (*OperationResult, error) {
	result, err := s.repo.AddUserToFarm(ctx, req.UserID, farmID, req.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to farm: %w", err)
	}

	if !result.Success {
		return nil, errors.New(result.Message)
	}

	return result, nil
}

func (s *service) RemoveUserFromFarm(ctx context.Context, farmID, userID uuid.UUID) (*OperationResult, error) {
	result, err := s.repo.RemoveUserFromFarm(ctx, userID, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove user from farm: %w", err)
	}

	if !result.Success {
		return nil, errors.New(result.Message)
	}

	return result, nil
}

func (s *service) CheckUserPermission(ctx context.Context, userID, farmID uuid.UUID, requiredRole string) (bool, error) {
	hasPermission, err := s.repo.CheckUserPermission(ctx, userID, farmID, requiredRole)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPermission, nil
}
