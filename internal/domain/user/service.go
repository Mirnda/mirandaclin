package user

import (
	"context"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/internal/domain/tenant"
	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/Mirnda/mirandaclin/pkg/config"
	"github.com/Mirnda/mirandaclin/pkg/mailer"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	TenantName            string
	Email                 string
	Password              string
	FullName              string
	Document              string
	Phone                 string
	HasWhatsapp           bool
	EmergencyContactName  string
	EmergencyContactPhone string
}

type CreateRequest struct {
	TenantID              uuid.UUID
	Email                 string
	Password              string
	Role                  string
	FullName              string
	Document              string
	Phone                 string
	HasWhatsapp           bool
	EmergencyContactName  string
	EmergencyContactPhone string
}

type AcceptInviteRequest struct {
	Token    string
	Password string
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

type UpdatePatientRequest struct {
	FullName              string
	Document              string
	BirthDate             *time.Time
	Phone                 string
	HasWhatsapp           bool
	EmergencyContactName  string
	EmergencyContactPhone string
	Address               shared.Address
}

type Service struct {
	db          *gorm.DB
	userRepo    Repository
	profileRepo profile.Repository
	inviteRepo  invite.Repository
	tenantRepo  tenant.Repository
	cfg         config.App
	mailer      mailer.Mailer
	cache       cache.Cache
	jwtSecret   string
}

func NewService(
	db *gorm.DB,
	ur Repository,
	pr profile.Repository,
	ir invite.Repository,
	tr tenant.Repository,
	cfg config.App,
	ml mailer.Mailer,
	c cache.Cache,
	secret string,
) *Service {
	return &Service{
		db:          db,
		userRepo:    ur,
		profileRepo: pr,
		inviteRepo:  ir,
		tenantRepo:  tr,
		cfg:         cfg,
		mailer:      ml,
		cache:       c,
		jwtSecret:   secret,
	}
}

// AcceptInvite aceita um convite: cria usuário com a senha fornecida,
// vincula ao profile do tenant e retorna LoginResponse com access + refresh token.
func (s *Service) AcceptInvite(ctx context.Context, req AcceptInviteRequest) (*LoginResponse, error) {
	inv, err := s.inviteRepo.FindByToken(ctx, s.db, req.Token)
	if err != nil {
		return nil, err
	}
	if inv == nil || inv.UsedAt != nil || time.Now().After(inv.ExpiresAt) {
		return nil, invite.ErrInvalidInvite
	}

	salt, hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var u *User
	p := &profile.Profile{TenantID: inv.TenantID, Role: inv.Role}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing, err := s.userRepo.FindByEmail(ctx, tx, inv.Email)
		if err != nil {
			return err
		}
		if existing != nil {
			existing.PasswordHash = hash
			existing.Salt = salt
			existing.LastTenantID = inv.TenantID
			if err := s.userRepo.Update(ctx, tx, existing); err != nil {
				return err
			}
			p.UserID = &existing.ID
			u = existing
		} else {
			newUser := &User{
				Email:           inv.Email,
				PasswordHash:    hash,
				Salt:            salt,
				LastTenantID:    inv.TenantID,
				EmailVerifiedAt: &now,
			}
			if err := s.userRepo.Create(ctx, tx, newUser); err != nil {
				return err
			}
			p.UserID = &newUser.ID
			u = newUser
		}
		if err := s.profileRepo.Create(ctx, tx, p); err != nil {
			return err
		}
		return s.inviteRepo.MarkUsed(ctx, tx, inv.ID)
	})
	if err != nil {
		return nil, err
	}

	accessToken, expiresAt, err := s.issueJWT(u, p, defaultTokenDuration, "access")
	if err != nil {
		return nil, err
	}
	refreshToken, refreshExpiresAt, err := s.issueJWT(u, p, 11*time.Hour, "refresh")
	if err != nil {
		return nil, err
	}
	refreshCookie := s.generateRefreshCookie(refreshToken, s.cfg.Env == "production", refreshExpiresAt)

	return &LoginResponse{
		AccessToken:          accessToken,
		AccessExpirationTime: expiresAt,
		RefreshCookie:        refreshCookie,
	}, nil
}
