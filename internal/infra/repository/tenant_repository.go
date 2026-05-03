package repository

import (
	"context"

	"github.com/Mirnda/mirandaclin/internal/domain/tenant"
	"gorm.io/gorm"
)

type tenantRepository struct{}

func NewTenantRepository() tenant.Repository {
	return &tenantRepository{}
}

func (r *tenantRepository) Create(ctx context.Context, db *gorm.DB, t *tenant.Tenant) error {
	return db.WithContext(ctx).Create(t).Error
}
