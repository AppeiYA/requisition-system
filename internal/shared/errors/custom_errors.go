package customerrors

import "net/http"

type CustomError struct {
	Message string
	Code    int
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewCustomError(message string, code int) *CustomError {
	return &CustomError{
		Message: message,
		Code:    code,
	}
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

func NewInternalServerError(message string) *CustomError {
	return &CustomError{
		Message: message,
		Code:    http.StatusInternalServerError,
	}
}

func NewUnauthorizedError(message string) *CustomError {
	return &CustomError{
		Message: message,
		Code:    http.StatusUnauthorized,
	}
}

func NewForbiddenError(message string) *CustomError {
	return &CustomError{
		Message: message,
		Code:    http.StatusForbidden,
	}
}
