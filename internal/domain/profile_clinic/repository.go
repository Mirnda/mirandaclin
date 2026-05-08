package profileclinic

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, pc *ProfileClinic) error
	FindByProfileAndClinic(ctx context.Context, db *gorm.DB, tenantID, profileID, clinicID uuid.UUID) (*ProfileClinic, error)
	ListByProfile(ctx context.Context, db *gorm.DB, tenantID, profileID uuid.UUID) ([]ProfileClinic, error)
	ListByClinic(ctx context.Context, db *gorm.DB, tenantID, clinicID uuid.UUID) ([]ProfileClinic, error)
	Update(ctx context.Context, db *gorm.DB, pc *ProfileClinic) error
}
