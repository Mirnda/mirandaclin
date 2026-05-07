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

type Handler struct {
	svc    *Service
	appEnv string
}

func NewHandler(svc *Service, appEnv string) *Handler {
	return &Handler{svc: svc, appEnv: appEnv}
}

// type registerRequest struct {
// 	TenantName            string `json:"tenant_name"`
// 	Email                 string `json:"email"                   validate:"required,email"`
// 	Password              string `json:"password"                validate:"required,min=8"`
// 	FullName              string `json:"full_name"               validate:"required"`
// 	Document              string `json:"document"`
// 	Phone                 string `json:"phone"`
// 	HasWhatsapp           bool   `json:"has_whatsapp"`
// 	EmergencyContactName  string `json:"emergency_contact_name"`
// 	EmergencyContactPhone string `json:"emergency_contact_phone"`
// }

// type createUserRequest struct {
// 	Email                 string `json:"email"                  validate:"required,email"`
// 	Password              string `json:"password"               validate:"required,min=8"`
// 	Role                  string `json:"role"                   validate:"required,oneof=admin dentist secretary"`
// 	Phone                 string `json:"phone"`
// 	HasWhatsapp           bool   `json:"has_whatsapp"`
// 	EmergencyContactName  string `json:"emergency_contact_name"`
// 	EmergencyContactPhone string `json:"emergency_contact_phone"`
// 	FullName              string `json:"full_name"              validate:"required"`
// 	Document              string `json:"document"`
// }

// type createPatientRequest struct {
// 	FullName              string         `json:"full_name"               validate:"required"`
// 	Document              string         `json:"document"`
// 	BirthDate             *time.Time     `json:"birth_date"`
// 	Phone                 string         `json:"phone"`
// 	HasWhatsapp           bool           `json:"has_whatsapp"`
// 	EmergencyContactName  string         `json:"emergency_contact_name"`
// 	EmergencyContactPhone string         `json:"emergency_contact_phone"`
// 	Address               shared.Address `json:"address"`
// }

// type updatePatientRequest struct {
// 	FullName              string         `json:"full_name"               validate:"required"`
// 	Document              string         `json:"document"`
// 	BirthDate             *time.Time     `json:"birth_date"`
// 	Phone                 string         `json:"phone"`
// 	HasWhatsapp           bool           `json:"has_whatsapp"`
// 	EmergencyContactName  string         `json:"emergency_contact_name"`
// 	EmergencyContactPhone string         `json:"emergency_contact_phone"`
// 	Address               shared.Address `json:"address"`
// }

// type loginRequest struct {
// 	Email    string    `json:"email"     validate:"required,email"`
// 	Password string    `json:"password"  validate:"required"`
// 	TenantID uuid.UUID `json:"tenant_id"`
// }

type acceptInviteRequest struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// // @Summary     Registro de nova clínica (cria tenant + usuário admin)
// // @Tags        auth
// // @Accept      json
// // @Produce     json
// // @Param       body body registerRequest true "Dados do registro"
// // @Success     201 {object} response.Response{data=map[string]string}
// // @Failure     400 {object} response.Response
// // @Failure     409 {object} response.Response
// // @Router      /v1/api/auth/register [post]
// func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	log := logger.FromContext(ctx)

// 	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
// 	var req registerRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		log.WithField(logger.Err(err)).Warn("payload inválido")
// 		response.Error(w, http.StatusBadRequest, "payload inválido")
// 		return
// 	}
// 	if errs := validator.Validate(req); errs != nil {
// 		log.WithField(logger.String("validate", fmt.Sprintf("%#v", errs))).Warn("dados inválidos")
// 		response.Error(w, http.StatusBadRequest, "dados inválidos")
// 		return
// 	}

// 	err := h.svc.Register(ctx, RegisterRequest(req))
// 	if errors.Is(err, ErrEmailConflict) {
// 		response.Error(w, http.StatusConflict, err.Error())
// 		return
// 	}
// 	if errors.Is(err, ErrTenantConflict) {
// 		response.Error(w, http.StatusConflict, err.Error())
// 		return
// 	}
// 	if err != nil {
// 		log.Error("erro ao registrar usuário", logger.Err(err))
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.Created(w, "usuário registrado com sucesso", map[string]string{"message": "verifique seu e-mail"})
// }

// // @Summary     Criar usuário staff no tenant (admin, dentist, secretary)
// // @Tags        users
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       body body createUserRequest true "Dados do usuário"
// // @Success     201 {object} response.Response{data=UserWithProfile}
// // @Failure     400 {object} response.Response
// // @Failure     409 {object} response.Response
// // @Router      /v1/api/users [post]
// func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
// 	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
// 	var req createUserRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		response.Error(w, http.StatusBadRequest, "payload inválido")
// 		return
// 	}
// 	if errs := validator.Validate(req); errs != nil {
// 		response.Error(w, http.StatusBadRequest, "dados inválidos")
// 		return
// 	}

// 	tenantID := middleware.TenantFromContext(r.Context())
// 	u, err := h.svc.Create(r.Context(), CreateRequest{
// 		TenantID:              tenantID,
// 		Email:                 req.Email,
// 		Password:              req.Password,
// 		Role:                  req.Role,
// 		Phone:                 req.Phone,
// 		HasWhatsapp:           req.HasWhatsapp,
// 		EmergencyContactName:  req.EmergencyContactName,
// 		EmergencyContactPhone: req.EmergencyContactPhone,
// 		FullName:              req.FullName,
// 		Document:              req.Document,
// 	})
// 	if errors.Is(err, ErrEmailConflict) {
// 		response.Error(w, http.StatusConflict, err.Error())
// 		return
// 	}
// 	if errors.Is(err, ErrInvalidRole) {
// 		response.Error(w, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao criar usuário",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.Created(w, "usuário criado com sucesso", u)
// }

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
	if errs := validator.Validate(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "dados inválidos")
		return
	}

	loginResp, err := h.svc.AcceptInvite(r.Context(), AcceptInviteRequest{
		Token:    req.Token,
		Password: req.Password,
	})
	if errors.Is(err, invite.ErrInvalidInvite) {
		response.Error(w, http.StatusUnprocessableEntity, "convite inválido ou expirado")
		return
	}
	if err != nil {
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

// @Summary     Verificar email
// @Tags        auth
// @Produce     json
// @Param       token query string true "Token de verificação"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Failure     422 {object} response.Response
// @Router      /v1/api/auth/verify-email [get]
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	token := r.URL.Query().Get("token")
	if token == "" {
		response.Error(w, http.StatusBadRequest, "token obrigatório")
		return
	}

	err := h.svc.VerifyEmail(ctx, token)
	if errors.Is(err, ErrInvalidToken) {
		log.Error("erro ao verificar email", logger.Err(err))
		response.Error(w, http.StatusUnprocessableEntity, "token inválido ou expirado")
		return
	}
	if err != nil {
		log.Error("erro ao verificar email", logger.Err(err))
		response.Error(w, http.StatusInternalServerError, "erro interno")
		return
	}
	response.OK(w, "email verificado com sucesso", nil)
}

// @Summary     Login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body loginRequest true "Credenciais"
// @Success     200 {object} response.Response{data=map[string]interface{}}
// @Failure     401 {object} response.Response
// @Failure     422 {object} response.Response
// @Router      /v1/api/auth/login [post]
// func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	log := logger.FromContext(ctx)

// 	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
// 	var req loginRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		response.Error(w, http.StatusBadRequest, "payload inválido")
// 		return
// 	}
// 	if errs := validator.Validate(req); errs != nil {
// 		response.Error(w, http.StatusBadRequest, "dados inválidos")
// 		return
// 	}

// 	token, refreshToken, expiresAt, err := h.svc.Login(ctx, LoginRequest(req))
// 	if errors.Is(err, ErrInvalidCreds) {
// 		log.Warn("credenciais inválidas", logger.String("email", req.Email))
// 		response.Error(w, http.StatusUnauthorized, "credenciais inválidas")
// 		return
// 	}
// 	if errors.Is(err, ErrTenantRequired) {
// 		response.Error(w, http.StatusUnprocessableEntity, ErrTenantRequired.Error())
// 		return
// 	}
// 	if errors.Is(err, ErrTenantForbidden) {
// 		response.Error(w, http.StatusForbidden, ErrTenantForbidden.Error())
// 		return
// 	}
// 	if err != nil {
// 		log.Error("erro ao autenticar usuário", logger.Err(err))
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	h.setRefreshCookie(w, refreshToken)
// 	response.OK(w, "autenticado com sucesso", map[string]any{
// 		"token":      token,
// 		"expires_at": expiresAt,
// 	})
// }

// // @Summary     Obter usuário staff por ID
// // @Tags        users
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "User ID"
// // @Success     200 {object} response.Response{data=UserWithProfile}
// // @Failure     404 {object} response.Response
// // @Router      /v1/api/users/{id} [get]
// func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
// 	id, err := parseUUID(r.PathValue("id"))
// 	if err != nil {
// 		response.Error(w, http.StatusBadRequest, "id inválido")
// 		return
// 	}
// 	tenantID := middleware.TenantFromContext(r.Context())
// 	u, err := h.svc.GetByID(r.Context(), tenantID, id)
// 	if errors.Is(err, ErrUserNotFound) {
// 		response.Error(w, http.StatusNotFound, "usuário não encontrado")
// 		return
// 	}
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao buscar usuário",
// 			logger.String("user_id", id.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "ok", u)
// }

// // @Summary     Listar pacientes do tenant
// // @Tags        patients
// // @Security    BearerAuth
// // @Produce     json
// // @Success     200 {object} response.Response{data=[]profile.Profile}
// // @Router      /v1/api/patients [get]
// func (h *Handler) ListPatients(w http.ResponseWriter, r *http.Request) {
// 	tenantID := middleware.TenantFromContext(r.Context())
// 	patients, err := h.svc.ListPatients(r.Context(), tenantID)
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao buscar pacientes",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "ok", patients)
// }

// // @Summary     Criar paciente no tenant
// // @Tags        patients
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       body body createPatientRequest true "Dados do paciente"
// // @Success     201 {object} response.Response{data=profile.Profile}
// // @Failure     400 {object} response.Response
// // @Router      /v1/api/patients [post]
// func (h *Handler) CreatePatient(w http.ResponseWriter, r *http.Request) {
// 	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
// 	var req createPatientRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		response.Error(w, http.StatusBadRequest, "payload inválido")
// 		return
// 	}
// 	if errs := validator.Validate(req); errs != nil {
// 		response.Error(w, http.StatusBadRequest, "dados inválidos")
// 		return
// 	}

// 	tenantID := middleware.TenantFromContext(r.Context())
// 	p, err := h.svc.CreatePatient(r.Context(), CreatePatientRequest{
// 		TenantID:              tenantID,
// 		FullName:              req.FullName,
// 		Document:              req.Document,
// 		BirthDate:             req.BirthDate,
// 		Phone:                 req.Phone,
// 		HasWhatsapp:           req.HasWhatsapp,
// 		EmergencyContactName:  req.EmergencyContactName,
// 		EmergencyContactPhone: req.EmergencyContactPhone,
// 		Address:               req.Address,
// 	})
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao criar paciente",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.Created(w, "paciente criado com sucesso", p)
// }

// // @Summary     Obter paciente por ID
// // @Tags        patients
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "Patient ID"
// // @Success     200 {object} response.Response{data=profile.Profile}
// // @Failure     404 {object} response.Response
// // @Router      /v1/api/patients/{id} [get]
// func (h *Handler) GetPatient(w http.ResponseWriter, r *http.Request) {
// 	id, err := parseUUID(r.PathValue("id"))
// 	if err != nil {
// 		response.Error(w, http.StatusBadRequest, "id inválido")
// 		return
// 	}
// 	tenantID := middleware.TenantFromContext(r.Context())
// 	p, err := h.svc.GetPatient(r.Context(), tenantID, id)
// 	if errors.Is(err, ErrPatientNotFound) {
// 		response.Error(w, http.StatusNotFound, "paciente não encontrado")
// 		return
// 	}
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao buscar paciente",
// 			logger.String("patient_id", id.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "ok", p)
// }

// // @Summary     Atualizar paciente
// // @Tags        patients
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       id path string true "Patient ID"
// // @Param       body body updatePatientRequest true "Dados do paciente"
// // @Success     200 {object} response.Response{data=profile.Profile}
// // @Failure     400 {object} response.Response
// // @Failure     404 {object} response.Response
// // @Router      /v1/api/patients/{id} [put]
// func (h *Handler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
// 	id, err := parseUUID(r.PathValue("id"))
// 	if err != nil {
// 		response.Error(w, http.StatusBadRequest, "id inválido")
// 		return
// 	}

// 	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
// 	var req updatePatientRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		response.Error(w, http.StatusBadRequest, "payload inválido")
// 		return
// 	}
// 	if errs := validator.Validate(req); errs != nil {
// 		response.Error(w, http.StatusBadRequest, "dados inválidos")
// 		return
// 	}

// 	tenantID := middleware.TenantFromContext(r.Context())
// 	p, err := h.svc.UpdatePatient(r.Context(), tenantID, id, UpdatePatientRequest(req))
// 	if errors.Is(err, ErrPatientNotFound) {
// 		response.Error(w, http.StatusNotFound, "paciente não encontrado")
// 		return
// 	}
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao atualizar paciente",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.String("patient_id", id.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "paciente atualizado com sucesso", p)
// }

// // @Summary     Remover paciente do tenant
// // @Tags        patients
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "Patient ID"
// // @Success     200 {object} response.Response
// // @Failure     404 {object} response.Response
// // @Router      /v1/api/patients/{id} [delete]
// func (h *Handler) DeletePatient(w http.ResponseWriter, r *http.Request) {
// 	id, err := parseUUID(r.PathValue("id"))
// 	if err != nil {
// 		response.Error(w, http.StatusBadRequest, "id inválido")
// 		return
// 	}
// 	tenantID := middleware.TenantFromContext(r.Context())
// 	if err := h.svc.DeletePatient(r.Context(), tenantID, id); err != nil {
// 		if errors.Is(err, ErrPatientNotFound) {
// 			response.Error(w, http.StatusNotFound, "paciente não encontrado")
// 			return
// 		}
// 		logger.FromContext(r.Context()).Error("erro ao remover paciente",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.String("patient_id", id.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "paciente removido com sucesso", nil)
// }

// // @Summary     Buscar pacientes por nome (parcial, sem distinção de acentos/maiúsculas)
// // @Tags        patients
// // @Security    BearerAuth
// // @Produce     json
// // @Param       q query string true "Termo de busca (ex: \"ale\")"
// // @Success     200 {object} response.Response{data=[]profile.Profile}
// // @Failure     400 {object} response.Response
// // @Router      /v1/api/patients/search [get]
// func (h *Handler) SearchPatients(w http.ResponseWriter, r *http.Request) {
// 	q := r.URL.Query().Get("q")
// 	if q == "" {
// 		response.Error(w, http.StatusBadRequest, "nome do paciente é obrigatório")
// 		return
// 	}

// 	tenantID := middleware.TenantFromContext(r.Context())
// 	patients, err := h.svc.SearchPatients(r.Context(), tenantID, q)
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao buscar pacientes",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "ok", patients)
// }

// // @Summary     Listar usuários staff do tenant
// // @Tags        users
// // @Security    BearerAuth
// // @Produce     json
// // @Success     200 {object} response.Response{data=[]UserWithProfile}
// // @Router      /v1/api/users [get]
// func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
// 	tenantID := middleware.TenantFromContext(r.Context())
// 	users, err := h.svc.List(r.Context(), tenantID)
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao buscar usuários",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "ok", users)
// }

// type updateUserRequest struct {
// 	FullName              string `json:"full_name"               validate:"required"`
// 	Phone                 string `json:"phone"`
// 	Document              string `json:"document"`
// 	HasWhatsapp           bool   `json:"has_whatsapp"`
// 	EmergencyContactName  string `json:"emergency_contact_name"`
// 	EmergencyContactPhone string `json:"emergency_contact_phone"`
// 	Password              string `json:"password"`
// }

// // @Summary     Atualizar usuário staff
// // @Tags        users
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       id path string true "User ID"
// // @Param       body body updateUserRequest true "Dados do usuário"
// // @Success     200 {object} response.Response{data=UserWithProfile}
// // @Failure     400 {object} response.Response
// // @Failure     404 {object} response.Response
// // @Router      /v1/api/users/{id} [put]
// func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
// 	id, err := parseUUID(r.PathValue("id"))
// 	if err != nil {
// 		response.Error(w, http.StatusBadRequest, "id inválido")
// 		return
// 	}

// 	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
// 	var req updateUserRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		response.Error(w, http.StatusBadRequest, "payload inválido")
// 		return
// 	}
// 	if errs := validator.Validate(req); errs != nil {
// 		response.Error(w, http.StatusBadRequest, "dados inválidos")
// 		return
// 	}

// 	tenantID := middleware.TenantFromContext(r.Context())
// 	u, err := h.svc.Update(r.Context(), tenantID, id, UpdateRequest(req))
// 	if errors.Is(err, ErrUserNotFound) {
// 		response.Error(w, http.StatusNotFound, "usuário não encontrado")
// 		return
// 	}
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao atualizar usuário",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.String("user_id", id.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "usuário atualizado com sucesso", u)
// }

// // @Summary     Remover usuário staff do tenant
// // @Tags        users
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "User ID"
// // @Success     200 {object} response.Response
// // @Failure     404 {object} response.Response
// // @Router      /v1/api/users/{id} [delete]
// func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
// 	id, err := parseUUID(r.PathValue("id"))
// 	if err != nil {
// 		response.Error(w, http.StatusBadRequest, "id inválido")
// 		return
// 	}
// 	tenantID := middleware.TenantFromContext(r.Context())
// 	if err := h.svc.Delete(r.Context(), tenantID, id); err != nil {
// 		if errors.Is(err, ErrUserNotFound) {
// 			response.Error(w, http.StatusNotFound, "usuário não encontrado")
// 			return
// 		}
// 		logger.FromContext(r.Context()).Error("erro ao remover usuário",
// 			logger.String("tenant_id", tenantID.String()),
// 			logger.String("user_id", id.String()),
// 			logger.Err(err),
// 		)
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}
// 	response.OK(w, "usuário removido com sucesso", nil)
// }

// // @Summary     Renovar access token via refresh token em cookie
// // @Tags        auth
// // @Produce     json
// // @Success     200 {object} response.Response{data=map[string]interface{}}
// // @Failure     401 {object} response.Response
// // @Router      /v1/api/auth/refresh [post]
// func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
// 	cookie, err := r.Cookie("refresh_token")
// 	if err != nil {
// 		response.Error(w, http.StatusUnauthorized, "refresh token ausente")
// 		return
// 	}

// 	newToken, newRefresh, expiresAt, err := h.svc.Refresh(r.Context(), cookie.Value)
// 	if errors.Is(err, ErrInvalidCreds) || errors.Is(err, ErrTenantForbidden) {
// 		response.Error(w, http.StatusUnauthorized, "sessão inválida")
// 		return
// 	}
// 	if err != nil {
// 		logger.FromContext(r.Context()).Error("erro ao renovar token", logger.Err(err))
// 		response.Error(w, http.StatusInternalServerError, "erro interno")
// 		return
// 	}

// 	h.setRefreshCookie(w, newRefresh, cookie.MaxAge)
// 	response.OK(w, "token renovado com sucesso", map[string]any{
// 		"token":      newToken,
// 		"expires_at": expiresAt,
// 	})
// }
