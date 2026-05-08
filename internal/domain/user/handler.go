package user

type Handler struct {
	svc    *Service
	appEnv string
}

func NewHandler(svc *Service, appEnv string) *Handler {
	return &Handler{svc: svc, appEnv: appEnv}
}
