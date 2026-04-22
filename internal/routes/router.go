// internal/routes/routes.go
package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"

	"mailForgeApi/internal/middleware"
	"mailForgeApi/pkg/logger"
)

func NewRouter(log *logger.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)       // attaches X-Request-Id to every request
	r.Use(chimiddleware.Recoverer)       // recovers from panics
	r.Use(middleware.RequestLogger(log)) //structured logger

	r.Get("/health", healthCheck)

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
