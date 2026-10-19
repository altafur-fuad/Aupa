package delivery

import (
	"context"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
)

type mockOrderProvider struct {
	order *orders.Order
}

func (m *mockOrderProvider) GetOrder(ctx context.Context, orderID, customerID string, isPrivileged bool) (*orders.Order, error) {
	if m.order == nil || m.order.ID != orderID {
		return nil, orders.ErrOrderNotFound
	}
	if !isPrivileged && m.order.CustomerID != customerID {
		return nil, orders.ErrUnauthorizedOrderAccess
	}
	return m.order, nil
}

func TestGetTrackingInfo_Placed(t *testing.T) {
	order := &orders.Order{
		ID:         "ord_placed_1",
		CustomerID: "cust_1",
		Status:     orders.StatusPlaced,
		UpdatedAt:  time.Now().UTC(),
	}

	orderProv := &mockOrderProvider{order: order}
	outboxRepo := NewMemoryOutboxRepository()
	svc := NewService(orderProv, outboxRepo)

	info, err := svc.GetTrackingInfo(context.Background(), "ord_placed_1", "cust_1", false)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if info.StageIndex != 0 {
		t.Errorf("expected stage 0, got %d", info.StageIndex)
	}
	if info.EstimatedMinutesRemaining != 30 {
		t.Errorf("expected 30 mins, got %d", info.EstimatedMinutesRemaining)
	}
	if info.Rider != nil {
		t.Error("expected no rider assigned yet for placed order")
	}
}

func TestGetTrackingInfo_OutForDelivery_HasRider(t *testing.T) {
	order := &orders.Order{
		ID:         "ord_en_route",
		CustomerID: "cust_1",
		Status:     orders.StatusOutForDelivery,
		UpdatedAt:  time.Now().UTC(),
	}

	orderProv := &mockOrderProvider{order: order}
	outboxRepo := NewMemoryOutboxRepository()
	svc := NewService(orderProv, outboxRepo)

	info, err := svc.GetTrackingInfo(context.Background(), "ord_en_route", "cust_1", false)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if info.StageIndex != 3 {
		t.Errorf("expected stage 3, got %d", info.StageIndex)
	}
	if info.EstimatedMinutesRemaining != 8 {
		t.Errorf("expected 8 mins, got %d", info.EstimatedMinutesRemaining)
	}
	if info.Rider == nil {
		t.Fatal("expected rider assigned for out_for_delivery order")
	}
	if info.Rider.Name != "Rafiqul Islam" {
		t.Errorf("expected Rafiqul Islam, got %s", info.Rider.Name)
	}
	if info.Rider.CurrentLocation == nil || info.Rider.CurrentLocation.Latitude == 0 {
		t.Error("expected valid rider coordinates")
	}
}

func TestGetTrackingInfo_Unauthorized(t *testing.T) {
	order := &orders.Order{
		ID:         "ord_secret",
		CustomerID: "cust_owner",
		Status:     orders.StatusAccepted,
	}

	orderProv := &mockOrderProvider{order: order}
	outboxRepo := NewMemoryOutboxRepository()
	svc := NewService(orderProv, outboxRepo)

	_, err := svc.GetTrackingInfo(context.Background(), "ord_secret", "cust_intruder", false)
	if err != ErrUnauthorizedTracking {
		t.Errorf("expected ErrUnauthorizedTracking, got: %v", err)
	}
}

func TestOutbox_SaveAndRetrieve(t *testing.T) {
	outboxRepo := NewMemoryOutboxRepository()
	svc := NewService(&mockOrderProvider{}, outboxRepo)

	err := svc.PublishOutboxEvent(context.Background(), "order", "ord_999", "order.placed", map[string]any{
		"amountPoisha": 50000,
	})
	if err != nil {
		t.Fatalf("failed to publish outbox event: %v", err)
	}

	pending, err := outboxRepo.GetPending(context.Background(), 10)
	if err != nil || len(pending) != 1 {
		t.Fatalf("expected 1 pending outbox event, got %d, err: %v", len(pending), err)
	}
	if pending[0].EventType != "order.placed" {
		t.Errorf("expected eventType order.placed, got %s", pending[0].EventType)
	}

	if err := outboxRepo.MarkProcessed(context.Background(), pending[0].ID); err != nil {
		t.Fatalf("failed to mark processed: %v", err)
	}

	pendingAfter, _ := outboxRepo.GetPending(context.Background(), 10)
	if len(pendingAfter) != 0 {
		t.Errorf("expected 0 pending events after mark, got %d", len(pendingAfter))
	}
}
