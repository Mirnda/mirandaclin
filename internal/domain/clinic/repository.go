package clinic

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, c *Clinic) error
	FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*Clinic, error)
	List(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) ([]Clinic, error)
	Update(ctx context.Context, db *gorm.DB, c *Clinic) error
	SoftDelete(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) error
}
