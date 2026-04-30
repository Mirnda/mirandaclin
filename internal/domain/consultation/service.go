package consultation

import (
	"context"
	"errors"

	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrConsultationNotFound = errors.New("consulta não encontrada")

type CreateRequest struct {
	TenantID      uuid.UUID
	AppointmentID uuid.UUID
	PatientID     uuid.UUID
	DentistID     uuid.UUID
	Diagnosis     string
	Treatment     string
}

type Service struct {
	db   *gorm.DB
	repo Repository
}

func NewService(db *gorm.DB, r Repository) *Service {
	return &Service{db: db, repo: r}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Consultation, error) {
	c := &Consultation{
		TenantID:      req.TenantID,
		AppointmentID: req.AppointmentID,
		PatientID:     req.PatientID,
		DentistID:     req.DentistID,
		Diagnosis:     req.Diagnosis,
		Treatment:     req.Treatment,
	}
	if err := s.repo.Create(ctx, s.db, c); err != nil {
		logger.FromContext(ctx).Error("erro ao criar consulta",
			logger.String("tenant_id", req.TenantID.String()),
			logger.Err(err),
		)
		return nil, err
	}
	return c, nil
}

func (s *Service) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]Consultation, error) {
	return s.repo.ListByPatient(ctx, s.db, tenantID, patientID)
}

func (s *Service) ListByDentist(ctx context.Context, tenantID, dentistID uuid.UUID) ([]Consultation, error) {
	return s.repo.ListByDentist(ctx, s.db, tenantID, dentistID)
}
