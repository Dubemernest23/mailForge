// internal/routes/auth.go
package routes

import (
	"github.com/go-chi/chi/v5"

	"mailForgeApi/internal/modules/auth"
	"mailForgeApi/internal/response"
)

func registerAuthRoutes(r chi.Router, h *auth.Handler) {
	r.Post("/auth/register", response.Handler(h.Register))
	r.Post("/auth/login", response.Handler(h.Login))
	r.Post("/auth/refresh", response.Handler(h.Refresh))
	r.Post("/auth/logout", response.Handler(h.Logout))
}
