package errdefs

import (
	"errors"
	"net"
	"net/http"
	"syscall"
)

type CustomError struct {
	Message string
	Code    int
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewNotFoundError(message string) *CustomError {
	return &CustomError{
		Message: message,
		Code:    http.StatusNotFound,
	}
}

func NewBadRequestError(message string) *CustomError {
	return &CustomError{
		Message: message,
		Code:    http.StatusBadRequest,
	}
}

func IsConnectionRefused(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return errors.Is(opErr.Err, syscall.ECONNREFUSED)
	}
	return false
}
