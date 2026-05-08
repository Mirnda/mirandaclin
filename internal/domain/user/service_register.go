package user

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/internal/domain/tenant"
	"github.com/Mirnda/mirandaclin/pkg/logger"

	"gorm.io/gorm"
)

//go:embed registry_verification_email_template.html
var verificationEmailTmpl string

var (
	ErrTenantConflict = errors.New("namespace inválido")
	ErrInvalidToken   = errors.New("token inválido ou expirado")

	tokenDefaultLength      int64 = 6
	tokenDefaultOnlyNumbers bool  = true
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

// Register cria um novo tenant e seu primeiro usuário admin, depois envia email de confirmação.
func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	existing, err := s.userRepo.FindByEmail(ctx, s.db, req.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrEmailConflict
	}

	if req.TenantName == "" {
		req.TenantName = req.Email
	}

	existingTenant, err := s.tenantRepo.FindByName(ctx, s.db, req.TenantName)
	if err != nil {
		return err
	}
	if existingTenant != nil {
		return ErrTenantConflict
	}

	salt, hash, err := hashPassword(req.Password)
	if err != nil {
		return err
	}

	verifyToken, err := generateToken(tokenDefaultLength, tokenDefaultOnlyNumbers)
	if err != nil {
		return err
	}

	t := &tenant.Tenant{Name: req.TenantName}
	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Salt:         salt,
	}
	p := &profile.Profile{
		Role:                  shared.RoleAdmin,
		FullName:              req.FullName,
		Document:              req.Document,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
	}
	inv := &invite.Invite{
		Token:        verifyToken,
		Email:        req.Email,
		Role:         shared.RoleAdmin,
		PasswordHash: hash,
		Salt:         salt,
		EventId:      logger.GetRequestID(ctx),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	log := logger.FromContext(ctx)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.tenantRepo.Create(ctx, tx, t); err != nil {
			return err
		}
		u.LastTenantID = t.ID
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		p.UserID = &u.ID
		p.TenantID = t.ID
		if err := s.profileRepo.Create(ctx, tx, p); err != nil {
			return err
		}
		inv.TenantID = t.ID
		inv.UserID = &u.ID
		if err := s.inviteRepo.Create(ctx, tx, inv); err != nil {
			return err
		}

		uName := p.FullName
		parts := strings.Fields(uName)
		if len(parts) > 0 {
			uName = parts[0]
			if len(parts[0]) < 5 && len(parts) > 1 {
				uName = strings.Join(parts[:2], " ")
			}
		}

		return s.sendVerificationEmail(ctx, s.cfg.Name, uName, req.Email, verifyToken, s.cfg.VerifyEmailUrl(verifyToken))
	})
	if err != nil {
		log.Error("erro ao registrar usuário/tenant", logger.Err(err))
		return err
	}

	return nil
}

func generateToken(length int64, onlyNumbers bool) (string, error) {
	var charset = "9173546280"
	if !onlyNumbers {
		charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	result := make([]byte, length)

	for i := range result {
		n := rand.IntN(len(charset))
		result[i] = charset[n]
	}

	return string(result), nil
}

func (s *Service) sendVerificationEmail(ctx context.Context, appName, username, to, token, directUrl string) error {

	mailData := struct {
		AppName          string
		UserName         string
		VerificationCode string
		VerificationURL  string
	}{
		appName,
		username,
		token,
		directUrl,
	}

	tmpl, err := template.New("verification").Parse(verificationEmailTmpl)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err = tmpl.Execute(&body, mailData); err != nil {
		return err
	}

	subject := fmt.Sprintf("%s Código de verificação", appName)

	return s.mailer.Send(ctx, to, subject, body.String())
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	inv, err := s.inviteRepo.FindByToken(ctx, s.db, token)
	if err != nil {
		return err
	}
	if inv == nil || inv.UsedAt != nil || time.Now().After(inv.ExpiresAt) || inv.UserID == nil {
		return ErrInvalidToken
	}

	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).Where("id = ?", *inv.UserID).Update("email_verified_at", now).Error; err != nil {
			return err
		}
		return s.inviteRepo.MarkUsed(ctx, tx, inv.ID)
	})
}
