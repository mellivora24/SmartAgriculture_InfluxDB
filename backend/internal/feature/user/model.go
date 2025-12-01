package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone        *string   `json:"phone,omitempty" gorm:"type:varchar(20)"`
	PasswordHash string    `json:"-" gorm:"type:varchar(255);not null"`
	FullName     *string   `json:"full_name,omitempty" gorm:"type:varchar(255)"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "tbl_users"
}

func (u User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
		FullName:  u.FullName,
		CreatedAt: u.CreatedAt,
	}
}

type CreateUserRequest struct {
	Username string  `json:"username" validate:"required,min=3,max=100"`
	Email    string  `json:"email" validate:"required,email"`
	Phone    *string `json:"phone,omitempty"`
	Password string  `json:"password" validate:"required,min=6"`
	FullName *string `json:"full_name,omitempty"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone    *string `json:"phone,omitempty"`
	FullName *string `json:"full_name,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	FullName  *string   `json:"full_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Token     string    `json:"token,omitempty"`
}
