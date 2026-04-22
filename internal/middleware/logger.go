package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"mailForgeApi/pkg/logger"
)

func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// chi's response writer wrapper gives us the status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// process the request
			next.ServeHTTP(ww, r)

			latency := time.Since(start)
			status := ww.Status()

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", status),
				zap.String("latency", fmt.Sprintf("%dms", latency.Milliseconds())),
				zap.Int("bytes", ww.BytesWritten()),
				zap.String("ip", r.RemoteAddr),
				zap.String("request_id", middleware.GetReqID(r.Context())),
			}

			// log at appropriate level based on status code
			switch {
			case status >= 500:
				log.Error("server error", fields...)
			case status >= 400:
				log.Warn("client error", fields...)
			default:
				log.Info("request", fields...)
			}
		})
	}
}
