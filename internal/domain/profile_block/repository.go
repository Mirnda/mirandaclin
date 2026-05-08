package profileblock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, b *ProfileBlock) error
	Delete(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) error
	FindBlocksForSlot(ctx context.Context, db *gorm.DB, tenantID, profileID uuid.UUID, clinicID *uuid.UUID, date time.Time, start, end string) ([]ProfileBlock, error)
}
