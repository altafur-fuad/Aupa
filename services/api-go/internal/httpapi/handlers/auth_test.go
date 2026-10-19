package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
)

func setupTestAuthHandler() (*AuthHandler, *auth.Service) {
	userRepo := auth.NewMemoryUserRepository()
	sessionRepo := auth.NewMemorySessionRepository()
	authService := auth.NewService(
		userRepo,
		sessionRepo,
		"test-secret-key-must-be-at-least-32-chars-long",
		"aupa.com.bd",
		"aupa-client",
		15*time.Minute,
		24*time.Hour,
	)
	return NewAuthHandler(authService), authService
}

func TestAuthHandlerRegisterAndLogin(t *testing.T) {
	h, _ := setupTestAuthHandler()

	t.Run("register returns 201 Created", func(t *testing.T) {
		payload := []byte(`{"fullName":"Rahim Ahmed","email":"rahim@example.com","password":"SecurePassword123"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("login returns 200 with tokens", func(t *testing.T) {
		payload := []byte(`{"email":"rahim@example.com","password":"SecurePassword123"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		data := resp["data"].(map[string]any)
		tokens := data["tokens"].(map[string]any)

		if tokens["accessToken"] == "" || tokens["refreshToken"] == "" {
			t.Errorf("missing tokens in response: %v", tokens)
		}
	})

	t.Run("login with wrong password returns 401", func(t *testing.T) {
		payload := []byte(`{"email":"rahim@example.com","password":"WrongPassword!"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})
}

func TestAuthHandlerMe(t *testing.T) {
	h, authService := setupTestAuthHandler()

	_, tokens, err := authService.Register(t.Context(), "Test User", "testme@example.com", "Password123", "", auth.RoleCustomer)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("valid Bearer token returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		rec := httptest.NewRecorder()

		h.GetMe(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing Bearer token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
		rec := httptest.NewRecorder()

		h.GetMe(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})
}
