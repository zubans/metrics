package errdefs

import "net/http"

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
