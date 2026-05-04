package repository

import (
	"context"
	"errors"

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

func (r *tenantRepository) FindByName(ctx context.Context, db *gorm.DB, name string) (*tenant.Tenant, error) {
	var t tenant.Tenant

	err := db.WithContext(ctx).Where("name = ?", name).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}
