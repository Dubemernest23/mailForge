// internal/routes/routes.go
package routes

import (
	"crypto/rsa"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/venosm/http-swagger"

	"mailForgeApi/internal/config"
	_ "mailForgeApi/internal/docs"
	"mailForgeApi/internal/middleware"
	"mailForgeApi/internal/modules/auth"
	"mailForgeApi/internal/shared/constants"
	"mailForgeApi/internal/shared/response"
	"mailForgeApi/pkg/logger"
)

func NewRouter(log *logger.Logger, authHandler *auth.Handler, publicKey *rsa.PublicKey) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.Recoverer(log))

	r.Get("/health", healthCheck)
	r.NotFound(response.NotFound)
	r.MethodNotAllowed(response.MethodNotAllowed)
	if config.NewInitConfig().Server.AppEnv != "production" {
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
			httpSwagger.TryItOutEnabled(true),
			httpSwagger.PersistAuthorization(true),
		))
	}

	registerAuthRoutes(r, authHandler)

	// Protected route group is empty for now. Phase C mounts business routes
	// here so router wiring doesn't need to change again.
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.JWTMiddleware(publicKey))
	})

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, constants.StatusOK, map[string]string{"status": "ok"})
}
