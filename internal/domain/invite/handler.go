package invite

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

type createInviteRequest struct {
	Email     string    `json:"email"      validate:"required,email"`
	ProfileID uuid.UUID `json:"profile_id" validate:"required"`
}

// @Summary     Enviar convite por email para um perfil sem usuário
// @Tags        invites
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body createInviteRequest true "Dados do convite"
// @Success     201 {object} response.Response
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Failure     409 {object} response.Response
// @Failure     422 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /v1/api/invites [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if errs := validator.Validate(req); errs != nil {
		log.Error("erro ao gerar convite", logger.String("validator", fmt.Sprintf("%v", errs)))
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID := middleware.TenantFromContext(ctx)
	log = log.WithField(logger.String("tenant_id", tenantID.String()))

	err := h.svc.Create(ctx, CreateRequest{
		TenantID:  tenantID,
		Email:     req.Email,
		ProfileID: req.ProfileID,
	})
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			response.Error(w, http.StatusNotFound, "perfil não encontrado")
			return
		}
		if errors.Is(err, ErrProfileAlreadyLinked) {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrInvalidProfileRole) {
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		log.Error("erro ao gerar convite", logger.Err(err))
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Info("convite enviado com sucesso")
	response.Created(w, "convite enviado com sucesso", nil)
}
