package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mirandev/mirandaclin/internal/domain/clinic"
	"gorm.io/gorm"
)

type clinicRepository struct{}

func NewClinicRepository() clinic.Repository {
	return &clinicRepository{}
}

func (r *clinicRepository) Create(ctx context.Context, db *gorm.DB, c *clinic.Clinic) error {
	return db.WithContext(ctx).Create(c).Error
}

func (r *clinicRepository) FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*clinic.Clinic, error) {
	var c clinic.Clinic
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *clinicRepository) List(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) ([]clinic.Clinic, error) {
	var clinics []clinic.Clinic
	err := db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&clinics).Error
	return clinics, err
}

func (r *clinicRepository) Update(ctx context.Context, db *gorm.DB, c *clinic.Clinic) error {
	return db.WithContext(ctx).
		Where("tenant_id = ?", c.TenantID).
		Save(c).Error
}

func (r *clinicRepository) SoftDelete(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) error {
	return db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&clinic.Clinic{}).Error
}
