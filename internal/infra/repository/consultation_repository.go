package repository

import (
	"context"
	"errors"

	"github.com/Mirnda/mirandaclin/internal/domain/consultation"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type consultationRepository struct{}

func NewConsultationRepository() consultation.Repository {
	return &consultationRepository{}
}

func (r *consultationRepository) Create(ctx context.Context, db *gorm.DB, c *consultation.Consultation) error {
	return db.WithContext(ctx).Create(c).Error
}

func (r *consultationRepository) FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*consultation.Consultation, error) {
	var c consultation.Consultation
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *consultationRepository) ListByPatient(ctx context.Context, db *gorm.DB, tenantID, patientID uuid.UUID) ([]consultation.Consultation, error) {
	var items []consultation.Consultation
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND patient_id = ?", tenantID, patientID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *consultationRepository) ListByDentist(ctx context.Context, db *gorm.DB, tenantID, dentistID uuid.UUID) ([]consultation.Consultation, error) {
	var items []consultation.Consultation
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND dentist_id = ?", tenantID, dentistID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}
