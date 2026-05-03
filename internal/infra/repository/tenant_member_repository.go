package repository

import (
	"context"
	"errors"

	tenantmember "github.com/Mirnda/mirandaclin/internal/domain/tenant_member"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tenantMemberRepository struct{}

func NewTenantMemberRepository() tenantmember.Repository {
	return &tenantMemberRepository{}
}

func (r *tenantMemberRepository) Create(ctx context.Context, db *gorm.DB, m *tenantmember.TenantMember) error {
	return db.WithContext(ctx).Create(m).Error
}

func (r *tenantMemberRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) ([]*tenantmember.TenantMember, error) {
	var members []*tenantmember.TenantMember
	err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&members).Error
	return members, err
}

func (r *tenantMemberRepository) FindByUserAndTenant(ctx context.Context, db *gorm.DB, userID, tenantID uuid.UUID) (*tenantmember.TenantMember, error) {
	var m tenantmember.TenantMember
	err := db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}
