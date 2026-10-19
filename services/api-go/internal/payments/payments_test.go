package payments

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
)

type mockOrderProvider struct {
	order        *orders.Order
	paidCount    int
	failedCount  int
	orderErr     error
	updatedState orders.PaymentStatus
}

func (m *mockOrderProvider) GetOrder(ctx context.Context, orderID, customerID string, isPrivileged bool) (*orders.Order, error) {
	if m.orderErr != nil {
		return nil, m.orderErr
	}
	return m.order, nil
}

func (m *mockOrderProvider) RecordPaymentSuccess(ctx context.Context, orderID string) error {
	m.paidCount++
	m.updatedState = orders.PaymentStatusPaid
	if m.order != nil {
		m.order.PaymentStatus = orders.PaymentStatusPaid
		if m.order.Status == orders.StatusPlaced {
			m.order.Status = orders.StatusAccepted
		}
	}
	return nil
}

func (m *mockOrderProvider) RecordPaymentFailure(ctx context.Context, orderID string) error {
	m.failedCount++
	m.updatedState = orders.PaymentStatusFailed
	if m.order != nil {
		m.order.PaymentStatus = orders.PaymentStatusFailed
	}
	return nil
}

func setupTestService(order *orders.Order) (*Service, *mockOrderProvider) {
	orderProv := &mockOrderProvider{order: order}
	attemptRepo := NewMemoryPaymentAttemptRepository()
	auditRepo := NewMemoryWebhookAuditRepository()
	secret := "test_webhook_secret_key_123"

	svc := NewService(orderProv, attemptRepo, auditRepo, secret)
	return svc, orderProv
}

func TestInitiatePayment_Success(t *testing.T) {
	order := &orders.Order{
		ID:                "ord_test1",
		CustomerID:        "cust_1",
		Status:            orders.StatusPlaced,
		PaymentStatus:     orders.PaymentStatusPending,
		TotalAmountPoisha: 54000,
		Currency:          "BDT",
	}

	svc, _ := setupTestService(order)

	session, err := svc.InitiatePayment(context.Background(), "cust_1", "ord_test1", "mock")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if session.CheckoutURL == "" {
		t.Error("expected non-empty checkout URL")
	}
	if session.PaymentAttemptID == "" {
		t.Error("expected non-empty attempt ID")
	}

	attempts, err := svc.GetPaymentAttempts(context.Background(), "ord_test1")
	if err != nil || len(attempts) != 1 {
		t.Fatalf("expected 1 attempt saved, got %d, err: %v", len(attempts), err)
	}
	if attempts[0].Status != AttemptPending {
		t.Errorf("expected status pending, got %s", attempts[0].Status)
	}
}

func TestInitiatePayment_AlreadyPaid(t *testing.T) {
	order := &orders.Order{
		ID:            "ord_paid",
		CustomerID:    "cust_1",
		Status:        orders.StatusAccepted,
		PaymentStatus: orders.PaymentStatusPaid,
	}

	svc, _ := setupTestService(order)

	_, err := svc.InitiatePayment(context.Background(), "cust_1", "ord_paid", "mock")
	if err != ErrPaymentAlreadyCompleted {
		t.Errorf("expected ErrPaymentAlreadyCompleted, got: %v", err)
	}
}

func TestInitiatePayment_CancelledOrder(t *testing.T) {
	order := &orders.Order{
		ID:            "ord_cancelled",
		CustomerID:    "cust_1",
		Status:        orders.StatusCancelled,
		PaymentStatus: orders.PaymentStatusPending,
	}

	svc, _ := setupTestService(order)

	_, err := svc.InitiatePayment(context.Background(), "cust_1", "ord_cancelled", "mock")
	if err != ErrInvalidOrderState {
		t.Errorf("expected ErrInvalidOrderState, got: %v", err)
	}
}

func TestInitiatePayment_InvalidProvider(t *testing.T) {
	order := &orders.Order{
		ID:            "ord_1",
		CustomerID:    "cust_1",
		Status:        orders.StatusPlaced,
		PaymentStatus: orders.PaymentStatusPending,
	}

	svc, _ := setupTestService(order)

	_, err := svc.InitiatePayment(context.Background(), "cust_1", "ord_1", "nonexistent")
	if err != ErrProviderNotFound {
		t.Errorf("expected ErrProviderNotFound, got: %v", err)
	}
}

func TestProcessWebhook_ValidSuccess(t *testing.T) {
	order := &orders.Order{
		ID:                "ord_webhook1",
		CustomerID:        "cust_1",
		Status:            orders.StatusPlaced,
		PaymentStatus:     orders.PaymentStatusPending,
		TotalAmountPoisha: 65000,
	}

	svc, orderProv := setupTestService(order)

	event := WebhookEvent{
		EventID:          "evt_123",
		Provider:         "mock",
		OrderID:          "ord_webhook1",
		PaymentAttemptID: "pay_att_1",
		TransactionID:    "trx_bank_999",
		Status:           AttemptSucceeded,
		AmountPoisha:     65000,
		Currency:         "BDT",
		Timestamp:        time.Now().UTC(),
	}

	payload, _ := json.Marshal(event)
	sig := GenerateSignature(payload, "test_webhook_secret_key_123")

	req := httptest.NewRequest("POST", "/v1/webhooks/payments/mock", nil)
	req.Header.Set("X-Aupa-Signature", sig)

	processedEvent, isDup, err := svc.ProcessWebhook(context.Background(), "mock", req, payload)
	if err != nil {
		t.Fatalf("expected webhook to succeed, got: %v", err)
	}
	if isDup {
		t.Error("expected first delivery not to be duplicate")
	}
	if processedEvent.TransactionID != "trx_bank_999" {
		t.Errorf("expected trx_bank_999, got %s", processedEvent.TransactionID)
	}

	if orderProv.paidCount != 1 {
		t.Errorf("expected 1 paid record, got %d", orderProv.paidCount)
	}
	if order.Status != orders.StatusAccepted {
		t.Errorf("expected order status to advance to accepted, got %s", order.Status)
	}

	// Test duplicate event delivery (idempotency)
	_, isDup2, err2 := svc.ProcessWebhook(context.Background(), "mock", req, payload)
	if err2 != nil {
		t.Fatalf("expected duplicate webhook to succeed cleanly, got: %v", err2)
	}
	if !isDup2 {
		t.Error("expected duplicate event to be identified as duplicate")
	}
	if orderProv.paidCount != 1 {
		t.Errorf("expected paidCount to stay 1 on duplicate, got %d", orderProv.paidCount)
	}
}

func TestProcessWebhook_InvalidSignature(t *testing.T) {
	order := &orders.Order{
		ID:                "ord_sig_fail",
		TotalAmountPoisha: 30000,
	}

	svc, _ := setupTestService(order)

	payload := []byte(`{"eventId":"evt_forged","orderId":"ord_sig_fail","status":"succeeded"}`)
	req := httptest.NewRequest("POST", "/v1/webhooks/payments/mock", nil)
	req.Header.Set("X-Aupa-Signature", "forged_signature_hex_value")

	_, _, err := svc.ProcessWebhook(context.Background(), "mock", req, payload)
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestProcessWebhook_AmountMismatch(t *testing.T) {
	order := &orders.Order{
		ID:                "ord_mismatch",
		TotalAmountPoisha: 50000,
	}

	svc, _ := setupTestService(order)

	event := WebhookEvent{
		EventID:      "evt_mismatch",
		OrderID:      "ord_mismatch",
		AmountPoisha: 1000, // Less than order total 50000
		Status:       AttemptSucceeded,
	}
	payload, _ := json.Marshal(event)
	sig := GenerateSignature(payload, "test_webhook_secret_key_123")

	req := httptest.NewRequest("POST", "/v1/webhooks/payments/mock", nil)
	req.Header.Set("X-Aupa-Signature", sig)

	_, _, err := svc.ProcessWebhook(context.Background(), "mock", req, payload)
	if err != ErrAmountMismatch {
		t.Errorf("expected ErrAmountMismatch, got: %v", err)
	}
}
