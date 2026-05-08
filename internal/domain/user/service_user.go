package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

var (
	ErrInvalidRole   = errors.New("role inválido para este endpoint")
	ErrEmailConflict = errors.New("email já cadastrado")
	ErrUserNotFound  = errors.New("usuário não encontrado")
)

// List retorna todos os usuários staff (não pacientes) do tenant.
func (s *Service) List(ctx context.Context, tenantID uuid.UUID) ([]UserWithProfile, error) {
	profiles, err := s.profileRepo.List(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]UserWithProfile, 0, len(profiles))
	for i := range profiles {
		p := &profiles[i]
		if p.Role == shared.RolePatient || p.UserID == nil {
			continue
		}
		u, err := s.userRepo.FindByID(ctx, s.db, *p.UserID)
		if err != nil || u == nil {
			continue
		}
		result = append(result, *mergeUserProfile(u, p))
	}
	return result, nil
}

func (s *Service) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*UserWithProfile, error) {
	u, err := s.userRepo.FindByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	p, err := s.profileRepo.FindByUserAndTenant(ctx, s.db, id, tenantID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrUserNotFound
	}
	return mergeUserProfile(u, p), nil
}

type UpdateRequest struct {
	FullName              string
	Phone                 string
	Document              string
	HasWhatsapp           bool
	EmergencyContactName  string
	EmergencyContactPhone string
	Password              string
}

func (s *Service) Update(ctx context.Context, tenantID, userID uuid.UUID, req UpdateRequest) (*UserWithProfile, error) {
	p, err := s.profileRepo.FindByUserAndTenant(ctx, s.db, userID, tenantID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrUserNotFound
	}

	p.FullName = req.FullName
	p.Phone = req.Phone
	p.Document = req.Document
	p.HasWhatsapp = req.HasWhatsapp
	p.EmergencyContactName = req.EmergencyContactName
	p.EmergencyContactPhone = req.EmergencyContactPhone

	if req.Password != "" {
		u, err := s.userRepo.FindByID(ctx, s.db, userID)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, ErrUserNotFound
		}
		salt, hash, err := hashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		u.Salt = salt
		u.PasswordHash = hash
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := s.userRepo.Update(ctx, tx, u); err != nil {
				return err
			}
			return s.profileRepo.Update(ctx, tx, p)
		})
		if err != nil {
			return nil, err
		}
		return mergeUserProfile(u, p), nil
	}

	if err := s.profileRepo.Update(ctx, s.db, p); err != nil {
		return nil, err
	}
	u, err := s.userRepo.FindByID(ctx, s.db, userID)
	if err != nil || u == nil {
		return nil, ErrUserNotFound
	}
	return mergeUserProfile(u, p), nil
}

func hashPassword(password string) (salt, hash string, err error) {
	b := make([]byte, 16)
	if _, err = rand.Read(b); err != nil {
		return
	}
	salt = hex.EncodeToString(b)
	h, err := bcrypt.GenerateFromPassword([]byte(password+salt), 12)
	if err != nil {
		return
	}
	hash = string(h)
	return
}

func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	p, err := s.profileRepo.FindByUserAndTenant(ctx, s.db, id, tenantID)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrUserNotFound
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.profileRepo.SoftDelete(ctx, tx, tenantID, p.ID); err != nil {
			return err
		}

		allProfilesFromUser, err := s.profileRepo.FindByUserID(ctx, tx, id)
		if err != nil {
			return err
		}

		if len(allProfilesFromUser) > 0 {
			u, err := s.userRepo.FindByID(ctx, tx, id)
			if err != nil {
				return err
			}

			if u.LastTenantID == tenantID {
				u.LastTenantID = allProfilesFromUser[0].TenantID

				return s.userRepo.Update(ctx, tx, u)
			}

			return nil
		}

		return s.userRepo.SoftDelete(ctx, tx, id)
	})
}
