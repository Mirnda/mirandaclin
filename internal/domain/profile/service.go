package profile

import (
	"context"
	"errors"

	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

var ErrProfileNotFound = errors.New("perfil não encontrado")

type Service struct {
	db    *gorm.DB
	cache cache.Cache
	repo  Repository
}

func NewService(db *gorm.DB, cache cache.Cache, repo Repository) *Service {
	return &Service{db, cache, repo}
}

func (s *Service) Create(ctx context.Context, req CreateProfile) (*Profile, error) {

	if !shared.IsValidRole(req.Role) {
		return nil, shared.ErrInvalidRole
	}

	p := &Profile{
		TenantID:              req.TenantID,
		Role:                  req.Role,
		FullName:              req.FullName,
		Document:              req.Document,
		BirthDate:             req.BirthDate,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		Address:               req.Address,
	}

	s.invalidateProfileCache(ctx, req.TenantID)
	err := s.repo.Create(ctx, s.db, p)
	return p, err
}

func (s *Service) ListByRole(ctx context.Context, tenantID uuid.UUID, role string) ([]Profile, error) {
	profiles, err := s.getProfileListByRoleCache(ctx, tenantID, role)
	if err == nil && profiles != nil {
		return profiles, err
	}

	profiles, err = s.repo.ListByRole(ctx, s.db, tenantID, role)
	if err != nil {
		return nil, err
	}

	s.setProfileListByRoleCache(ctx, tenantID, role, profiles)
	return profiles, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*Profile, error) {
	p, err := s.getProfileIDCache(ctx, tenantID, id)
	if err == nil && p != nil {
		return p, err
	}

	p, err = s.repo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProfileNotFound
	}

	s.setProfileIDCache(ctx, tenantID, p)
	return p, nil
}

func (s *Service) Update(ctx context.Context, tenantID, id uuid.UUID, req UpdateProfile) (*Profile, error) {
	p, err := s.repo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, ErrProfileNotFound
	}

	req.UpdateToProfile(p)

	if err := s.repo.Update(ctx, s.db, p); err != nil {
		return nil, err
	}

	s.invalidateProfileCache(ctx, tenantID)
	return p, nil
}

func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	s.invalidateProfileCache(ctx, tenantID)
	return s.repo.SoftDelete(ctx, s.db, tenantID, id)
}
