package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	emailverification "github.com/Mirnda/mirandaclin/internal/domain/email_verification"
	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/middleware"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
	"github.com/Mirnda/mirandaclin/pkg/validator"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type registerRequest struct {
	TenantName            string `json:"tenant_name"`
	Email                 string `json:"email"                   validate:"required,email"`
	Password              string `json:"password"                validate:"required,min=8"`
	FullName              string `json:"full_name"               validate:"required"`
	Document              string `json:"document"`
	Phone                 string `json:"phone"`
	HasWhatsapp           bool   `json:"has_whatsapp"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
}

type createUserRequest struct {
	Email                 string `json:"email"                  validate:"required,email"`
	Password              string `json:"password"               validate:"required,min=8"`
	Role                  string `json:"role"                   validate:"required,oneof=admin dentist secretary patient"`
	Phone                 string `json:"phone"`
	HasWhatsapp           bool   `json:"has_whatsapp"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	FullName              string `json:"full_name"              validate:"required"`
	Document              string `json:"document"`
}

type loginRequest struct {
	Email    string    `json:"email"     validate:"required,email"`
	Password string    `json:"password"  validate:"required"`
	TenantID uuid.UUID `json:"tenant_id"` // opcional quando pertence a múltiplos tenants
}

type acceptInviteRequest struct {
	Token string `json:"token" validate:"required"`
}

// @Summary     Registro de nova clínica (cria tenant + usuário admin)
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body registerRequest true "Dados do registro"
// @Success     201 {object} response.Response{data=map[string]string}
// @Failure     400 {object} response.Response
// @Failure     409 {object} response.Response
// @Router      /v1/api/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.With(logger.Err(err)).Warn("payload inválido")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		log.With(logger.String("validate", fmt.Sprintf("%#v", errs))).Warn("dados inválidos")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	err := h.svc.Register(r.Context(), RegisterRequest(req))
	if errors.Is(err, ErrEmailConflict) {
		log.With(logger.Err(err)).Warn("Register err")
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, ErrTenantConflict) {
		log.With(logger.Err(err)).Warn("Register err")
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		log.Error("erro ao registrar usuário", logger.Err(err))
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.Created(w, "usuário registrado com sucesso", map[string]string{"message": "verifique seu e-mail"})
}

// @Summary     Criar usuário no tenant
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createUserRequest true "Dados do usuário"
// @Success     201 {object} response.Response{data=User}
// @Failure     400 {object} response.Response
// @Failure     409 {object} response.Response
// @Router      /v1/api/users [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(r.Context())
	u, err := h.svc.Create(r.Context(), CreateRequest{
		TenantID:              tenantID,
		Email:                 req.Email,
		Password:              req.Password,
		Role:                  req.Role,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		FullName:              req.FullName,
		Document:              req.Document,
	})
	if errors.Is(err, ErrEmailConflict) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		logger.FromContext(r.Context()).Error("erro ao criar usuário",
			logger.String("tenant_id", tenantID.String()),
			logger.Err(err),
		)
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.Created(w, "usuário criado com sucesso", u)
}

// @Summary     Aceitar convite e criar conta
// @Tags        invites
// @Accept      json
// @Produce     json
// @Param       body body acceptInviteRequest true "Token do convite"
// @Success     201 {object} response.Response{data=map[string]string}
// @Failure     400 {object} response.Response
// @Failure     409 {object} response.Response
// @Failure     422 {object} response.Response
// @Router      /v1/api/invites/accept [post]
func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	token, err := h.svc.AcceptInvite(r.Context(), AcceptInviteRequest(req))
	if errors.Is(err, invite.ErrInvalidInvite) {
		response.Error(w, http.StatusUnprocessableEntity, "convite inválido ou expirado")
		return
	}
	if errors.Is(err, ErrEmailConflict) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		logger.FromContext(r.Context()).Error("erro ao aceitar convite", logger.Err(err))
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.Created(w, "cadastro realizado com sucesso", map[string]string{"token": token})
}

// @Summary     Verificar email
// @Tags        auth
// @Produce     json
// @Param       token query string true "Token de verificação"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Failure     422 {object} response.Response
// @Router      /v1/api/auth/verify-email [get]
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	token := r.URL.Query().Get("token")
	if token == "" {
		response.Error(w, http.StatusBadRequest, "token obrigatório")
		return
	}

	err := h.svc.VerifyEmail(ctx, token)
	if errors.Is(err, emailverification.ErrInvalidToken) {
		log.Error("erro ao verificar email", logger.Err(err))
		response.Error(w, http.StatusUnprocessableEntity, "token inválido ou expirado")
		return
	}
	if err != nil {
		log.Error("erro ao verificar email", logger.Err(err))
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "email verificado com sucesso", nil)
}

// @Summary     Login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body loginRequest true "Credenciais"
// @Success     200 {object} response.Response{data=map[string]string}
// @Failure     401 {object} response.Response
// @Failure     422 {object} response.Response
// @Router      /v1/api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	token, err := h.svc.Login(ctx, LoginRequest(req))
	if errors.Is(err, ErrInvalidCreds) {
		log.Warn("credenciais inválidas", logger.String("email", req.Email))
		response.Error(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}
	if errors.Is(err, ErrTenantRequired) {
		response.Error(w, http.StatusUnprocessableEntity, ErrTenantRequired.Error())
		return
	}
	if errors.Is(err, ErrTenantForbidden) {
		response.Error(w, http.StatusForbidden, ErrTenantForbidden.Error())
		return
	}
	if err != nil {
		log.Error("erro ao autenticar usuário", logger.Err(err))
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "autenticado com sucesso", map[string]string{"token": token})
}

// @Summary     Obter usuário por ID
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "User ID"
// @Success     200 {object} response.Response{data=User}
// @Failure     404 {object} response.Response
// @Router      /v1/api/users/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	u, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, ErrUserNotFound) {
		response.Error(w, http.StatusNotFound, "usuário não encontrado")
		return
	}
	if err != nil {
		logger.FromContext(r.Context()).Error("erro ao buscar usuário",
			logger.String("user_id", id.String()),
			logger.Err(err),
		)
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "ok", u)
}
