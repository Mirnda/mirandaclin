package user

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
)

type createPatientRequest struct {
	FullName              string         `json:"full_name"               validate:"required"`
	Document              string         `json:"document"`
	BirthDate             *time.Time     `json:"birth_date"`
	Phone                 string         `json:"phone"`
	HasWhatsapp           bool           `json:"has_whatsapp"`
	EmergencyContactName  string         `json:"emergency_contact_name"`
	EmergencyContactPhone string         `json:"emergency_contact_phone"`
	Address               shared.Address `json:"address"`
}

type updatePatientRequest struct {
	FullName              string         `json:"full_name"               validate:"required"`
	Document              string         `json:"document"`
	BirthDate             *time.Time     `json:"birth_date"`
	Phone                 string         `json:"phone"`
	HasWhatsapp           bool           `json:"has_whatsapp"`
	EmergencyContactName  string         `json:"emergency_contact_name"`
	EmergencyContactPhone string         `json:"emergency_contact_phone"`
	Address               shared.Address `json:"address"`
}

// @Summary     Criar paciente no tenant
// @Tags        patients
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createPatientRequest true "Dados do paciente"
// @Success     201 {object} response.Response{data=profile.Profile}
// @Failure     400 {object} response.Response
// @Router      /v1/api/patients [post]
func (h *Handler) CreatePatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createPatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("invalid payload")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("patient validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(ctx)
	p, err := h.svc.CreatePatient(ctx, CreatePatientRequest{
		TenantID:              tenantID,
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
		log.WithErr(err).Error("failed to create patient")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.With("patient_id", p.ID.String()).Debug("patient created")
	response.Created(w, "paciente criado com sucesso", p)
}

// @Summary     Listar pacientes do tenant
// @Tags        patients
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response{data=[]profile.Profile}
// @Router      /v1/api/patients [get]
func (h *Handler) ListPatients(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	tenantID := middleware.TenantFromContext(ctx)
	patients, err := h.svc.ListPatients(ctx, tenantID)
	if err != nil {
		log.WithErr(err).Error("failed to list patients")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("patients listed successfully")
	response.OK(w, "ok", patients)
}

// @Summary     Buscar pacientes por nome (parcial, sem distinção de acentos/maiúsculas)
// @Tags        patients
// @Security    BearerAuth
// @Produce     json
// @Param       q query string true "Termo de busca (ex: \"ale\")"
// @Success     200 {object} response.Response{data=[]profile.Profile}
// @Failure     400 {object} response.Response
// @Router      /v1/api/patients/search [get]
func (h *Handler) SearchPatients(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	q := r.URL.Query().Get("q")
	if q == "" {
		log.Warn("patient name is required")
		response.Error(w, http.StatusBadRequest, "nome do paciente é obrigatório")
		return
	}

	tenantID := middleware.TenantFromContext(ctx)
	patients, err := h.svc.SearchPatients(ctx, tenantID, q)
	if err != nil {
		log.WithErr(err).Error("failed to search patients")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("patients listed by name successfully")
	response.OK(w, "ok", patients)
}

// @Summary     Obter paciente por ID
// @Tags        patients
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Patient ID"
// @Success     200 {object} response.Response{data=profile.Profile}
// @Failure     404 {object} response.Response
// @Router      /v1/api/patients/{id} [get]
func (h *Handler) GetPatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("invalid patient id")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	log = log.With("patient_id", id.String())

	tenantID := middleware.TenantFromContext(ctx)
	p, err := h.svc.GetPatient(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, ErrPatientNotFound) {
			log.WithErr(err).Warn("patient not found")
			response.Error(w, http.StatusNotFound, "paciente não encontrado")
			return
		}
		log.WithErr(err).Error("failed to get patient")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("patient found")
	response.OK(w, "ok", p)
}

// @Summary     Atualizar paciente
// @Tags        patients
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Patient ID"
// @Param       body body updatePatientRequest true "Dados do paciente"
// @Success     200 {object} response.Response{data=profile.Profile}
// @Failure     400 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/patients/{id} [put]
func (h *Handler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("patient id is required")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("patient decode error")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("patient validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(ctx)
	p, err := h.svc.UpdatePatient(ctx, tenantID, id, UpdatePatientRequest(req))
	if err != nil {
		if errors.Is(err, ErrPatientNotFound) {
			log.WithErr(err).Warn("patient not found")
			response.Error(w, http.StatusNotFound, "paciente não encontrado")
			return
		}
		log.WithErr(err).Error("failed to update patient")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("patient updated")
	response.OK(w, "paciente atualizado com sucesso", p)
}

// @Summary     Remover paciente do tenant
// @Tags        patients
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Patient ID"
// @Success     200 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/patients/{id} [delete]
func (h *Handler) DeletePatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		log.WithErr(err).Warn("patient id is required")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	tenantID := middleware.TenantFromContext(ctx)
	if err := h.svc.DeletePatient(ctx, tenantID, id); err != nil {
		if errors.Is(err, ErrPatientNotFound) {
			log.WithErr(err).Warn("patient not found")
			response.Error(w, http.StatusNotFound, "paciente não encontrado")
			return
		}
		log.WithErr(err).Error("failed to delete patient")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("patient deleted")
	response.OK(w, "paciente removido com sucesso", nil)
}
