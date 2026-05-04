package clinic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/Mirnda/mirandaclin/internal/middleware"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
	"github.com/Mirnda/mirandaclin/pkg/validator"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createClinicRequest struct {
	Name          string         `json:"name"           validate:"required"`
	Phone         string         `json:"phone"`
	Address       shared.Address `json:"address"`
	OperatingDays []string       `json:"operating_days"`
	OpenTime      string         `json:"open_time"`
	CloseTime     string         `json:"close_time"`
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
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

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

	tenantID := middleware.TenantFromContext(ctx)
	c := &Clinic{
		TenantID:      tenantID,
		Name:          req.Name,
		Phone:         req.Phone,
		Address:       req.Address,
		OperatingDays: pq.StringArray(req.OperatingDays),
		OpenTime:      req.OpenTime,
		CloseTime:     req.CloseTime,
	}
	if err := h.svc.Create(ctx, c); err != nil {
		log.Error("erro ao criar clínica",
			logger.String("tenant_id", tenantID.String()),
			logger.Err(err),
		)
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
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	tenantID := middleware.TenantFromContext(ctx)
	items, err := h.svc.List(ctx, tenantID)
	if err != nil {
		log.Error("erro ao listar clínicas",
			logger.String("tenant_id", tenantID.String()),
			logger.Err(err),
		)
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
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	tenantID := middleware.TenantFromContext(ctx)
	c, err := h.svc.GetByID(ctx, tenantID, id)
	if errors.Is(err, ErrClinicNotFound) {
		response.Error(w, http.StatusNotFound, "clínica não encontrada")
		return
	}
	if err != nil {
		log.Error("erro ao buscar clínica",
			logger.String("tenant_id", tenantID.String()),
			logger.String("clinic_id", id.String()),
			logger.Err(err),
		)
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "ok", c)
}

type updateClinicRequest struct {
	Name          *string              `json:"name"`
	Phone         *string              `json:"phone"`
	Address       *shared.AddressInput `json:"address"`
	OperatingDays *[]string            `json:"operating_days"`
	OpenTime      *string              `json:"open_time"`
	CloseTime     *string              `json:"close_time"`
}

// @Summary     Atualizar clínica
// @Tags        clinics
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string              true "Clinic ID"
// @Param       body body updateClinicRequest true "Campos a atualizar"
// @Success     200 {object} response.Response{data=Clinic}
// @Failure     400 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /v1/api/clinics/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		log.With(logger.Err(err)).Warn("erro ao atualizar clínica")
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateClinicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.With(logger.Err(err)).Warn("erro ao atualizar clínica")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		log.With(logger.String("validate", fmt.Sprintf("%#v", errs))).Warn("dados inválidos")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(ctx)

	c, err := h.svc.Update(ctx, tenantID, id, UpdateRequest(req))
	if errors.Is(err, ErrClinicNotFound) {
		log.With(logger.Err(err)).Warn("erro ao atualizar clínica")
		response.Error(w, http.StatusNotFound, "clínica não encontrada")
		return
	}
	if err != nil {
		log.Error("erro ao atualizar clínica",
			logger.String("tenant_id", tenantID.String()),
			logger.String("clinic_id", id.String()),
			logger.Err(err),
		)
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "clínica atualizada com sucesso", c)
}

// @Summary     Remover clínica
// @Tags        clinics
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Clinic ID"
// @Success     200 {object} response.Response
// @Router      /v1/api/clinics/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "id inválido")
		return
	}
	tenantID := middleware.TenantFromContext(ctx)
	if err := h.svc.Delete(ctx, tenantID, id); err != nil {
		log.Error("erro ao remover clínica",
			logger.String("tenant_id", tenantID.String()),
			logger.String("clinic_id", id.String()),
			logger.Err(err),
		)
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "clínica removida com sucesso", nil)
}
