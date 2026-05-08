package collaborator

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	profileclinic "github.com/Mirnda/mirandaclin/internal/domain/profile_clinic"
	"github.com/Mirnda/mirandaclin/internal/domain/shared"
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

type createCollaboratorRequest struct {
	Role                  string                                     `json:"role"                    validate:"required"`
	FullName              string                                     `json:"full_name"               validate:"required"`
	Document              string                                     `json:"document"`
	BirthDate             *time.Time                                 `json:"birth_date"`
	Phone                 string                                     `json:"phone"`
	HasWhatsapp           bool                                       `json:"has_whatsapp"`
	EmergencyContactName  string                                     `json:"emergency_contact_name"`
	EmergencyContactPhone string                                     `json:"emergency_contact_phone"`
	Address               shared.Address                             `json:"address"`
	ProfileClinics        []profileclinic.CreateProfileClinicRequest `json:"profile_clinics"`
}

// @Summary     Criar colaborador no tenant
// @Tags        collaborator
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createCollaboratorRequest true "Dados do colaborador"
// @Success     201 {object} response.Response{data=Collaborator}
// @Failure     400 {object} response.Response
// @Router      /v1/api/collaborator [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createCollaboratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("collaborator decode error")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("collaborator validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(ctx)
	collaborator, err := h.svc.Create(ctx, CreateCollaboratorRequest{
		CreateProfile: profile.CreateProfile{
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
		},
		ProfileClinics: req.ProfileClinics,
	})
	if err != nil {
		if errors.Is(err, shared.ErrInvalidRole) {
			log.WithErr(err).With("role", req.Role).Warn("invalid role")
			response.Error(w, http.StatusBadRequest, shared.ErrInvalidRole.Error())
			return
		}
		if errors.Is(err, profileclinic.ErrClinicRequired) {
			log.WithErr(err).Warn("clinic id is required")
			response.Error(w, http.StatusBadRequest, profileclinic.ErrClinicRequired.Error())
			return
		}
		log.WithErr(err).Error("failed to create patient")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.With("profile_id", collaborator.Profile.ID.String()).Debug("collaborator created")
	response.Created(w, "colaborador criado com sucesso", collaborator)
}

// @Summary     Listar colaboradores
// @Tags        collaborators
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response{data=[]Collaborator}
// @Router      /v1/api/collaborator [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	tenantID := middleware.TenantFromContext(ctx)

	collaborators, err := h.svc.List(ctx, tenantID)
	if err != nil {
		log.WithErr(err).Warn("failed to list users")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("collaborators listed")
	response.OK(w, "ok", collaborators)
}
