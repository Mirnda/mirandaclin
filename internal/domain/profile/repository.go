package profile

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, p *Profile) error
	FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*Profile, error)
	FindByUserAndTenant(ctx context.Context, db *gorm.DB, userID, tenantID uuid.UUID) (*Profile, error)
	// FindByUserID retorna todos os perfis do usuário (multi-tenant login).
	FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) ([]*Profile, error)
	List(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) ([]Profile, error)
	ListCollaborators(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) ([]Profile, error)
	ListByRole(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, role string) ([]Profile, error)
	Update(ctx context.Context, db *gorm.DB, p *Profile) error
	SoftDelete(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) error
}
