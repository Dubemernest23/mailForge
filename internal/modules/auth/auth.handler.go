package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"mailForgeApi/internal/apperrors"
	"mailForgeApi/internal/constants"
	"mailForgeApi/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// decodeAndValidate is shared across all four handlers — decode JSON body,
// then run the DTO's own Validate(). One place for this pattern instead of
// repeating decode+validate boilerplate four times.
func decodeAndValidate[T interface{ Validate() error }](r *http.Request, dst *T) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return response.NewAppError(constants.StatusBadRequest, response.CodeBadRequest, "invalid request body")
	}
	if err := (*dst).Validate(); err != nil {
		return response.NewAppError(constants.StatusBadRequest, response.CodeBadRequest, err.Error())
	}
	return nil
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrDuplicate):
		return response.WrapAppError(err, constants.StatusConflict, response.CodeConflict, "email already registered")
	case errors.Is(err, apperrors.ErrUnauthorized):
		return response.WrapAppError(err, constants.StatusUnauthorized, response.CodeUnauthorized, "invalid credentials")
	case errors.Is(err, apperrors.ErrInvalidRefreshToken):
		return response.WrapAppError(err, constants.StatusUnauthorized, response.CodeUnauthorized, "invalid or expired refresh token")
	default:
		return err // unmapped -> HandleError's fallback -> 500
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var req RegisterRequest
	if err := decodeAndValidate(r, &req); err != nil {
		return err
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		return mapServiceError(err)
	}

	response.WriteJSON(w, constants.StatusCreated, resp)
	return nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := decodeAndValidate(r, &req); err != nil {
		return err
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		return mapServiceError(err)
	}

	response.WriteJSON(w, constants.StatusOK, resp)
	return nil
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) error {
	var req RefreshRequest
	if err := decodeAndValidate(r, &req); err != nil {
		return err
	}

	resp, err := h.svc.Refresh(r.Context(), req)
	if err != nil {
		return mapServiceError(err)
	}

	response.WriteJSON(w, constants.StatusOK, resp)
	return nil
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	var req LogoutRequest
	if err := decodeAndValidate(r, &req); err != nil {
		return err
	}

	if err := h.svc.Logout(r.Context(), req); err != nil {
		return mapServiceError(err)
	}

	w.WriteHeader(constants.StatusNoContent)
	return nil
}
