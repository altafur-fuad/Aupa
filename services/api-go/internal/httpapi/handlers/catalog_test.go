package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
)

func setupCatalogHandler() (*CatalogHandler, *http.ServeMux) {
	repo := catalog.NewMemoryCatalogRepository()
	svc := catalog.NewService(repo, repo)
	h := NewCatalogHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/restaurants", h.ListRestaurants)
	mux.HandleFunc("GET /v1/restaurants/{id}", h.GetRestaurant)
	mux.HandleFunc("GET /v1/restaurants/{id}/menu", h.GetMenu)

	return h, mux
}

func TestCatalogHandlers(t *testing.T) {
	_, mux := setupCatalogHandler()

	t.Run("list restaurants returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/restaurants?q=Burger", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		data := resp["data"].([]any)
		if len(data) == 0 {
			t.Errorf("expected matching restaurants, got 0")
		}
	})

	t.Run("get restaurant by id returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/restaurants/rest_01", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("get unknown restaurant returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/restaurants/rest_unknown", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("get menu returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/restaurants/rest_01/menu", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
}
