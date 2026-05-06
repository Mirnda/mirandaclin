package user

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrUnexpectedAlgorithm = fmt.Errorf("algoritmo inesperado")

const defaultTokenDuration time.Duration = 15 * time.Minute

type LoginRequest struct {
	Email     string
	Password  string
	TenantID  uuid.UUID // opcional quando usuário pertence a múltiplos tenants
	KeepAlive bool
}

type LoginResponse struct {
	AccessToken          string
	AccessExpirationTime time.Time
	RefreshCookie        *http.Cookie
}

// *Service) Login(ctx context.Context, req LoginRequest) (token string, refreshToken string, maxAgeSecRefresh int, expiresAt time.Time, err error) {
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	log := logger.FromContext(ctx)

	u, err := s.userRepo.FindByEmail(ctx, s.db, req.Email)
	if err != nil {
		log.WithErr(err).Warn("failed to get user by email")
		return nil, err
	}
	if u == nil {
		log.WithErr(ErrInvalidCreds).Warn("provided email does not have a registered user")
		return nil, ErrInvalidCreds
	}

	log = log.With("user_id", u.ID.String())

	if u.EmailVerifiedAt.IsZero() || u.EmailVerifiedAt.After(time.Now()) {
		log.WithErr(ErrUserNotVerified).Warn("not allowed to log in before email confirmation")
		return nil, ErrUserNotVerified
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password+u.Salt)); err != nil {
		log.WithErr(ErrUserNotVerified).Warn("invalid password")
		return nil, ErrUserNotVerified
	}

	profiles, err := s.profileRepo.FindByUserID(ctx, s.db, u.ID)
	if err != nil {
		log.WithErr(err).Warn("failed to get profile by user")
		return nil, err
	}
	if len(profiles) == 0 {
		log.WithErr(ErrUserNotVerified).Warn("user does not have a registered profile")
		return nil, ErrUserNotVerified
	}

	var prof *profile.Profile
	if req.TenantID != uuid.Nil {
		log.Debug("use request tenant")
		for _, p := range profiles {
			if p.TenantID == req.TenantID {
				prof = p
				break
			}
		}

	} else if u.LastTenantID != uuid.Nil {
		log.Debug("use user tenant")
		for _, p := range profiles {
			if p.TenantID == u.LastTenantID {
				prof = p
				break
			}
		}

	} else if len(profiles) == 1 {
		log.Debug("use single profile tenant")
		prof = profiles[0]

	} else {
		log.WithErr(ErrTenantRequired).Warn("tenant not found")
		return nil, ErrTenantRequired
	}

	if prof == nil {
		log.WithErr(ErrTenantForbidden).Warn("profile not found")
		return nil, ErrTenantForbidden
	}

	log = log.With("profile_id", prof.ID.String())

	if u.LastTenantID != prof.TenantID {
		log.Debug("update user tenant id as profile tenant id")
		u.LastTenantID = prof.TenantID
		if err = s.userRepo.Update(ctx, s.db, u); err != nil {
			log.WithErr(err).Warn("failed to update user tenant id")
		}
	}

	accessToken, expiresAt, err := s.issueJWT(u, prof, defaultTokenDuration, "access")
	if err != nil {
		log.WithErr(err).Warn("failed to issue access token")
		return nil, err
	}

	var refreshExpirationTime = 11 * time.Hour
	if req.KeepAlive {
		refreshExpirationTime = 14 * 24 * time.Hour
	}

	refreshToken, refreshExpiresAt, err := s.issueJWT(u, prof, refreshExpirationTime, "refresh")
	if err != nil {
		log.WithErr(err).Warn("failed to issue refresh token")
		return nil, err
	}

	refreshCookie := s.generateRefreshCookie(refreshToken, s.appURL == "production", refreshExpiresAt)

	response := LoginResponse{
		AccessToken:          accessToken,
		AccessExpirationTime: expiresAt,
		RefreshCookie:        refreshCookie,
	}

	log.Debug("login service OK")
	return &response, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, time.Time, error) {
	log := logger.FromContext(ctx)

	parsed, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.With("algorithm", fmt.Sprintf("%v", t.Header["alg"])).Warn("unexpected algorithm")
			return nil, ErrUnexpectedAlgorithm
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !parsed.Valid {
		log.WithErr(ErrInvalidCreds).Warn("invalid token")
		return "", time.Time{}, ErrInvalidCreds
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" {
		log.WithErr(ErrInvalidCreds).With("type", fmt.Sprintf("%v", claims["type"])).Warn("invalid claims")
		return "", time.Time{}, ErrInvalidCreds
	}
	userID, err := uuid.Parse(fmt.Sprintf("%v", claims["sub"]))
	if err != nil {
		log.WithErr(ErrInvalidCreds).With("sub", fmt.Sprintf("%v", claims["sub"])).Warn("failed to parse user id from claims")
		return "", time.Time{}, ErrInvalidCreds
	}
	tID, err := uuid.Parse(fmt.Sprintf("%v", claims["tenant_id"]))
	if err != nil {
		log.WithErr(ErrInvalidCreds).With("tenant_id", fmt.Sprintf("%v", claims["tenant_id"])).Warn("failed to parse tenant id from claims")
		return "", time.Time{}, ErrInvalidCreds
	}

	log = log.With("user_id", userID.String())
	log = log.With("tenant_id", tID.String())

	u, err := s.userRepo.FindByID(ctx, s.db, userID)
	if err != nil {
		log.WithErr(err).Warn("failed to find user by id")
		return "", time.Time{}, err
	}
	if u == nil {
		log.WithErr(ErrInvalidCreds).Warn("user not found")
		return "", time.Time{}, ErrInvalidCreds
	}

	prof, err := s.profileRepo.FindByUserAndTenant(ctx, s.db, userID, tID)
	if err != nil {
		log.WithErr(err).Warn("failed to find profile by user id")
		return "", time.Time{}, err
	}
	if prof == nil {
		log.WithErr(ErrTenantForbidden).Warn("profile not found")
		return "", time.Time{}, ErrTenantForbidden
	}

	log = log.With("profile_id", prof.ID.String())

	newaccessToken, expiresAt, err := s.issueJWT(u, prof, defaultTokenDuration, "access")
	if err != nil {
		log.WithErr(err).Warn("failed to issue access token")
		return "", time.Time{}, err
	}

	return newaccessToken, expiresAt, err
}

func (s *Service) issueJWT(u *User, p *profile.Profile, expirationTime time.Duration, tokenType string) (string, time.Time, error) {
	expiresAt := time.Now().Add(expirationTime)

	claims := jwt.MapClaims{
		"jti":       uuid.New().String(),
		"sub":       u.ID.String(),
		"tenant_id": p.TenantID.String(),
		"role":      p.Role,
		"scope":     ScopeForRole(p.Role),
		"exp":       expiresAt.Unix(),
		"iat":       time.Now().Unix(),
		"type":      tokenType, // access | refresh
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
	return token, expiresAt, err
}

func (s *Service) generateRefreshCookie(refreshToken string, secure bool, expireTime time.Time) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	return &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/v1/api/auth/refresh",
		Expires:  expireTime,
	}
}
