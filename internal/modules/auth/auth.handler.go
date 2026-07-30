package auth

import (
	"net/http"

	"mailForgeApi/internal/shared/apperrors"
	"mailForgeApi/internal/shared/constants"
	httpx "mailForgeApi/internal/shared/http"
	"mailForgeApi/internal/shared/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var req RegisterRequest
	if err := httpx.DecodeAndValidate(r, &req); err != nil {
		return err
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		return apperrors.MapServiceError(err)
	}

	response.WriteJSON(w, constants.StatusCreated, resp)
	return nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := httpx.DecodeAndValidate(r, &req); err != nil {
		return err
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		return apperrors.MapServiceError(err)
	}

	response.WriteJSON(w, constants.StatusOK, resp)
	return nil
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) error {
	var req RefreshRequest
	if err := httpx.DecodeAndValidate(r, &req); err != nil {
		return err
	}

	resp, err := h.svc.Refresh(r.Context(), req)
	if err != nil {
		return apperrors.MapServiceError(err)
	}

	response.WriteJSON(w, constants.StatusOK, resp)
	return nil
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	var req LogoutRequest
	if err := httpx.DecodeAndValidate(r, &req); err != nil {
		return err
	}

	if err := h.svc.Logout(r.Context(), req); err != nil {
		return apperrors.MapServiceError(err)
	}

	w.WriteHeader(constants.StatusNoContent)
	return nil
}
