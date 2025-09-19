package helpers

import (
	"readmeow/internal/delivery/apierr"
	"readmeow/internal/dto"
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

func ParseAndValidateRequest[T any](c *fiber.Ctx, request *T, requestType RequestType, v *validator.Validator) error {
	switch requestType.(type) {
	case Body:
		if err := c.BodyParser(request); err != nil {
			return apierr.InvalidRequest()
		}
	case Query:
		if err := c.QueryParser(request); err != nil {
			return apierr.InvalidRequest()
		}
	}
	if errs := v.ValidateStruct(request); len(errs) > 0 {
		return apierr.ValidationError(errs)
	}
	return nil
}

func ValidateId(c *fiber.Ctx, id string) error {
	if err := uuid.Validate(id); err != nil {
		return apierr.InvalidRequest()
	}
	return nil
}

func SuccessResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Code:    fiber.StatusOK,
		Message: "success",
	})
}
