package patient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPatientNotFound = errors.New("paciente não encontrado")
)

type Service struct {
	db          *gorm.DB
	cache       cache.Cache
	profileRepo profile.Repository
	// userRepo    Repository
	// inviteRepo  invite.Repository
	// tenantRepo  tenant.Repository
	// cfg         config.App
	// mailer      mailer.Mailer
	// jwtSecret   string
}

func NewService(db *gorm.DB, c cache.Cache, pr profile.Repository) *Service {
	return &Service{
		db:          db,
		cache:       c,
		profileRepo: pr,
		// userRepo:    ur,
		// inviteRepo:  ir,
		// tenantRepo:  tr,
		// cfg:         cfg,
		// mailer:      ml,
		// jwtSecret:   secret,
	}
}

type CreatePatientRequest struct {
	TenantID              uuid.UUID
	FullName              string
	Document              string
	BirthDate             *time.Time
	Phone                 string
	HasWhatsapp           bool
	EmergencyContactName  string
	EmergencyContactPhone string
	Address               shared.Address
}

func patientsCacheKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s:profile:patients", tenantID)
}

func (s *Service) invalidatePatientCache(ctx context.Context, tenantID uuid.UUID) {
	_ = s.cache.Del(ctx, patientsCacheKey(tenantID))

	_, _ = s.cache.DelWithIndex(ctx, "*:profile:*")
	_, _ = s.cache.DelWithIndex(ctx, "*:patients:*")
}

// CreatePatient cria um paciente (Profile sem User vinculado) e invalida o cache.
func (s *Service) CreatePatient(ctx context.Context, req CreatePatientRequest) (*profile.Profile, error) {
	p := &profile.Profile{
		TenantID:              req.TenantID,
		Role:                  shared.RolePatient,
		FullName:              req.FullName,
		Document:              req.Document,
		BirthDate:             req.BirthDate,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		Address:               req.Address,
	}
	if err := s.profileRepo.Create(ctx, s.db, p); err != nil {
		return nil, err
	}
	s.invalidatePatientCache(ctx, req.TenantID)
	return p, nil
}
