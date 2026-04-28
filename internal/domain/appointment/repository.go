package appointment

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, a *Appointment) error
	FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*Appointment, error)
	ListByPatient(ctx context.Context, db *gorm.DB, tenantID, patientID uuid.UUID) ([]Appointment, error)
	ListByDentist(ctx context.Context, db *gorm.DB, tenantID, dentistID uuid.UUID, date time.Time) ([]Appointment, error)
	ListByClinic(ctx context.Context, db *gorm.DB, tenantID, clinicID uuid.UUID, date time.Time) ([]Appointment, error)
	Update(ctx context.Context, db *gorm.DB, a *Appointment) error
}
