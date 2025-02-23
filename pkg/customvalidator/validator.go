package customvalidator

import "github.com/go-playground/validator/v10"

type StructValidator struct {
	Validation *validator.Validate
}

func (v *StructValidator) Validate(out any) error {
	return v.Validation.Struct(out)
}
