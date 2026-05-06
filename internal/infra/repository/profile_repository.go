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

func (r *profileRepository) FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*profile.Profile, error) {
	var p profile.Profile
	err := db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *profileRepository) FindByUserAndTenant(ctx context.Context, db *gorm.DB, userID, tenantID uuid.UUID) (*profile.Profile, error) {
	var p profile.Profile
	err := db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND deleted_at IS NULL", userID, tenantID).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *profileRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) ([]*profile.Profile, error) {
	var profiles []*profile.Profile
	err := db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Find(&profiles).Error
	return profiles, err
}

func (r *profileRepository) List(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) ([]profile.Profile, error) {
	var profiles []profile.Profile
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Find(&profiles).Error
	return profiles, err
}

func (r *profileRepository) ListByRole(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, role string) ([]profile.Profile, error) {
	var profiles []profile.Profile
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND role = ? AND deleted_at IS NULL", tenantID, role).
		Find(&profiles).Error
	return profiles, err
}

func (r *profileRepository) Update(ctx context.Context, db *gorm.DB, p *profile.Profile) error {
	return db.WithContext(ctx).
		Where("tenant_id = ?", p.TenantID).
		Save(p).Error
}

func (r *profileRepository) SoftDelete(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) error {
	return db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&profile.Profile{}).Error
}
