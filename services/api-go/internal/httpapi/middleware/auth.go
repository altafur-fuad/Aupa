package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
)

const AuthUserKey contextKey = "AuthUser"

// Authenticate extracts Bearer JWT token, validates it against auth service,
// and injects authenticated user claims into request context.
func Authenticate(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := GetRequestID(r.Context())
			authHeader := r.Header.Get("Authorization")

			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid bearer authorization token.", nil, reqID)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := authService.VerifyAccessToken(tokenString)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Access token is invalid or expired.", nil, reqID)
				return
			}

			ctx := context.WithValue(r.Context(), AuthUserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole checks whether the authenticated user has at least one of the specified roles.
func RequireRole(allowedRoles ...auth.Role) func(http.Handler) http.Handler {
	allowedMap := make(map[string]bool)
	for _, r := range allowedRoles {
		allowedMap[string(r)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := GetRequestID(r.Context())
			claims := GetClaims(r.Context())

			if claims == nil {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.", nil, reqID)
				return
			}

			if !allowedMap[claims.Role] {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource.", nil, reqID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClaims retrieves CustomClaims from context if authenticated.
func GetClaims(ctx context.Context) *auth.CustomClaims {
	if val, ok := ctx.Value(AuthUserKey).(*auth.CustomClaims); ok {
		return val
	}
	return nil
}
