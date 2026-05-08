package user

import (
	"context"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"gorm.io/gorm"
)

type AcceptInviteRequest struct {
	Token    string
	Password string
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
