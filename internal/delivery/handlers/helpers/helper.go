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

func SuccessResponse(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func ParseAndValidateRequest[T any](c *fiber.Ctx, request *T, requestType RequestType, v *validator.Validator) error {
	errParseFunc := func() error {
		return InvalidRequest()
	}
	switch requestType.(type) {
	case Body:
		if err := c.BodyParser(request); err != nil {
			return errParseFunc()
		}
	case Query:
		if err := c.QueryParser(request); err != nil {
			return errParseFunc()
		}
	}
	if err := v.Validate.Struct(request); err != nil {
		return InvalidJSON()
	}
	return nil
}

func ValidateId(c *fiber.Ctx, id string) error {
	if err := uuid.Validate(id); err != nil {
		return InvalidRequest()
	}
	return nil
}
