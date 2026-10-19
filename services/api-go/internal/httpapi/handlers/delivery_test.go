package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/delivery"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
	"github.com/golang-jwt/jwt/v5"
)

type dummyTrackingOrderProvider struct {
	order *orders.Order
}

func (d *dummyTrackingOrderProvider) GetOrder(ctx context.Context, orderID, customerID string, isPrivileged bool) (*orders.Order, error) {
	if d.order == nil || d.order.ID != orderID {
		return nil, orders.ErrOrderNotFound
	}
	if !isPrivileged && d.order.CustomerID != customerID {
		return nil, orders.ErrUnauthorizedOrderAccess
	}
	return d.order, nil
}

func deliveryContextWithUser(req *http.Request, userID string, role auth.Role) *http.Request {
	claims := &auth.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
		Role: string(role),
	}
	ctx := context.WithValue(req.Context(), middleware.AuthUserKey, claims)
	return req.WithContext(ctx)
}

func setupDeliveryHandler(order *orders.Order) *DeliveryHandler {
	prov := &dummyTrackingOrderProvider{order: order}
	outbox := delivery.NewMemoryOutboxRepository()
	svc := delivery.NewService(prov, outbox)
	return NewDeliveryHandler(svc)
}

func TestGetTracking_Unauthorized(t *testing.T) {
	h := setupDeliveryHandler(nil)

	req := httptest.NewRequest("GET", "/v1/orders/ord_123/tracking", nil)
	w := httptest.NewRecorder()

	h.GetTracking(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got: %d", w.Code)
	}
}

func TestGetTracking_Success(t *testing.T) {
	order := &orders.Order{
		ID:         "ord_track_ok",
		CustomerID: "cust_track_1",
		Status:     orders.StatusOutForDelivery,
		UpdatedAt:  time.Now().UTC(),
	}
	h := setupDeliveryHandler(order)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/orders/{id}/tracking", func(w http.ResponseWriter, r *http.Request) {
		req := deliveryContextWithUser(r, "cust_track_1", auth.RoleCustomer)
		h.GetTracking(w, req)
	})

	req := httptest.NewRequest("GET", "/v1/orders/ord_track_ok/tracking", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d (%s)", w.Code, w.Body.String())
	}

	var res struct {
		Data delivery.TrackingInfo `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if res.Data.OrderID != "ord_track_ok" {
		t.Errorf("expected orderId ord_track_ok, got %s", res.Data.OrderID)
	}
	if res.Data.StageIndex != 3 {
		t.Errorf("expected stage 3 for out_for_delivery, got %d", res.Data.StageIndex)
	}
	if res.Data.Rider == nil || res.Data.Rider.Name != "Rafiqul Islam" {
		t.Errorf("expected Rafiqul Islam, got %+v", res.Data.Rider)
	}
}

func TestGetTracking_ForbiddenOtherUser(t *testing.T) {
	order := &orders.Order{
		ID:         "ord_track_other",
		CustomerID: "cust_rightful_owner",
		Status:     orders.StatusPreparing,
	}
	h := setupDeliveryHandler(order)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/orders/{id}/tracking", func(w http.ResponseWriter, r *http.Request) {
		req := deliveryContextWithUser(r, "cust_intruder", auth.RoleCustomer)
		h.GetTracking(w, req)
	})

	req := httptest.NewRequest("GET", "/v1/orders/ord_track_other/tracking", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got: %d", w.Code)
	}
}

func TestGetTracking_NotFound(t *testing.T) {
	h := setupDeliveryHandler(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/orders/{id}/tracking", func(w http.ResponseWriter, r *http.Request) {
		req := deliveryContextWithUser(r, "cust_1", auth.RoleCustomer)
		h.GetTracking(w, req)
	})

	req := httptest.NewRequest("GET", "/v1/orders/ord_nonexistent/tracking", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got: %d", w.Code)
	}
}
