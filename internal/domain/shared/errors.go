package shared

import "errors"

var (
	ErrTenantRequired = errors.New("tenant id deve ser informado")

	ErrEmailRequired = errors.New("email deve ser informado")
	ErrEmailConflict = errors.New("email já cadastrado")

	ErrProfileRequired = errors.New("profile id deve ser informado")
	ErrClinicRequired  = errors.New("clinic id deve ser informado")
)
