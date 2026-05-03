package tenant_member

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, m *TenantMember) error
	FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) ([]*TenantMember, error)
	FindByUserAndTenant(ctx context.Context, db *gorm.DB, userID, tenantID uuid.UUID) (*TenantMember, error)
}
