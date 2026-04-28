package repository

import (
	"context"
	"errors"

	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type profileRepository struct{}

func NewProfileRepository() profile.Repository {
	return &profileRepository{}
}

func (r *profileRepository) Create(ctx context.Context, db *gorm.DB, p *profile.Profile) error {
	return db.WithContext(ctx).Create(p).Error
}

func (r *profileRepository) FindByUserID(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) (*profile.Profile, error) {
	var p profile.Profile
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *profileRepository) Update(ctx context.Context, db *gorm.DB, p *profile.Profile) error {
	return db.WithContext(ctx).
		Where("tenant_id = ?", p.TenantID).
		Save(p).Error
}
