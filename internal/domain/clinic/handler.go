package clinic

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mirandev/mirandaclin/internal/domain/shared"
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

type createClinicRequest struct {
	Name          string        `json:"name"           validate:"required"`
	Phone         string        `json:"phone"`
	Address       shared.Address `json:"address"`
	OperatingDays []string      `json:"operating_days"`
	OpenTime      string        `json:"open_time"`
	CloseTime     string        `json:"close_time"`
}

// @Summary     Criar clínica
// @Tags        clinics
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createClinicRequest true "Dados da clínica"
// @Success     201 {object} response.Response{data=Clinic}
// @Router      /v1/api/clinics [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createClinicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(r.Context())
	c := &Clinic{
		TenantID:      tenantID,
		Name:          req.Name,
		Phone:         req.Phone,
		Address:       req.Address,
		OperatingDays: pq.StringArray(req.OperatingDays),
		OpenTime:      req.OpenTime,
		CloseTime:     req.CloseTime,
	}
	if err := h.svc.Create(r.Context(), c); err != nil {
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.Created(w, "clínica criada com sucesso", c)
}

// @Summary     Listar clínicas
// @Tags        clinics
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response{data=[]Clinic}
// @Router      /v1/api/clinics [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantFromContext(r.Context())
	items, err := h.svc.List(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "ok", items)
}

// @Summary     Obter clínica por ID
// @Tags        clinics
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Clinic ID"
// @Success     200 {object} response.Response{data=Clinic}
// @Failure     404 {object} response.Response
// @Router      /v1/api/clinics/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	tenantID := middleware.TenantFromContext(r.Context())
	c, err := h.svc.GetByID(r.Context(), tenantID, id)
	if errors.Is(err, ErrClinicNotFound) {
		response.Error(w, http.StatusNotFound, "clínica não encontrada")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "ok", c)
}

// @Summary     Remover clínica
// @Tags        clinics
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Clinic ID"
// @Success     200 {object} response.Response
// @Router      /v1/api/clinics/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	tenantID := middleware.TenantFromContext(r.Context())
	if err := h.svc.Delete(r.Context(), tenantID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "clínica removida com sucesso", nil)
}
