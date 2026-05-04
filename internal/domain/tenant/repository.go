package tenant

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, t *Tenant) error
	FindByName(ctx context.Context, db *gorm.DB, tenantName string) (*Tenant, error)
}
