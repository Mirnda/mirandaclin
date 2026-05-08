package shared

import "errors"

var (
	ErrTenantRequired = errors.New("tenant id deve ser informado")
	ErrNameRequired   = errors.New("nome deve ser informado")

	ErrInvalidRole  = errors.New("função não permitida")
	ErrRoleRequires = errors.New("função deve ser informada")

	ErrEmailRequired = errors.New("email deve ser informado")
	ErrEmailConflict = errors.New("email já cadastrado")

	ErrProfileRequired = errors.New("profile id deve ser informado")
	ErrClinicRequired  = errors.New("clinic id deve ser informado")
)
