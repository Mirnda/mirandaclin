package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailConflict = errors.New("email já cadastrado neste tenant")
	ErrInvalidCreds  = errors.New("credenciais inválidas")
	ErrUserNotFound  = errors.New("usuário não encontrado")
)

type CreateRequest struct {
	TenantID              uuid.UUID
	Email                 string
	Password              string
	Role                  string
	Phone                 string
	HasWhatsapp           bool
	EmergencyContactName  string
	EmergencyContactPhone string
	FullName              string
	Document              string
}

type LoginRequest struct {
	TenantID uuid.UUID
	Email    string
	Password string
}

type AcceptInviteRequest struct {
	Token string
}

type Service struct {
	db          *gorm.DB
	userRepo    Repository
	profileRepo profile.Repository
	inviteRepo  invite.Repository
	cache       cache.Cache
	jwtSecret   string
}

func NewService(db *gorm.DB, ur Repository, pr profile.Repository, ir invite.Repository, c cache.Cache, secret string) *Service {
	return &Service{db: db, userRepo: ur, profileRepo: pr, inviteRepo: ir, cache: c, jwtSecret: secret}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*User, error) {
	existing, err := s.userRepo.FindByEmail(ctx, s.db, req.TenantID, req.Email)
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

	u := &User{
		TenantID:              req.TenantID,
		Email:                 req.Email,
		PasswordHash:          hash,
		Salt:                  salt,
		Role:                  req.Role,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
	}

	p := &profile.Profile{
		TenantID: req.TenantID,
		FullName: req.FullName,
		Document: req.Document,
	}

	var log = logger.FromContext(ctx)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		p.UserID = u.ID
		return s.profileRepo.Create(ctx, tx, p)
	})
	if err != nil {
		log.Error("erro ao criar usuário", logger.String("tenant_id", req.TenantID.String()), logger.Err(err))
		return nil, err
	}

	log.Info("usuário criado", logger.String("tenant_id", req.TenantID.String()), logger.String("user_id", u.ID.String()))
	return u, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	var log = logger.FromContext(ctx)
	u, err := s.userRepo.FindByEmail(ctx, s.db, req.TenantID, req.Email)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password+u.Salt)); err != nil {
		log.Warn("tentativa de login com senha inválida", logger.String("tenant_id", req.TenantID.String()))
		return "", ErrInvalidCreds
	}

	token, err := s.issueJWT(u)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*User, error) {
	u, err := s.userRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Service) AcceptInvite(ctx context.Context, req AcceptInviteRequest) (string, error) {
	inv, err := s.inviteRepo.FindByToken(ctx, s.db, req.Token)
	if err != nil {
		return "", err
	}
	if inv == nil || inv.UsedAt != nil || time.Now().After(inv.ExpiresAt) {
		return "", invite.ErrInvalidInvite
	}

	existing, err := s.userRepo.FindByEmail(ctx, s.db, inv.TenantID, inv.Email)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", ErrEmailConflict
	}

	u := &User{
		TenantID:     inv.TenantID,
		Email:        inv.Email,
		PasswordHash: inv.PasswordHash,
		Salt:         inv.Salt,
		Role:         inv.Role,
	}
	p := &profile.Profile{
		TenantID: inv.TenantID,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		p.UserID = u.ID
		if err := s.profileRepo.Create(ctx, tx, p); err != nil {
			return err
		}
		return s.inviteRepo.MarkUsed(ctx, tx, inv.ID)
	})
	if err != nil {
		return "", err
	}

	return s.issueJWT(u)
}

func (s *Service) issueJWT(u *User) (string, error) {
	claims := jwt.MapClaims{
		"sub":       u.ID.String(),
		"tenant_id": u.TenantID.String(),
		"role":      u.Role,
		"scope":     ScopeForRole(u.Role),
		"exp":       time.Now().Add(time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
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
