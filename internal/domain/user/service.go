package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	emailverification "github.com/Mirnda/mirandaclin/internal/domain/email_verification"
	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/internal/domain/tenant"
	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/mailer"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailConflict   = errors.New("email já cadastrado")
	ErrTenantConflict  = errors.New("namespace inválido")
	ErrInvalidCreds    = errors.New("credenciais inválidas")
	ErrUserNotVerified = errors.New("usuário não confirmou email")
	ErrUserNotFound    = errors.New("usuário não encontrado")
	ErrPatientNotFound = errors.New("paciente não encontrado")
	ErrTenantRequired  = errors.New("informe o tenant_id para autenticar")
	ErrTenantForbidden = errors.New("usuário não pertence a este tenant")
	ErrInvalidRole     = errors.New("role inválido para este endpoint")
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

// type LoginRequest struct {
// 	Email     string
// 	Password  string
// 	TenantID  uuid.UUID // opcional quando usuário pertence a múltiplos tenants
// 	KeepAlive bool
// }

// type LoginResponse struct {
// 	accessToken string

// 	refreshToken string
// }

type AcceptInviteRequest struct {
	Token string
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

type CreatePatientRequest struct {
	TenantID              uuid.UUID
	FullName              string
	Document              string
	BirthDate             *time.Time
	Phone                 string
	HasWhatsapp           bool
	EmergencyContactName  string
	EmergencyContactPhone string
	Address               shared.Address
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
	db             *gorm.DB
	userRepo       Repository
	profileRepo    profile.Repository
	inviteRepo     invite.Repository
	tenantRepo     tenant.Repository
	emailVerifRepo emailverification.Repository
	mailer         mailer.Mailer
	cache          cache.Cache
	jwtSecret      string
	appURL         string
}

func NewService(
	db *gorm.DB,
	ur Repository,
	pr profile.Repository,
	ir invite.Repository,
	tr tenant.Repository,
	evr emailverification.Repository,
	ml mailer.Mailer,
	c cache.Cache,
	secret string,
	appURL string,
) *Service {
	return &Service{
		db:             db,
		userRepo:       ur,
		profileRepo:    pr,
		inviteRepo:     ir,
		tenantRepo:     tr,
		emailVerifRepo: evr,
		mailer:         ml,
		cache:          c,
		jwtSecret:      secret,
		appURL:         appURL,
	}
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

	verifyToken, err := generateToken()
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
		Role:                  RoleAdmin,
		FullName:              req.FullName,
		Document:              req.Document,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
	}
	ev := &emailverification.EmailVerification{
		Token:     verifyToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	log := logger.FromContext(ctx)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.tenantRepo.Create(ctx, tx, t); err != nil {
			return err
		}
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		p.UserID = &u.ID
		p.TenantID = t.ID
		if err := s.profileRepo.Create(ctx, tx, p); err != nil {
			return err
		}
		ev.UserID = u.ID
		return s.emailVerifRepo.Create(ctx, tx, ev)
	})
	if err != nil {
		log.Error("erro ao registrar usuário/tenant", logger.Err(err))
		return err
	}

	log.Info("tenant registrado", logger.String("tenant_id", t.ID.String()), logger.String("user_id", u.ID.String()))
	return s.sendVerificationEmail(ctx, req.Email, verifyToken)
}

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

	verifyToken, err := generateToken()
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
	ev := &emailverification.EmailVerification{
		Token:     verifyToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	log := logger.FromContext(ctx)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, u); err != nil {
			return err
		}
		p.UserID = &u.ID
		if err := s.profileRepo.Create(ctx, tx, p); err != nil {
			return err
		}
		ev.UserID = u.ID
		return s.emailVerifRepo.Create(ctx, tx, ev)
	})
	if err != nil {
		log.Error("erro ao criar usuário", logger.String("tenant_id", req.TenantID.String()), logger.Err(err))
		return nil, err
	}

	log.Info("usuário criado", logger.String("tenant_id", req.TenantID.String()), logger.String("user_id", u.ID.String()))
	_ = s.sendVerificationEmail(ctx, req.Email, verifyToken)
	return mergeUserProfile(u, p), nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	ev, err := s.emailVerifRepo.FindByToken(ctx, s.db, token)
	if err != nil {
		return err
	}
	if ev == nil || ev.UsedAt != nil || time.Now().After(ev.ExpiresAt) {
		return emailverification.ErrInvalidToken
	}

	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).Where("id = ?", ev.UserID).Update("email_verified_at", now).Error; err != nil {
			return err
		}
		return s.emailVerifRepo.MarkUsed(ctx, tx, ev.ID)
	})
}

// // func (s *Service) Login(ctx context.Context, req LoginRequest) (token string, refreshToken string, maxAgeSecRefresh int, expiresAt time.Time, err error) {
// func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
// 	log := logger.FromContext(ctx)

// 	u, err := s.userRepo.FindByEmail(ctx, s.db, req.Email)
// 	if err != nil {
// 		log.WithErr(err).Warn("failed to get user by email")
// 		return nil, err
// 	}
// 	if u == nil {
// 		log.WithErr(ErrInvalidCreds).Warn("provided email does not have a registered user")
// 		return nil, ErrInvalidCreds
// 	}

// 	log = log.With("user_id", u.ID.String())

// 	if u.EmailVerifiedAt.IsZero() || u.EmailVerifiedAt.After(time.Now()) {
// 		err = ErrUserNotVerified
// 		log.WithErr(err).Warn("not allowed to log in before email confirmation")
// 		return nil, err
// 	}

// 	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password+u.Salt)); err != nil {
// 		log.WithErr(err).Warn("invalid password")
// 		return nil, err "", "", 0, time.Time{}, ErrUserNotVerified
// 	}

// 	profiles, err := s.profileRepo.FindByUserID(ctx, s.db, u.ID)
// 	if err != nil {
// 		log.WithErr(err).Warn("failed to get profile by user")
// 		return nil, err
// 	}
// 	if len(profiles) == 0 {
// 		err = ErrUserNotVerified
// 		log.WithErr(err).Warn("user does not have a registered profile")
// 		return nil, err
// 	}

// 	var prof *profile.Profile
// 	if req.TenantID != uuid.Nil {
// 		log.Debug("use request tenant")

// 		for _, p := range profiles {
// 			if p.TenantID == req.TenantID {
// 				prof = p
// 				break
// 			}
// 		}

// 	} else if u.LastTenantID != uuid.Nil {
// 		log.Debug("use user tenant")

// 		for _, p := range profiles {
// 			if p.TenantID == u.LastTenantID {
// 				prof = p
// 				break
// 			}
// 		}

// 	} else if len(profiles) == 1 {
// 		log.Debug("use single profile tenant")
// 		prof = profiles[0]

// 	} else {
// 		err = ErrTenantRequired
// 		log.WithErr(err).Warn("tenant not found")
// 		return nil, err
// 	}

// 	if prof == nil {
// 		err = ErrTenantForbidden
// 		log.WithErr(err).Warn("profile not found")
// 		return nil, err
// 	}

// 	if u.LastTenantID != prof.TenantID {
// 		log.Debug("update user tenant id as profile tenant id")
// 		u.LastTenantID = prof.TenantID
// 		_ = s.userRepo.Update(ctx, s.db, u)
// 	}

// 	token, expiresAt, err = s.issueJWT(u, prof)
// 	if err != nil {
// 		return nil, err
// 	}
// 	refreshToken, maxAgeSecRefresh, err = s.issueRefreshJWT(u, prof, req.KeepAlive)
// 	return token, refreshToken, maxAgeSecRefresh, expiresAt, err
// }

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

// ListPatients retorna todos os pacientes do tenant (via cache).
func (s *Service) ListPatients(ctx context.Context, tenantID uuid.UUID) ([]profile.Profile, error) {
	return s.loadPatients(ctx, tenantID)
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

// loadPatients carrega todos os pacientes do tenant do cache; em caso de miss, busca no banco.
func (s *Service) loadPatients(ctx context.Context, tenantID uuid.UUID) ([]profile.Profile, error) {
	key := patientsCacheKey(tenantID)
	if cached, err := s.cache.Get(ctx, key); err == nil && cached != "" {
		var patients []profile.Profile
		if err := json.Unmarshal([]byte(cached), &patients); err == nil {
			return patients, nil
		}
	}

	patients, err := s.profileRepo.ListByRole(ctx, s.db, tenantID, RolePatient)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(patients); err == nil {
		ttl := 6 * time.Hour
		_ = s.cache.Set(ctx, key, string(data), ttl)
	}

	return patients, nil
}

func (s *Service) invalidatePatientCache(ctx context.Context, tenantID uuid.UUID) {
	errDel := s.cache.Del(ctx, patientsCacheKey(tenantID))

	keyPrefix, errDelPrefix := s.cache.DelWithPrefix(ctx, fmt.Sprintf("%s:patients:*", tenantID))

	logger.FromContext(ctx).
		WithField(logger.String("err Del", errDel.Error())).
		WithField(logger.String("err errDelPrefix", errDelPrefix.Error())).
		WithField(logger.String("keyPrefix", fmt.Sprintf("%v", keyPrefix))).Info("invalidatePatientCache")
}

// CreatePatient cria um paciente (Profile sem User vinculado) e invalida o cache.
func (s *Service) CreatePatient(ctx context.Context, req CreatePatientRequest) (*profile.Profile, error) {
	p := &profile.Profile{
		TenantID:              req.TenantID,
		Role:                  RolePatient,
		FullName:              req.FullName,
		Document:              req.Document,
		BirthDate:             req.BirthDate,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		Address:               req.Address,
	}
	if err := s.profileRepo.Create(ctx, s.db, p); err != nil {
		return nil, err
	}
	s.invalidatePatientCache(ctx, req.TenantID)
	return p, nil
}

// GetPatient retorna um paciente pelo ID do perfil.
func (s *Service) GetPatient(ctx context.Context, tenantID, id uuid.UUID) (*profile.Profile, error) {
	p, err := s.profileRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Role != RolePatient {
		return nil, ErrPatientNotFound
	}
	return p, nil
}

// UpdatePatient atualiza os dados pessoais de um paciente e invalida o cache.
func (s *Service) UpdatePatient(ctx context.Context, tenantID, id uuid.UUID, req UpdatePatientRequest) (*profile.Profile, error) {
	p, err := s.profileRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Role != RolePatient {
		return nil, ErrPatientNotFound
	}

	p.FullName = req.FullName
	p.Document = req.Document
	p.BirthDate = req.BirthDate
	p.Phone = req.Phone
	p.HasWhatsapp = req.HasWhatsapp
	p.EmergencyContactName = req.EmergencyContactName
	p.EmergencyContactPhone = req.EmergencyContactPhone
	p.Address = req.Address

	if err := s.profileRepo.Update(ctx, s.db, p); err != nil {
		return nil, err
	}
	s.invalidatePatientCache(ctx, tenantID)
	return p, nil
}

// DeletePatient remove um paciente (soft-delete somente do Profile) e invalida o cache.
func (s *Service) DeletePatient(ctx context.Context, tenantID, id uuid.UUID) error {
	p, err := s.profileRepo.FindByID(ctx, s.db, tenantID, id)
	if err != nil {
		return err
	}
	if p == nil || p.Role != RolePatient {
		return ErrPatientNotFound
	}
	if err := s.profileRepo.SoftDelete(ctx, s.db, tenantID, id); err != nil {
		return err
	}
	s.invalidatePatientCache(ctx, tenantID)
	return nil
}

// SearchPatients busca pacientes por nome parcial, insensível a maiúsculas e acentos.
// Usa a lista cacheada do tenant para filtrar em memória — sem roundtrip ao banco por busca.
func (s *Service) SearchPatients(ctx context.Context, tenantID uuid.UUID, query string) ([]profile.Profile, error) {
	q := normalizeForSearch(query)

	//busca search no cache
	cSearchkey := fmt.Sprintf("%s:patients:search:%s", tenantID, query)
	if cached, err := s.cache.Get(ctx, cSearchkey); err == nil && cached != "" {
		var cachedSearchPatients []profile.Profile
		if err := json.Unmarshal([]byte(cached), &cachedSearchPatients); err == nil {
			return cachedSearchPatients, nil
		}
	}

	patients, err := s.loadPatients(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]profile.Profile, 0)
	for _, p := range patients {
		if strings.Contains(normalizeForSearch(p.FullName), q) {
			result = append(result, p)
		}
	}

	if data, err := json.Marshal(result); err == nil {
		ttl := 40 * time.Minute
		_ = s.cache.Set(ctx, cSearchkey, string(data), ttl)
	}
	return result, nil
}

// func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
// 	parsed, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
// 		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("algoritmo inesperado: %v", t.Header["alg"])
// 		}
// 		return []byte(s.jwtSecret), nil
// 	})
// 	if err != nil || !parsed.Valid {
// 		return "", "", time.Time{}, ErrInvalidCreds
// 	}

// 	claims, ok := parsed.Claims.(jwt.MapClaims)
// 	if !ok || claims["type"] != "refresh" {
// 		return "", "", time.Time{}, ErrInvalidCreds
// 	}

// 	userID, err := uuid.Parse(fmt.Sprintf("%v", claims["sub"]))
// 	if err != nil {
// 		return "", "", time.Time{}, ErrInvalidCreds
// 	}
// 	tID, err := uuid.Parse(fmt.Sprintf("%v", claims["tenant_id"]))
// 	if err != nil {
// 		return "", "", time.Time{}, ErrInvalidCreds
// 	}

// 	u, err := s.userRepo.FindByID(ctx, s.db, userID)
// 	if err != nil {
// 		return "", "", time.Time{}, err
// 	}
// 	if u == nil {
// 		return "", "", time.Time{}, ErrInvalidCreds
// 	}

// 	prof, err := s.profileRepo.FindByUserAndTenant(ctx, s.db, userID, tID)
// 	if err != nil {
// 		return "", "", time.Time{}, err
// 	}
// 	if prof == nil {
// 		return "", "", time.Time{}, ErrTenantForbidden
// 	}

// 	newToken, expiresAt, err := s.issueJWT(u, prof)
// 	if err != nil {
// 		return "", "", time.Time{}, err
// 	}
// 	newRefresh, _, err := s.issueRefreshJWT(u, prof, false)

// 	return newToken, newRefresh, expiresAt, err
// }

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
	p := &profile.Profile{TenantID: inv.TenantID, Role: inv.Role}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if existing != nil {
			p.UserID = &existing.ID
			u = existing
		} else {
			newUser := &User{
				Email:        inv.Email,
				PasswordHash: inv.PasswordHash,
				Salt:         inv.Salt,
				LastTenantID: inv.TenantID,
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
		return "", err
	}

	token, _, err := s.issueJWT(u, p, defaultTokenDuration, "access")
	return token, err
}

func (s *Service) sendVerificationEmail(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("https://localhost:8080/v1/api/auth/verify-email?token=%s", token) //TODO

	body := fmt.Sprintf(
		`<p>Bem-vindo ao mirandaclin! Confirme seu email para ativar sua conta.</p>`+
			`<p><a href="%s">Clique aqui para verificar seu email</a></p>`+
			`<p>O link expira em 24 horas.</p>`, link)

	err := s.mailer.Send(ctx, to, "Confirme seu email — mirandaclin", body)
	if err != nil {
		logger.FromContext(ctx).Error("falha ao enviar email de verificação",
			logger.String("to", to),
			logger.Err(err),
		)
	}
	return err
}

// func (s *Service) issueRefreshJWT(u *User, p *profile.Profile, keep bool) (string, int, error) {
// 	var maxAgeSecRefresh int = 12 * 60 * 60
// 	if keep {
// 		maxAgeSecRefresh = 7 * 24 * 60 * 60
// 	}

// 	claims := jwt.MapClaims{
// 		"jti":       uuid.New().String(),
// 		"sub":       u.ID.String(),
// 		"tenant_id": p.TenantID.String(),
// 		"type":      "refresh",
// 		"exp":       time.Now().Add(time.Duration(maxAgeSecRefresh) * time.Second).Unix(),
// 		"iat":       time.Now().Unix(),
// 	}

// 	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
// 	return token, maxAgeSecRefresh, err
// }

// func (s *Service) issueJWT(u *User, p *profile.Profile) (string, time.Time, error) {
// 	expiresAt := time.Now().Add(time.Hour)
// 	claims := jwt.MapClaims{
// 		"jti":            uuid.New().String(),
// 		"sub":            u.ID.String(),
// 		"tenant_id":      p.TenantID.String(),
// 		"role":           p.Role,
// 		"scope":          ScopeForRole(p.Role),
// 		"email_verified": u.EmailVerifiedAt != nil,
// 		"exp":            expiresAt.Unix(),
// 		"iat":            time.Now().Unix(),
// 	}
// 	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
// 	return token, expiresAt, err
// }

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
