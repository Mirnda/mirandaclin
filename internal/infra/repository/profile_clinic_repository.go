package repository

import (
	"context"
	"errors"

	profileclinic "github.com/Mirnda/mirandaclin/internal/domain/profile_clinic"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type profileClinicRepository struct{}

func NewProfileClinicRepository() profileclinic.Repository {
	return &profileClinicRepository{}
}

func (r *profileClinicRepository) Create(ctx context.Context, db *gorm.DB, dc *profileclinic.ProfileClinic) error {
	return db.WithContext(ctx).Create(dc).Error
}

func (r *profileClinicRepository) FindByProfileAndClinic(ctx context.Context, db *gorm.DB, tenantID, profileID, clinicID uuid.UUID) (*profileclinic.ProfileClinic, error) {
	var dc profileclinic.ProfileClinic
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND profile_id = ? AND clinic_id = ?", tenantID, profileID, clinicID).
		First(&dc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &dc, err
}

func (r *profileClinicRepository) ListByProfile(ctx context.Context, db *gorm.DB, tenantID, profileID uuid.UUID) ([]profileclinic.ProfileClinic, error) {
	var items []profileclinic.ProfileClinic
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND profile_id = ?", tenantID, profileID).
		Find(&items).Error
	return items, err
}

func (r *profileClinicRepository) ListByClinic(ctx context.Context, db *gorm.DB, tenantID, clinicID uuid.UUID) ([]profileclinic.ProfileClinic, error) {
	var items []profileclinic.ProfileClinic
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND clinic_id = ?", tenantID, clinicID).
		Find(&items).Error
	return items, err
}

func (r *profileClinicRepository) Update(ctx context.Context, db *gorm.DB, dc *profileclinic.ProfileClinic) error {
	return db.WithContext(ctx).
		Where("tenant_id = ?", dc.TenantID).
		Save(dc).Error
}
