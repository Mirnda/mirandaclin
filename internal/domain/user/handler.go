package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/middleware"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
	"github.com/Mirnda/mirandaclin/pkg/validator"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
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
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type acceptInviteRequest struct {
	Token string `json:"token" validate:"required"`
}

// @Summary     Criar usuário
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

	token, err := h.svc.AcceptInvite(r.Context(), AcceptInviteRequest{Token: req.Token})
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

// @Summary     Login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body loginRequest true "Credenciais"
// @Success     200 {object} response.Response{data=map[string]string}
// @Failure     401 {object} response.Response
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

	tenantID := middleware.TenantFromContext(ctx)
	token, err := h.svc.Login(ctx, LoginRequest{
		TenantID: tenantID,
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, ErrInvalidCreds) {
		log.Error("credenciais inválidas", logger.Err(err))
		response.Error(w, http.StatusUnauthorized, "credenciais inválidas")
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
	tenantID := middleware.TenantFromContext(r.Context())
	u, err := h.svc.GetByID(r.Context(), tenantID, id)
	if errors.Is(err, ErrUserNotFound) {
		response.Error(w, http.StatusNotFound, "usuário não encontrado")
		return
	}
	if err != nil {
		logger.FromContext(r.Context()).Error("erro ao buscar usuário",
			logger.String("tenant_id", tenantID.String()),
			logger.String("user_id", id.String()),
			logger.Err(err),
		)
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "ok", u)
}
