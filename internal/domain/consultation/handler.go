package consultation

import (
	"encoding/json"
	"net/http"

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

type createConsultationRequest struct {
	AppointmentID string `json:"appointment_id" validate:"required,uuid"`
	PatientID     string `json:"patient_id"     validate:"required,uuid"`
	Diagnosis     string `json:"diagnosis"      validate:"required"`
	Treatment     string `json:"treatment"      validate:"required"`
}

// @Summary     Criar relatório de consulta
// @Tags        consultations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createConsultationRequest true "Dados da consulta"
// @Success     201 {object} response.Response{data=Consultation}
// @Failure     400 {object} response.Response
// @Router      /v1/api/consultations [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createConsultationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(r.Context())
	dentistID := middleware.UserIDFromContext(r.Context())
	appointmentID, _ := uuid.Parse(req.AppointmentID)
	patientID, _ := uuid.Parse(req.PatientID)

	c, err := h.svc.Create(r.Context(), CreateRequest{
		TenantID:      tenantID,
		AppointmentID: appointmentID,
		PatientID:     patientID,
		DentistID:     dentistID,
		Diagnosis:     req.Diagnosis,
		Treatment:     req.Treatment,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.Created(w, "consulta registrada com sucesso", c)
}

// @Summary     Relatório de consultas do paciente
// @Tags        consultations
// @Security    BearerAuth
// @Produce     json
// @Param       patient_id path string true "Patient ID"
// @Success     200 {object} response.Response{data=[]Consultation}
// @Router      /v1/api/consultations/patient/{patient_id} [get]
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

// @Summary     Relatório de consultas do dentista
// @Tags        consultations
// @Security    BearerAuth
// @Produce     json
// @Param       dentist_id path string true "Dentist ID"
// @Success     200 {object} response.Response{data=[]Consultation}
// @Router      /v1/api/consultations/dentist/{dentist_id} [get]
func (h *Handler) ListByDentist(w http.ResponseWriter, r *http.Request) {
	dentistID, err := uuid.Parse(r.PathValue("dentist_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "dentist_id inválido")
		return
	}
	tenantID := middleware.TenantFromContext(r.Context())
	items, err := h.svc.ListByDentist(r.Context(), tenantID, dentistID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "ok", items)
}
