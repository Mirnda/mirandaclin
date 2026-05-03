package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/tenant"
	tenantmember "github.com/Mirnda/mirandaclin/internal/domain/tenant_member"
	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailConflict   = errors.New("email já cadastrado")
	ErrInvalidCreds    = errors.New("credenciais inválidas")
	ErrUserNotFound    = errors.New("usuário não encontrado")
	ErrTenantRequired  = errors.New("informe o tenant_id para autenticar")
	ErrTenantForbidden = errors.New("usuário não pertence a este tenant")
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

type LoginRequest struct {
	Email    string
	Password string
	TenantID uuid.UUID // opcional quando usuário pertence a múltiplos tenants
}

type AcceptInviteRequest struct {
	Token string
}

type Service struct {
	db         *gorm.DB
	userRepo   Repository
	inviteRepo invite.Repository
	tenantRepo tenant.Repository
	memberRepo tenantmember.Repository
	cache      cache.Cache
	jwtSecret  string
}

func NewService(
	db *gorm.DB,
	ur Repository,
	ir invite.Repository,
	tr tenant.Repository,
	mr tenantmember.Repository,
	c cache.Cache,
	secret string,
) *Service {
	return &Service{
		db:         db,
		userRepo:   ur,
		inviteRepo: ir,
		tenantRepo: tr,
		memberRepo: mr,
		cache:      c,
		jwtSecret:  secret,
	}
}

// Register cria um novo tenant e seu primeiro usuário (admin) em transação única.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (string, error) {
	existing, err := s.userRepo.FindByEmail(ctx, s.db, req.Email)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", ErrEmailConflict
	}

	salt, hash, err := hashPassword(req.Password)
	if err != nil {
		return "", err
	}

	t := &tenant.Tenant{Name: req.TenantName}
	u := &User{
		Email:                 req.Email,
		PasswordHash:          hash,
		Salt:                  salt,
		FullName:              req.FullName,
		Document:              req.Document,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
	}
	m := &tenantmember.TenantMember{Role: RoleAdmin}

	var log = logger.FromContext(ctx)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.tenantRepo.Create(ctx, tx, t); err != nil {
			return err
		}
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		m.UserID = u.ID
		m.TenantID = t.ID
		return s.memberRepo.Create(ctx, tx, m)
	})
	if err != nil {
		log.Error("erro ao registrar clínica", logger.Err(err))
		return "", err
	}

	log.Info("tenant registrado", logger.String("tenant_id", t.ID.String()), logger.String("user_id", u.ID.String()))
	return s.issueJWT(u, m)
}

// Create adiciona um novo usuário a um tenant existente.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*User, error) {
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

	u := &User{
		Email:                 req.Email,
		PasswordHash:          hash,
		Salt:                  salt,
		FullName:              req.FullName,
		Document:              req.Document,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
	}
	m := &tenantmember.TenantMember{TenantID: req.TenantID, Role: req.Role}

	var log = logger.FromContext(ctx)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		m.UserID = u.ID
		return s.memberRepo.Create(ctx, tx, m)
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

	u, err := s.userRepo.FindByEmail(ctx, s.db, req.Email)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password+u.Salt)); err != nil {
		log.Warn("tentativa de login com senha inválida", logger.String("user_id", u.ID.String()))
		return "", ErrInvalidCreds
	}

	members, err := s.memberRepo.FindByUserID(ctx, s.db, u.ID)
	if err != nil {
		return "", err
	}
	if len(members) == 0 {
		return "", ErrInvalidCreds
	}

	var member *tenantmember.TenantMember
	if req.TenantID != uuid.Nil {
		for _, m := range members {
			if m.TenantID == req.TenantID {
				member = m
				break
			}
		}
		if member == nil {
			return "", ErrTenantForbidden
		}
	} else if len(members) == 1 {
		member = members[0]
	} else {
		return "", ErrTenantRequired
	}

	return s.issueJWT(u, member)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.userRepo.FindByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// AcceptInvite aceita um convite: cria usuário novo se o email não existe,
// ou apenas adiciona ao tenant se o email já existe globalmente.
func (s *Service) AcceptInvite(ctx context.Context, req AcceptInviteRequest) (string, error) {
	inv, err := s.inviteRepo.FindByToken(ctx, s.db, req.Token)
	if err != nil {
		return "", err
	}
	if inv == nil || inv.UsedAt != nil || time.Now().After(inv.ExpiresAt) {
		return "", invite.ErrInvalidInvite
	}

	existing, err := s.userRepo.FindByEmail(ctx, s.db, inv.Email)
	if err != nil {
		return "", err
	}

	var u *User
	m := &tenantmember.TenantMember{TenantID: inv.TenantID, Role: inv.Role}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if existing != nil {
			m.UserID = existing.ID
			u = existing
		} else {
			newUser := &User{
				Email:        inv.Email,
				PasswordHash: inv.PasswordHash,
				Salt:         inv.Salt,
			}
			if err := s.userRepo.Create(ctx, tx, newUser); err != nil {
				return err
			}
			m.UserID = newUser.ID
			u = newUser
		}
		if err := s.memberRepo.Create(ctx, tx, m); err != nil {
			return err
		}
		return s.inviteRepo.MarkUsed(ctx, tx, inv.ID)
	})
	if err != nil {
		return "", err
	}

	return s.issueJWT(u, m)
}

func (s *Service) issueJWT(u *User, m *tenantmember.TenantMember) (string, error) {
	claims := jwt.MapClaims{
		"sub":       u.ID.String(),
		"tenant_id": m.TenantID.String(),
		"role":      m.Role,
		"scope":     ScopeForRole(m.Role),
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
