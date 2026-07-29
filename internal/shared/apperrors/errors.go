package apperrors

import (
	"errors"
	"mailForgeApi/internal/shared/constants"
	"mailForgeApi/internal/shared/response"
)

var (
	ErrNotFound            = errors.New("resource not found")
	ErrDuplicate           = errors.New("resource already exists")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrMethodNotAllowed    = errors.New("method not allowed")
	ErrConflict            = errors.New("conflict")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrInternalServerError = errors.New("internal server error")
	ErrNotImplemented      = errors.New("not implemented")
	ErrServiceUnavailable  = errors.New("service unavailable")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

func MapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrDuplicate):
		return response.WrapAppError(err, constants.StatusConflict, response.CodeConflict, "email already registered")
	case errors.Is(err, ErrUnauthorized):
		return response.WrapAppError(err, constants.StatusUnauthorized, response.CodeUnauthorized, "invalid credentials")
	case errors.Is(err, ErrInvalidRefreshToken):
		return response.WrapAppError(err, constants.StatusUnauthorized, response.CodeUnauthorized, "invalid or expired refresh token")
	default:
		return err // unmapped -> HandleError's fallback -> 500
	}
}
