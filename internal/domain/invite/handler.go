package invite

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	TenantID string `json:"tenant_id" validate:"required,uuid"`
	Email    string `json:"email"     validate:"required,email"`
	Role     string `json:"role"      validate:"required,oneof=admin dentist secretary patient"`
	Password string `json:"password"  validate:"required,min=8"`
}

// @Summary     Gerar convite por email
// @Tags        invites
// @Security    ApiKeyAuth
// @Accept      json
// @Produce     json
// @Param       body body createInviteRequest true "Dados do convite"
// @Success     201 {object} response.Response{data=Invite}
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /v1/api/invites [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	var log = logger.FromContext(ctx)

	log.Info("init createInviteRequest")

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("erro ao gerar convite",
			logger.Err(err),
		)
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}

	log = log.With(logger.String("tenant_id", req.TenantID))

	if errs := validator.Validate(req); errs != nil {
		log.Error("erro ao gerar convite",
			logger.String("validator", fmt.Sprintf("%v", errs)),
		)
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	tenantID, _ := uuid.Parse(req.TenantID)

	inv, err := h.svc.Create(ctx, CreateRequest{
		TenantID: tenantID,
		Email:    req.Email,
		Role:     req.Role,
		Password: req.Password,
	})
	if err != nil {
		log.Error("erro ao gerar convite",
			logger.Err(err),
		)
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Info("convite enviado com sucesso")
	response.Created(w, "convite enviado com sucesso", inv)
}
