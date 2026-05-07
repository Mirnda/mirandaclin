package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/pkg/logger"

	"github.com/google/uuid"
)

var (
	ErrPatientNotFound = errors.New("paciente não encontrado")
)

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

// CreatePatient cria um paciente (Profile sem User vinculado) e invalida o cache.
func (s *Service) CreatePatient(ctx context.Context, req CreatePatientRequest) (*profile.Profile, error) {
	p := &profile.Profile{
		TenantID:              req.TenantID,
		Role:                  RolePatient,
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

// ListPatients retorna todos os pacientes do tenant (via cache).
func (s *Service) ListPatients(ctx context.Context, tenantID uuid.UUID) ([]profile.Profile, error) {
	return s.loadPatients(ctx, tenantID)
}

// SearchPatients busca pacientes por nome parcial, insensível a maiúsculas e acentos.
// Usa a lista cacheada do tenant para filtrar em memória — sem roundtrip ao banco por busca.
func (s *Service) SearchPatients(ctx context.Context, tenantID uuid.UUID, query string) ([]profile.Profile, error) {
	q := normalizeForSearch(query)

	//busca search no cache
	cSearchkey := fmt.Sprintf("%s:patients:search:%s", tenantID, query)
	if cached, err := s.cache.Get(ctx, cSearchkey); err == nil && cached != "" {
		var cachedSearchPatients []profile.Profile
		if err := json.Unmarshal([]byte(cached), &cachedSearchPatients); err == nil {
			return cachedSearchPatients, nil
		}
	}

	patients, err := s.loadPatients(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]profile.Profile, 0)
	for _, p := range patients {
		if strings.Contains(normalizeForSearch(p.FullName), q) {
			result = append(result, p)
		}
	}

	if data, err := json.Marshal(result); err == nil {
		ttl := 40 * time.Minute
		_ = s.cache.Set(ctx, cSearchkey, string(data), ttl)
	}
	return result, nil
}

// loadPatients carrega todos os pacientes do tenant do cache; em caso de miss, busca no banco.
func (s *Service) loadPatients(ctx context.Context, tenantID uuid.UUID) ([]profile.Profile, error) {
	log := logger.FromContext(ctx)

	key := patientsCacheKey(tenantID)
	if cached, err := s.cache.Get(ctx, key); err == nil && cached != "" {
		var patients []profile.Profile
		if err := json.Unmarshal([]byte(cached), &patients); err == nil {
			log.Debug("cache hits list patients")
			return patients, nil
		}
	}

	patients, err := s.profileRepo.ListByRole(ctx, s.db, tenantID, RolePatient)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(patients); err == nil {
		log.Debug("cache receive list patients")
		ttl := 6 * time.Hour
		_ = s.cache.Set(ctx, key, string(data), ttl)
	}

	return patients, nil
}

// GetPatient retorna um paciente pelo ID do perfil.
func (s *Service) GetPatient(ctx context.Context, tenantID, id uuid.UUID) (*profile.Profile, error) {
	p, err := s.profileRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Role != RolePatient {
		return nil, ErrPatientNotFound
	}
	return p, nil
}

// UpdatePatient atualiza os dados pessoais de um paciente e invalida o cache.
func (s *Service) UpdatePatient(ctx context.Context, tenantID, id uuid.UUID, req UpdatePatientRequest) (*profile.Profile, error) {
	p, err := s.profileRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Role != RolePatient {
		return nil, ErrPatientNotFound
	}

	p.FullName = req.FullName
	p.Document = req.Document
	p.BirthDate = req.BirthDate
	p.Phone = req.Phone
	p.HasWhatsapp = req.HasWhatsapp
	p.EmergencyContactName = req.EmergencyContactName
	p.EmergencyContactPhone = req.EmergencyContactPhone
	p.Address = req.Address

	if err := s.profileRepo.Update(ctx, s.db, p); err != nil {
		return nil, err
	}
	s.invalidatePatientCache(ctx, tenantID)
	return p, nil
}

// DeletePatient remove um paciente (soft-delete somente do Profile) e invalida o cache.
func (s *Service) DeletePatient(ctx context.Context, tenantID, id uuid.UUID) error {
	p, err := s.profileRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return err
	}
	if p == nil || p.Role != RolePatient {
		return ErrPatientNotFound
	}
	if err := s.profileRepo.SoftDelete(ctx, s.db, tenantID, id); err != nil {
		return err
	}
	s.invalidatePatientCache(ctx, tenantID)
	return nil
}

func (s *Service) invalidatePatientCache(ctx context.Context, tenantID uuid.UUID) {
	log := logger.FromContext(ctx)

	err := s.cache.Del(ctx, patientsCacheKey(tenantID))
	if err != nil {
		log.WithErr(err).Warn("failed to remove patients from cache")
	}

	keyPrefix, err := s.cache.DelWithPrefix(ctx, fmt.Sprintf("%s:patients:*", tenantID))
	if err != nil {
		log.WithErr(err).Warn("failed to remove patients.search from cache")
	}

	log.WithField(logger.String("keyPrefix", fmt.Sprintf("%v", keyPrefix))).Info("invalidatePatientCache")
}
