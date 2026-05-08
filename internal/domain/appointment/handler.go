package appointment

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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

type createAppointmentRequest struct {
	PatientID   string `json:"patient_id"    validate:"required,uuid"`
	DentistID   string `json:"dentist_id"    validate:"required,uuid"`
	ClinicID    string `json:"clinic_id"     validate:"required,uuid"`
	ScheduledAt string `json:"scheduled_at"  validate:"required"`
	Notes       string `json:"notes"`
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
	var ctx = r.Context()
	var log = logger.FromContext(ctx)
	tenantID := middleware.TenantFromContext(ctx)
	loggedUserID := middleware.UserIDFromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("appointment decode error")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("appointment validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		log.WithErr(err).Warn("invalid format scheduled_at")
		response.Error(w, http.StatusBadRequest, "scheduled_at deve estar no formato RFC3339")
		return
	}

	patientID, _ := uuid.Parse(req.PatientID)
	dentistID, _ := uuid.Parse(req.DentistID)
	clinicID, _ := uuid.Parse(req.ClinicID)

	a, err := h.svc.Create(r.Context(), CreateRequest{
		TenantID:    tenantID,
		PatientID:   patientID,
		DentistID:   dentistID,
		ClinicID:    clinicID,
		SecretaryID: loggedUserID,
		ScheduledAt: scheduledAt,
		Notes:       req.Notes,
	})

	log.WithErr(err).
		Debug("appointment returns")

	if err != nil {
		log.WithErr(err).Warn("failed to create appointment")

		switch {
		case errors.Is(err, ErrDentistNotActive):
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, ErrOutsideWorkingDays):
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, ErrDentistBlocked):
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "erro interno")
		}
		return
	}

	log.With("appointment_id", a.ID.String()).Debug("appointment created")
	response.Created(w, "agendamento criado com sucesso", a)
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
