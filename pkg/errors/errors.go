package errors

import "errors"

type ApiError struct {
	Message string
	Base    error
}

func (e *ApiError) Error() string {
	return e.Message
}

func (e *ApiError) Unwrap() error {
	return e.Base
}

func New(base error, message string) *ApiError {
	return &ApiError{
		Base:    base,
		Message: message,
	}
}

func IsApiError(err error) (*ApiError, bool) {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
