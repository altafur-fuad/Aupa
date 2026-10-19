package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/payments"
	"github.com/golang-jwt/jwt/v5"
)

type dummyOrderProvider struct {
	order *orders.Order
}

func (d *dummyOrderProvider) GetOrder(ctx context.Context, orderID, customerID string, isPrivileged bool) (*orders.Order, error) {
	if d.order == nil || d.order.ID != orderID {
		return nil, orders.ErrOrderNotFound
	}
	if !isPrivileged && d.order.CustomerID != customerID {
		return nil, orders.ErrUnauthorizedOrderAccess
	}
	return d.order, nil
}

func (d *dummyOrderProvider) RecordPaymentSuccess(ctx context.Context, orderID string) error {
	if d.order != nil && d.order.ID == orderID {
		d.order.PaymentStatus = orders.PaymentStatusPaid
		if d.order.Status == orders.StatusPlaced {
			d.order.Status = orders.StatusAccepted
		}
	}
	return nil
}

func (d *dummyOrderProvider) RecordPaymentFailure(ctx context.Context, orderID string) error {
	if d.order != nil && d.order.ID == orderID {
		d.order.PaymentStatus = orders.PaymentStatusFailed
	}
	return nil
}

func setupPaymentsHandlerTest(order *orders.Order, secret string) (*PaymentsHandler, *payments.Service) {
	dummyProv := &dummyOrderProvider{order: order}
	attemptRepo := payments.NewMemoryPaymentAttemptRepository()
	auditRepo := payments.NewMemoryWebhookAuditRepository()
	svc := payments.NewService(dummyProv, attemptRepo, auditRepo, secret)
	h := NewPaymentsHandler(svc)
	return h, svc
}

func paymentsContextWithUser(req *http.Request, userID string, role auth.Role) *http.Request {
	claims := &auth.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
		Role: string(role),
	}
	ctx := context.WithValue(req.Context(), middleware.AuthUserKey, claims)
	return req.WithContext(ctx)
}

func TestInitiatePayment_Unauthorized(t *testing.T) {
	h, _ := setupPaymentsHandlerTest(nil, "secret")

	req := httptest.NewRequest("POST", "/v1/orders/ord_123/pay", nil)
	w := httptest.NewRecorder()

	h.InitiatePayment(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got: %d", w.Code)
	}
}

func TestInitiatePayment_Success(t *testing.T) {
	order := &orders.Order{
		ID:                "ord_pay_ok",
		CustomerID:        "cust_1",
		Status:            orders.StatusPlaced,
		PaymentStatus:     orders.PaymentStatusPending,
		TotalAmountPoisha: 45000,
		Currency:          "BDT",
	}
	h, _ := setupPaymentsHandlerTest(order, "secret")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/orders/{id}/pay", func(w http.ResponseWriter, r *http.Request) {
		req := paymentsContextWithUser(r, "cust_1", auth.RoleCustomer)
		h.InitiatePayment(w, req)
	})

	body := []byte(`{"provider":"mock"}`)
	req := httptest.NewRequest("POST", "/v1/orders/ord_pay_ok/pay", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d (%s)", w.Code, w.Body.String())
	}

	var res struct {
		Data struct {
			CheckoutURL string `json:"checkoutUrl"`
			AttemptID   string `json:"attemptId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Data.CheckoutURL == "" || res.Data.AttemptID == "" {
		t.Errorf("expected non-empty checkoutUrl and attemptId, got: %+v", res.Data)
	}
}

func TestInitiatePayment_ForbiddenOtherUser(t *testing.T) {
	order := &orders.Order{
		ID:                "ord_pay_other",
		CustomerID:        "cust_different",
		Status:            orders.StatusPlaced,
		PaymentStatus:     orders.PaymentStatusPending,
		TotalAmountPoisha: 45000,
	}
	h, _ := setupPaymentsHandlerTest(order, "secret")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/orders/{id}/pay", func(w http.ResponseWriter, r *http.Request) {
		req := paymentsContextWithUser(r, "cust_attacker", auth.RoleCustomer)
		h.InitiatePayment(w, req)
	})

	req := httptest.NewRequest("POST", "/v1/orders/ord_pay_other/pay", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got: %d", w.Code)
	}
}

func TestHandleWebhook_MissingAndInvalidSignature(t *testing.T) {
	secret := "webhook_secret_key"
	h, _ := setupPaymentsHandlerTest(nil, secret)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/webhooks/payments/{provider}", h.HandleWebhook)

	// 1. Missing signature
	req := httptest.NewRequest("POST", "/v1/webhooks/payments/mock", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing signature, got: %d", w.Code)
	}

	// 2. Invalid signature
	req2 := httptest.NewRequest("POST", "/v1/webhooks/payments/mock", bytes.NewReader([]byte("{}")))
	req2.Header.Set("X-Aupa-Signature", "invalid_forged_sig")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid signature, got: %d", w2.Code)
	}
}

func TestHandleWebhook_SuccessAndReplay(t *testing.T) {
	secret := "aupa_live_webhook_secret"
	order := &orders.Order{
		ID:                "ord_webhook_target",
		CustomerID:        "cust_1",
		Status:            orders.StatusPlaced,
		PaymentStatus:     orders.PaymentStatusPending,
		TotalAmountPoisha: 72000,
	}
	h, _ := setupPaymentsHandlerTest(order, secret)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/webhooks/payments/{provider}", h.HandleWebhook)

	event := payments.WebhookEvent{
		EventID:          "evt_wh_001",
		Provider:         "mock",
		OrderID:          "ord_webhook_target",
		PaymentAttemptID: "pay_att_123",
		TransactionID:    "trx_live_abc",
		Status:           payments.AttemptSucceeded,
		AmountPoisha:     72000,
		Currency:         "BDT",
		Timestamp:        time.Now().UTC(),
	}
	payload, _ := json.Marshal(event)
	sig := payments.GenerateSignature(payload, secret)

	req := httptest.NewRequest("POST", "/v1/webhooks/payments/mock", bytes.NewReader(payload))
	req.Header.Set("X-Aupa-Signature", sig)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d (%s)", w.Code, w.Body.String())
	}

	var res struct {
		Data struct {
			Received    bool   `json:"received"`
			IsDuplicate bool   `json:"isDuplicate"`
			Status      string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !res.Data.Received || res.Data.IsDuplicate {
		t.Errorf("expected received=true, isDuplicate=false, got: %+v", res.Data)
	}

	// Verify order was transitioned
	if order.PaymentStatus != orders.PaymentStatusPaid {
		t.Errorf("expected order paymentStatus=paid, got: %s", order.PaymentStatus)
	}
	if order.Status != orders.StatusAccepted {
		t.Errorf("expected order status=accepted, got: %s", order.Status)
	}

	// Test duplicate webhook replay
	req2 := httptest.NewRequest("POST", "/v1/webhooks/payments/mock", bytes.NewReader(payload))
	req2.Header.Set("X-Aupa-Signature", sig)
	w2 := httptest.NewRecorder()

	mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on replay, got: %d", w2.Code)
	}
	var res2 struct {
		Data struct {
			Received    bool `json:"received"`
			IsDuplicate bool `json:"isDuplicate"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &res2)
	if !res2.Data.IsDuplicate {
		t.Errorf("expected isDuplicate=true on webhook replay, got: %+v", res2.Data)
	}
}
