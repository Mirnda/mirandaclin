package cep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
)

var nonDigit = regexp.MustCompile(`\D`)

type AddressResponse struct {
	PostalCode   string `json:"postal_code"`
	Street       string `json:"street"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

type Handler struct {
	cache cache.Cache
}

func NewHandler(c cache.Cache) *Handler {
	return &Handler{cache: c}
}

// @Summary     Consultar endereço por CEP
// @Tags        utils
// @Security    BearerAuth
// @Produce     json
// @Param       cep path string true "CEP (somente dígitos ou com hífen)"
// @Success     200 {object} response.Response{data=AddressResponse}
// @Failure     400 {object} response.Response
// @Failure     404 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /v1/api/cep/{cep} [get]
func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	cep := nonDigit.ReplaceAllString(r.PathValue("cep"), "")
	if len(cep) != 8 {
		log.Warn("CEP inválido — informe 8 dígitos")
		response.Error(w, http.StatusBadRequest, "CEP inválido — informe 8 dígitos")
		return
	}

	addr := h.getCache(ctx, cep)
	if addr != nil {
		log.Info("cache hit CEP")
		response.OK(w, "ok", addr)
		return
	}

	addr, err := h.fetchViaCEP(r.Context(), cep)
	if err != nil || addr == nil {
		addr, err = h.fetchBrasilAPI(r.Context(), cep)
	}
	if err != nil {
		logger.FromContext(r.Context()).Error("falha ao consultar CEP",
			logger.String("cep", cep),
			logger.Err(err),
		)
		response.Error(w, http.StatusInternalServerError, "erro ao consultar CEP")
		return
	}
	if addr == nil {
		response.Error(w, http.StatusNotFound, "CEP não encontrado")
		return
	}

	h.setCache(ctx, addr)
	response.OK(w, "ok", addr)
}

// --- ViaCEP ---

type viaCEPResponse struct {
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Erro        bool   `json:"erro"`
}

func (h *Handler) fetchViaCEP(ctx context.Context, cep string) (*AddressResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep), nil)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var v viaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	if v.Erro {
		return nil, nil
	}

	return &AddressResponse{
		PostalCode:   cep,
		Street:       v.Logradouro,
		Complement:   v.Complemento,
		Neighborhood: v.Bairro,
		City:         v.Localidade,
		State:        v.Uf,
	}, nil
}

// --- BrasilAPI ---

type brasilAPIResponse struct {
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

func (h *Handler) fetchBrasilAPI(ctx context.Context, cep string) (*AddressResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://brasilapi.com.br/api/cep/v2/%s", cep), nil)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("brasilapi: status %d", resp.StatusCode)
	}

	var b brasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}

	return &AddressResponse{
		PostalCode:   cep,
		Street:       b.Street,
		Neighborhood: b.Neighborhood,
		City:         b.City,
		State:        b.State,
	}, nil
}
