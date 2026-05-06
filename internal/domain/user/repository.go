package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, u *User) error
	FindByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, db *gorm.DB, email string) (*User, error)
	Update(ctx context.Context, db *gorm.DB, u *User) error
	SoftDelete(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}
