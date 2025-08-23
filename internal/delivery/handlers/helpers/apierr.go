package helpers

import (
	"errors"
	"fmt"
	"readmeow/pkg/errs"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("alredy exists")
	ErrBadRequest     = errors.New("bad request")
	ErrInvalidJSON    = errors.New("invalid request json data")
	ErrInternalServer = errors.New("internal server error")
)

type ApiErr struct {
	Code    int
	Message any
}

func (ae ApiErr) Error() string {
	return fmt.Sprintf("error: %s, code: %d", ae.Message, ae.Code)
}

func NewApiError(code int, err error) ApiErr {
	return ApiErr{
		Code:    code,
		Message: err.Error(),
	}
}

func ToApiError(err error) ApiErr {
	switch {
	case errors.Is(err, errs.ErrNotFoundBase):
		return NotFound()
	case errors.Is(err, errs.ErrAlreadyExistsBase):
		return AlreadyExists()
	case errors.Is(err, errs.ErrInvalidFieldsBase):
		return InvalidRequest()
	case errors.Is(err, errs.ErrInvalidValuesBase):
		return InvalidRequest()
	default:
		return InternalServerError()
	}
}

func InternalServerError() ApiErr {
	return NewApiError(fiber.StatusInternalServerError, ErrInternalServer)
}

func InvalidRequest() ApiErr {
	return NewApiError(fiber.StatusUnprocessableEntity, ErrBadRequest)
}

func NotFound() ApiErr {
	return NewApiError(fiber.StatusNotFound, ErrNotFound)
}

func AlreadyExists() ApiErr {
	return NewApiError(fiber.StatusConflict, ErrAlreadyExists)
}
