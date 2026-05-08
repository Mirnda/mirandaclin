package collaborator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/clinic"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	profileblock "github.com/Mirnda/mirandaclin/internal/domain/profile_block"
	profileclinic "github.com/Mirnda/mirandaclin/internal/domain/profile_clinic"
	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	cache       cache.Cache
	profileRepo profile.Repository
	clinicRepo  clinic.Repository
	pcRepo      profileclinic.Repository
	blockRepo   profileblock.Repository
}

func NewService(db *gorm.DB, cache cache.Cache, profileRepo profile.Repository, clinicRepo clinic.Repository, pcRepo profileclinic.Repository, blockRepo profileblock.Repository) *Service {
	return &Service{db, cache, profileRepo, clinicRepo, pcRepo, blockRepo}
}

type CreateCollaboratorRequest struct {
	profile.CreateProfile
	ProfileClinics []profileclinic.CreateProfileClinicRequest
}

func (s *Service) Create(ctx context.Context, req CreateCollaboratorRequest) (*Collaborator, error) {

	if !shared.IsValidCollaboratorRole(req.Role) {
		return nil, shared.ErrInvalidRole
	}

	p := &profile.Profile{
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

	var collaboratorClinics []CollaboratorClinic

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.profileRepo.Create(ctx, tx, p); err != nil {
			return err
		}

		if len(req.ProfileClinics) == 0 {
			return nil
		}

		for _, pcReq := range req.ProfileClinics {

			if pcReq.ClinicID == uuid.Nil {
				return profileclinic.ErrClinicRequired
			}

			c, err := s.clinicRepo.FindByID(ctx, tx, req.TenantID, pcReq.ClinicID)
			if err != nil {
				return err
			}
			if c == nil {
				return profileclinic.ErrClinicRequired
			}

			pc := &profileclinic.ProfileClinic{
				ProfileID:           p.ID,
				TenantID:            req.TenantID,
				ClinicID:            pcReq.ClinicID,
				WorkingDays:         pcReq.WorkingDays,
				SlotDurationMinutes: pcReq.SlotDurationMinutes,
			}

			if err := s.pcRepo.Create(ctx, tx, pc); err != nil {
				return err
			}

			collaboratorClinics = append(collaboratorClinics, CollaboratorClinic{Clinic: *c, ProfileClinic: *pc})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.invalidateColaboratorsCache(ctx, req.TenantID)

	return &Collaborator{
		Profile:             *p,
		CollaboratorClinics: collaboratorClinics,
	}, nil
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID) ([]Collaborator, error) {
	log := logger.FromContext(ctx)

	log.Debug("init list internal/domain/collaborator/service.go:115")

	var collaborators []Collaborator

	collaboratorsCacheKey := collaboratorsCacheKey(tenantID)
	if cached, err := s.cache.Get(ctx, collaboratorsCacheKey); err == nil && cached != "" {
		if err := json.Unmarshal([]byte(cached), &collaborators); err == nil {
			return collaborators, nil
		}
	}

	log.With("collaboratorsCacheKey", collaboratorsCacheKey).Debug("nao achou nada")

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		log.Debug("Transaction")

		profiles, err := s.profileRepo.ListCollaborators(ctx, s.db, tenantID)
		if err != nil {
			log.WithErr(err).Debug("err ListCollaborators")
			return err
		}

		for _, p := range profiles {
			if p.Role == shared.RolePatient {
				continue
			}

			profileclinics, err := s.pcRepo.ListByProfile(ctx, s.db, tenantID, p.ID)
			if err != nil {
				log.WithErr(err).Debug("err ListByProfile")
				return err
			}

			var collaborator = Collaborator{Profile: p}
			var allBlocks []profileblock.ProfileBlock

			for _, pc := range profileclinics {
				pcClinic, err := s.clinicRepo.FindByID(ctx, s.db, tenantID, pc.ClinicID)
				if err != nil {
					log.WithErr(err).Debug("err clinicRepo.FindByID")
					return err
				}

				pcBlocks, err := s.blockRepo.FindBlocksForSlot(ctx, s.db, tenantID, p.ID, &pc.ClinicID, time.Time{}, "", "")
				if err != nil {
					log.WithErr(err).Debug("err FindBlocksForSlot")
					return err
				}
				allBlocks = append(allBlocks, pcBlocks...)

				collaborator.CollaboratorClinics = append(collaborator.CollaboratorClinics, CollaboratorClinic{
					Clinic:        *pcClinic,
					ProfileClinic: pc,
				})
			}

			pBlocks, err := s.blockRepo.FindBlocksForSlot(ctx, s.db, tenantID, p.ID, nil, time.Time{}, "", "")
			if err != nil {
				log.WithErr(err).Debug("err pBlocks, err := s.blockRepo.FindBlocksForSlot")
				return err
			}
			allBlocks = append(allBlocks, pBlocks...)

			collaborator.ProfileBlocks = profileblock.RemoveDuplicates(allBlocks)

			collaborators = append(collaborators, collaborator)
			log.Debug("colaborators fim")
		}

		return nil
	})
	if err != nil {
		log.WithErr(err).Debug("fora da transaction")
		return nil, err
	}

	if data, err := json.Marshal(collaborators); err == nil {
		ttl := 6 * time.Hour
		_ = s.cache.Set(ctx, collaboratorsCacheKey, string(data), ttl)
	}

	log.Debug("fcollaboratorsdsdsdsdssd")
	return collaborators, nil
}

func collaboratorsCacheKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s:profile:clinic:profileclinic:profileblock:collaborators:", tenantID)
}

func (s *Service) invalidateColaboratorsCache(ctx context.Context, tenantID uuid.UUID) {
	_ = s.cache.Del(ctx, collaboratorsCacheKey(tenantID))

	_, _ = s.cache.DelWithIndex(ctx, "*:profile:*")
	_, _ = s.cache.DelWithIndex(ctx, "*:profileclinic:*")
	_, _ = s.cache.DelWithIndex(ctx, "*:profileblock:*")
}
