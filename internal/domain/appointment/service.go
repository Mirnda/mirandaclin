package appointment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	dentistblock "github.com/mirandev/mirandaclin/internal/domain/dentist_block"
	dentistclinic "github.com/mirandev/mirandaclin/internal/domain/dentist_clinic"
	"github.com/mirandev/mirandaclin/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrDentistNotActive    = errors.New("dentista não está ativo nesta clínica")
	ErrOutsideWorkingDays  = errors.New("horário fora dos dias de trabalho do dentista")
	ErrDentistBlocked      = errors.New("dentista indisponível neste horário")
	ErrAppointmentNotFound = errors.New("agendamento não encontrado")
)

type CreateRequest struct {
	TenantID    uuid.UUID
	PatientID   uuid.UUID
	DentistID   uuid.UUID
	ClinicID    uuid.UUID
	SecretaryID *uuid.UUID
	ScheduledAt time.Time
	Notes       string
}

type Service struct {
	db              *gorm.DB
	appointmentRepo Repository
	dcRepo          dentistclinic.Repository
	blockRepo       dentistblock.Repository
	log             logger.Logger
}

func NewService(db *gorm.DB, ar Repository, dcr dentistclinic.Repository, br dentistblock.Repository, log logger.Logger) *Service {
	return &Service{db: db, appointmentRepo: ar, dcRepo: dcr, blockRepo: br, log: log}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Appointment, error) {
	dc, err := s.dcRepo.FindByDentistAndClinic(ctx, s.db, req.TenantID, req.DentistID, req.ClinicID)
	if err != nil {
		return nil, err
	}
	if dc == nil || !dc.Active {
		return nil, ErrDentistNotActive
	}

	weekday := strings.ToLower(req.ScheduledAt.Weekday().String())
	if !containsDay(dc.WorkingDays, weekday) {
		return nil, ErrOutsideWorkingDays
	}

	slotStart := req.ScheduledAt.Format("15:04")
	slotEnd := req.ScheduledAt.Add(time.Duration(dc.SlotDurationMinutes) * time.Minute).Format("15:04")

	blocks, err := s.blockRepo.FindBlocksForSlot(ctx, s.db, req.TenantID, req.DentistID, &req.ClinicID, req.ScheduledAt, slotStart, slotEnd)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		return nil, ErrDentistBlocked
	}

	a := &Appointment{
		TenantID:    req.TenantID,
		PatientID:   req.PatientID,
		DentistID:   req.DentistID,
		ClinicID:    req.ClinicID,
		SecretaryID: req.SecretaryID,
		ScheduledAt: req.ScheduledAt,
		Notes:       req.Notes,
		Status:      StatusScheduled,
	}
	if err := s.appointmentRepo.Create(ctx, s.db, a); err != nil {
		s.log.Error("erro ao criar agendamento", zap.String("tenant_id", req.TenantID.String()), zap.Error(err))
		return nil, err
	}
	return a, nil
}

func (s *Service) Cancel(ctx context.Context, tenantID, id uuid.UUID) error {
	a, err := s.appointmentRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return err
	}
	if a == nil {
		return ErrAppointmentNotFound
	}
	now := time.Now()
	a.Status = StatusCancelled
	a.CanceledAt = &now
	return s.appointmentRepo.Update(ctx, s.db, a)
}

func (s *Service) Complete(ctx context.Context, tenantID, id uuid.UUID) error {
	a, err := s.appointmentRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return err
	}
	if a == nil {
		return ErrAppointmentNotFound
	}
	a.Status = StatusCompleted
	return s.appointmentRepo.Update(ctx, s.db, a)
}

func (s *Service) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]Appointment, error) {
	return s.appointmentRepo.ListByPatient(ctx, s.db, tenantID, patientID)
}

func (s *Service) ListByDentist(ctx context.Context, tenantID, dentistID uuid.UUID, date time.Time) ([]Appointment, error) {
	return s.appointmentRepo.ListByDentist(ctx, s.db, tenantID, dentistID, date)
}

func containsDay(days []string, day string) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}
