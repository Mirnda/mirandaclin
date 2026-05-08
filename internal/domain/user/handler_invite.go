package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
	"github.com/Mirnda/mirandaclin/pkg/validator"
)

type acceptInviteRequest struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// @Summary     Aceitar convite e criar conta
// @Tags        invites
// @Accept      json
// @Produce     json
// @Param       body body acceptInviteRequest true "Token do convite"
// @Success     201 {object} response.Response{data=map[string]string}
// @Failure     400 {object} response.Response
// @Failure     422 {object} response.Response
// @Router      /v1/api/invites/accept [post]
func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "payload inválido")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	loginResp, err := h.svc.AcceptInvite(r.Context(), AcceptInviteRequest{
		Token:    req.Token,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, invite.ErrInvalidInvite) {
			response.Error(w, http.StatusUnprocessableEntity, "convite inválido ou expirado")
			return
		}

		logger.FromContext(r.Context()).Error("erro ao aceitar convite", logger.Err(err))
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}

	http.SetCookie(w, loginResp.RefreshCookie)
	response.Created(w, "cadastro realizado com sucesso", map[string]any{
		"token":      loginResp.AccessToken,
		"expires_at": loginResp.AccessExpirationTime,
	})
}
