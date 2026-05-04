package clinic

import (
	"context"
	"errors"

	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var ErrClinicNotFound = errors.New("clínica não encontrada")

type Service struct {
	db   *gorm.DB
	repo Repository
}

func NewService(db *gorm.DB, r Repository) *Service {
	return &Service{db: db, repo: r}
}

func (s *Service) Create(ctx context.Context, c *Clinic) error {
	if err := s.repo.Create(ctx, s.db, c); err != nil {
		logger.FromContext(ctx).Error("erro ao criar clínica",
			logger.String("tenant_id", c.TenantID.String()),
			logger.Err(err),
		)
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

type UpdateRequest struct {
	Name          *string
	Phone         *string
	Address       *shared.AddressInput
	OperatingDays *[]string
	OpenTime      *string
	CloseTime     *string
}

func (s *Service) Update(ctx context.Context, tenantID, id uuid.UUID, req UpdateRequest) (*Clinic, error) {
	c, err := s.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Phone != nil {
		c.Phone = *req.Phone
	}
	if req.Address != nil {
		c.Address = req.Address.ToAddress()
	}
	if req.OperatingDays != nil {
		c.OperatingDays = pq.StringArray(*req.OperatingDays)
	}
	if req.OpenTime != nil {
		c.OpenTime = *req.OpenTime
	}
	if req.CloseTime != nil {
		c.CloseTime = *req.CloseTime
	}

	if err := s.repo.Update(ctx, s.db, c); err != nil {
		logger.FromContext(ctx).Error("erro ao atualizar clínica",
			logger.String("tenant_id", tenantID.String()),
			logger.String("clinic_id", id.String()),
			logger.Err(err),
		)
		return nil, err
	}

	return c, nil
}

func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, s.db, tenantID, id)
}
