package response

import (
	"encoding/json"
	"errors"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"mailForgeApi/internal/constants"
)

const ContentTypeJSON = "application/json"

const (
	CodeBadRequest          = "bad_request"
	CodeUnauthorized        = "unauthorized"
	CodeForbidden           = "forbidden"
	CodeRouteNotFound       = "route_not_found"
	CodeMethodNotAllowed    = "method_not_allowed"
	CodeConflict            = "conflict"
	CodeUnprocessableEntity = "unprocessable_entity"
	CodeTooManyRequests     = "too_many_requests"
	CodeInternalServerError = "internal_server_error"
	CodeNotImplemented      = "not_implemented"
	CodeServiceUnavailable  = "service_unavailable"
)

type ErrorBody struct {
	Success   bool        `json:"success"`
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		HandleError(w, r, err)
	}
}

func Handler(fn HandlerFunc) http.HandlerFunc {
	return fn.ServeHTTP
}

func NewAppError(statusCode int, code string, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

func WrapAppError(err error, statusCode int, code string, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	WriteJSON(w, statusCode, ErrorBody{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
		RequestID: chimiddleware.GetReqID(r.Context()),
	})
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		WriteError(w, r, appErr.StatusCode, appErr.Code, appErr.Message)
		return
	}

	InternalServerError(w, r)
}

func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, constants.StatusBadRequest, CodeBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, constants.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, constants.StatusForbidden, CodeForbidden, message)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, constants.StatusNotFound, CodeRouteNotFound, "route not found")
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, constants.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
}

func Conflict(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, constants.StatusConflict, CodeConflict, message)
}

func UnprocessableEntity(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, constants.StatusUnprocessableEntity, CodeUnprocessableEntity, message)
}

func TooManyRequests(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, constants.StatusTooManyRequests, CodeTooManyRequests, message)
}

func InternalServerError(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, constants.StatusInternalServerError, CodeInternalServerError, "an unexpected error occurred")
}

func NotImplemented(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, constants.StatusNotImplemented, CodeNotImplemented, "not implemented")
}

func ServiceUnavailable(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, constants.StatusServiceUnavailable, CodeServiceUnavailable, "service unavailable")
}
