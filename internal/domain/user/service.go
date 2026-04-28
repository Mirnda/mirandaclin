package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mirandev/mirandaclin/internal/domain/profile"
	"github.com/mirandev/mirandaclin/internal/infra/cache"
	"github.com/mirandev/mirandaclin/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailConflict  = errors.New("email já cadastrado neste tenant")
	ErrInvalidCreds   = errors.New("credenciais inválidas")
	ErrUserNotFound   = errors.New("usuário não encontrado")
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

type Service struct {
	db          *gorm.DB
	userRepo    Repository
	profileRepo profile.Repository
	cache       cache.Cache
	jwtSecret   string
	log         logger.Logger
}

func NewService(db *gorm.DB, ur Repository, pr profile.Repository, c cache.Cache, secret string, log logger.Logger) *Service {
	return &Service{db: db, userRepo: ur, profileRepo: pr, cache: c, jwtSecret: secret, log: log}
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

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		p.UserID = u.ID
		return s.profileRepo.Create(ctx, tx, p)
	})
	if err != nil {
		s.log.Error("erro ao criar usuário", zap.String("tenant_id", req.TenantID.String()), zap.Error(err))
		return nil, err
	}

	s.log.Info("usuário criado", zap.String("tenant_id", req.TenantID.String()), zap.String("user_id", u.ID.String()))
	return u, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	u, err := s.userRepo.FindByEmail(ctx, s.db, req.TenantID, req.Email)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password+u.Salt)); err != nil {
		s.log.Warn("tentativa de login com senha inválida", zap.String("tenant_id", req.TenantID.String()))
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
	h, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	hash = string(h)
	return
}
