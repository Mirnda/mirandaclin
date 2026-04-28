package profile

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, p *Profile) error
	FindByUserID(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) (*Profile, error)
	Update(ctx context.Context, db *gorm.DB, p *Profile) error
}
