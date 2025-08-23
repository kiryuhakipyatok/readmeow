package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Validate *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &Validator{
		Validate: v,
	}
}

func (v *Validator) ValidateStruct(s any) map[string]string {
	err := v.Validate.Struct(s)
	if err == nil {
		return nil
	}
	errors := make(map[string]string)
	validationError := err.(validator.ValidationErrors)
	for _, e := range validationError {
		errors[e.Field()] = v.translate(e)
	}
	fmt.Println(errors)
	return errors
}

func (v *Validator) translate(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "email":
		return "must be a valid email"
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", e.Param())
	default:
		return "is invalid"
	}
}
