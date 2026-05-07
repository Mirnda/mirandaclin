package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
	"github.com/Mirnda/mirandaclin/pkg/validator"
)

type registerRequest struct {
	TenantName            string `json:"tenant_name"`
	Email                 string `json:"email"                   validate:"required,email"`
	Password              string `json:"password"                validate:"required,min=8"`
	FullName              string `json:"full_name"               validate:"required"`
	Document              string `json:"document"`
	Phone                 string `json:"phone"`
	HasWhatsapp           bool   `json:"has_whatsapp"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
}

// @Summary     Registro de nova clínica (cria tenant + usuário admin)
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body registerRequest true "Dados do registro"
// @Success     201 {object} response.Response{data=map[string]string}
// @Failure     400 {object} response.Response
// @Failure     409 {object} response.Response
// @Router      /v1/api/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithErr(err).Warn("register decode error")
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		log.WithErr(err).Warn("register validation returned an error")
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	err := h.svc.Register(ctx, RegisterRequest(req))
	if err != nil {
		if errors.Is(err, ErrEmailConflict) {
			log.WithErr(err).Warn("failed to list patients")
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrTenantConflict) {
			log.WithErr(err).Warn("failed to list patients")
			response.Error(w, http.StatusConflict, err.Error())
			return
		}

		log.WithErr(err).Error("failed to register patients")
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	log.Debug("user registred")
	response.Created(w, "usuário registrado com sucesso", map[string]string{"message": "verifique seu e-mail"})
}
