package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
)

// Recovery catches any downstream panics, logs stack traces securely server-side,
// and returns a standardized, opaque 500 error envelope to the caller.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqID := GetRequestID(r.Context())
					stack := string(debug.Stack())

					logger.Error("panic recovered in HTTP handler",
						"error", fmt.Sprintf("%v", err),
						"requestId", reqID,
						"path", r.URL.Path,
						"method", r.Method,
						"stack", stack,
					)

					response.Error(w, http.StatusInternalServerError,
						"INTERNAL_SERVER_ERROR",
						"An unexpected server error occurred. Please contact support.",
						nil,
						reqID,
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
