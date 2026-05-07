package invite

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/pkg/config"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/mailer"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultTTL = 7 * 24 * time.Hour

var (
	ErrInvalidInvite        = errors.New("convite inválido ou expirado")
	ErrProfileNotFound      = errors.New("perfil não encontrado")
	ErrProfileAlreadyLinked = errors.New("perfil já possui usuário vinculado")
	ErrInvalidProfileRole   = errors.New("perfil de paciente não pode receber convite")
)

type CreateRequest struct {
	TenantID  uuid.UUID
	Email     string
	ProfileID uuid.UUID
}

type Service struct {
	db          *gorm.DB
	repo        Repository
	profileRepo profile.Repository
	mailer      mailer.Mailer
	cfg         config.App
}

func NewService(db *gorm.DB, r Repository, pr profile.Repository, m mailer.Mailer, cfg config.App) *Service {
	return &Service{db: db, repo: r, profileRepo: pr, mailer: m, cfg: cfg}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) error {
	prof, err := s.profileRepo.FindByID(ctx, s.db, req.TenantID, req.ProfileID)
	if err != nil {
		return err
	}
	if prof == nil {
		return ErrProfileNotFound
	}
	if prof.UserID != nil {
		return ErrProfileAlreadyLinked
	}
	if prof.Role == "patient" {
		return ErrInvalidProfileRole
	}

	token, err := generateToken()
	if err != nil {
		return err
	}

	inv := &Invite{
		TenantID:  req.TenantID,
		Token:     token,
		Email:     req.Email,
		Role:      prof.Role,
		EventId:   logger.GetRequestID(ctx),
		ExpiresAt: time.Now().Add(defaultTTL),
	}
	if err := s.repo.Create(ctx, s.db, inv); err != nil {
		return err
	}

	link := s.cfg.VerifyEmailUrl(token)
	body := fmt.Sprintf(
		`<p>Você foi convidado para a plataforma mirandaclin.</p>`+
			`<p><a href="%s">Clique aqui para confirmar seu cadastro</a></p>`+
			`<p>O link expira em 7 dias.</p>`, link)
	if err := s.mailer.Send(ctx, req.Email, "Convite para mirandaclin", body); err != nil {
		logger.FromContext(ctx).Error("falha ao enviar email de convite",
			logger.String("tenant_id", req.TenantID.String()),
			logger.String("to", req.Email),
			logger.Err(err),
		)
	}

	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
