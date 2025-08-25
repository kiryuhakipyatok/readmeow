package helpers

import (
	"readmeow/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RequestType interface {
	isRequestType()
}

type Query struct{}
type Body struct{}

func (Query) isRequestType() {}
func (Body) isRequestType()  {}

func ValidateStruct(s any, v *validator.Validator) map[string]string {
	if errs := v.ValidateStruct(s); errs != nil {
		return errs
	}
	return nil
}

func ParseAndValidateRequest[T any](c *fiber.Ctx, request *T, requestType RequestType, v *validator.Validator) error {
	switch requestType.(type) {
	case Body:
		if err := c.BodyParser(request); err != nil {
			return InvalidRequest()
		}
	case Query:
		if err := c.QueryParser(request); err != nil {
			return InvalidRequest()
		}
	}
	if errs := ValidateStruct(request, v); errs != nil {
		return ValidationError(errs)
	}
	return nil
}

func SuccessResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "success",
	})
}

func ValidateId(c *fiber.Ctx, id string) error {
	if err := uuid.Validate(id); err != nil {
		return InvalidRequest()
	}
	return nil
}

func AlreadyLoggined(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user already loggined",
	})
}
