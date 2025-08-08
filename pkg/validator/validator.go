package validator

import "github.com/go-playground/validator/v10"

func NewValidator() *validator.Validate {
	validator := validator.New()
	return validator
}
