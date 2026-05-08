package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
	"github.com/Mirnda/mirandaclin/pkg/validator"

	"github.com/google/uuid"
)

type loginRequest struct {
	Email     string    `json:"email"     validate:"required,email"`
	Password  string    `json:"password"  validate:"required"`
	TenantID  uuid.UUID `json:"tenant_id"`
	KeepAlive bool      `json:"keep_alive"`
}

// @Summary     Login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body loginRequest true "Credenciais"
// @Success     200 {object} response.Response{data=map[string]interface{}}
// @Failure     401 {object} response.Response
// @Failure     422 {object} response.Response
// @Router      /v1/api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Debug("invalid payload")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("login validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	loginResponse, err := h.svc.Login(ctx, LoginRequest(req))
	if err != nil {
		if errors.Is(err, ErrInvalidCreds) {
			log.WithErr(err).With("email", req.Email).Debug("invalid credentials")
			response.Error(w, http.StatusUnauthorized, "credenciais inválidas")
			return
		}
		if errors.Is(err, ErrTenantRequired) {
			log.WithErr(err).Warn("error tenant required")
			response.Error(w, http.StatusUnprocessableEntity, ErrTenantRequired.Error())
			return
		}
		if errors.Is(err, ErrTenantForbidden) {
			log.WithErr(err).Warn("error tenant forbidden")
			response.Error(w, http.StatusForbidden, ErrTenantForbidden.Error())
			return
		}
		if errors.Is(err, ErrUserNotVerified) {
			log.WithErr(err).With("email", req.Email).Debug("verify email is required")
			response.Error(w, http.StatusUnauthorized, "usuário não confirmou email")
			return
		}

		log.WithErr(err).Warn("internal error")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	http.SetCookie(w, loginResponse.RefreshCookie)

	log.Debug("login success")
	response.OK(w, "autenticado com sucesso", map[string]any{
		"token":      loginResponse.AccessToken,
		"expires_at": loginResponse.AccessExpirationTime,
	})
}

// @Summary     Renovar access token via refresh token em cookie
// @Tags        auth
// @Produce     json
// @Success     200 {object} response.Response{data=map[string]interface{}}
// @Failure     401 {object} response.Response
// @Router      /v1/api/auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		log.WithErr(err).Warn("missing token")
		response.Error(w, http.StatusUnauthorized, "refresh token ausente")
		return
	}

	newAccessToken, expiresAt, err := h.svc.Refresh(ctx, cookie.Value)
	if errors.Is(err, ErrInvalidCreds) || errors.Is(err, ErrTenantForbidden) {
		log.WithErr(err).Warn("invalid session")
		response.Error(w, http.StatusUnauthorized, "sessão inválida")
		return
	}
	if err != nil {
		log.WithErr(err).Warn("failed to renew token")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("token updated")
	response.OK(w, "token renovado com sucesso", map[string]any{
		"token":      newAccessToken,
		"expires_at": expiresAt,
	})
}
