package profile

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/shared"
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

type createProfileRequest struct {
	Role                  string         `json:"role" validate:"required"`
	FullName              string         `json:"full_name" validate:"required,min=4"`
	Document              string         `json:"document" validate:"omitempty,min=8"`
	BirthDate             *time.Time     `json:"birth_date"`
	Phone                 string         `json:"phone" validate:"omitempty,min=8"`
	HasWhatsapp           bool           `json:"has_whatsapp"`
	EmergencyContactName  string         `json:"emergency_contact_name" validate:"omitempty,min=4"`
	EmergencyContactPhone string         `json:"emergency_contact_phone" validate:"omitempty,min=8"`
	Address               shared.Address `json:"address"`
}

// @Summary     Criar perfil no tenant
// @Tags        profiles
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createProfileRequest true "Dados do perfil"
// @Success     201 {object} response.Response{data=Profile}
// @Failure     400 {object} response.Response
// @Router      /v1/api/profiles [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	tenantID := middleware.TenantFromContext(ctx)
	loggedUserID := middleware.UserIDFromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("profile decode error")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("profile validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	p, err := h.svc.Create(ctx, CreateProfile{
		TenantID:              tenantID,
		Role:                  req.Role,
		FullName:              req.FullName,
		Document:              req.Document,
		BirthDate:             req.BirthDate,
		Phone:                 req.Phone,
		HasWhatsapp:           req.HasWhatsapp,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		Address:               req.Address,
	})
	if err != nil {
		if errors.Is(err, shared.ErrInvalidRole) {
			log.WithErr(err).Warn("invalid profile role")
			response.Error(w, http.StatusNotFound, "função / cargo do perfil é inválida")
			return
		}
		log.WithErr(err).Error("failed to create profile")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.With("profile_id", p.ID.String()).
		With("created_by", loggedUserID.String()).
		Debug("profile created")
	response.Created(w, "consulta registrada com sucesso", p)
}

// @Summary     Buscar perfis por função
// @Tags        profiles
// @Security    BearerAuth
// @Produce     json
// @Param       role path string true "Profile Role"
// @Success     200 {object} response.Response{data=[]Profile}
// @Failure     400 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/profiles/role/{role} [get]
func (h *Handler) ListByRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	tenantID := middleware.TenantFromContext(ctx)

	role := r.PathValue("role")

	profiles, err := h.svc.ListByRole(ctx, tenantID, role)
	if err != nil {
		log.WithErr(err).Error("failed to list profiles")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("profiles by role listed successfully")
	response.OK(w, "ok", profiles)
}

// @Summary     Obter perfil por ID
// @Tags        profiles
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Profile ID"
// @Success     200 {object} response.Response{data=Profile}
// @Failure     400 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/profiles/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	tenantID := middleware.TenantFromContext(ctx)

	profileID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("invalid profile id")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	log = log.With("profile_id", profileID.String())

	p, err := h.svc.Get(ctx, tenantID, profileID)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			log.WithErr(err).Warn("profile not found")
			response.Error(w, http.StatusNotFound, "perfil não encontrado")
			return
		}
		log.WithErr(err).Error("failed to get profile")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("profile found")
	response.OK(w, "ok", p)
}

type updateProfileRequest struct {
	UserID                **uuid.UUID
	Role                  *string
	FullName              *string
	Document              *string
	BirthDate             **time.Time
	Phone                 *string
	HasWhatsapp           *bool
	EmergencyContactName  *string
	EmergencyContactPhone *string
	Address               *shared.Address
}

// @Summary     Atualizar perfil
// @Tags        profiles
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Profile ID"
// @Param       body body updateProfileRequest true "Dados do perfil"
// @Success     200 {object} response.Response{data=Profile}
// @Failure     400 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/profiles/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	tenantID := middleware.TenantFromContext(ctx)

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("profile id is required")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("profile decode error")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("profile validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	p, err := h.svc.Update(ctx, tenantID, id, UpdateProfile(req))
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			log.WithErr(err).Warn("profile not found")
			response.Error(w, http.StatusNotFound, "perfil não encontrado")
			return
		}
		log.WithErr(err).Error("failed to update profile")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("profile updated")
	response.OK(w, "perfil atualizado com sucesso", p)
}

// @Summary     Remover perfil no tenant
// @Tags        profiles
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Profile ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/profiles/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	tenantID := middleware.TenantFromContext(ctx)

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("profile id is required")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	if err := h.svc.Delete(ctx, tenantID, id); err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			log.WithErr(err).Warn("profile not found")
			response.Error(w, http.StatusNotFound, "perfil não encontrado")
			return
		}
		log.WithErr(err).Error("failed to delete profile")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("profile deleted")
	response.OK(w, "perfil removido com sucesso", nil)
}
