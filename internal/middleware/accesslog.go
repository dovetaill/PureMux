package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// AccessLog 记录请求访问日志。
func AccessLog(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			if logger != nil {
				logger.Info("http_access",
					"method", r.Method,
					"path", r.URL.Path,
					"duration_ms", time.Since(start).Milliseconds(),
				)
			}
		})
	}
}
