package handlers

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
	errParseFunc := func(err error) error {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	switch requestType.(type) {
	case Body:
		if err := c.BodyParser(request); err != nil {
			return errParseFunc(err)
		}
	case Query:
		if err := c.QueryParser(request); err != nil {
			return errParseFunc(err)
		}
	}
	if err := v.Validate.Struct(request); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	return nil
}

func ValidateId(c *fiber.Ctx, id string) error {
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	return nil
}
