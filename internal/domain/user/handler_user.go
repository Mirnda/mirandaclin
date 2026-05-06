package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mirnda/mirandaclin/internal/middleware"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
	"github.com/Mirnda/mirandaclin/pkg/validator"
)

type createUserRequest struct {
	Email                 string `json:"email"                  validate:"required,email"`
	Password              string `json:"password"               validate:"required,min=8"`
	Role                  string `json:"role"                   validate:"required,oneof=admin dentist secretary"`
	Phone                 string `json:"phone"`
	HasWhatsapp           bool   `json:"has_whatsapp"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	FullName              string `json:"full_name"              validate:"required"`
	Document              string `json:"document"`
}

type updateUserRequest struct {
	FullName              string `json:"full_name"               validate:"required"`
	Phone                 string `json:"phone"`
	Document              string `json:"document"`
	HasWhatsapp           bool   `json:"has_whatsapp"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	Password              string `json:"password"`
}

// @Summary     Criar usuário staff no tenant (admin, dentist, secretary)
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createUserRequest true "Dados do usuário"
// @Success     201 {object} response.Response{data=UserWithProfile}
// @Failure     400 {object} response.Response
// @Failure     409 {object} response.Response
// @Router      /v1/api/users [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("user decode error")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("user validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(ctx)
	u, err := h.svc.Create(ctx, CreateRequest{
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

	if err != nil {
		if errors.Is(err, ErrEmailConflict) {
			log.WithErr(err).Warn("failed to create user")
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidRole) {
			log.WithErr(err).Warn("failed to create user")
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		log.WithErr(err).Error("failed to create user")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("user created")
	response.Created(w, "usuário criado com sucesso", u)
}

// @Summary     Listar usuários staff do tenant
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response{data=[]UserWithProfile}
// @Router      /v1/api/users [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	tenantID := middleware.TenantFromContext(ctx)
	users, err := h.svc.List(ctx, tenantID)
	if err != nil {
		log.WithErr(err).Error("failed to list users")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("users listed")
	response.OK(w, "ok", users)
}

// @Summary     Obter usuário staff por ID
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "User ID"
// @Success     200 {object} response.Response{data=UserWithProfile}
// @Failure     404 {object} response.Response
// @Router      /v1/api/users/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("user id is required")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	log = log.With("user_id", id.String())

	tenantID := middleware.TenantFromContext(ctx)
	u, err := h.svc.GetByID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.WithErr(err).Warn("user not found")
			response.Error(w, http.StatusNotFound, "usuário não encontrado")
			return
		}

		log.WithErr(err).Error("failed to get user")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("user found")
	response.OK(w, "ok", u)
}

// @Summary     Atualizar usuário staff
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "User ID"
// @Param       body body updateUserRequest true "Dados do usuário"
// @Success     200 {object} response.Response{data=UserWithProfile}
// @Failure     400 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/users/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("user id is required")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	log = log.With("user_id", id.String())

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("user decode error")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		log.WithErr(err).Warn("user validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(ctx)
	u, err := h.svc.Update(ctx, tenantID, id, UpdateRequest(req))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.WithErr(err).Warn("user not found")
			response.Error(w, http.StatusNotFound, "usuário não encontrado")
			return
		}

		log.WithErr(err).Error("failed to update user")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("user updated")
	response.OK(w, "usuário atualizado com sucesso", u)
}

// @Summary     Remover usuário staff do tenant
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "User ID"
// @Success     200 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/users/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("user id is required")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	log = log.With("user_id", id.String())

	tenantID := middleware.TenantFromContext(ctx)
	if err := h.svc.Delete(ctx, tenantID, id); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.WithErr(err).Warn("user not found")
			response.Error(w, http.StatusNotFound, "usuário não encontrado")
			return
		}

		log.WithErr(err).Error("failed to delete user")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("user deleted")
	response.OK(w, "usuário removido com sucesso", nil)
}
