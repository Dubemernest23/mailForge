// internal/routes/routes.go
package routes

import (
	"net/http"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"

	"mailForgeApi/internal/constants"
	"mailForgeApi/internal/middleware"
	"mailForgeApi/internal/response"
	"mailForgeApi/pkg/logger"
)

func NewRouter(log *logger.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)       // attaches X-Request-Id to every request
	r.Use(middleware.RequestLogger(log)) //structured logger
	r.Use(middleware.Recoverer(log))     // recovers from panics with a JSON response

	r.Get("/health", healthCheck)
	r.NotFound(response.NotFound)
	r.MethodNotAllowed(response.MethodNotAllowed)

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, constants.StatusOK, map[string]string{"status": "ok"})
}
