package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var result struct {
		UserID       uuid.UUID `gorm:"column:user_id"`
		Username     string    `gorm:"column:username"`
		Email        string    `gorm:"column:email"`
		Phone        *string   `gorm:"column:phone"`
		PasswordHash string    `gorm:"column:password_hash"`
		FullName     *string   `gorm:"column:full_name"`
		CreatedAt    string    `gorm:"column:created_at"`
	}

	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_user_by_username(?)", username).
		Scan(&result).Error

	if err != nil || result.UserID == uuid.Nil {
		return nil, err
	}

	return &User{
		ID:           result.UserID,
		Username:     result.Username,
		Email:        result.Email,
		Phone:        result.Phone,
		PasswordHash: result.PasswordHash,
		FullName:     result.FullName,
	}, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var result struct {
		UserID       uuid.UUID `gorm:"column:user_id"`
		Username     string    `gorm:"column:username"`
		Email        string    `gorm:"column:email"`
		Phone        *string   `gorm:"column:phone"`
		PasswordHash string    `gorm:"column:password_hash"`
		FullName     *string   `gorm:"column:full_name"`
		CreatedAt    string    `gorm:"column:created_at"`
	}

	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM get_user_by_email(?)", email).
		Scan(&result).Error

	if err != nil || result.UserID == uuid.Nil {
		return nil, err
	}

	return &User{
		ID:           result.UserID,
		Username:     result.Username,
		Email:        result.Email,
		Phone:        result.Phone,
		PasswordHash: result.PasswordHash,
		FullName:     result.FullName,
	}, nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&User{}).Error
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	var users []*User
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
