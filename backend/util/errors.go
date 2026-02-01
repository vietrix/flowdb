package util

import "github.com/google/uuid"

type AppError struct {
	ID      string
	Message string
	Err     error
}

func NewAppError(message string, err error) *AppError {
	return &AppError{
		ID:      uuid.NewString(),
		Message: message,
		Err:     err,
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}
