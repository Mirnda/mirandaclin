package repository

import (
	"context"
	"errors"

	dentistclinic "github.com/Mirnda/mirandaclin/internal/domain/dentist_clinic"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dentistClinicRepository struct{}

func NewDentistClinicRepository() dentistclinic.Repository {
	return &dentistClinicRepository{}
}

func (r *dentistClinicRepository) Create(ctx context.Context, db *gorm.DB, dc *dentistclinic.DentistClinic) error {
	return db.WithContext(ctx).Create(dc).Error
}

func (r *dentistClinicRepository) FindByDentistAndClinic(ctx context.Context, db *gorm.DB, tenantID, dentistID, clinicID uuid.UUID) (*dentistclinic.DentistClinic, error) {
	var dc dentistclinic.DentistClinic
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND dentist_id = ? AND clinic_id = ?", tenantID, dentistID, clinicID).
		First(&dc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &dc, err
}

func (r *dentistClinicRepository) ListByDentist(ctx context.Context, db *gorm.DB, tenantID, dentistID uuid.UUID) ([]dentistclinic.DentistClinic, error) {
	var items []dentistclinic.DentistClinic
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND dentist_id = ?", tenantID, dentistID).
		Find(&items).Error
	return items, err
}

func (r *dentistClinicRepository) ListByClinic(ctx context.Context, db *gorm.DB, tenantID, clinicID uuid.UUID) ([]dentistclinic.DentistClinic, error) {
	var items []dentistclinic.DentistClinic
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND clinic_id = ?", tenantID, clinicID).
		Find(&items).Error
	return items, err
}

func (r *dentistClinicRepository) Update(ctx context.Context, db *gorm.DB, dc *dentistclinic.DentistClinic) error {
	return db.WithContext(ctx).
		Where("tenant_id = ?", dc.TenantID).
		Save(dc).Error
}
