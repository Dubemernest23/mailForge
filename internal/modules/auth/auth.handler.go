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

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and returns an access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth.RegisterRequest  true  "Registration details"
// @Success      201      {object}  auth.AuthResponse
// @Failure      400      {object}  response.ErrorBody  "validation error (missing/invalid field)"
// @Failure      409      {object}  response.ErrorBody  "email already registered"
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Login a user
// @Description  Authenticates a user and returns an access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth.LoginRequest  true  "Login details"
// @Success      200      {object}  auth.AuthResponse
// @Failure      400      {object}  response.ErrorBody  "validation error (missing/invalid field)"
// @Failure      401      {object}  response.ErrorBody  "invalid credentials"
// @Router       /auth/login [post]
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

// Refresh godoc
// @Summary      Refresh an access token
// @Description  Generates a new access token using a valid refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth.RefreshRequest  true  "Refresh details"
// @Success      200      {object}  auth.AuthResponse
// @Failure      400      {object}  response.ErrorBody  "validation error (missing/invalid field)"
// @Failure      401      {object}  response.ErrorBody  "invalid refresh token"
// @Router       /auth/refresh [post]
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

// Logout godoc
// @Summary      Logout a user
// @Description  Invalidates the given refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      auth.LogoutRequest  true  "Logout details"
// @Success      204      "no content"
// @Failure      400      {object}  response.ErrorBody  "validation error (missing/invalid field)"
// @Router       /auth/logout [post]
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
