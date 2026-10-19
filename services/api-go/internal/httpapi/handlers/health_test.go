package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	h := NewHealthHandler(func() bool { return true })

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Healthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Errorf("expected status ok, got %v", payload["status"])
	}
	if payload["time"] == nil {
		t.Errorf("expected timestamp in healthz payload")
	}
}

func TestReadyz(t *testing.T) {
	t.Run("ready returns 200", func(t *testing.T) {
		h := NewHealthHandler(func() bool { return true })

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		h.Readyz(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("unready returns 503", func(t *testing.T) {
		h := NewHealthHandler(func() bool { return false })

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		h.Readyz(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503, got %d", rec.Code)
		}
	})
}
