package cart

import (
	"context"
	"testing"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
)

func setupCalculator() *Calculator {
	repo := catalog.NewMemoryCatalogRepository()
	catalogService := catalog.NewService(repo, repo)
	return NewCalculator(catalogService)
}

func TestCalculateQuote(t *testing.T) {
	calc := setupCalculator()
	ctx := context.Background()

	t.Run("calculates subtotal, delivery fee, vat and total accurately in poisha", func(t *testing.T) {
		// Classic Cheese Burger (৳350 = 35000 poisha) x 2 = 70000 poisha
		// Delivery fee for Burger Barn is ৳60 = 6000 poisha
		// No discount
		// VAT = 70000 * 5 / 100 = 3500 poisha
		// Total = 70000 + 6000 + 3500 = 79500 poisha
		req := QuoteRequest{
			RestaurantID: "rest_01",
			Items: []ItemInput{
				{ItemID: "item_01", Quantity: 2},
			},
		}

		quote, err := calc.CalculateQuote(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if quote.SubtotalPoisha != 70000 {
			t.Errorf("expected subtotal 70000, got %d", quote.SubtotalPoisha)
		}
		if quote.DeliveryFeePoisha != 6000 {
			t.Errorf("expected delivery fee 6000, got %d", quote.DeliveryFeePoisha)
		}
		if quote.VATPoisha != 3500 {
			t.Errorf("expected vat 3500, got %d", quote.VATPoisha)
		}
		if quote.TotalPoisha != 79500 {
			t.Errorf("expected total 79500, got %d", quote.TotalPoisha)
		}
		if quote.ItemCount != 2 {
			t.Errorf("expected item count 2, got %d", quote.ItemCount)
		}
	})

	t.Run("applies AUPA50 promo discount with ৳200 cap", func(t *testing.T) {
		// Subtotal = 70000 poisha (>= 50000)
		// 50% would be 35000, but capped at 20000 poisha
		// Taxable = 70000 - 20000 = 50000 poisha
		// VAT = 50000 * 5 / 100 = 2500 poisha
		// Total = 50000 + 6000 + 2500 = 58500 poisha
		req := QuoteRequest{
			RestaurantID: "rest_01",
			PromoCode:    "AUPA50",
			Items: []ItemInput{
				{ItemID: "item_01", Quantity: 2},
			},
		}

		quote, err := calc.CalculateQuote(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if quote.DiscountPoisha != 20000 {
			t.Errorf("expected discount 20000, got %d", quote.DiscountPoisha)
		}
		if quote.TotalPoisha != 58500 {
			t.Errorf("expected total 58500, got %d", quote.TotalPoisha)
		}
	})

	t.Run("calculates option modifier pricing", func(t *testing.T) {
		// Classic Cheese Burger base 35000 + Double Patty (+15000) = 50000 poisha x 1
		req := QuoteRequest{
			RestaurantID: "rest_01",
			Items: []ItemInput{
				{ItemID: "item_01", Quantity: 1, OptionItemIDs: []string{"opt_double"}},
			},
		}

		quote, err := calc.CalculateQuote(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if quote.SubtotalPoisha != 50000 {
			t.Errorf("expected subtotal 50000, got %d", quote.SubtotalPoisha)
		}
	})

	t.Run("empty cart returns error", func(t *testing.T) {
		req := QuoteRequest{
			RestaurantID: "rest_01",
			Items:        []ItemInput{},
		}

		_, err := calc.CalculateQuote(ctx, req)
		if err != ErrEmptyCart {
			t.Errorf("expected ErrEmptyCart, got %v", err)
		}
	})
}
