package dentistclinic

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, dc *DentistClinic) error
	FindByDentistAndClinic(ctx context.Context, db *gorm.DB, tenantID, dentistID, clinicID uuid.UUID) (*DentistClinic, error)
	ListByDentist(ctx context.Context, db *gorm.DB, tenantID, dentistID uuid.UUID) ([]DentistClinic, error)
	ListByClinic(ctx context.Context, db *gorm.DB, tenantID, clinicID uuid.UUID) ([]DentistClinic, error)
	Update(ctx context.Context, db *gorm.DB, dc *DentistClinic) error
}
