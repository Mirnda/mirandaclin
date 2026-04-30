package invite

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/mailer"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const defaultTTL = 7 * 24 * time.Hour

var ErrInvalidInvite = errors.New("convite inválido ou expirado")

type CreateRequest struct {
	TenantID uuid.UUID
	Email    string
	Role     string
	Password string
}

type Service struct {
	db     *gorm.DB
	repo   Repository
	mailer mailer.Mailer
	appURL string
}

func NewService(db *gorm.DB, r Repository, m mailer.Mailer, appURL string) *Service {
	return &Service{db: db, repo: r, mailer: m, appURL: appURL}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Invite, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	salt, hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	requestID := logger.GetRequestID(ctx)

	inv := &Invite{
		TenantID:     req.TenantID,
		Token:        token,
		Email:        req.Email,
		Role:         req.Role,
		PasswordHash: hash,
		Salt:         salt,
		EventId:      requestID,
		ExpiresAt:    time.Now().Add(defaultTTL),
	}
	if err := s.repo.Create(ctx, s.db, inv); err != nil {
		return nil, err
	}

	link := fmt.Sprintf("%s/convite/aceitar?token=%s", s.appURL, token)
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

	return inv, nil
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

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
