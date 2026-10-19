package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// Logger logs each incoming HTTP request with duration, status code, and request ID.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			reqID := GetRequestID(r.Context())

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.statusCode,
				"durationMs", duration.Milliseconds(),
				"requestId", reqID,
				"remoteAddr", r.RemoteAddr,
			)
		})
	}
}
