package consultation

import (
	"context"
	"errors"

	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
	log  logger.Logger
}

func NewService(db *gorm.DB, r Repository, log logger.Logger) *Service {
	return &Service{db: db, repo: r, log: log}
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
		s.log.Error("erro ao criar consulta", zap.String("tenant_id", req.TenantID.String()), zap.Error(err))
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
