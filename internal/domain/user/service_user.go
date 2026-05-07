package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

var (
	ErrInvalidRole   = errors.New("role inválido para este endpoint")
	ErrEmailConflict = errors.New("email já cadastrado")
	ErrUserNotFound  = errors.New("usuário não encontrado")
)

// Create adiciona um novo usuário (staff) a um tenant existente.
// Rejeita role=patient — pacientes são criados via CreatePatient.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*UserWithProfile, error) {
	if req.Role == RolePatient {
		return nil, ErrInvalidRole
	}

	existing, err := s.userRepo.FindByEmail(ctx, s.db, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailConflict
	}

	salt, hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	verifyToken, err := generateToken(tokenDefaultLength, tokenDefaultOnlyNumbers)
	if err != nil {
		return nil, err
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Salt:         salt,
		LastTenantID: req.TenantID,
	}
	p := &profile.Profile{
		TenantID:              req.TenantID,
		Role:                  req.Role,
		FullName:              req.FullName,
		Document:              req.Document,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
	}
	inv := &invite.Invite{
		TenantID:  req.TenantID,
		Token:     verifyToken,
		Email:     req.Email,
		Role:      req.Role,
		EventId:   logger.GetRequestID(ctx),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		p.UserID = &u.ID
		if err := s.profileRepo.Create(ctx, tx, p); err != nil {
			return err
		}
		inv.UserID = &u.ID
		return s.inviteRepo.Create(ctx, tx, inv)
	})
	if err != nil {
		return nil, err
	}

	uName := p.FullName
	parts := strings.Fields(uName)
	if len(parts) > 0 {
		uName = parts[0]
		if len(parts[0]) < 5 && len(parts) > 1 {
			uName = strings.Join(parts[:2], " ")
		}
	}

	return mergeUserProfile(u, p), s.sendVerificationEmail(ctx, s.cfg.Name, uName, req.Email, verifyToken, s.cfg.VerifyEmailUrl(verifyToken))
}

// List retorna todos os usuários staff (não pacientes) do tenant.
func (s *Service) List(ctx context.Context, tenantID uuid.UUID) ([]UserWithProfile, error) {
	profiles, err := s.profileRepo.List(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]UserWithProfile, 0, len(profiles))
	for i := range profiles {
		p := &profiles[i]
		if p.Role == RolePatient || p.UserID == nil {
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
