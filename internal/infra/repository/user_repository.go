package repository

import (
	"context"
	"errors"

	"github.com/Mirnda/mirandaclin/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct{}

func NewUserRepository() user.Repository {
	return &userRepository{}
}

func (r *userRepository) Create(ctx context.Context, db *gorm.DB, u *user.User) error {
	return db.WithContext(ctx).Create(u).Error
}

func (r *userRepository) FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*user.User, error) {
	var u user.User
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *userRepository) FindByEmail(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, email string) (*user.User, error) {
	var u user.User
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND email = ?", tenantID, email).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *userRepository) Update(ctx context.Context, db *gorm.DB, u *user.User) error {
	return db.WithContext(ctx).
		Where("tenant_id = ?", u.TenantID).
		Save(u).Error
}

func (r *userRepository) SoftDelete(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) error {
	return db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&user.User{}).Error
}
