package errs

import (
	"errors"
	"fmt"
)

var (
	ErrNotFoundBase      = errors.New("not found")
	ErrAlreadyExistsBase = errors.New("already exists")
	ErrInvalidFieldsBase = errors.New("invalid fields")
	ErrInvalidValuesBase = errors.New("invalid values")
	ErrInvalidCodeBase   = errors.New("invalid code")
	ErrZeroAttemptsBase  = errors.New("zero attempts")
	ErrCodeIsExpiredBase = errors.New("code is expired")
)

type AppError struct {
	Operation string
	Err       error
}

func (ae AppError) Error() string {
	return fmt.Sprintf("%s : %v", ae.Operation, ae.Err)
}

func (ae AppError) Unwrap() error {
	return ae.Err
}

func NewAppError(op string, err error) AppError {
	return AppError{
		Operation: op,
		Err:       err,
	}
}

func ErrAlreadyExists(op string, err error) AppError {
	return NewAppError(op, fmt.Errorf("%w : %w", ErrAlreadyExistsBase, err))
}

func ErrNotFound(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrNotFoundBase))
}

func ErrInvalidFields(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrInvalidFieldsBase))
}

func ErrInvalidValues(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrInvalidValuesBase))
}

func ErrInvalidCode(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrInvalidCodeBase))
}

func ErrZeroAttempts(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrZeroAttemptsBase))
}

func ErrCodeIsExpired(op string) AppError {
	return NewAppError(op, fmt.Errorf("%w", ErrCodeIsExpiredBase))
}
