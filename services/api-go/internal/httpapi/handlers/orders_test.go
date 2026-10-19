package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/cart"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
	"github.com/golang-jwt/jwt/v5"
)

func setupOrdersHandler() (*OrdersHandler, *orders.Service) {
	catRepo := catalog.NewMemoryCatalogRepository()
	catSvc := catalog.NewService(catRepo, catRepo)
	calc := cart.NewCalculator(catSvc)

	orderRepo := orders.NewMemoryOrderRepository()
	idempRepo := orders.NewMemoryIdempotencyRepository()

	svc := orders.NewService(orderRepo, idempRepo, calc)
	return NewOrdersHandler(svc), svc
}

func contextWithUser(req *http.Request, userID string, role auth.Role) *http.Request {
	claims := &auth.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
		Role:  string(role),
		Email: userID + "@example.com",
	}
	ctx := context.WithValue(req.Context(), middleware.AuthUserKey, claims)
	return req.WithContext(ctx)
}

func TestOrdersHandler(t *testing.T) {
	h, _ := setupOrdersHandler()

	t.Run("create order without idempotency key fails with 400", func(t *testing.T) {
		payload := []byte(`{"restaurantId":"rest_01","items":[{"itemId":"item_01","quantity":1}]}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(payload))
		req = contextWithUser(req, "cust_01", auth.RoleCustomer)
		rec := httptest.NewRecorder()

		h.CreateOrder(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("create order with valid payload returns 201 Created", func(t *testing.T) {
		payload := []byte(`{
			"restaurantId": "rest_01",
			"deliveryAddressId": "addr_01",
			"paymentMethod": "cash_on_delivery",
			"items": [{"itemId": "item_01", "quantity": 1}]
		}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(payload))
		req.Header.Set("Idempotency-Key", "idem_order_h1")
		req = contextWithUser(req, "cust_01", auth.RoleCustomer)
		rec := httptest.NewRecorder()

		h.CreateOrder(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		// Replay with exact same key returns 200 OK
		reqReplay := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(payload))
		reqReplay.Header.Set("Idempotency-Key", "idem_order_h1")
		reqReplay = contextWithUser(reqReplay, "cust_01", auth.RoleCustomer)
		recReplay := httptest.NewRecorder()

		h.CreateOrder(recReplay, reqReplay)

		if recReplay.Code != http.StatusOK {
			t.Fatalf("expected 200 on replay, got %d", recReplay.Code)
		}
	})

	t.Run("list orders returns customer's orders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/orders", nil)
		req = contextWithUser(req, "cust_01", auth.RoleCustomer)
		rec := httptest.NewRecorder()

		h.ListOrders(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		data := resp["data"].([]any)
		if len(data) == 0 {
			t.Errorf("expected past orders, got 0")
		}
	})
}
