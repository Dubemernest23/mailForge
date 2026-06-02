package middleware

import (
	"net/http"

	chimiddleware "github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"mailForgeApi/internal/response"
	"mailForgeApi/pkg/logger"
)

func Recoverer(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered",
						zap.Any("error", err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("request_id", chimiddleware.GetReqID(r.Context())),
					)
					response.InternalServerError(w, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
