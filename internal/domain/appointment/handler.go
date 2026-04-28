package appointment

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mirandev/mirandaclin/internal/middleware"
	"github.com/mirandev/mirandaclin/pkg/response"
	"github.com/mirandev/mirandaclin/pkg/validator"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createAppointmentRequest struct {
	PatientID   string  `json:"patient_id"    validate:"required,uuid"`
	DentistID   string  `json:"dentist_id"    validate:"required,uuid"`
	ClinicID    string  `json:"clinic_id"     validate:"required,uuid"`
	SecretaryID *string `json:"secretary_id"`
	ScheduledAt string  `json:"scheduled_at"  validate:"required"`
	Notes       string  `json:"notes"`
}

// @Summary     Criar agendamento
// @Tags        appointments
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createAppointmentRequest true "Dados do agendamento"
// @Success     201 {object} response.Response{data=Appointment}
// @Failure     400 {object} response.Response
// @Failure     422 {object} response.Response
// @Router      /v1/api/appointments [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "scheduled_at deve estar no formato RFC3339")
		return
	}

	tenantID := middleware.TenantFromContext(r.Context())
	patientID, _ := uuid.Parse(req.PatientID)
	dentistID, _ := uuid.Parse(req.DentistID)
	clinicID, _ := uuid.Parse(req.ClinicID)

	var secretaryID *uuid.UUID
	if req.SecretaryID != nil {
		id, _ := uuid.Parse(*req.SecretaryID)
		secretaryID = &id
	}

	a, err := h.svc.Create(r.Context(), CreateRequest{
		TenantID:    tenantID,
		PatientID:   patientID,
		DentistID:   dentistID,
		ClinicID:    clinicID,
		SecretaryID: secretaryID,
		ScheduledAt: scheduledAt,
		Notes:       req.Notes,
	})

	switch {
	case errors.Is(err, ErrDentistNotActive):
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrOutsideWorkingDays):
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrDentistBlocked):
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	case err != nil:
		response.Error(w, http.StatusInternalServerError, "erro interno")
	default:
		response.Created(w, "agendamento criado com sucesso", a)
	}
}

// @Summary     Listar agendamentos do paciente
// @Tags        appointments
// @Security    BearerAuth
// @Produce     json
// @Param       patient_id path string true "Patient ID"
// @Success     200 {object} response.Response{data=[]Appointment}
// @Router      /v1/api/appointments/patient/{patient_id} [get]
func (h *Handler) ListByPatient(w http.ResponseWriter, r *http.Request) {
	patientID, err := uuid.Parse(r.PathValue("patient_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "patient_id inválido")
		return
	}
	tenantID := middleware.TenantFromContext(r.Context())
	items, err := h.svc.ListByPatient(r.Context(), tenantID, patientID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "ok", items)
}

// @Summary     Cancelar agendamento
// @Tags        appointments
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Appointment ID"
// @Success     200 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/appointments/{id}/cancel [patch]
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	tenantID := middleware.TenantFromContext(r.Context())
	if err := h.svc.Cancel(r.Context(), tenantID, id); err != nil {
		if errors.Is(err, ErrAppointmentNotFound) {
			response.Error(w, http.StatusNotFound, "agendamento não encontrado")
			return
		}
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "agendamento cancelado", nil)
}
