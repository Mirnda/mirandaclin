package clinic

import (
	"context"
	"errors"

	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrClinicNotFound = errors.New("clínica não encontrada")

type Service struct {
	db   *gorm.DB
	repo Repository
	log  logger.Logger
}

func NewService(db *gorm.DB, r Repository, log logger.Logger) *Service {
	return &Service{db: db, repo: r, log: log}
}

func (s *Service) Create(ctx context.Context, c *Clinic) error {
	if err := s.repo.Create(ctx, s.db, c); err != nil {
		s.log.Error("erro ao criar clínica", zap.String("tenant_id", c.TenantID.String()), zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Clinic, error) {
	c, err := s.repo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrClinicNotFound
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID) ([]Clinic, error) {
	return s.repo.List(ctx, s.db, tenantID)
}

func (s *Service) Update(ctx context.Context, c *Clinic) error {
	return s.repo.Update(ctx, s.db, c)
}

func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, s.db, tenantID, id)
}
