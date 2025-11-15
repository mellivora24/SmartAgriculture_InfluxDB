package farm

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, farm *Farm) error
	GetByID(ctx context.Context, id uuid.UUID) (*Farm, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Farm, error)

	GetUserFarms(ctx context.Context, userID uuid.UUID) ([]*UserFarm, error)
	GetFarmOverview(ctx context.Context, farmID uuid.UUID) (*FarmOverview, error)
	GetFarmStructure(ctx context.Context, farmID uuid.UUID) ([]*FarmStructure, error)
	CreateFarmWithOwner(ctx context.Context, userID uuid.UUID, name string, description, location *string) (*OperationResult, error)
	AddUserToFarm(ctx context.Context, userID, farmID uuid.UUID, role string) (*OperationResult, error)
	RemoveUserFromFarm(ctx context.Context, userID, farmID uuid.UUID) (*OperationResult, error)
	CheckUserPermission(ctx context.Context, userID, farmID uuid.UUID, requiredRole string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, farm *Farm) error {
	return r.db.WithContext(ctx).Create(farm).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Farm, error) {
	var farm Farm
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&farm).Error
	if err != nil {
		return nil, err
	}
	return &farm, nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&Farm{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&Farm{}).Error
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]*Farm, error) {
	var farms []*Farm
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&farms).Error
	if err != nil {
		return nil, err
	}
	return farms, nil
}

func (r *repository) GetUserFarms(ctx context.Context, userID uuid.UUID) ([]*UserFarm, error) {
	var farms []*UserFarm
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_user_farms(?)", userID).
		Scan(&farms).Error
	if err != nil {
		return nil, err
	}
	return farms, nil
}

func (r *repository) GetFarmOverview(ctx context.Context, farmID uuid.UUID) (*FarmOverview, error) {
	var overview FarmOverview
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_farm_overview(?)", farmID).
		Scan(&overview).Error
	if err != nil {
		return nil, err
	}
	return &overview, nil
}

func (r *repository) GetFarmStructure(ctx context.Context, farmID uuid.UUID) ([]*FarmStructure, error) {
	var structure []*FarmStructure
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_farm_structure(?)", farmID).
		Scan(&structure).Error
	if err != nil {
		return nil, err
	}
	return structure, nil
}

func (r *repository) CreateFarmWithOwner(ctx context.Context, userID uuid.UUID, name string, description, location *string) (*OperationResult, error) {
	var result OperationResult
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM create_farm_with_owner(?, ?, ?, ?)", userID, name, description, location).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) AddUserToFarm(ctx context.Context, userID, farmID uuid.UUID, role string) (*OperationResult, error) {
	var result OperationResult
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM add_user_to_farm(?, ?, ?)", userID, farmID, role).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) RemoveUserFromFarm(ctx context.Context, userID, farmID uuid.UUID) (*OperationResult, error) {
	var result OperationResult
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM remove_user_from_farm(?, ?)", userID, farmID).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) CheckUserPermission(ctx context.Context, userID, farmID uuid.UUID, requiredRole string) (bool, error) {
	var hasPermission bool
	err := r.db.WithContext(ctx).
		Raw("SELECT check_user_farm_permission(?, ?, ?)", userID, farmID, requiredRole).
		Scan(&hasPermission).Error
	if err != nil {
		return false, err
	}
	return hasPermission, nil
}
