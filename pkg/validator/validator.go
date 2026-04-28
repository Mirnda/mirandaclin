package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// ValidationError representa um erro de validação de campo individual.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate executa a validação da struct e retorna os erros por campo, ou nil se válida.
func Validate(s any) []ValidationError {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	var errs []ValidationError
	for _, e := range err.(validator.ValidationErrors) {
		errs = append(errs, ValidationError{
			Field:   strings.ToLower(e.Field()),
			Message: fmt.Sprintf("falhou na validação '%s'", e.Tag()),
		})
	}
	return errs
}
