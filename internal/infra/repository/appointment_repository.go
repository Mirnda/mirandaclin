package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/appointment"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type appointmentRepository struct{}

func NewAppointmentRepository() appointment.Repository {
	return &appointmentRepository{}
}

func (r *appointmentRepository) Create(ctx context.Context, db *gorm.DB, a *appointment.Appointment) error {
	return db.WithContext(ctx).Create(a).Error
}

func (r *appointmentRepository) FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*appointment.Appointment, error) {
	var a appointment.Appointment
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &a, err
}

func (r *appointmentRepository) ListByPatient(ctx context.Context, db *gorm.DB, tenantID, patientID uuid.UUID) ([]appointment.Appointment, error) {
	var items []appointment.Appointment
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND patient_id = ?", tenantID, patientID).
		Order("scheduled_at DESC").
		Find(&items).Error
	return items, err
}

func (r *appointmentRepository) ListByDentist(ctx context.Context, db *gorm.DB, tenantID, dentistID uuid.UUID, date time.Time) ([]appointment.Appointment, error) {
	var items []appointment.Appointment
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND dentist_id = ? AND scheduled_at >= ? AND scheduled_at < ?", tenantID, dentistID, start, end).
		Order("scheduled_at ASC").
		Find(&items).Error
	return items, err
}

func (r *appointmentRepository) ListByClinic(ctx context.Context, db *gorm.DB, tenantID, clinicID uuid.UUID, date time.Time) ([]appointment.Appointment, error) {
	var items []appointment.Appointment
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND clinic_id = ? AND scheduled_at >= ? AND scheduled_at < ?", tenantID, clinicID, start, end).
		Order("scheduled_at ASC").
		Find(&items).Error
	return items, err
}

func (r *appointmentRepository) Update(ctx context.Context, db *gorm.DB, a *appointment.Appointment) error {
	return db.WithContext(ctx).
		Where("tenant_id = ?", a.TenantID).
		Save(a).Error
}
