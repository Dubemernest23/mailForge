package response

import (
	"encoding/json"
	"errors"
	"net/http"

	chimiddleware "github.com/go-chi/chi/middleware"

	"mailForgeApi/internal/constants"
)

const ContentTypeJSON = "application/json"

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
	WriteError(w, r, constants.StatusBadRequest, "bad_request", message)
}

func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, constants.StatusUnauthorized, "unauthorized", message)
}

func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, constants.StatusForbidden, "forbidden", message)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, constants.StatusNotFound, "route_not_found", "route not found")
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, constants.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func InternalServerError(w http.ResponseWriter, r *http.Request) {
	WriteError(w, r, constants.StatusInternalServerError, "internal_server_error", "an unexpected error occurred")
}
