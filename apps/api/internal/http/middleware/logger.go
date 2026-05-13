package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			next.ServeHTTP(w, r)
			log.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(started).String(),
			)
		})
	}
}
