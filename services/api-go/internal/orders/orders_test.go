package orders

import (
	"context"
	"testing"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/cart"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
)

func setupOrderService() *Service {
	catRepo := catalog.NewMemoryCatalogRepository()
	catSvc := catalog.NewService(catRepo, catRepo)
	calc := cart.NewCalculator(catSvc)

	orderRepo := NewMemoryOrderRepository()
	idempRepo := NewMemoryIdempotencyRepository()

	return NewService(orderRepo, idempRepo, calc)
}

func TestOrderCreationAndIdempotency(t *testing.T) {
	svc := setupOrderService()
	ctx := context.Background()

	input := CreateOrderInput{
		CustomerID:        "cust_123",
		RestaurantID:      "rest_01",
		DeliveryAddressID: "addr_01",
		PaymentMethod:     PaymentMethodCOD,
		Items: []cart.ItemInput{
			{ItemID: "item_01", Quantity: 2},
		},
		IdempotencyKey: "idem_key_abc_1",
		RawPayloadHash: HashPayload([]byte(`{"items":[{"itemId":"item_01","quantity":2}]}`)),
	}

	t.Run("first order creation succeeds", func(t *testing.T) {
		order, isReplay, err := svc.CreateOrder(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isReplay {
			t.Errorf("expected isReplay to be false on first call")
		}
		if order.Status != StatusPlaced {
			t.Errorf("expected StatusPlaced, got %s", order.Status)
		}
		if order.TotalAmountPoisha != 79500 {
			t.Errorf("expected 79500, got %d", order.TotalAmountPoisha)
		}

		// Replay exact same request with same idempotency key
		replayOrder, isReplay2, err := svc.CreateOrder(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error on replay: %v", err)
		}
		if !isReplay2 {
			t.Errorf("expected isReplay to be true on duplicate submission")
		}
		if replayOrder.ID != order.ID {
			t.Errorf("expected same order ID %s, got %s", order.ID, replayOrder.ID)
		}
	})

	t.Run("reusing idempotency key with different payload returns conflict", func(t *testing.T) {
		conflictingInput := input
		conflictingInput.RawPayloadHash = HashPayload([]byte(`{"items":[{"itemId":"item_01","quantity":5}]}`))

		_, _, err := svc.CreateOrder(ctx, conflictingInput)
		if err != ErrIdempotencyConflict {
			t.Errorf("expected ErrIdempotencyConflict, got %v", err)
		}
	})
}

func TestOrderStateMachineAndTransitions(t *testing.T) {
	svc := setupOrderService()
	ctx := context.Background()

	input := CreateOrderInput{
		CustomerID:        "cust_456",
		RestaurantID:      "rest_01",
		DeliveryAddressID: "addr_01",
		PaymentMethod:     PaymentMethodCOD,
		Items: []cart.ItemInput{
			{ItemID: "item_01", Quantity: 1},
		},
		IdempotencyKey: "idem_lifecycle_test",
		RawPayloadHash: "hash123",
	}

	order, _, err := svc.CreateOrder(ctx, input)
	if err != nil {
		t.Fatalf("failed setup: %v", err)
	}

	t.Run("valid sequential lifecycle transitions", func(t *testing.T) {
		// placed -> accepted
		order, err = svc.TransitionStatus(ctx, order.ID, StatusAccepted)
		if err != nil {
			t.Fatalf("transition to accepted failed: %v", err)
		}

		// accepted -> preparing
		order, err = svc.TransitionStatus(ctx, order.ID, StatusPreparing)
		if err != nil {
			t.Fatalf("transition to preparing failed: %v", err)
		}

		// preparing -> ready
		order, err = svc.TransitionStatus(ctx, order.ID, StatusReady)
		if err != nil {
			t.Fatalf("transition to ready failed: %v", err)
		}

		// ready -> out_for_delivery
		order, err = svc.TransitionStatus(ctx, order.ID, StatusOutForDelivery)
		if err != nil {
			t.Fatalf("transition to out_for_delivery failed: %v", err)
		}

		// out_for_delivery -> delivered
		order, err = svc.TransitionStatus(ctx, order.ID, StatusDelivered)
		if err != nil {
			t.Fatalf("transition to delivered failed: %v", err)
		}

		if order.Status != StatusDelivered {
			t.Errorf("expected delivered status, got %s", order.Status)
		}
		if order.PaymentStatus != PaymentStatusPaid {
			t.Errorf("expected COD order to be marked paid upon delivery, got %s", order.PaymentStatus)
		}

		// Transition out of terminal state must fail
		_, invalidErr := svc.TransitionStatus(ctx, order.ID, StatusPlaced)
		if invalidErr == nil {
			t.Errorf("expected error transitioning out of delivered terminal state, got nil")
		}
	})
}

func TestOrderOwnership(t *testing.T) {
	svc := setupOrderService()
	ctx := context.Background()

	input := CreateOrderInput{
		CustomerID:        "cust_owner_A",
		RestaurantID:      "rest_01",
		DeliveryAddressID: "addr_01",
		PaymentMethod:     PaymentMethodCOD,
		Items: []cart.ItemInput{
			{ItemID: "item_01", Quantity: 1},
		},
		IdempotencyKey: "idem_owner_test",
		RawPayloadHash: "hashowner",
	}

	order, _, _ := svc.CreateOrder(ctx, input)

	t.Run("owner can view own order", func(t *testing.T) {
		fetched, err := svc.GetOrder(ctx, order.ID, "cust_owner_A", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fetched.ID != order.ID {
			t.Errorf("expected %s, got %s", order.ID, fetched.ID)
		}
	})

	t.Run("other customer cannot access order", func(t *testing.T) {
		_, err := svc.GetOrder(ctx, order.ID, "cust_intruder_B", false)
		if err != ErrUnauthorizedOrderAccess {
			t.Errorf("expected ErrUnauthorizedOrderAccess, got %v", err)
		}
	})

	t.Run("privileged admin can access order", func(t *testing.T) {
		fetched, err := svc.GetOrder(ctx, order.ID, "admin_user", true)
		if err != nil {
			t.Fatalf("expected admin access, got error: %v", err)
		}
		if fetched.ID != order.ID {
			t.Errorf("expected %s, got %s", order.ID, fetched.ID)
		}
	})
}
