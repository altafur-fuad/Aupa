package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
)

func setupTestAuthService() *auth.Service {
	return auth.NewService(
		auth.NewMemoryUserRepository(),
		auth.NewMemorySessionRepository(),
		"a-very-strong-secret-key-32-bytes-long!",
		"aupa.com.bd",
		"aupa-client",
		15*time.Minute,
		24*time.Hour,
	)
}

func TestAuthenticateAndRequireRole(t *testing.T) {
	authService := setupTestAuthService()
	ctx := context.Background()

	// Register test users with different roles
	_, customerTokens, _ := authService.Register(ctx, "Customer User", "cust@example.com", "Password123", "", auth.RoleCustomer)
	_, merchantTokens, _ := authService.Register(ctx, "Merchant User", "merch@example.com", "Password123", "", auth.RoleMerchant)
	_, adminTokens, _ := authService.Register(ctx, "Admin User", "admin@example.com", "Password123", "", auth.RoleAdmin)

	merchantEndpoint := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			w.Header().Set("X-User-Id", claims.Subject)
			w.WriteHeader(http.StatusOK)
		}),
		Authenticate(authService),
		RequireRole(auth.RoleMerchant, auth.RoleAdmin),
	)

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/merchant/dashboard", nil)
		rec := httptest.NewRecorder()

		merchantEndpoint.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/merchant/dashboard", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-string")
		rec := httptest.NewRecorder()

		merchantEndpoint.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("customer accessing merchant endpoint returns 403 Forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/merchant/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+customerTokens.AccessToken)
		rec := httptest.NewRecorder()

		merchantEndpoint.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %d", rec.Code)
		}
	})

	t.Run("merchant accessing merchant endpoint returns 200 OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/merchant/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+merchantTokens.AccessToken)
		rec := httptest.NewRecorder()

		merchantEndpoint.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
		if rec.Header().Get("X-User-Id") == "" {
			t.Errorf("expected authenticated user ID in response header")
		}
	})

	t.Run("admin accessing merchant endpoint returns 200 OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/merchant/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+adminTokens.AccessToken)
		rec := httptest.NewRecorder()

		merchantEndpoint.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
	})
}
