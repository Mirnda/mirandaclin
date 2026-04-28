package consultation

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, c *Consultation) error
	FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*Consultation, error)
	ListByPatient(ctx context.Context, db *gorm.DB, tenantID, patientID uuid.UUID) ([]Consultation, error)
	ListByDentist(ctx context.Context, db *gorm.DB, tenantID, dentistID uuid.UUID) ([]Consultation, error)
}
