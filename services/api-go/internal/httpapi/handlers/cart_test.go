package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/cart"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
)

func setupCartHandler() *CartHandler {
	repo := catalog.NewMemoryCatalogRepository()
	catalogSvc := catalog.NewService(repo, repo)
	calc := cart.NewCalculator(catalogSvc)
	return NewCartHandler(calc)
}

func TestCartHandlerCalculateQuote(t *testing.T) {
	h := setupCartHandler()

	t.Run("calculates quote returning 200 OK", func(t *testing.T) {
		payload := []byte(`{
			"restaurantId": "rest_01",
			"deliveryAddressId": "addr_01",
			"promoCode": "AUPA50",
			"items": [
				{"itemId": "item_01", "quantity": 2}
			]
		}`)

		req := httptest.NewRequest(http.MethodPost, "/v1/checkout/quote", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		h.CalculateQuote(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		data := resp["data"].(map[string]any)

		if data["currency"] != "BDT" {
			t.Errorf("expected currency BDT, got %v", data["currency"])
		}
		if int64(data["totalPoisha"].(float64)) != 58500 {
			t.Errorf("expected total 58500, got %v", data["totalPoisha"])
		}
	})

	t.Run("empty items returns 400 Bad Request", func(t *testing.T) {
		payload := []byte(`{
			"restaurantId": "rest_01",
			"deliveryAddressId": "addr_01",
			"items": []
		}`)

		req := httptest.NewRequest(http.MethodPost, "/v1/checkout/quote", bytes.NewReader(payload))
		rec := httptest.NewRecorder()

		h.CalculateQuote(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}
