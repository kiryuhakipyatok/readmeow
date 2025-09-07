package apierr

import (
	"errors"
	"fmt"
	"readmeow/pkg/errs"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrBadRequest        = errors.New("bad request")
	ErrInternalServer    = errors.New("internal server error")
	ErrInvalidVerifyCode = errors.New("invalid verification code")
	ErrZeroAttempts      = errors.New("verification code has zero attempts")
	ErrCodeIsExpired     = errors.New("verification code is expired")
	ErrAlreadyLoggined   = errors.New("already loggined")
	ErrRequestTimeout    = errors.New("request timeout")
	ErrToManeRequests    = errors.New("to many requests")
	ErrForbidden         = errors.New("forbidden")
	ErrUnauthorized      = errors.New("unauthorized")
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

func ValidationError(errors map[string]string) ApiErr {
	return ApiErr{
		Code:    fiber.StatusUnprocessableEntity,
		Message: errors,
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
	case errors.Is(err, errs.ErrInvalidCodeBase):
		return InvlidVerifyCode()
	case errors.Is(err, errs.ErrZeroAttemptsBase):
		return ZeroAttempts()
	case errors.Is(err, errs.ErrCodeIsExpiredBase):
		return CodeIsExpired()
	default:
		return InternalServerError()
	}
}

func InternalServerError() ApiErr {
	return NewApiError(fiber.StatusInternalServerError, ErrInternalServer)
}

func InvalidRequest() ApiErr {
	return NewApiError(fiber.StatusBadRequest, ErrBadRequest)
}

func NotFound() ApiErr {
	return NewApiError(fiber.StatusNotFound, ErrNotFound)
}

func AlreadyExists() ApiErr {
	return NewApiError(fiber.StatusConflict, ErrAlreadyExists)
}

func InvlidVerifyCode() ApiErr {
	return NewApiError(fiber.StatusOK, ErrInvalidVerifyCode)
}

func ZeroAttempts() ApiErr {
	return NewApiError(fiber.StatusOK, ErrZeroAttempts)
}

func CodeIsExpired() ApiErr {
	return NewApiError(fiber.StatusOK, ErrCodeIsExpired)
}

func AlreadyLoggined() ApiErr {
	return NewApiError(fiber.StatusConflict, ErrAlreadyLoggined)
}

func RequestTimeout() ApiErr {
	return NewApiError(fiber.StatusRequestTimeout, ErrRequestTimeout)
}

func TooManyRequests() ApiErr {
	return NewApiError(fiber.StatusTooManyRequests, ErrToManeRequests)
}

func Forbidden() ApiErr {
	return NewApiError(fiber.StatusForbidden, ErrForbidden)
}

func Unauthorized() ApiErr {
	return NewApiError(fiber.StatusUnauthorized, ErrUnauthorized)
}
