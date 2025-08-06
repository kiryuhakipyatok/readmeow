package validator

import "github.com/go-playground/validator"

func NewValidator() *validator.Validate {
	validator := validator.New()
	return validator
}
